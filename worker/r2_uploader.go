package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mengonten-api/config"
)

type R2Uploader struct {
	Client    *s3.Client
	Bucket    string
	PublicURL string
}

func NewR2Uploader(apis *config.ExternalAPIs) *R2Uploader {
	if apis.R2AccountID == "" || apis.R2AccessKeyID == "" {
		log.Println("R2 credentials not configured, R2 upload disabled")
		return nil
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", apis.R2AccountID)

	log.Printf("R2 init: AccountID=%s, Bucket=%s, PublicURL=%s, Endpoint=%s",
		apis.R2AccountID, apis.R2Bucket, apis.R2PublicURL, endpoint)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			apis.R2AccessKeyID,
			apis.R2SecretKey,
			"",
		)),
	)
	if err != nil {
		log.Printf("R2 config error: %v", err)
		return nil
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	publicURL := apis.R2PublicURL
	publicURL = strings.TrimRight(publicURL, "/")

	return &R2Uploader{
		Client:    client,
		Bucket:    apis.R2Bucket,
		PublicURL: publicURL,
	}
}

func (r *R2Uploader) Upload(filePath, key string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("R2 uploader not initialized")
	}

	log.Printf("R2 uploading: %s -> bucket=%s key=%s", filePath, r.Bucket, key)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	contentType := getContentType(filePath)

	_, err = r.Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("R2 upload error: %v", err)
	}

	publicURL := fmt.Sprintf("%s/%s", r.PublicURL, key)
	log.Printf("R2 uploaded OK: %s", publicURL)
	return publicURL, nil
}

func (r *R2Uploader) Delete(key string) error {
	if r == nil {
		return fmt.Errorf("R2 uploader not initialized")
	}

	_, err := r.Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("R2 delete error: %v", err)
	}

	log.Printf("File deleted from R2: %s", key)
	return nil
}

func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func UploadToR2Static(filePath, key string) (string, error) {
	apis := config.InitExternalAPIs()
	uploader := NewR2Uploader(apis)
	if uploader == nil {
		return "", fmt.Errorf("R2 not configured")
	}
	return uploader.Upload(filePath, key)
}

func DeleteFromR2(key string) error {
	apis := config.InitExternalAPIs()
	uploader := NewR2Uploader(apis)
	if uploader == nil {
		return fmt.Errorf("R2 not configured")
	}
	return uploader.Delete(key)
}
