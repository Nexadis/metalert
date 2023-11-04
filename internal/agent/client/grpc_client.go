package client

import (
	"context"

	"github.com/Nexadis/metalert/internal/models"
	pb "github.com/Nexadis/metalert/proto/metrics/v1"
	"google.golang.org/grpc"
)

type GRPCClient struct{}

func NewGRPC(server string) (*GRPCClient, error) {
	conn, err := grpc.Dial("")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewMetricsCollectorServiceClient(conn)
	_ = c
	return nil, nil
}

func (gc *GRPCClient) Post(ctx context.Context, server string, m models.Metric) error {
	return nil
}
