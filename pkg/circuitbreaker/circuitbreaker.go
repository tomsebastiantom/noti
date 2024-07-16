package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

const (
	StateClosed = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	state            int32
	failureCount     int32
	successCount     int32
	failureThreshold int32
	successThreshold int32
	resetTimeout     time.Duration
	lastFailureTime  time.Time
	mutex            sync.RWMutex
}

func NewCircuitBreaker(failureThreshold, successThreshold int32, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		resetTimeout:     resetTimeout,
	}
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.allowRequest() {
		return errors.New("circuit breaker is open")
	}

	err := operation()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		return cb.recordFailure(err)
	}

	return cb.recordSuccess()
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = StateHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordFailure(err error) error {
	cb.failureCount++
	cb.successCount = 0
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen || cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	}

	return err
}

func (cb *CircuitBreaker) recordSuccess() error {
	cb.successCount++

	if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
		cb.state = StateClosed
		cb.failureCount = 0
		cb.successCount = 0
	}

	return nil
}

// Metrics methods

func (cb *CircuitBreaker) GetState() int32 {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) GetFailureCount() int32 {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}

func (cb *CircuitBreaker) GetSuccessCount() int32 {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.successCount
}
