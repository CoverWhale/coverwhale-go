package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	Metrics []prometheus.Collector
}

func NewExporter() *Exporter {
	return &Exporter{}
}

func NewCounterVec(name, help string, labels []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        name,
			Help:        help,
			ConstLabels: nil,
		},
		labels,
	)
}

func NewHistogramVec(name, help string, labels []string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        name,
			Help:        help,
			ConstLabels: nil,
			Buckets:     []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		labels,
	)
}
