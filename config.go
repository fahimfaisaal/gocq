package varmq

import (
	"time"

	"github.com/goptics/varmq/utils"
)

type ConfigFunc func(*configs)

type configs struct {
	Concurrency          uint32
	Cache                ICache
	CleanupCacheInterval time.Duration
	JobIdGenerator       func() string
	IdleWorkerTimeout    time.Duration
	KeepIdleWorkers      uint16
}

func newConfig() configs {
	return configs{
		Concurrency: 1,
		Cache:       getCache(),
		JobIdGenerator: func() string {
			return ""
		},
	}
}

func loadConfigs(config ...any) configs {
	c := newConfig()

	return mergeConfigs(c, config...)
}

func mergeConfigs(c configs, cs ...any) configs {
	for _, config := range cs {
		switch config := config.(type) {
		case ConfigFunc:
			config(&c)
		case int:
			c.Concurrency = withSafeConcurrency(config)
		}
	}

	return c
}

func WithIdleWorkerTimeout(duration time.Duration) ConfigFunc {
	return func(c *configs) {
		c.IdleWorkerTimeout = duration
	}
}

// WithKeepIdleWorkers sets the percentage of idle workers to keep in the pool
// relative to the concurrency level.
// This helps scale the idle worker pool with concurrency changes.
// Example: 20 means keep 20% of the concurrency level as idle workers.
// The actual number of idle workers will be calculated as: concurrency * percentage / 100.
// The percentage must be between 1 and 100. Values outside this range will be clamped.
func WithKeepIdleWorkers(percentage uint16) ConfigFunc {
	// Clamp percentage between 1 and 100
	if percentage == 0 {
		percentage = 1
	} else if percentage > 100 {
		percentage = 100
	}

	return func(c *configs) {
		c.KeepIdleWorkers = percentage
	}
}

func WithCache(cache ICache) ConfigFunc {
	return func(c *configs) {
		c.Cache = cache
	}
}

func WithConcurrency(concurrency int) ConfigFunc {
	return func(c *configs) {
		c.Concurrency = withSafeConcurrency(concurrency)
	}
}

func WithAutoCleanupCache(duration time.Duration) ConfigFunc {
	return func(c *configs) {
		c.CleanupCacheInterval = duration
	}
}

func WithJobIdGenerator(fn func() string) ConfigFunc {
	return func(c *configs) {
		c.JobIdGenerator = fn
	}
}

func withSafeConcurrency(concurrency int) uint32 {
	// If concurrency is less than 1, use the number of CPUs as the concurrency
	if concurrency < 1 {
		return utils.Cpus()
	}
	return uint32(concurrency)
}
