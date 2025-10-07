package utils

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// AlertLevel represents the severity of an alert
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelError
	AlertLevelCritical
)

// Alert represents a system alert
type Alert struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Level        AlertLevel             `json:"level"`
	Component    string                 `json:"component"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details"`
	Acknowledged bool                   `json:"acknowledged"`
}

// HealthStatus represents the health status of a component
type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusDegraded
	HealthStatusUnhealthy
	HealthStatusUnknown
)

// ComponentHealth represents the health of a system component
type ComponentHealth struct {
	Name       string                 `json:"name"`
	Status     HealthStatus           `json:"status"`
	LastCheck  time.Time              `json:"last_check"`
	Details    map[string]interface{} `json:"details"`
	ErrorCount int64                  `json:"error_count"`
	LastError  time.Time              `json:"last_error"`
}

// Monitor manages system monitoring and alerting
type Monitor struct {
	components map[string]*ComponentHealth
	alerts     []Alert
	handlers   []AlertHandler
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// AlertHandler is a function that handles alerts
type AlertHandler func(Alert)

// HealthChecker is a function that checks component health
type HealthChecker func() (HealthStatus, map[string]interface{}, error)

var (
	globalMonitor *Monitor
	monitorOnce   sync.Once
)

// GetMonitor returns the global monitor instance
func GetMonitor() *Monitor {
	monitorOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		globalMonitor = &Monitor{
			components: make(map[string]*ComponentHealth),
			alerts:     make([]Alert, 0),
			handlers:   make([]AlertHandler, 0),
			ctx:        ctx,
			cancel:     cancel,
		}
	})
	return globalMonitor
}

// RegisterComponent registers a component for health monitoring
func (m *Monitor) RegisterComponent(name string, checker HealthChecker, checkInterval time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.components[name] = &ComponentHealth{
		Name:      name,
		Status:    HealthStatusUnknown,
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Start health checking goroutine
	go m.healthCheckLoop(name, checker, checkInterval)
}

// healthCheckLoop continuously checks component health
func (m *Monitor) healthCheckLoop(name string, checker HealthChecker, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			status, details, err := checker()

			m.mutex.Lock()
			component := m.components[name]
			component.Status = status
			component.LastCheck = time.Now()
			component.Details = details

			if err != nil {
				component.ErrorCount++
				component.LastError = time.Now()

				// Create alert for health check failure
				alert := Alert{
					ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
					Timestamp: time.Now(),
					Level:     AlertLevelError,
					Component: name,
					Message:   fmt.Sprintf("Health check failed: %v", err),
					Details: map[string]interface{}{
						"error":  err.Error(),
						"status": status,
					},
				}

				m.alerts = append(m.alerts, alert)

				// Notify handlers
				for _, handler := range m.handlers {
					go handler(alert)
				}
			}
			m.mutex.Unlock()
		}
	}
}

// AddAlertHandler adds a custom alert handler
func (m *Monitor) AddAlertHandler(handler AlertHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.handlers = append(m.handlers, handler)
}

// GetComponentHealth returns the health status of a component
func (m *Monitor) GetComponentHealth(name string) (*ComponentHealth, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	component, exists := m.components[name]
	return component, exists
}

// GetAllComponents returns all component health statuses
func (m *Monitor) GetAllComponents() map[string]*ComponentHealth {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*ComponentHealth)
	for name, component := range m.components {
		result[name] = component
	}
	return result
}

// GetAlerts returns recent alerts
func (m *Monitor) GetAlerts(limit int) []Alert {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	start := len(m.alerts) - limit
	if start < 0 {
		start = 0
	}

	return m.alerts[start:]
}

// AcknowledgeAlert acknowledges an alert
func (m *Monitor) AcknowledgeAlert(alertID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, alert := range m.alerts {
		if alert.ID == alertID {
			m.alerts[i].Acknowledged = true
			return true
		}
	}
	return false
}

// CreateAlert creates a manual alert
func (m *Monitor) CreateAlert(level AlertLevel, component string, message string, details map[string]interface{}) Alert {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	alert := Alert{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
		Details:   details,
	}

	m.alerts = append(m.alerts, alert)

	// Notify handlers
	for _, handler := range m.handlers {
		go handler(alert)
	}

	return alert
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	m.cancel()
}

// Convenience functions for common monitoring patterns

// MonitorDatabaseHealth monitors database health
func MonitorDatabaseHealth() (HealthStatus, map[string]interface{}, error) {
	// This would typically check database connection, query performance, etc.
	// For now, we'll simulate a basic check
	return HealthStatusHealthy, map[string]interface{}{
		"connection_pool_size": 25,
		"active_connections":   5,
	}, nil
}

// MonitorNetworkHealth monitors network connectivity
func MonitorNetworkHealth() (HealthStatus, map[string]interface{}, error) {
	// This would typically check network connectivity to external services
	return HealthStatusHealthy, map[string]interface{}{
		"bitcoin_rpc_available":   true,
		"layeredge_rpc_available": true,
	}, nil
}

// MonitorSystemHealth monitors system resources
func MonitorSystemHealth() (HealthStatus, map[string]interface{}, error) {
	// This would typically check CPU, memory, disk usage, etc.
	return HealthStatusHealthy, map[string]interface{}{
		"cpu_usage":    45.2,
		"memory_usage": 67.8,
		"disk_usage":   23.1,
	}, nil
}

// DefaultAlertHandler is a default alert handler that logs alerts
func DefaultAlertHandler(alert Alert) {
	level := "INFO"
	switch alert.Level {
	case AlertLevelInfo:
		level = "INFO"
	case AlertLevelWarning:
		level = "WARN"
	case AlertLevelError:
		level = "ERROR"
	case AlertLevelCritical:
		level = "CRITICAL"
	}

	log.Printf("[%s] [%s] %s: %s", level, alert.Component, alert.Message, alert.Details)
}

// InitializeMonitoring initializes the monitoring system with default components
func InitializeMonitoring() *Monitor {
	monitor := GetMonitor()

	// Add default alert handler
	monitor.AddAlertHandler(DefaultAlertHandler)

	// Register default components
	monitor.RegisterComponent("database", MonitorDatabaseHealth, 30*time.Second)
	monitor.RegisterComponent("network", MonitorNetworkHealth, 60*time.Second)
	monitor.RegisterComponent("system", MonitorSystemHealth, 30*time.Second)

	return monitor
}
