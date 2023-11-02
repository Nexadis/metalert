package controller

import (
	"github.com/Nexadis/metalert/internal/models"
	pb "github.com/Nexadis/metalert/proto/metrics/v1"
)

func MetricsFromPB(ms *pb.Metrics) (models.Metrics, error) {
	result := make(models.Metrics, 0, len(ms.Metrics))
	for _, m := range ms.Metrics {
		newm, err := MetricFromPB(m)
		if err != nil {
			return nil, err
		}
		result = append(result, newm)
	}
	return result, nil
}

func typeFromPB(t pb.Metric_MType) (string, error) {
	switch t {
	case pb.Metric_M_TYPE_GAUGE:
		return models.GaugeType, nil
	case pb.Metric_M_TYPE_COUNTER:
		return models.CounterType, nil
	}
	return "", models.ErrorType
}

func typeToPB(t string) (pb.Metric_MType, error) {
	switch t {
	case models.GaugeType:
		return pb.Metric_M_TYPE_GAUGE, nil
	case models.CounterType:
		return pb.Metric_M_TYPE_COUNTER, nil
	}
	return pb.Metric_M_TYPE_UNSPECIFIED, models.ErrorType
}

func MetricsToPB(ms models.Metrics) (*pb.Metrics, error) {
	var result pb.Metrics
	metrics := make([]*pb.Metric, 0, len(ms))
	for _, m := range ms {
		pm, err := MetricToPB(m)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, pm)
	}
	result.Metrics = metrics
	return &result, nil
}

func MetricFromPB(m *pb.Metric) (models.Metric, error) {
	t, err := typeFromPB(m.GetType())
	if err != nil {
		return models.Metric{}, err
	}
	newm, err := models.NewMetric(m.GetId(), t, m.GetValue())
	if err != nil {
		return models.Metric{}, err
	}
	return newm, nil
}

func MetricToPB(m models.Metric) (*pb.Metric, error) {
	var pm pb.Metric
	t, err := typeToPB(m.MType)
	if err != nil {
		return nil, err
	}
	pm.Id = m.ID
	pm.Type = t
	v, err := m.GetValue()
	if err != nil {
		return nil, err
	}
	pm.Value = v
	return &pm, nil
}
