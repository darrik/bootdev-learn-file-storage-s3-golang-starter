package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

const maxMemory = 10 << 20 // 10mb

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, 500, "failure parsing multipart form", err)
		return
	}

	file, fileheader, err := r.FormFile("thumbnail")
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

	if mt != "image/jpeg" && mt != "image/png" {
		respondWithError(w, 400, "invalid media type", nil)
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
	filename := filepath.Join(cfg.assetsRoot, rndb64 + "." + ext)
	fp, err := os.Create(filename)
	if err != nil {
		respondWithError(w, 500, "error creating file", err)
		return
	}
	defer fp.Close()

	n, err := io.Copy(fp, file)
	if err != nil || n != fileheader.Size {
		respondWithError(w, 500, "error writing to file", err)
		return
	}

	videoInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "error getting video information", err)
		return
	}

	if videoInfo.UserID != userID {
		respondWithJSON(w, http.StatusUnauthorized, struct{}{})
	}

	url := fmt.Sprintf("http://localhost:%s/%s", os.Getenv("PORT"), filename)
	videoInfo.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(w, 500, "error updating video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoInfo)
}
