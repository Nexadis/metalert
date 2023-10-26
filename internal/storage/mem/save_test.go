package mem

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSaveRestore(t *testing.T) {
	ms := NewMetricsStorage()
	m, err := models.NewMetric("id", models.CounterType, "1")
	assert.NoError(t, err)
	ms.Set(context.TODO(), m)
	f := os.TempDir() + "/save_mem.json"
	os.Remove(f)
	err = ms.Save(context.TODO(), f)
	assert.NoError(t, err)
	restored := NewMetricsStorage()
	err = restored.Restore(context.TODO(), f)
	assert.NoError(t, err)
	assert.Equal(t, ms, restored)
}

func TestSaveTimer(t *testing.T) {
	ms := NewMetricsStorage()
	m, err := models.NewMetric("id", models.CounterType, "1")
	assert.NoError(t, err)
	ms.Set(context.TODO(), m)
	f := os.TempDir() + "/savetimer_mem.json"
	os.Remove(f)
	ctx, cancel := context.WithCancel(context.Background())
	go ms.SaveTimer(ctx, f, 0)
	assert.NoError(t, err)
	restored := NewMetricsStorage()
	time.Sleep(1500 * time.Millisecond)
	err = restored.Restore(ctx, f)
	assert.NoError(t, err)
	assert.Equal(t, ms, restored)
	m, err = models.NewMetric("test", models.GaugeType, "1")
	assert.NoError(t, err)
	err = ms.Set(ctx, m)
	assert.NoError(t, err)
	cancel()
	<-ctx.Done()
	time.Sleep(500 * time.Millisecond)
	restored = NewMetricsStorage()
	err = restored.Restore(context.TODO(), f)
	assert.NoError(t, err)
	assert.Equal(t, ms, restored)
}
