package da

import (
	// "encoding/hex"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Layer-Edge/bitcoin-da/utils"
	"gopkg.in/zeromq/goczmq.v4"
)

type Lambda func([][]byte) bool
type Lambda2 func([][]byte) ([]byte, error)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements a circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	state        CircuitBreakerState
	failureCount int
	lastFailTime time.Time
	timeout      time.Duration
	maxFailures  int
	mutex        sync.RWMutex
}

// RetryConfig holds configuration for retry mechanisms
type RetryConfig struct {
	MaxRetries    int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

type BlockSubscriber struct {
	channeler      *goczmq.Channeler
	circuitBreaker *CircuitBreaker
	retryConfig    *RetryConfig
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
}

// NewBlockSubscriber creates a new BlockSubscriber with circuit breaker and retry mechanisms
func NewBlockSubscriber() *BlockSubscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockSubscriber{
		circuitBreaker: &CircuitBreaker{
			state:        CircuitClosed,
			failureCount: 0,
			timeout:      30 * time.Second,
			maxFailures:  5,
		},
		retryConfig: &RetryConfig{
			MaxRetries:    3,
			BaseDelay:     1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return time.Since(cb.lastFailTime) > cb.timeout
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
		log.Printf("Circuit breaker opened due to %d failures", cb.failureCount)
	}
}

// RetryWithBackoff executes a function with exponential backoff retry
func (subr *BlockSubscriber) RetryWithBackoff(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= subr.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(subr.retryConfig.BaseDelay) *
				utils.PowFloat(subr.retryConfig.BackoffFactor, float64(attempt-1)))
			if delay > subr.retryConfig.MaxDelay {
				delay = subr.retryConfig.MaxDelay
			}

			log.Printf("Retry attempt %d after %v delay", attempt, delay)

			select {
			case <-time.After(delay):
			case <-subr.ctx.Done():
				return fmt.Errorf("operation cancelled: %w", subr.ctx.Err())
			}
		}

		if !subr.circuitBreaker.CanExecute() {
			return fmt.Errorf("circuit breaker is open, operation rejected")
		}

		err := operation()
		if err == nil {
			subr.circuitBreaker.RecordSuccess()
			return nil
		}

		lastErr = err
		subr.circuitBreaker.RecordFailure()
		log.Printf("Operation failed (attempt %d/%d): %v", attempt+1, subr.retryConfig.MaxRetries+1, err)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", subr.retryConfig.MaxRetries+1, lastErr)
}

func (subr *BlockSubscriber) Subscribe(endpoint string, filter string) bool {
	log.Println("Subscribe:", endpoint, filter)

	return subr.RetryWithBackoff(func() error {
		subr.mu.Lock()
		defer subr.mu.Unlock()

		// Clean up existing channeler if any
		if subr.channeler != nil {
			subr.channeler.Destroy()
		}

		subr.channeler = goczmq.NewSubChanneler(endpoint, filter)
		if subr.channeler == nil {
			return fmt.Errorf("failed to create subscribe channeler for endpoint %s with filter %s", endpoint, filter)
		}

		log.Println("Successfully created subscribe channeler")
		return nil
	}) == nil
}

func (subr *BlockSubscriber) Replier(endpoint string) bool {
	log.Println("Replier:", endpoint)

	return subr.RetryWithBackoff(func() error {
		subr.mu.Lock()
		defer subr.mu.Unlock()

		// Clean up existing channeler if any
		if subr.channeler != nil {
			subr.channeler.Destroy()
		}

		subr.channeler = goczmq.NewRepChanneler(endpoint)
		if subr.channeler == nil {
			return fmt.Errorf("failed to create reply channeler for endpoint %s", endpoint)
		}

		log.Println("Successfully created reply channeler")
		return nil
	}) == nil
}

func (subr *BlockSubscriber) Reset() {
	subr.mu.Lock()
	defer subr.mu.Unlock()

	if subr.channeler != nil {
		subr.channeler.Destroy()
		subr.channeler = nil
	}
}

// Close gracefully shuts down the BlockSubscriber
func (subr *BlockSubscriber) Close() error {
	subr.cancel()
	subr.Reset()
	return nil
}

func (subr *BlockSubscriber) GetMessage() (bool, [][]byte) {
	subr.mu.RLock()
	defer subr.mu.RUnlock()

	if subr.channeler == nil {
		log.Println("Channeler is nil, cannot get message")
		return false, nil
	}

	select {
	case msg, ok := <-subr.channeler.RecvChan:
		return ok, msg
	case <-subr.ctx.Done():
		log.Println("Context cancelled, stopping message retrieval")
		return false, nil
	}
}

func (subr *BlockSubscriber) Validate(ok bool, msg [][]byte) bool {
	// Validate
	if !ok {
		log.Println("Failed to receive message")
		return false
	}
	if len(msg) != 3 {
		log.Printf("Received message with unexpected number of parts: expected 3, got %d", len(msg))
		return false
	}

	// Additional validation for message content
	for i, part := range msg {
		if len(part) == 0 {
			log.Printf("Message part %d is empty", i)
			return false
		}
	}

	return true
}

func (subr *BlockSubscriber) Process(fn Lambda, msg [][]byte) bool {
	log.Println("Processing message")

	// Add timeout protection for processing
	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in message processing: %v", r)
				done <- false
			}
		}()
		done <- fn(msg)
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(30 * time.Second):
		log.Println("Message processing timeout")
		return false
	case <-subr.ctx.Done():
		log.Println("Context cancelled during message processing")
		return false
	}
}

func (subr *BlockSubscriber) ProcessOutTuple(fn Lambda2, msg [][]byte) ([]byte, error) {
	log.Println("Processing message")

	// Add timeout protection for processing
	done := make(chan struct {
		result []byte
		err    error
	}, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in message processing: %v", r)
				done <- struct {
					result []byte
					err    error
				}{nil, fmt.Errorf("panic in processing: %v", r)}
			}
		}()

		result, err := fn(msg)
		done <- struct {
			result []byte
			err    error
		}{result, err}
	}()

	select {
	case res := <-done:
		return res.result, res.err
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("message processing timeout")
	case <-subr.ctx.Done():
		return nil, fmt.Errorf("context cancelled during message processing: %w", subr.ctx.Err())
	}
}
