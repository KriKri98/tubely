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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid format", err)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid format", err)
		return
	}

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get Video Metadata from DB", err)
		return
	}

	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Uploader is not owner of video", err)
		return
	}

	fileExtension := ""
	mediaTypeSplit := strings.Split(mediaType, "/")
	if len(mediaTypeSplit) != 2 {
		fileExtension = "bin"
	} else {
		fileExtension = mediaTypeSplit[1]
	}

	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	base64Encoder := base64.URLEncoding
	randomString := base64Encoder.EncodeToString(randomBytes)

	path := filepath.Join(cfg.assetsRoot, randomString)
	thumbnail, err := os.Create(fmt.Sprintf("%v.%v", path, fileExtension))
	_, err = io.Copy(thumbnail, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to save thumbnail", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%v/assets/%v.%v", cfg.port, randomString, fileExtension)
	videoMetadata.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(videoMetadata)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to create thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)

}
