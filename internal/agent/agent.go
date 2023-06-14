package agent

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Watcher interface {
	Run() error
	Pull()
	Report()
}

type httpAgent struct {
	listener       string
	pullInterval   int64
	reportInterval int64
	storage        metrx.MemStorage
	client         *resty.Client
}

func NewAgent(listener string, pullInterval, reportInterval int64) Watcher {
	storage := metrx.NewMetricsStorage()
	return &httpAgent{
		listener:       listener,
		pullInterval:   pullInterval,
		reportInterval: reportInterval,
		storage:        storage,
		client:         resty.New(),
	}
}

func (ha *httpAgent) Run() error {
	pullTicker := time.NewTicker(time.Duration(ha.pullInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(ha.reportInterval) * time.Second)
	for {
		select {
		case <-pullTicker.C:
			ha.Pull()
		case <-reportTicker.C:
			ha.Report()
		}
	}
}

func (ha *httpAgent) pullCustom() {
	randValue := strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
	err := ha.storage.Set(metrx.GaugeType, "RandomValue", randValue)
	if err != nil {
		panic(err)
	}
	err = ha.storage.Set(metrx.CounterType, "PollCount", "1")
	if err != nil {
		panic(err)
	}
}

func (ha *httpAgent) pullRuntime() {
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
			panic(err)
		}
	}
	logger.Debug("Metrics pulled")

}

func (ha *httpAgent) Pull() {
	ha.pullCustom()
	ha.pullRuntime()

}

func (ha *httpAgent) Report() {

	values, err := ha.storage.Values()
	if err != nil {
		panic(err)
	}
	const UpdateURL = "/update/{valType}/{name}/{value}"
	path := fmt.Sprintf("http://%s%s", ha.listener, UpdateURL)
	for _, m := range values {
		resp, err := ha.client.R().
			SetHeader("Content-type", "text/plain").
			SetPathParams(map[string]string{
				"valType": m.ValType,
				"name":    m.Name,
				"value":   m.Value,
			}).
			Post(path)
		if err != nil {
			logger.Error("Can't report metrics")
			break
		}
		logger.Debug("Metric: %s , status:%s", m.Name, resp.Status())
	}
}
