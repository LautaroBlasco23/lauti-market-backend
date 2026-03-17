package infrastructure

import (
	"context"
	"fmt"

	imageDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/image/domain"
	imagestorev1 "github.com/lautaroblasco23/imagestore/proto/imagestore/v1"
	"google.golang.org/grpc"
)

const chunkSize = 32 * 1024

type GRPCImageClient struct {
	client imagestorev1.ImageServiceClient
}

func NewGRPCImageClient(conn *grpc.ClientConn) *GRPCImageClient {
	return &GRPCImageClient{client: imagestorev1.NewImageServiceClient(conn)}
}

func (c *GRPCImageClient) UploadImage(ctx context.Context, input imageDomain.UploadImageInput) (*imageDomain.UploadImageResult, error) {
	stream, err := c.client.UploadImage(ctx)
	if err != nil {
		return nil, fmt.Errorf("opening upload stream: %w", err)
	}

	if sendErr := stream.Send(&imagestorev1.UploadImageRequest{
		Data: &imagestorev1.UploadImageRequest_Metadata{
			Metadata: &imagestorev1.ImageMetadataInput{
				UserId: input.UserID, Filename: input.Filename, ContentType: input.ContentType,
			},
		},
	}); sendErr != nil {
		return nil, fmt.Errorf("sending metadata: %w", sendErr)
	}

	data := input.Data
	for len(data) > 0 {
		n := chunkSize
		if n > len(data) {
			n = len(data)
		}
		if sendErr := stream.Send(&imagestorev1.UploadImageRequest{
			Data: &imagestorev1.UploadImageRequest_Chunk{Chunk: data[:n]},
		}); sendErr != nil {
			return nil, fmt.Errorf("sending chunk: %w", sendErr)
		}
		data = data[n:]
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("closing stream: %w", err)
	}
	return &imageDomain.UploadImageResult{
		ImageID:      resp.GetImageId(),
		URL:          resp.GetUrl(),
		ThumbnailURL: resp.GetThumbnailUrl(),
		SizeBytes:    resp.GetSizeBytes(),
	}, nil
}
