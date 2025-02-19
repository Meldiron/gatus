package store

import (
	"github.com/Meldiron/gatus/core"
	"github.com/Meldiron/gatus/storage/store/memory"
)

// Store is the interface that each stores should implement
type Store interface {
	// GetAllAsJSON returns the JSON encoding of all monitored core.ServiceStatus
	// with a subset of core.Result defined by the page and pageSize parameters
	GetAllServiceStatusesWithResultPagination(page, pageSize int) map[string]*core.ServiceStatus

	// GetServiceStatus returns the service status for a given service name in the given group
	GetServiceStatus(groupName, serviceName string) *core.ServiceStatus

	// GetServiceStatusByKey returns the service status for a given key
	GetServiceStatusByKey(key string) *core.ServiceStatus

	// Insert adds the observed result for the specified service into the store
	Insert(service *core.Service, result *core.Result)

	// DeleteAllServiceStatusesNotInKeys removes all ServiceStatus that are not within the keys provided
	//
	// Used to delete services that have been persisted but are no longer part of the configured services
	DeleteAllServiceStatusesNotInKeys(keys []string) int

	// Clear deletes everything from the store
	Clear()

	// Save persists the data if and where it needs to be persisted
	Save() error
}

var (
	// Validate interface implementation on compile
	_ Store = (*memory.Store)(nil)
)
