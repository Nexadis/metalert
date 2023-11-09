package server

import (
	"context"
	"testing"
	"time"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/models/controller"
	"github.com/Nexadis/metalert/internal/storage/mem"
	pb "github.com/Nexadis/metalert/proto/metrics/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewGRPCServer(t *testing.T) {
	c := NewConfig()
	c.SetDefault()
	s := mem.NewMetricsStorage()
	gs, err := NewGRPCServer(c, s)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	err = gs.Run(ctx)
	assert.NoError(t, err)
}

func TestGetPost(t *testing.T) {
	c := NewConfig()
	s := mem.NewMetricsStorage()
	gs, err := NewGRPCServer(c, s)
	assert.NoError(t, err)
	m, err := models.NewMetric("name", models.GaugeType, "123.123")
	assert.NoError(t, err)
	ms := models.Metrics{m}

	pbms, err := controller.MetricsToPB(ms)
	assert.NoError(t, err)
	postreq := pb.PostRequest{
		Metrics: pbms,
	}
	_, err = gs.Post(context.TODO(), &postreq)
	assert.NoError(t, err)

	greq := pb.GetRequest{}
	getresp, err := gs.Get(context.TODO(), &greq)
	assert.NoError(t, err)
	gotms, err := controller.MetricsFromPB(getresp.Metrics)
	assert.NoError(t, err)
	assert.Equal(t, ms, gotms)
}
