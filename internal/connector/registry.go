package connector

import (
	"fmt"
	"sync"
)

// Registry manages connector registration and lookup
type Registry struct {
	mu                sync.RWMutex
	externalConnectors map[string]ExternalResourceConnector
	cloudConnectors    map[string]CloudServiceConnector
}

// NewRegistry creates a new connector registry
func NewRegistry() *Registry {
	return &Registry{
		externalConnectors: make(map[string]ExternalResourceConnector),
		cloudConnectors:    make(map[string]CloudServiceConnector),
	}
}

// RegisterExternal adds an external resource connector to the registry
func (r *Registry) RegisterExternal(conn ExternalResourceConnector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.externalConnectors[conn.Type()] = conn
}

// RegisterCloud adds a cloud service connector to the registry
func (r *Registry) RegisterCloud(conn CloudServiceConnector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cloudConnectors[conn.Type()] = conn
}

// GetExternal retrieves an external resource connector
func (r *Registry) GetExternal(connType string) (ExternalResourceConnector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, ok := r.externalConnectors[connType]
	if !ok {
		return nil, fmt.Errorf("no external connector registered for type: %s", connType)
	}
	return conn, nil
}

// GetCloud retrieves a cloud service connector
func (r *Registry) GetCloud(connType string) (CloudServiceConnector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, ok := r.cloudConnectors[connType]
	if !ok {
		return nil, fmt.Errorf("no cloud connector registered for type: %s", connType)
	}
	return conn, nil
}

// HasExternal checks if an external connector is registered
func (r *Registry) HasExternal(connType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.externalConnectors[connType]
	return ok
}

// HasCloud checks if a cloud connector is registered
func (r *Registry) HasCloud(connType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.cloudConnectors[connType]
	return ok
}

// ExternalTypes returns all registered external connector types
func (r *Registry) ExternalTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.externalConnectors))
	for t := range r.externalConnectors {
		types = append(types, t)
	}
	return types
}

// CloudTypes returns all registered cloud connector types
func (r *Registry) CloudTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.cloudConnectors))
	for t := range r.cloudConnectors {
		types = append(types, t)
	}
	return types
}
