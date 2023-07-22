package agent

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"

	"github.com/Nexadis/metalert/internal/agent/client"
	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Watcher interface {
	Run() error
	Pull() chan *metrx.MetricsString
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

func (ha *httpAgent) Run() error {
	errs := make(chan error)
	reportTicker := time.NewTicker(time.Duration(ha.config.ReportInterval) * time.Second)
	for {
		select {
		case <-reportTicker.C:
			mchan := ha.Pull()
			go ha.Report(mchan, errs)
		case err := <-errs:
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

func (ha *httpAgent) pullCustom(mchan chan *metrx.MetricsString) {
	randValue := metrx.Gauge(rand.Float64()).String()
	mchan <- &metrx.MetricsString{
		ID:    "RandomValue",
		MType: metrx.GaugeType,
		Value: randValue,
	}
	ha.counter += 1
	mchan <- &metrx.MetricsString{
		ID:    "PollCount",
		MType: metrx.CounterType,
		Value: ha.counter.String(),
	}
}

func (ha *httpAgent) pullRuntime(mchan chan *metrx.MetricsString) {
	var val string
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	mstruct := reflect.ValueOf(*m)
	for _, gaugeName := range RuntimeNames {
		value := mstruct.FieldByName(gaugeName)
		switch value.Kind() {
		case reflect.Float64:
			val = metrx.Gauge(value.Float()).String()
		case reflect.Uint32, reflect.Uint64:
			val = metrx.Gauge(value.Uint()).String()
		}
		mchan <- &metrx.MetricsString{
			ID:    gaugeName,
			MType: metrx.GaugeType,
			Value: val,
		}
	}
	logger.Info("Metrics pulled")
}

func (ha *httpAgent) Pull() chan *metrx.MetricsString {
	logger.Info("Pull Custom")
	mchan := make(chan *metrx.MetricsString)
	go func() {
		ha.pullCustom(mchan)
		ha.pullRuntime(mchan)
		close(mchan)
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
