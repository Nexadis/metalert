package client

import (
	"context"
	"testing"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/server"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/stretchr/testify/assert"
)

func TestNewGRPCClient(t *testing.T) {
	c := NewGRPC("server")
	_, err := c.Get(context.TODO())
	assert.Error(t, err)
	m, err := models.NewMetric("name", models.GaugeType, "123.123")
	assert.NoError(t, err)
	err = c.Post(context.TODO(), m)
	assert.Error(t, err)

	c = NewGRPC("")
	_, err = c.Get(context.TODO())
	assert.Error(t, err)
	err = c.Post(context.TODO(), m)
	assert.Error(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = c.Post(ctx, m)
	assert.Error(t, err)
}

func TestPostGet(t *testing.T) {
	c := server.NewConfig()
	c.SetDefault()
	c.GRPC = "localhost:3355"
	s := mem.NewMetricsStorage()
	gs, err := server.NewGRPCServer(c, s)
	assert.NoError(t, err)
	go gs.Run(context.TODO())
	client := NewGRPC(c.GRPC)
	m, err := models.NewMetric("name", models.GaugeType, "123.123")
	assert.NoError(t, err)
	err = client.Post(context.TODO(), m)
	assert.NoError(t, err)
	ms, err := client.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ms))
	assert.Equal(t, m, ms[0])
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = client.Post(ctx, m)
	assert.Error(t, err)
	_, err = client.Get(ctx)
	assert.Error(t, err)
}
