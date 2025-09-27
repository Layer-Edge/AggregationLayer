package utils

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	ErrorSeverityLow ErrorSeverity = iota
	ErrorSeverityMedium
	ErrorSeverityHigh
	ErrorSeverityCritical
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeDatabase   ErrorType = "database"
	ErrorTypeBlockchain ErrorType = "blockchain"
	ErrorTypeProcessing ErrorType = "processing"
	ErrorTypeSystem     ErrorType = "system"
	ErrorTypeUnknown    ErrorType = "unknown"
)

// ErrorInfo contains detailed information about an error
type ErrorInfo struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Severity   ErrorSeverity          `json:"severity"`
	Type       ErrorType              `json:"type"`
	Component  string                 `json:"component"`
	Message    string                 `json:"message"`
	Error      error                  `json:"error"`
	Stack      string                 `json:"stack"`
	Context    map[string]interface{} `json:"context"`
	RetryCount int                    `json:"retry_count"`
	Recovered  bool                   `json:"recovered"`
}

// Metrics holds system metrics
type Metrics struct {
	TotalErrors       int64                   `json:"total_errors"`
	ErrorsByType      map[ErrorType]int64     `json:"errors_by_type"`
	ErrorsBySeverity  map[ErrorSeverity]int64 `json:"errors_by_severity"`
	ErrorsByComponent map[string]int64        `json:"errors_by_component"`
	RecoveryRate      float64                 `json:"recovery_rate"`
	LastErrorTime     time.Time               `json:"last_error_time"`
	mutex             sync.RWMutex
}

// ErrorHandler manages error logging and metrics collection
type ErrorHandler struct {
	errors   []ErrorInfo
	metrics  Metrics
	handlers []ErrorHandlerFunc
	mutex    sync.RWMutex
}

// ErrorHandlerFunc is a function that handles errors
type ErrorHandlerFunc func(ErrorInfo)

var (
	globalErrorHandler *ErrorHandler
	once               sync.Once
)

// GetErrorHandler returns the global error handler instance
func GetErrorHandler() *ErrorHandler {
	once.Do(func() {
		globalErrorHandler = &ErrorHandler{
			errors: make([]ErrorInfo, 0),
			metrics: Metrics{
				ErrorsByType:      make(map[ErrorType]int64),
				ErrorsBySeverity:  make(map[ErrorSeverity]int64),
				ErrorsByComponent: make(map[string]int64),
			},
			handlers: make([]ErrorHandlerFunc, 0),
		}
	})
	return globalErrorHandler
}

// LogError logs an error with detailed information
func (eh *ErrorHandler) LogError(severity ErrorSeverity, errorType ErrorType, component string, message string, err error, context map[string]interface{}) ErrorInfo {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	// Generate stack trace
	stack := make([]byte, 1024)
	length := runtime.Stack(stack, false)
	stackTrace := string(stack[:length])

	errorInfo := ErrorInfo{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Severity:  severity,
		Type:      errorType,
		Component: component,
		Message:   message,
		Error:     err,
		Stack:     stackTrace,
		Context:   context,
	}

	// Add to errors list
	eh.errors = append(eh.errors, errorInfo)

	// Update metrics
	eh.metrics.TotalErrors++
	eh.metrics.ErrorsByType[errorType]++
	eh.metrics.ErrorsBySeverity[severity]++
	eh.metrics.ErrorsByComponent[component]++
	eh.metrics.LastErrorTime = time.Now()

	// Log to standard logger
	logLevel := "INFO"
	switch severity {
	case ErrorSeverityLow:
		logLevel = "INFO"
	case ErrorSeverityMedium:
		logLevel = "WARN"
	case ErrorSeverityHigh:
		logLevel = "ERROR"
	case ErrorSeverityCritical:
		logLevel = "FATAL"
	}

	log.Printf("[%s] [%s] [%s] %s: %v", logLevel, errorType, component, message, err)

	// Call registered handlers
	for _, handler := range eh.handlers {
		go handler(errorInfo)
	}

	return errorInfo
}

// LogErrorWithRetry logs an error with retry information
func (eh *ErrorHandler) LogErrorWithRetry(severity ErrorSeverity, errorType ErrorType, component string, message string, err error, retryCount int, recovered bool, context map[string]interface{}) ErrorInfo {
	errorInfo := eh.LogError(severity, errorType, component, message, err, context)
	errorInfo.RetryCount = retryCount
	errorInfo.Recovered = recovered

	// Update recovery rate
	eh.mutex.Lock()
	if recovered {
		eh.metrics.RecoveryRate = float64(eh.metrics.TotalErrors) / float64(eh.metrics.TotalErrors+1)
	}
	eh.mutex.Unlock()

	return errorInfo
}

// AddHandler adds a custom error handler
func (eh *ErrorHandler) AddHandler(handler ErrorHandlerFunc) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	eh.handlers = append(eh.handlers, handler)
}

// GetMetrics returns current metrics
func (eh *ErrorHandler) GetMetrics() Metrics {
	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	// Create a copy to avoid returning a struct with a mutex
	metrics := eh.metrics
	metrics.ErrorsByType = make(map[ErrorType]int64)
	metrics.ErrorsBySeverity = make(map[ErrorSeverity]int64)
	metrics.ErrorsByComponent = make(map[string]int64)

	// Copy maps
	for k, v := range eh.metrics.ErrorsByType {
		metrics.ErrorsByType[k] = v
	}
	for k, v := range eh.metrics.ErrorsBySeverity {
		metrics.ErrorsBySeverity[k] = v
	}
	for k, v := range eh.metrics.ErrorsByComponent {
		metrics.ErrorsByComponent[k] = v
	}

	return metrics
}

// GetErrors returns recent errors
func (eh *ErrorHandler) GetErrors(limit int) []ErrorInfo {
	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	if limit <= 0 || limit > len(eh.errors) {
		limit = len(eh.errors)
	}

	start := len(eh.errors) - limit
	if start < 0 {
		start = 0
	}

	return eh.errors[start:]
}

// ClearErrors clears old errors (keeps last 1000)
func (eh *ErrorHandler) ClearErrors() {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	if len(eh.errors) > 1000 {
		eh.errors = eh.errors[len(eh.errors)-1000:]
	}
}

// Convenience functions for common error logging patterns

// LogNetworkError logs a network-related error
func LogNetworkError(component string, message string, err error, context map[string]interface{}) ErrorInfo {
	return GetErrorHandler().LogError(ErrorSeverityMedium, ErrorTypeNetwork, component, message, err, context)
}

// LogDatabaseError logs a database-related error
func LogDatabaseError(component string, message string, err error, context map[string]interface{}) ErrorInfo {
	return GetErrorHandler().LogError(ErrorSeverityHigh, ErrorTypeDatabase, component, message, err, context)
}

// LogBlockchainError logs a blockchain-related error
func LogBlockchainError(component string, message string, err error, context map[string]interface{}) ErrorInfo {
	return GetErrorHandler().LogError(ErrorSeverityHigh, ErrorTypeBlockchain, component, message, err, context)
}

// LogProcessingError logs a processing-related error
func LogProcessingError(component string, message string, err error, context map[string]interface{}) ErrorInfo {
	return GetErrorHandler().LogError(ErrorSeverityMedium, ErrorTypeProcessing, component, message, err, context)
}

// LogSystemError logs a system-related error
func LogSystemError(component string, message string, err error, context map[string]interface{}) ErrorInfo {
	return GetErrorHandler().LogError(ErrorSeverityCritical, ErrorTypeSystem, component, message, err, context)
}

// LogCriticalError logs a critical error
func LogCriticalError(component string, message string, err error, context map[string]interface{}) ErrorInfo {
	return GetErrorHandler().LogError(ErrorSeverityCritical, ErrorTypeUnknown, component, message, err, context)
}

// LogErrorWithContext logs an error with additional context
func LogErrorWithContext(severity ErrorSeverity, errorType ErrorType, component string, message string, err error, ctx context.Context) ErrorInfo {
	context := map[string]interface{}{
		"context_deadline": ctx.Err(),
	}
	return GetErrorHandler().LogError(severity, errorType, component, message, err, context)
}

// RecoverFromPanic recovers from a panic and logs it
func RecoverFromPanic(component string) {
	if r := recover(); r != nil {
		LogSystemError(component, "Panic recovered", fmt.Errorf("panic: %v", r), map[string]interface{}{
			"panic_value": r,
		})
	}
}

// MonitorErrorRate monitors error rate and triggers alerts
func (eh *ErrorHandler) MonitorErrorRate(threshold float64, window time.Duration) {
	ticker := time.NewTicker(window)
	defer ticker.Stop()

	for range ticker.C {
		eh.mutex.RLock()
		recentErrors := 0
		cutoff := time.Now().Add(-window)

		for _, errorInfo := range eh.errors {
			if errorInfo.Timestamp.After(cutoff) {
				recentErrors++
			}
		}

		errorRate := float64(recentErrors) / window.Minutes()
		eh.mutex.RUnlock()

		if errorRate > threshold {
			LogSystemError("ErrorMonitor", "High error rate detected",
				fmt.Errorf("error rate %.2f exceeds threshold %.2f", errorRate, threshold),
				map[string]interface{}{
					"error_rate": errorRate,
					"threshold":  threshold,
					"window":     window,
				})
		}
	}
}
