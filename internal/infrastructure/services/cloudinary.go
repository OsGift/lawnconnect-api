package services

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadService defines the interface for uploading files to a cloud storage.
type UploadService interface {
	UploadFile(ctx context.Context, file io.Reader, filename string) (string, error)
}

// uploadService implements UploadService using Cloudinary.
type uploadService struct {
	cld *cloudinary.Cloudinary
}

// NewUploadService creates a new UploadService instance configured for Cloudinary.
func NewUploadService(cld *cloudinary.Cloudinary) UploadService {
	return &uploadService{cld: cld}
}

// UploadFile uploads a file to Cloudinary.
func (s *uploadService) UploadFile(ctx context.Context, file io.Reader, filename string) (string, error) {
	if s.cld == nil {
		return "", fmt.Errorf("cloudinary service not initialized")
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID: filename,
		Folder:   "lawnconnect_uploads",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %w", err)
	}

	log.Printf("File '%s' uploaded to Cloudinary. URL: %s", filename, uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}