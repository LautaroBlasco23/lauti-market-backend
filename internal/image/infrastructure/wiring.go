package infrastructure

import (
	"fmt"

	imageDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/image/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ImageModule struct {
	Client imageDomain.ImageClient
	conn   *grpc.ClientConn
}

func Wire(addr string) (*ImageModule, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("creating grpc client: %w", err)
	}
	return &ImageModule{Client: NewGRPCImageClient(conn), conn: conn}, nil
}

func (m *ImageModule) Close() error { return m.conn.Close() }
