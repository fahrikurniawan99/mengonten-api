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
	if apis.CloudinaryName == "" || apis.CloudinaryKey == "" || apis.CloudinarySecret == "" {
		log.Printf("Cloudinary credentials incomplete: Name=%s, Key=%s", apis.CloudinaryName, apis.CloudinaryKey)
		return nil
	}

	cldURL := fmt.Sprintf("cloudinary://%s:%s@%s",
		apis.CloudinaryKey,
		apis.CloudinarySecret,
		apis.CloudinaryName)

	log.Printf("Cloudinary init with URL: cloudinary://%s:***@%s", apis.CloudinaryKey, apis.CloudinaryName)

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
		return "", fmt.Errorf("cloudinary client not initialized - check CLOUDINARY_NAME, CLOUDINARY_KEY, CLOUDINARY_SECRET in .env")
	}

	log.Printf("Uploading clip to Cloudinary: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}
	log.Printf("File size: %d bytes", fileInfo.Size())

	ctx := context.Background()
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "video",
	}

	resp, err := cu.Client.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload error: %v", err)
	}

	log.Printf("Cloudinary response - PublicID: %s, Format: %s, Version: %v, SecureURL: %s, URL: %s",
		resp.PublicID, resp.Format, resp.Version, resp.SecureURL, resp.URL)

	if resp.SecureURL == "" && resp.URL == "" && resp.PublicID == "" {
		return "", fmt.Errorf("cloudinary returned empty response - check credentials and account settings")
	}

	secureURL := resp.SecureURL
	if secureURL == "" {
		secureURL = resp.URL
	}

	log.Printf("Video uploaded to Cloudinary: %s", secureURL)
	return secureURL, nil
}
