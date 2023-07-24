package agent

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	memStat "github.com/shirou/gopsutil/v3/mem"

	"github.com/Nexadis/metalert/internal/agent/client"
	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Watcher interface {
	Run(ctx context.Context) error
	Pull(ctx context.Context) chan *metrx.MetricsString
	Report(input chan *metrx.MetricsString, errs chan error)
}

const (
	UpdateURL      = "/update/{valType}/{name}/{value}"
	JSONUpdateURL  = "/update/"
	JSONUpdatesURL = "/updates/"
)

var RuntimeNames []string

type httpAgent struct {
	config  *Config
	storage mem.MetricsStorage
	counter metrx.Counter
	client  client.MetricPoster
}

func New(config *Config) Watcher {
	defineRuntimes()
	storage := mem.NewMetricsStorage()
	client := client.NewHTTP(client.SetKey(config.Key))
	agent := &httpAgent{
		config:  config,
		storage: storage,
		client:  client,
	}
	return agent
}

func (ha *httpAgent) Run(ctx context.Context) error {
	errs := make(chan error)
	reportTicker := time.NewTicker(time.Duration(ha.config.ReportInterval) * time.Second)
	for {
		select {
		case <-reportTicker.C:
			mchan := ha.Pull(ctx)
			go ha.Report(mchan, errs)
		case err := <-errs:
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

func (ha *httpAgent) pullCustom(ctx context.Context, mchan chan *metrx.MetricsString) {
	var customMetrics []metrx.MetricsString

	randValue := metrx.Gauge(rand.Float64()).String()
	customMetrics = append(customMetrics, metrx.MetricsString{
		ID:    "RandomValue",
		MType: metrx.GaugeType,
		Value: randValue,
	})
	ha.counter += 1
	customMetrics = append(customMetrics, metrx.MetricsString{
		ID:    "PollCount",
		MType: metrx.CounterType,
		Value: ha.counter.String(),
	})
	v, _ := memStat.VirtualMemory()
	totalMemory := metrx.Gauge(v.Total)
	customMetrics = append(customMetrics, metrx.MetricsString{
		ID:    "TotalMemory",
		MType: metrx.GaugeType,
		Value: totalMemory.String(),
	})
	freeMemory := metrx.Gauge(v.Free)
	customMetrics = append(customMetrics, metrx.MetricsString{
		ID:    "FreeMemory",
		MType: metrx.GaugeType,
		Value: freeMemory.String(),
	})
	c, _ := cpu.PercentWithContext(ctx, 0, false)
	CPUUtilization := metrx.Gauge(c[0])
	customMetrics = append(customMetrics, metrx.MetricsString{
		ID:    "CPUUtilization1",
		MType: metrx.GaugeType,
		Value: CPUUtilization.String(),
	})
	for _, m := range customMetrics {
		select {
		case mchan <- &m:
		case <-ctx.Done():
			return
		}
	}
}

func (ha *httpAgent) pullRuntime(ctx context.Context, mchan chan *metrx.MetricsString) {
	var val string
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	mstruct := reflect.ValueOf(*memStats)
	for _, gaugeName := range RuntimeNames {
		value := mstruct.FieldByName(gaugeName)
		switch value.Kind() {
		case reflect.Float64:
			val = metrx.Gauge(value.Float()).String()
		case reflect.Uint32, reflect.Uint64:
			val = metrx.Gauge(value.Uint()).String()
		}
		m := &metrx.MetricsString{
			ID:    gaugeName,
			MType: metrx.GaugeType,
			Value: val,
		}
		select {
		case mchan <- m:
		case <-ctx.Done():
			return
		}
	}
}

func (ha *httpAgent) Pull(ctx context.Context) chan *metrx.MetricsString {
	mchan := make(chan *metrx.MetricsString)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		ha.pullCustom(ctx, mchan)
		ha.pullRuntime(ctx, mchan)
		close(mchan)
		logger.Info("Metrics pulled")
	}()
	return mchan
}

func (ha *httpAgent) Report(input chan *metrx.MetricsString, errs chan error) {
	m := &metrx.Metrics{}
	path := fmt.Sprintf("http://%s%s", ha.config.Address, UpdateURL)
	pathJSON := fmt.Sprintf("http://%s%s", ha.config.Address, JSONUpdateURL)
	for ms := range input {
		err := ha.client.Post(path, ms.GetMType(), ms.GetID(), ms.GetValue())
		if err != nil {
			logger.Error("Can't report metrics")
			errs <- fmt.Errorf("can't report metrics: %w", err)
			return
		}
		m.ParseMetricsString(ms)
		err = ha.client.PostObj(pathJSON, m)
		if err != nil {
			logger.Error("Can't report metrics", err)
			errs <- fmt.Errorf("can't report metrics: %w", err)
			return
		}
	}
	errs <- nil
}

func defineRuntimes() {
	if RuntimeNames != nil {
		return
	}
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	mstruct := reflect.ValueOf(*m)
	RuntimeNames = make([]string, 0, mstruct.NumField())
	for i := 0; i < mstruct.NumField(); i++ {
		value := mstruct.Type().Field(i).Name
		RuntimeNames = append(RuntimeNames, value)
	}
}
