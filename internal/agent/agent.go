// agent для сбора метрик
//
// Cобирает метрики, Отправляет метрики. Всё это происходит в фоновом режим с заданным интервалом.
package agent

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	memStat "github.com/shirou/gopsutil/v3/mem"

	"github.com/Nexadis/metalert/internal/agent/client"
	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Watcher - интерфейс агента, собирающего и отправляющего метрики.
type Watcher interface {
	Run(ctx context.Context) error
	Pull(ctx context.Context, mchan chan metrx.Metric)
	Report(ctx context.Context, input chan metrx.Metric)
}

// Endpoint'ы для отправки метрик.
const (
	UpdateURL     = "/update/{valType}/{name}/{value}"
	JSONUpdateURL = "/update/"
	// Для отправки сразу пачки метрик
	JSONUpdatesURL = "/updates/"
)

const MetricsBufSize = 100

// RuntimeNames - набор Runtime метрик, список которых заполняется один раз с помощью reflect и многократно используется
var RuntimeNames []string

// HTTPAgent реализует интерфейс Watcher, собирает и отправляет метрики
type HTTPAgent struct {
	config     *Config
	counter    metrx.Counter
	clientREST client.MetricPoster
	clientJSON client.MetricPoster
}

// New - Конструктор для HTTPAgent
func New(config *Config) *HTTPAgent {
	defineRuntimes()
	clientREST := client.NewHTTP(client.SetKey(config.Key), client.SetTransport(client.RESTType))
	clientJSON := client.NewHTTP(client.SetKey(config.Key), client.SetTransport(client.JSONType))
	agent := &HTTPAgent{
		config:     config,
		clientREST: clientREST,
		clientJSON: clientJSON,
	}
	return agent
}

// Run запускает в фоне агент, начинает собирать и отправлять метрики с заданными интервалами
func (ha *HTTPAgent) Run(ctx context.Context) error {
	mchan := make(chan metrx.Metric, MetricsBufSize)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg := sync.WaitGroup{}
	wg.Add(int(ha.config.RateLimit))
	for i := 1; int64(i) <= ha.config.RateLimit; i++ {
		i := i
		logger.Info("Start reporter", i)
		go func() {
			defer wg.Done()
			ha.Report(ctx, mchan)
			logger.Info("End reporter", i)
		}()
	}
	pullTicker := time.NewTicker(time.Duration(ha.config.PollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			close(mchan)
			wg.Wait()
			return nil
		case <-pullTicker.C:
			ha.Pull(ctx, mchan)
		}
	}
}

// pullCustom получает нестандартные метрики, определенные разработчиком
func (ha *HTTPAgent) pullCustom(ctx context.Context, mchan chan metrx.Metric) {
	customMetrics := make([]metrx.Metric, 0, 5)

	randValue := metrx.Gauge(rand.Float64())
	m, err := metrx.NewMetric("RandomValue", metrx.GaugeType, randValue.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	ha.counter += 1
	m, err = metrx.NewMetric("PollCount", metrx.CounterType, ha.counter.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	v, _ := memStat.VirtualMemory()
	totalMemory := metrx.Gauge(v.Total)
	m, err = metrx.NewMetric("TotalMemory", metrx.GaugeType, totalMemory.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	freeMemory := metrx.Gauge(v.Free)
	m, err = metrx.NewMetric("FreeMemory", metrx.GaugeType, freeMemory.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	c, _ := cpu.PercentWithContext(ctx, 0, false)
	CPUUtilization := metrx.Gauge(c[0])
	m, err = metrx.NewMetric("CPUUtilization1", metrx.GaugeType, CPUUtilization.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	for _, m := range customMetrics {
		cm := m
		select {
		case mchan <- cm:
		case <-ctx.Done():
			return
		}
	}
	logger.Info("Got custom metrics")
}

// pullRuntime получает метрики из стандартной библиотеки runtime
func (ha *HTTPAgent) pullRuntime(ctx context.Context, mchan chan metrx.Metric) {
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
		m, err := metrx.NewMetric(gaugeName, metrx.GaugeType, val)
		if err != nil {
			logger.Error(err)
			return
		}
		select {
		case mchan <- m:
		case <-ctx.Done():
			return
		}
	}
}

// Pull внешняя функция для получения всех метрик
func (ha *HTTPAgent) Pull(ctx context.Context, mchan chan metrx.Metric) {
	ha.pullCustom(ctx, mchan)
	ha.pullRuntime(ctx, mchan)
	logger.Info("Metrics pulled")
}

// Report отправляет метрики на адрес, заданный в конфигурации, всеми доступными способами
func (ha *HTTPAgent) Report(ctx context.Context, input chan metrx.Metric) {
	path := fmt.Sprintf("http://%s%s", ha.config.Address, UpdateURL)
	pathJSON := fmt.Sprintf("http://%s%s", ha.config.Address, JSONUpdateURL)
	for m := range input {
		logger.Info("Post metric", m.ID)
		err := ha.clientREST.Post(ctx, path, m)
		if err != nil {
			logger.Error("Can't report metrics")
			// errs <- fmt.Errorf("can't report metrics: %w", err)
		}
		err = ha.clientJSON.Post(ctx, pathJSON, m)
		if err != nil {
			logger.Error("Can't report metrics", err)
			//	errs <- fmt.Errorf("can't report metrics: %w", err)
		}

	}
}

// defineRuntimes получает имена всех метрик из runtime
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
