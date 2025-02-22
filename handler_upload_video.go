package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	// "log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

const maxSize = 10 << 23 // 80mb

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get video", err)
		return
	}

	if video.UserID != userID {
		respondWithJSON(w, http.StatusUnauthorized, struct{}{})
		return
	}

	err = r.ParseMultipartForm(maxSize)
	if err != nil {
		respondWithError(w, 500, "failure parsing multipart form", err)
		return
	}

	file, fileheader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, 500, "error reading thumbnail", err)
	}
	defer file.Close()

	ct := fileheader.Header.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		respondWithError(w, 500, "failure parsing media type", err)
		return
	}
	if mt != "video/mp4" {
		respondWithError(w, 400, "invalid media type", nil)
		return
	}

	fp, err := os.CreateTemp("", "upload.mp4") // bara en i taget kan ladda upp...
	if err != nil {
		respondWithError(w, 500, "couldn't create storage file", err)
	}
	defer os.Remove(fp.Name())
	defer fp.Close()
	n, err := io.Copy(fp, file) // varför spara till disk istf bara ladda upp från ram direkt?
	if err != nil || n != fileheader.Size {
		respondWithError(w, 500, "error reading upload data", err)
		return
	}

	as, err := getVideoAspectRatio(fp.Name())
	if err != nil {
		respondWithError(w, 500, "getVideoAspectRatio", err)
		return
	}

	n, err = fp.Seek(0, io.SeekStart)
	if err != nil || n != 0 {
		respondWithError(w, 500, "something went wrong", err)
		return
	}

	parts := strings.Split(ct, "/")
	ext := parts[len(parts)-1]
	rnd := make([]byte, 32)
	n2, err := rand.Read(rnd)
	if err != nil || n2 != len(rnd) {
		respondWithError(w, 500, "something went wrong", err)
		return
	}
	rndb64 := base64.RawURLEncoding.EncodeToString(rnd)
	key := rndb64 + "." + ext
	if as == "16:9" {
		key = "landscape/" + key
	}	else if as == "9:16" {
		key = "portrait/" + key
	} else {
		key = "other/" + key
	}

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		// Bucket: aws.String(bucket),
		Bucket: aws.String(cfg.s3Bucket),
		Key: aws.String(key),
		Body: fp,
		ContentType: &ct,
	})
	if err != nil {
		respondWithError(w, 500, "couldn't upload media", err)
		return
	}
	log.Printf("Uploaded %s (%s) to s3 bucket\n", fileheader.Filename, key)

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
	video.VideoURL = &url
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, 500, "error updating video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
