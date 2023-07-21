package agent

import (
	"context"
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
	Pull(result chan error)
	Report(result chan error)
}

const (
	UpdateURL      = "/update/{valType}/{name}/{value}"
	JSONUpdateURL  = "/update/"
	JSONUpdatesURL = "/updates/"
)

var RuntimeNames []string

type httpAgent struct {
	listener       string
	pullInterval   int64
	reportInterval int64
	storage        mem.MetricsStorage
	mchan          chan *metrx.MetricsString
	counter        metrx.Counter
	client         client.MetricPoster
}

func NewAgent(listener string, pullInterval, reportInterval int64) Watcher {
	defineRuntimes()
	storage := mem.NewMetricsStorage()
	client := client.NewHTTP()
	mchan := make(chan *metrx.MetricsString)
	return &httpAgent{
		listener:       listener,
		pullInterval:   pullInterval,
		reportInterval: reportInterval,
		storage:        storage,
		mchan:          mchan,
		client:         client,
	}
}

func (ha *httpAgent) Run() error {
	pullerrs := make(chan error)
	reporterrs := make(chan error)
	reportTicker := time.NewTicker(time.Duration(ha.reportInterval) * time.Second)
	pullTicker := time.NewTicker(time.Duration(ha.pullInterval) * time.Second)
	for {
		select {
		case <-pullTicker.C:
			ha.mchan = make(chan *metrx.MetricsString)
			go ha.Pull(pullerrs)
		case <-reportTicker.C:
			go ha.Report(reporterrs)
		case err := <-pullerrs:
			if err != nil {
				logger.Error(err)
			}
			close(ha.mchan)
		case err := <-reporterrs:
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

func (ha *httpAgent) pullCustom() error {
	ctx := context.TODO()
	randValue := metrx.Gauge(rand.Float64()).String()
	err := ha.storage.Set(ctx, metrx.GaugeType, "RandomValue", randValue)
	logger.Info("Pull RandomValue")
	ha.mchan <- &metrx.MetricsString{
		ID:    "RandomValue",
		MType: metrx.GaugeType,
		Value: randValue,
	}
	if err != nil {
		return err
	}
	err = ha.storage.Set(ctx, metrx.CounterType, "PollCount", "1")
	ha.counter += 1
	logger.Info("Pull PollCount")
	ha.mchan <- &metrx.MetricsString{
		ID:    "PollCount",
		MType: metrx.CounterType,
		Value: ha.counter.String(),
	}
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
	ctx := context.TODO()
	for _, gaugeName := range RuntimeNames {
		value := mstruct.FieldByName(gaugeName)
		switch value.Kind() {
		case reflect.Float64:
			val = metrx.Gauge(value.Float()).String()
		case reflect.Uint32, reflect.Uint64:
			val = metrx.Gauge(value.Uint()).String()
		}
		err := ha.storage.Set(ctx, metrx.GaugeType, gaugeName, val)
		ha.mchan <- &metrx.MetricsString{
			ID:    gaugeName,
			MType: metrx.GaugeType,
			Value: val,
		}
		if err != nil {
			return err
		}
	}
	logger.Info("Metrics pulled")
	return nil
}

func (ha *httpAgent) Pull(errs chan error) {
	logger.Info("Pull Custom")
	err := ha.pullCustom()
	if err != nil {
		logger.Error(err)
	}
	err = ha.pullRuntime()
	if err != nil {
		logger.Error(err)
	}
	errs <- err
}

func (ha *httpAgent) Report(reported chan error) {
	m := &metrx.Metrics{}
	path := fmt.Sprintf("http://%s%s", ha.listener, UpdateURL)
	pathJSON := fmt.Sprintf("http://%s%s", ha.listener, JSONUpdateURL)
	for ms := range ha.mchan {
		err := ha.client.Post(path, ms.GetMType(), ms.GetID(), ms.GetValue())
		if err != nil {
			logger.Error("Can't report metrics")
			reported <- fmt.Errorf("can't report metrics: %w", err)
			return
		}
		m.ParseMetricsString(ms)
		err = ha.client.PostObj(pathJSON, m)
		if err != nil {
			logger.Error("Can't report metrics", err)
			reported <- fmt.Errorf("can't report metrics: %w", err)
			return
		}
		logger.Info("Metric", ms.GetID())

	}
	reported <- nil
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
