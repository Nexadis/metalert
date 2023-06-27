package agent

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
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

const UpdateURL = "/update/{valType}/{name}/{value}"

type httpAgent struct {
	listener       string
	pullInterval   int64
	reportInterval int64
	storage        metrx.MemStorage
	client         client.MetricPoster
}

func NewAgent(listener string, pullInterval, reportInterval int64) Watcher {
	storage := metrx.NewMetricsStorage()
	client := client.NewHttp()
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
	randValue := strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
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
	for i := 0; i < mstruct.NumField(); i++ {
		value := mstruct.Field(i)
		switch value.Kind() {
		case reflect.Uint32, reflect.Uint64:
			val = strconv.FormatUint(value.Uint(), 10)
		case reflect.Float64:
			val = strconv.FormatFloat(value.Float(), 'f', -1, 64)
		}
		err := ha.storage.Set(metrx.GaugeType, mstruct.Type().Field(i).Name, val)
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
	if err != nil {
		return err
	}
	path := fmt.Sprintf("http://%s%s", ha.listener, UpdateURL)
	for _, m := range values {
		err := ha.client.Post(path, m.ValType, m.Name, m.Value)
		if err != nil {
			logger.Error("Can't report metrics")
			break
		}
<<<<<<< Updated upstream
		logger.Debug("Metric: %s", m.Name)
=======
		logger.Info("Metric", m.Name, "status", resp.Status())
>>>>>>> Stashed changes
	}
	return nil
}
