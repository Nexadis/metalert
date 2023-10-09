// metrx реализует логику работы с метриками
package metrx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (
	Gauge   float64
	Counter int64
)

// Типы метрик
const (
	GaugeType   = `gauge`
	CounterType = `counter`
)

// Ошибки возникающие при обработке метрик
var (
	ErrorMetrics = errors.New("invalid metrics")
	ErrorType    = errors.New("invalid type")
)

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

// PrseCounter Получает Counter из строки
func ParseCounter(value string) (Counter, error) {
	val, err := strconv.Atoi(value)
	return Counter(val), err
}

// ParseGauge Получает Gauge из строки
func ParseGauge(value string) (Gauge, error) {
	val, err := strconv.ParseFloat(value, 64)
	return Gauge(val), err
}

// Metric - Структура для хранения метрики
type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *Counter `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *Gauge   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// NewMetric - Конструктор метрики, сам конвертирует строку в значение на основе типа
func NewMetric(id, mtype, value string) (Metric, error) {
	m := &Metric{
		ID:    id,
		MType: strings.ToLower(mtype),
	}
	err := m.SetValue(value)
	return *m, err
}

// SetValue() Парсит строку и сохраняет значение метрики. Определяет тип по MType
func (m *Metric) SetValue(value string) error {
	switch m.MType {
	case CounterType:
		v, err := ParseCounter(value)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrorMetrics, err)
		}
		m.Delta = &v
		m.Value = nil
	case GaugeType:
		v, err := ParseGauge(value)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrorMetrics, err)
		}
		m.Value = &v
		m.Delta = nil
	default:
		return fmt.Errorf("%v: %v", ErrorType, m)
	}
	return nil
}

// GetValue() Возвращает значение метрики в виде строки
func (m Metric) GetValue() (string, error) {
	switch m.MType {
	case CounterType:
		if m.Delta == nil {
			return "", ErrorMetrics
		}
		return m.Delta.String(), nil
	case GaugeType:
		if m.Value == nil {
			return "", ErrorMetrics
		}
		return m.Value.String(), nil
	}
	return "", fmt.Errorf("%v: %v", ErrorType, m.MType)
}
