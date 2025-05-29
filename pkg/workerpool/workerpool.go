package workerpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"getnoti.com/pkg/logger"
)

type Job interface {
	Process(ctx context.Context) error
}

type WorkerPoolConfig struct {
	Name           string
	InitialWorkers int
	MaxJobs        int
	MinWorkers     int
	MaxWorkers     int
	ScaleFactor    float64
	IdleTimeout    time.Duration
	ScaleInterval  time.Duration
}

type WorkerPool struct {
	config        WorkerPoolConfig
	jobs          chan Job
	workers       int
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	scaleMu       sync.Mutex
	lastJobTime   time.Time
	scalingTicker *time.Ticker
	logger        logger.Logger
	metrics       *Metrics
}

type Metrics struct {
	JobsProcessed  int64
	JobsInQueue    int64
	CurrentWorkers int
	FailedJobs     int64
	ProcessingTime time.Duration
}

func NewWorkerPool(config WorkerPoolConfig, logger logger.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		config:        config,
		jobs:          make(chan Job, config.MaxJobs),
		workers:       config.InitialWorkers,
		ctx:           ctx,
		cancel:        cancel,
		lastJobTime:   time.Now(),
		scalingTicker: time.NewTicker(config.ScaleInterval),
		logger:        logger,
		metrics:       &Metrics{},
	}
	wp.Start()
	go wp.monitorAndScale()
	return wp
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case job := <-wp.jobs:
			startTime := time.Now()
			err := job.Process(wp.ctx)
			wp.updateMetrics(err, time.Since(startTime))
			wp.updateLastJobTime()
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) updateMetrics(err error, processingTime time.Duration) {
	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()
	wp.metrics.JobsProcessed++
	wp.metrics.ProcessingTime += processingTime
	if err != nil {
		wp.metrics.FailedJobs++
	}
}

func (wp *WorkerPool) updateLastJobTime() {
	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()
	wp.lastJobTime = time.Now()
}

func (wp *WorkerPool) Submit(job Job) error {
	select {
	case wp.jobs <- job:
		wp.metrics.JobsInQueue++
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

func (wp *WorkerPool) monitorAndScale() {
	for {
		select {
		case <-wp.scalingTicker.C:
			wp.scale()
		case <-wp.ctx.Done():
			wp.scalingTicker.Stop()
			return
		}
	}
}

func (wp *WorkerPool) scale() {
    wp.scaleMu.Lock()
    defer wp.scaleMu.Unlock()

    queueSize := len(wp.jobs)
    currentWorkers := wp.workers

    desiredWorkers := int(float64(queueSize) * wp.config.ScaleFactor)
    if desiredWorkers < wp.config.MinWorkers {
        desiredWorkers = wp.config.MinWorkers
    } else if desiredWorkers > wp.config.MaxWorkers {
        desiredWorkers = wp.config.MaxWorkers
    }

    if time.Since(wp.lastJobTime) > wp.config.IdleTimeout && currentWorkers > wp.config.MinWorkers {
        desiredWorkers = wp.config.MinWorkers
    }

    delta := desiredWorkers - currentWorkers

    if delta != 0 && wp.checkSystemResources() {
        if delta > 0 {
            for i := 0; i < delta; i++ {
                wp.wg.Add(1)
                go wp.worker()
            }
        } else {
            for i := 0; i < -delta; i++ {
                wp.cancel()
            }
            ctx, cancel := context.WithCancel(wp.ctx)
            wp.ctx = ctx
            wp.cancel = cancel
        }
        wp.workers += delta
        // ✅ Fixed: Use structured logging
        wp.logger.Info("Scaled worker pool", 
            logger.Field{Key: "pool_name", Value: wp.config.Name},
            logger.Field{Key: "from_workers", Value: currentWorkers},
            logger.Field{Key: "to_workers", Value: wp.workers})
    }
}

func (wp *WorkerPool) checkSystemResources() bool {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(time.Second, false)
	return v.UsedPercent < 90 && c[0] < 80
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
}

func (wp *WorkerPool) GetMetrics() Metrics {
	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()
	wp.metrics.CurrentWorkers = wp.workers
	wp.metrics.JobsInQueue = int64(len(wp.jobs))
	return *wp.metrics
}

type WorkerPoolManager struct {
	pools  map[string]*WorkerPool
	mu     sync.Mutex
	logger logger.Logger
}

func NewWorkerPoolManager(logger logger.Logger) *WorkerPoolManager {
	return &WorkerPoolManager{
		pools:  make(map[string]*WorkerPool),
		logger: logger,
	}
}

func (wpm *WorkerPoolManager) GetOrCreatePool(config WorkerPoolConfig) *WorkerPool {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	if pool, exists := wpm.pools[config.Name]; exists {
		return pool
	}

	pool := NewWorkerPool(config, wpm.logger)
	wpm.pools[config.Name] = pool
	return pool
}

func (wpm *WorkerPoolManager) GetPool(name string) (*WorkerPool, bool) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()
	pool, exists := wpm.pools[name]
	return pool, exists
}

func (wpm *WorkerPoolManager) RemovePool(name string) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()
	if pool, exists := wpm.pools[name]; exists {
		pool.Stop()
		delete(wpm.pools, name)
	}
}

func (wpm *WorkerPoolManager) ScalePool(name string, newWorkerCount int) error {
    pool, exists := wpm.GetPool(name)
    if !exists {
        return fmt.Errorf("pool %s does not exist", name)
    }
    
    pool.scaleMu.Lock()
    defer pool.scaleMu.Unlock()

    if newWorkerCount < pool.config.MinWorkers || newWorkerCount > pool.config.MaxWorkers {
        return fmt.Errorf("new worker count %d is outside allowed range [%d, %d]", newWorkerCount, pool.config.MinWorkers, pool.config.MaxWorkers)
    }

    currentWorkers := pool.workers
    delta := newWorkerCount - currentWorkers

    if delta > 0 {
        for i := 0; i < delta; i++ {
            pool.wg.Add(1)
            go pool.worker()
        }
    } else if delta < 0 {
        for i := 0; i < -delta; i++ {
            pool.cancel()
        }
        ctx, cancel := context.WithCancel(pool.ctx)
        pool.ctx = ctx
        pool.cancel = cancel
    }

    pool.workers = newWorkerCount
    // ✅ Fixed: Use structured logging
    pool.logger.Info("Scaled worker pool", 
        logger.Field{Key: "pool_name", Value: name},
        logger.Field{Key: "from_workers", Value: currentWorkers},
        logger.Field{Key: "to_workers", Value: newWorkerCount})
    return nil
}

func (wpm *WorkerPoolManager) ScaleAllPools() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	for _, pool := range wpm.pools {
		pool.scale()
	}
}

func (wpm *WorkerPoolManager) ShutdownAll() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	for name, pool := range wpm.pools {
		pool.Stop()
		delete(wpm.pools, name)
	}
}

func (wpm *WorkerPoolManager) GetAllPoolMetrics() map[string]Metrics {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	metrics := make(map[string]Metrics)
	for name, pool := range wpm.pools {
		metrics[name] = pool.GetMetrics()
	}
	return metrics
}
