package client

import (
	"context"
	"errors"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/models/controller"
	"github.com/Nexadis/metalert/internal/utils/logger"
	pb "github.com/Nexadis/metalert/proto/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrConnection = errors.New("can't connect to server")

type GRPCClient struct {
	gc   pb.MetricsCollectorServiceClient
	conn *grpc.ClientConn
}

func NewGRPC(server string) *GRPCClient {
	if server == "" {
		logger.Error("empty address of server")
		return &GRPCClient{}

	}
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(err)
		return &GRPCClient{}
	}
	c := pb.NewMetricsCollectorServiceClient(conn)
	return &GRPCClient{
		gc:   c,
		conn: conn,
	}
}

func (c *GRPCClient) Post(ctx context.Context, m models.Metric) error {
	err := c.ctxClose(ctx)
	if err != nil {
		return err
	}
	var r pb.PostRequest
	if c.gc == nil {
		return ErrConnection
	}
	in, err := controller.MetricsToPB(models.Metrics{m})
	if err != nil {
		return err
	}
	r.Metrics = in
	_, err = c.gc.Post(ctx, &r)
	return err
}

func (c *GRPCClient) Get(ctx context.Context) (models.Metrics, error) {
	err := c.ctxClose(ctx)
	if err != nil {
		return nil, err
	}
	var r pb.GetRequest
	if c.gc == nil {
		return nil, ErrConnection
	}
	resp, err := c.gc.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return controller.MetricsFromPB(resp.Metrics)
}

func (c *GRPCClient) ctxClose(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		if c.conn != nil {
			c.conn.Close()
		}
		return err
	}
	return nil
}
