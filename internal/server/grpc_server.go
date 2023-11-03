package server

import (
	"context"

	"github.com/Nexadis/metalert/internal/models/controller"
	"github.com/Nexadis/metalert/internal/storage"
	pb "github.com/Nexadis/metalert/proto/metrics/v1"
)

type GRPCServer struct {
	pb.UnimplementedMetricsCollectorServiceServer
	storage storage.Storage
	config  *Config
}

func (s *GRPCServer) Get(ctx context.Context, r *pb.GetRequest) (*pb.GetResponse, error) {
	var resp pb.GetResponse
	metrics, err := s.storage.GetAll(ctx)
	if err != nil {
		return &resp, nil
	}
	resp.Metrics, err = controller.MetricsToPB(metrics)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *GRPCServer) Post(ctx context.Context, r *pb.PostRequest) (*pb.PostResponse, error) {
	var resp pb.PostResponse
	ms, err := controller.MetricsFromPB(r.Metrics)
	if err != nil {
		resp.Error = err.Error()
		return &resp, nil
	}
	for _, m := range ms {
		err = s.storage.Set(ctx, m)
		if err != nil {
			resp.Error = err.Error()
			return &resp, nil
		}
	}
	return &resp, nil
}
