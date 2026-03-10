package domain

import "context"

type UploadImageInput struct {
	UserID      string
	Filename    string
	ContentType string
	Data        []byte
}

type UploadImageResult struct {
	ImageID      string
	URL          string
	ThumbnailURL string
	SizeBytes    int64
}

type ImageClient interface {
	UploadImage(ctx context.Context, input UploadImageInput) (*UploadImageResult, error)
}
