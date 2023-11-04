// agent для сбора метрик
//
// Cобирает метрики, Отправляет метрики. Всё это происходит в фоновом режим с заданным интервалом.
package agent

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	memStat "github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"

	"github.com/Nexadis/metalert/internal/agent/client"
	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// TransportType создаёт тип для видов передачи метрик
type TransportType string

// Константы для определения способа передачи метрик
const (
	RESTType TransportType = "REST"
	JSONType TransportType = "JSON"
	GRPCType TransportType = "GRPC"
)

var Transports = []TransportType{
	RESTType,
	JSONType,
	GRPCType,
}

var ErrInvalidTransport = errors.New("invalid transport type")

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

// MetricPoster интерфейс для отправки метрик как через URL, так и JSON-объектами.
type MetricPoster interface {
	Post(ctx context.Context, m models.Metric) error
}

// Agent собирает и отправляет метрики
type Agent struct {
	config  *Config
	counter models.Counter
	client  MetricPoster
}

// New - Конструктор для Agent
func New(config *Config) *Agent {
	key, err := asymcrypt.ReadPem(config.CryptoKey)
	if err != nil {
		logger.Error(err)
	}

	defineRuntimes()

	generalOps := []client.FOption{
		client.SetSignKey(config.Key),
		client.SetPubKey(key),
	}
	c := chooseClient(config, generalOps)
	agent := &Agent{
		config: config,
		client: c,
	}
	return agent
}

func chooseClient(c *Config, ops []client.FOption) MetricPoster {
	var choosenClient MetricPoster
	switch c.Transport {
	case RESTType:
		choosenClient = client.NewREST(c.Address, ops...)
	case JSONType:
		choosenClient = client.NewJSON(c.Address, ops...)
	case GRPCType:
	}
	return choosenClient
}

// Run запускает в фоне агент, начинает собирать и отправлять метрики с заданными интервалами
func (ha *Agent) Run(ctx context.Context) error {
	mchan := make(chan models.Metric, MetricsBufSize)
	grp, ctx := errgroup.WithContext(ctx)
	for i := 1; int64(i) <= ha.config.RateLimit; i++ {
		i := i
		logger.Info("Start reporter", i)
		grp.Go(func() error {
			return ha.Report(ctx, mchan)
		})
	}
	pullTicker := time.NewTicker(time.Duration(ha.config.PollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			close(mchan)
			err := grp.Wait()
			return err
		case <-pullTicker.C:
			ha.Pull(ctx, mchan)
		}
	}
}

// pullCustom получает нестандартные метрики, определенные разработчиком
func (ha *Agent) pullCustom(ctx context.Context, mchan chan models.Metric) {
	customMetrics := make([]models.Metric, 0, 5)

	randValue := models.Gauge(rand.Float64())
	m, err := models.NewMetric("RandomValue", models.GaugeType, randValue.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	ha.counter += 1
	m, err = models.NewMetric("PollCount", models.CounterType, ha.counter.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	v, _ := memStat.VirtualMemory()
	totalMemory := models.Gauge(v.Total)
	m, err = models.NewMetric("TotalMemory", models.GaugeType, totalMemory.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	freeMemory := models.Gauge(v.Free)
	m, err = models.NewMetric("FreeMemory", models.GaugeType, freeMemory.String())
	if err != nil {
		logger.Error(err)
		return
	}
	customMetrics = append(customMetrics, m)
	c, _ := cpu.PercentWithContext(ctx, 0, false)
	CPUUtilization := models.Gauge(c[0])
	m, err = models.NewMetric("CPUUtilization1", models.GaugeType, CPUUtilization.String())
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
func (ha *Agent) pullRuntime(ctx context.Context, mchan chan models.Metric) {
	var val string
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	mstruct := reflect.ValueOf(*memStats)
	for _, gaugeName := range RuntimeNames {
		value := mstruct.FieldByName(gaugeName)
		switch value.Kind() {
		case reflect.Float64:
			val = models.Gauge(value.Float()).String()
		case reflect.Uint32, reflect.Uint64:
			val = models.Gauge(value.Uint()).String()
		}
		m, err := models.NewMetric(gaugeName, models.GaugeType, val)
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
func (ha *Agent) Pull(ctx context.Context, mchan chan models.Metric) {
	ha.pullCustom(ctx, mchan)
	ha.pullRuntime(ctx, mchan)
	logger.Info("Metrics pulled")
}

// Report отправляет метрики на адрес, заданный в конфигурации, всеми доступными способами
func (ha *Agent) Report(ctx context.Context, input chan models.Metric) error {
	for m := range input {
		logger.Info("Post metric", m.ID)
		err := ha.client.Post(ctx, m)
		if err != nil {
			logger.Error("Can't report metrics")
			return err
		}
	}
	return nil
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

func (t TransportType) String() string {
	return string(t)
}

func (t *TransportType) Set(value string) error {
	if value == "" {
		*t = GRPCType
		return nil
	}
	for _, v := range Transports {
		if string(v) == value {
			*t = v
			return nil
		}
	}
	return ErrInvalidTransport
}
