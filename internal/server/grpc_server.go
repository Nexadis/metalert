package server

import (
	"context"
	"net"

	"github.com/Nexadis/metalert/internal/models/controller"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
	pb "github.com/Nexadis/metalert/proto/metrics/v1"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.UnimplementedMetricsCollectorServiceServer
	storage storage.Storage
	config  *Config
}

func NewGRPCServer(config *Config, storage storage.Storage) (*grpcServer, error) {
	return &grpcServer{
		storage: storage,
		config:  config,
	}, nil
}

func (s *grpcServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.config.GRPC)
	if err != nil {
		return err
	}
	gs := grpc.NewServer()
	pb.RegisterMetricsCollectorServiceServer(gs, s)
	go func() {
		logger.Info("Grpc Server at ", s.config.GRPC)
		err = gs.Serve(lis)
	}()
	<-ctx.Done()
	gs.Stop()
	return err
}

func (s *grpcServer) Get(ctx context.Context, r *pb.GetRequest) (*pb.GetResponse, error) {
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

func (s *grpcServer) Post(ctx context.Context, r *pb.PostRequest) (*pb.PostResponse, error) {
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
