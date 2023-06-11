package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/Nexadis/metalert/internal/metrx"
)

type Agent interface {
	Run() error
	Pull()
	Report()
}

type httpAgent struct {
	listener       string
	pullInterval   int64
	reportInterval int64
	storage        metrx.MemStorage
	client         *http.Client
}

func NewAgent(listener string, pullInterval, reportInterval int64) Agent {
	storage := metrx.NewMetricsStorage()
	return &httpAgent{
		listener:       listener,
		pullInterval:   pullInterval,
		reportInterval: reportInterval,
		storage:        storage,
		client:         &http.Client{},
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
	for _, m := range values {
		path := fmt.Sprintf("%s/update/%s/%s/%s", ha.listener, m.ValType, m.Name, m.Value)
		fmt.Printf("Send %s http request\n", path)
		req, err := http.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-type", "text/plain")
		if err != nil {
			panic(err)
		}
		_, err = ha.client.Do(req)
		if err != nil {
			panic(err)
		}

	}
}
