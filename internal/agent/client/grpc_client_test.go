package client

import (
	"context"
	"testing"

	"github.com/Nexadis/metalert/internal/models"
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
