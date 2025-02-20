package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

const maxMemory = 10 << 20 // 10mb
const ctmsg = "couldn't find Content-Type"

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
	file, fileheader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, 500, "error reading thumbnail", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, 500, "error reading file", err)
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

	ct := fileheader.Header.Get("Content-Type")
	b64 := base64.StdEncoding.EncodeToString(data)
	url := fmt.Sprintf("data:%s;base64,%s", ct, b64)
	videoInfo.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(w, 500, "error updating video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoInfo)
}
