package worker

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"mengonten-api/config"
)

type CloudinaryUploader struct {
	Client *cloudinary.Cloudinary
}

func NewCloudinaryUploader(apis *config.ExternalAPIs) *CloudinaryUploader {
	cldURL := fmt.Sprintf("cloudinary://%s:%s@%s",
		apis.CloudinaryKey,
		apis.CloudinarySecret,
		apis.CloudinaryName)

	cld, err := cloudinary.NewFromURL(cldURL)
	if err != nil {
		log.Printf("Cloudinary init error: %v", err)
		return nil
	}

	return &CloudinaryUploader{
		Client: cld,
	}
}

func (cu *CloudinaryUploader) Upload(filePath, folder string) (string, error) {
	if cu.Client == nil {
		return "", fmt.Errorf("cloudinary client not initialized")
	}

	log.Printf("Uploading clip to Cloudinary: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	ctx := context.Background()
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "video",
	}

	resp, err := cu.Client.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload error: %v", err)
	}

	log.Printf("Cloudinary response - PublicID: %s, Format: %s, Version: %v", resp.PublicID, resp.Format, resp.Version)

	if resp.SecureURL == "" && resp.URL == "" {
		return "", fmt.Errorf("cloudinary returned empty URL")
	}

	secureURL := resp.SecureURL
	if secureURL == "" {
		secureURL = resp.URL
	}

	log.Printf("Video uploaded to Cloudinary: %s", secureURL)
	return secureURL, nil
}
