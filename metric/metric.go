package metric

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Meldiron/gatus/config"
	"github.com/Meldiron/gatus/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gauges = map[string]*prometheus.GaugeVec{}
	rwLock sync.RWMutex
)

// PublishMetricsForService publishes metrics for the given service and its result.
// These metrics will be exposed at /metrics if the metrics are enabled
func PublishMetricsForService(service *core.Service, result *core.Result) {
	if config.Get().Metrics {
		rwLock.Lock()
		gauge, exists := gauges[fmt.Sprintf("%s_%s", service.Name, service.URL)]
		if !exists {
			gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
				Subsystem:   "gatus",
				Name:        "tasks",
				ConstLabels: prometheus.Labels{"service": service.Name, "url": service.URL},
			}, []string{"status", "success"})
			gauges[fmt.Sprintf("%s_%s", service.Name, service.URL)] = gauge
		}
		rwLock.Unlock()
		gauge.WithLabelValues(strconv.Itoa(result.HTTPStatus), strconv.FormatBool(result.Success)).Inc()
	}
}
