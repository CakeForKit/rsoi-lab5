package service

import (
	"sync"
	"time"
)

const (
	circuitFailureThreshold = 3
	circuitRetryTimeout     = 3 * time.Second
)

// circuitBreaker прекращает запросы к недоступному сервису после нескольких
// последовательных ошибок и периодически проверяет его восстановление.
type circuitBreaker struct {
	mu          sync.Mutex
	failures    int
	openedAt    time.Time
	halfOpenRun bool
}

func (c *circuitBreaker) allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.openedAt.IsZero() {
		return true
	}
	if time.Since(c.openedAt) < circuitRetryTimeout || c.halfOpenRun {
		return false
	}

	c.halfOpenRun = true
	return true
}

func (c *circuitBreaker) success() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures = 0
	c.openedAt = time.Time{}
	c.halfOpenRun = false
}

func (c *circuitBreaker) failure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.halfOpenRun {
		c.openedAt = time.Now()
		c.halfOpenRun = false
		return
	}

	c.failures++
	if c.failures >= circuitFailureThreshold {
		c.openedAt = time.Now()
	}
}
