package watchdog

import (
	"log"
	"sync"
	"time"

	"github.com/Meldiron/gatus/config"
	"github.com/Meldiron/gatus/core"
	"github.com/Meldiron/gatus/metric"
	"github.com/Meldiron/gatus/storage"
)

var (
	// monitoringMutex is used to prevent multiple services from being evaluated at the same time.
	// Without this, conditions using response time may become inaccurate.
	monitoringMutex sync.Mutex
)

// Monitor loops over each services and starts a goroutine to monitor each services separately
func Monitor(cfg *config.Config) {
	for _, service := range cfg.Services {
		// To prevent multiple requests from running at the same time, we'll wait for a little bit before each iteration
		time.Sleep(1111 * time.Millisecond)
		go monitor(service)
	}
}

// monitor monitors a single service in a loop
func monitor(service *core.Service) {
	cfg := config.Get()
	for {
		if !cfg.DisableMonitoringLock {
			// By placing the lock here, we prevent multiple services from being monitored at the exact same time, which
			// could cause performance issues and return inaccurate results
			monitoringMutex.Lock()
		}
		if cfg.Debug {
			log.Printf("[watchdog][monitor] Monitoring group=%s; service=%s", service.Group, service.Name)
		}
		result := service.EvaluateHealth()
		metric.PublishMetricsForService(service, result)
		UpdateServiceStatuses(service, result)
		log.Printf(
			"[watchdog][monitor] Monitored group=%s; service=%s; success=%v; errors=%d; duration=%s",
			service.Group,
			service.Name,
			result.Success,
			len(result.Errors),
			result.Duration.Round(time.Millisecond),
		)
		HandleAlerting(service, result)
		if cfg.Debug {
			log.Printf("[watchdog][monitor] Waiting for interval=%s before monitoring group=%s service=%s again", service.Interval, service.Group, service.Name)
		}
		if !cfg.DisableMonitoringLock {
			monitoringMutex.Unlock()
		}
		time.Sleep(service.Interval)
	}
}

// UpdateServiceStatuses updates the slice of service statuses
func UpdateServiceStatuses(service *core.Service, result *core.Result) {
	storage.Get().Insert(service, result)
}
