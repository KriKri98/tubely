package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	preSignClient := s3.NewPresignClient(s3Client)
	params := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	httpRequest, err := preSignClient.PresignGetObject(context.Background(), &params, s3.WithPresignExpires(expireTime))
	return httpRequest.URL, err
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil || *video.VideoURL == "" {
		return video, nil
	}
	videoURLParts := strings.Split(*video.VideoURL, ",")
	bucket := videoURLParts[0]
	key := videoURLParts[1]
	presignedUrl, err := generatePresignedURL(cfg.s3Client, bucket, key, 5*time.Minute)
	if err != nil {
		return video, err
	}

	video.VideoURL = &presignedUrl

	return video, nil
}
