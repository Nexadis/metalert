package agent

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"

	"github.com/Nexadis/metalert/internal/agent/client"
	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Watcher interface {
	Run() error
	Pull() error
	Report() error
}

const (
	UpdateURL     = "/update/{valType}/{name}/{value}"
	JSONUpdateURL = "/update/"
)

var RuntimeNames []string

type httpAgent struct {
	listener       string
	pullInterval   int64
	reportInterval int64
	storage        metrx.MemStorage
	client         client.MetricPoster
}

func NewAgent(listener string, pullInterval, reportInterval int64) Watcher {
	defineRuntimes()
	storage := metrx.NewMetricsStorage()
	client := client.NewHTTP()
	return &httpAgent{
		listener:       listener,
		pullInterval:   pullInterval,
		reportInterval: reportInterval,
		storage:        storage,
		client:         client,
	}
}

func (ha *httpAgent) Run() error {
	pullTicker := time.NewTicker(time.Duration(ha.pullInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(ha.reportInterval) * time.Second)
	for {
		select {
		case <-pullTicker.C:
			err := ha.Pull()
			if err != nil {
				return err
			}
		case <-reportTicker.C:
			err := ha.Report()
			if err != nil {
				return err
			}
		}
	}
}

func (ha *httpAgent) pullCustom() error {
	randValue := metrx.Gauge(rand.Float64()).String()
	err := ha.storage.Set(metrx.GaugeType, "RandomValue", randValue)
	if err != nil {
		return err
	}
	err = ha.storage.Set(metrx.CounterType, "PollCount", "1")
	if err != nil {
		return err
	}
	return nil
}

func (ha *httpAgent) pullRuntime() error {
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
		err := ha.storage.Set(metrx.GaugeType, gaugeName, val)
		if err != nil {
			return err
		}
	}
	logger.Info("Metrics pulled")
	return nil
}

func (ha *httpAgent) Pull() error {
	err := ha.pullCustom()
	if err != nil {
		return err
	}
	err = ha.pullRuntime()
	if err != nil {
		return err
	}
	return nil
}

func (ha *httpAgent) Report() error {
	values, err := ha.storage.Values()
	m := &metrx.Metrics{}
	if err != nil {
		return err
	}
	path := fmt.Sprintf("http://%s%s", ha.listener, UpdateURL)
	pathJSON := fmt.Sprintf("http://%s%s", ha.listener, JSONUpdateURL)
	for _, ms := range values {
		err := ha.client.Post(path, ms.MType, ms.ID, ms.Value)
		if err != nil {
			logger.Error("Can't report metrics")
			break
		}
		m.ParseMetricsString(ms)
		err = ha.client.PostJSON(pathJSON, m)
		if err != nil {
			logger.Error("Can't report metrics", err)
			break
		}
		logger.Info("Metric", ms.ID)
	}
	return nil
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
