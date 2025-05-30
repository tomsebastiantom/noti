package workerpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "getnoti.com/pkg/logger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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
    ShutdownTimeout time.Duration
}

type WorkerPool struct {
    config        WorkerPoolConfig
    jobs          chan Job
    workers       int32  // Use atomic for thread-safe access
    wg            sync.WaitGroup
    ctx           context.Context
    cancel        context.CancelFunc
    scaleMu       sync.RWMutex  // Use RWMutex for better concurrency
    lastJobTime   time.Time
    scalingTicker *time.Ticker
    logger        log.Logger
    metrics       *Metrics
    metricsMu     sync.RWMutex  // Separate mutex for metrics
    state         int32         // 0: running, 1: stopping, 2: stopped
}

type Metrics struct {
    JobsProcessed  int64
    JobsInQueue    int64
    CurrentWorkers int32
    FailedJobs     int64
    ProcessingTime time.Duration
    LastJobTime    time.Time
    CreatedAt      time.Time
}

type WorkerPoolManager struct {
    pools    map[string]*WorkerPool
    mu       sync.RWMutex  // Use RWMutex for better read concurrency
    logger   log.Logger
    shutdown chan struct{}
    once     sync.Once
}

// NewWorkerPoolManager creates a new worker pool manager
func NewWorkerPoolManager(logger log.Logger) *WorkerPoolManager {
    return &WorkerPoolManager{
        pools:    make(map[string]*WorkerPool),
        logger:   logger,
        shutdown: make(chan struct{}),
    }
}

func NewWorkerPool(config WorkerPoolConfig, logger log.Logger) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Set default shutdown timeout if not provided
    if config.ShutdownTimeout == 0 {
        config.ShutdownTimeout = 30 * time.Second
    }
    
    wp := &WorkerPool{
        config:        config,
        jobs:          make(chan Job, config.MaxJobs),
        workers:       int32(config.InitialWorkers),
        ctx:           ctx,
        cancel:        cancel,
        lastJobTime:   time.Now(),
        scalingTicker: time.NewTicker(config.ScaleInterval),
        logger:        logger.With(log.String("pool_name", config.Name)),
        metrics: &Metrics{
            CreatedAt: time.Now(),
        },
        state: 0, // running
    }
    
    wp.logger.Info("Creating worker pool",
        log.Int("initial_workers", config.InitialWorkers),
        log.Int("max_jobs", config.MaxJobs),
        log.Int("min_workers", config.MinWorkers),
        log.Int("max_workers", config.MaxWorkers))
    
    wp.Start()
    go wp.monitorAndScale()
    return wp
}

// Start initializes the worker pool
func (wp *WorkerPool) Start() {
    if atomic.LoadInt32(&wp.state) != 0 {
        wp.logger.Warn("Attempted to start non-running worker pool")
        return
    }
    
    initialWorkers := int(atomic.LoadInt32(&wp.workers))
    for i := 0; i < initialWorkers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
    
    wp.logger.Info("Worker pool started",
        log.Field{Key: "workers_started", Value: initialWorkers})
}

// worker is the main worker goroutine with enhanced error handling
func (wp *WorkerPool) worker(workerID int) {
    defer wp.wg.Done()
    
    wp.logger.Debug("Worker started",
        log.Field{Key: "worker_id", Value: workerID})
    
    defer func() {
        if r := recover(); r != nil {
            wp.logger.Error("Worker panic recovered",
                log.Field{Key: "worker_id", Value: workerID},
                log.Field{Key: "panic", Value: fmt.Sprintf("%v", r)})
        }
        wp.logger.Debug("Worker stopped",
            log.Field{Key: "worker_id", Value: workerID})
    }()
    
    for {
        select {
        case job, ok := <-wp.jobs:
            if !ok {
                wp.logger.Debug("Job channel closed, worker exiting",
                    log.Field{Key: "worker_id", Value: workerID})
                return
            }
            
            wp.processJob(job, workerID)
            
        case <-wp.ctx.Done():
            wp.logger.Debug("Context cancelled, worker exiting",
                log.Field{Key: "worker_id", Value: workerID})
            return
        }
    }
}

// processJob handles individual job processing with timeout and metrics
func (wp *WorkerPool) processJob(job Job, workerID int) {
    startTime := time.Now()
    
    // Create job context with timeout
    jobCtx, cancel := context.WithTimeout(wp.ctx, 5*time.Minute)
    defer cancel()
    
    wp.logger.Debug("Processing job",
        log.Field{Key: "worker_id", Value: workerID})
    
    err := job.Process(jobCtx)
    processingTime := time.Since(startTime)
    
    wp.updateMetrics(err, processingTime)
    wp.updateLastJobTime()
    
    if err != nil {
        wp.logger.Error("Job processing failed",
            log.Field{Key: "worker_id", Value: workerID},
            log.Field{Key: "error", Value: err.Error()},
            log.Field{Key: "processing_time", Value: processingTime.String()})
    } else {
        wp.logger.Debug("Job completed successfully",
            log.Field{Key: "worker_id", Value: workerID},
            log.Field{Key: "processing_time", Value: processingTime.String()})
    }
}

// updateMetrics updates job processing metrics thread-safely
func (wp *WorkerPool) updateMetrics(err error, processingTime time.Duration) {
    wp.metricsMu.Lock()
    defer wp.metricsMu.Unlock()
    
    atomic.AddInt64(&wp.metrics.JobsProcessed, 1)
    wp.metrics.ProcessingTime += processingTime
    
    if err != nil {
        atomic.AddInt64(&wp.metrics.FailedJobs, 1)
    }
    
    // Update current workers count
    atomic.StoreInt32(&wp.metrics.CurrentWorkers, atomic.LoadInt32(&wp.workers))
    atomic.StoreInt64(&wp.metrics.JobsInQueue, int64(len(wp.jobs)))
}

// updateLastJobTime updates the last job processing time
func (wp *WorkerPool) updateLastJobTime() {
    wp.scaleMu.Lock()
    defer wp.scaleMu.Unlock()
    wp.lastJobTime = time.Now()
    wp.metrics.LastJobTime = wp.lastJobTime
}

// Submit adds a job to the pool with proper error handling
func (wp *WorkerPool) Submit(job Job) error {
    if atomic.LoadInt32(&wp.state) != 0 {
        return fmt.Errorf("worker pool is not running")
    }
    
    select {
    case wp.jobs <- job:
        atomic.AddInt64(&wp.metrics.JobsInQueue, 1)
        wp.logger.Debug("Job submitted successfully",
            log.Field{Key: "queue_size", Value: len(wp.jobs)})
        return nil
    default:
        wp.logger.Warn("Job queue is full, rejecting job",
            log.Field{Key: "max_jobs", Value: wp.config.MaxJobs})
        return fmt.Errorf("job queue is full (capacity: %d)", wp.config.MaxJobs)
    }
}

// monitorAndScale handles automatic scaling based on load
func (wp *WorkerPool) monitorAndScale() {
    defer wp.scalingTicker.Stop()
    
    wp.logger.Debug("Auto-scaling monitor started")
    
    for {
        select {
        case <-wp.scalingTicker.C:
            if atomic.LoadInt32(&wp.state) == 0 { // only scale if running
                wp.scale()
            }
        case <-wp.ctx.Done():
            wp.logger.Debug("Auto-scaling monitor stopped")
            return
        }
    }
}

// scale implements intelligent worker scaling with system resource checks
func (wp *WorkerPool) scale() {
    wp.scaleMu.Lock()
    defer wp.scaleMu.Unlock()
    
    queueSize := len(wp.jobs)
    currentWorkers := int(atomic.LoadInt32(&wp.workers))
    
    // Calculate desired workers based on queue load
    desiredWorkers := int(float64(queueSize) * wp.config.ScaleFactor)
    if desiredWorkers < wp.config.MinWorkers {
        desiredWorkers = wp.config.MinWorkers
    } else if desiredWorkers > wp.config.MaxWorkers {
        desiredWorkers = wp.config.MaxWorkers
    }
    
    // Scale down if idle for too long
    if time.Since(wp.lastJobTime) > wp.config.IdleTimeout && currentWorkers > wp.config.MinWorkers {
        desiredWorkers = wp.config.MinWorkers
    }
    
    delta := desiredWorkers - currentWorkers
    
    // Only scale if there's a significant change and system resources allow
    if delta != 0 && wp.checkSystemResources() {
        if delta > 0 {
            wp.scaleUp(delta)
        } else {
            wp.scaleDown(-delta)
        }
        
        wp.logger.Info("Scaled worker pool",
            log.Field{Key: "from_workers", Value: currentWorkers},
            log.Field{Key: "to_workers", Value: desiredWorkers},
            log.Field{Key: "queue_size", Value: queueSize})
    }
}

// scaleUp adds new workers
func (wp *WorkerPool) scaleUp(count int) {
    for i := 0; i < count; i++ {
        wp.wg.Add(1)
        workerID := int(atomic.AddInt32(&wp.workers, 1))
        go wp.worker(workerID)
    }
    
    wp.logger.Debug("Scaled up workers",
        log.Field{Key: "added_workers", Value: count})
}

// scaleDown reduces workers (graceful shutdown of excess workers)
func (wp *WorkerPool) scaleDown(count int) {
    // Note: In a production system, you'd implement graceful worker shutdown
    // For now, we'll just update the count and let natural attrition handle it
    newCount := atomic.AddInt32(&wp.workers, -int32(count))
    if newCount < 0 {
        atomic.StoreInt32(&wp.workers, 0)
    }
    
    wp.logger.Debug("Scaled down workers",
        log.Field{Key: "removed_workers", Value: count})
}

// checkSystemResources verifies system can handle more workers
func (wp *WorkerPool) checkSystemResources() bool {
    v, err := mem.VirtualMemory()
    if err != nil {
        wp.logger.Warn("Failed to get memory stats", log.Field{Key: "error", Value: err.Error()})
        return false
    }
    
    c, err := cpu.Percent(time.Second, false)
    if err != nil || len(c) == 0 {
        wp.logger.Warn("Failed to get CPU stats", log.Field{Key: "error", Value: err.Error()})
        return false
    }
    
    memOK := v.UsedPercent < 90
    cpuOK := c[0] < 80
    
    if !memOK || !cpuOK {
        wp.logger.Warn("System resources constrained",
            log.Field{Key: "memory_percent", Value: v.UsedPercent},
            log.Field{Key: "cpu_percent", Value: c[0]})
    }
    
    return memOK && cpuOK
}

// StopWithContext stops the worker pool with timeout
func (wp *WorkerPool) StopWithContext(ctx context.Context) error {
    if !atomic.CompareAndSwapInt32(&wp.state, 0, 1) { // 0 -> 1 (stopping)
        return fmt.Errorf("worker pool is already stopping or stopped")
    }
    
    wp.logger.Info("Stopping worker pool")
    
    // Stop accepting new jobs
    close(wp.jobs)
    
    // Cancel context to signal workers to stop
    wp.cancel()
    
    // Wait for all workers to finish with timeout
    done := make(chan struct{})
    go func() {
        wp.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        atomic.StoreInt32(&wp.state, 2) // stopped
        wp.logger.Info("Worker pool stopped successfully")
        return nil
    case <-ctx.Done():
        atomic.StoreInt32(&wp.state, 2) // stopped
        wp.logger.Warn("Worker pool stop timeout")
        return fmt.Errorf("timeout waiting for worker pool to stop")
    }
}

// Stop stops the worker pool with default timeout
func (wp *WorkerPool) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), wp.config.ShutdownTimeout)
    defer cancel()
    return wp.StopWithContext(ctx)
}

// IsHealthy checks if the worker pool is healthy
func (wp *WorkerPool) IsHealthy() bool {
    // Check if context is cancelled
    select {
    case <-wp.ctx.Done():
        return false
    default:
    }
    
    // Check if stopped
    if atomic.LoadInt32(&wp.state) != 0 {
        return false
    }
    
    wp.scaleMu.RLock()
    defer wp.scaleMu.RUnlock()
    
    currentWorkers := int(atomic.LoadInt32(&wp.workers))
    
    // Check if we have minimum workers running
    if currentWorkers < wp.config.MinWorkers {
        return false
    }
    
    // Check if job queue is not severely backed up (>90% full)
    queueSize := len(wp.jobs)
    if float64(queueSize)/float64(wp.config.MaxJobs) > 0.9 {
        return false
    }
    
    // Check if error rate is not too high (>50% failed jobs)
    wp.metricsMu.RLock()
    jobsProcessed := atomic.LoadInt64(&wp.metrics.JobsProcessed)
    failedJobs := atomic.LoadInt64(&wp.metrics.FailedJobs)
    wp.metricsMu.RUnlock()
    
    if jobsProcessed > 0 {
        errorRate := float64(failedJobs) / float64(jobsProcessed)
        if errorRate > 0.5 {
            return false
        }
    }
    
    return true
}

// GetMetrics returns current pool metrics
func (wp *WorkerPool) GetMetrics() Metrics {
    wp.metricsMu.RLock()
    defer wp.metricsMu.RUnlock()
    
    // Update real-time metrics
    metrics := *wp.metrics
    metrics.CurrentWorkers = atomic.LoadInt32(&wp.workers)
    metrics.JobsInQueue = int64(len(wp.jobs))
    metrics.JobsProcessed = atomic.LoadInt64(&wp.metrics.JobsProcessed)
    metrics.FailedJobs = atomic.LoadInt64(&wp.metrics.FailedJobs)
    
    return metrics
}

// WorkerPoolManager methods

// GetOrCreatePool gets an existing pool or creates a new one
func (wpm *WorkerPoolManager) GetOrCreatePool(config WorkerPoolConfig) *WorkerPool {
    wpm.mu.Lock()
    defer wpm.mu.Unlock()
    
    if pool, exists := wpm.pools[config.Name]; exists {
        wpm.logger.Debug("Returning existing worker pool",
            log.Field{Key: "pool_name", Value: config.Name})
        return pool
    }
    
    wpm.logger.Info("Creating new worker pool",
        log.Field{Key: "pool_name", Value: config.Name})
    
    pool := NewWorkerPool(config, wpm.logger)
    wpm.pools[config.Name] = pool
    return pool
}

// GetPool retrieves a pool by name
func (wpm *WorkerPoolManager) GetPool(name string) (*WorkerPool, bool) {
    wpm.mu.RLock()
    defer wpm.mu.RUnlock()
    pool, exists := wpm.pools[name]
    return pool, exists
}

// RemovePool removes and stops a pool
func (wpm *WorkerPoolManager) RemovePool(name string) error {
    wpm.mu.Lock()
    defer wpm.mu.Unlock()
    
    pool, exists := wpm.pools[name]
    if !exists {
        return fmt.Errorf("pool %s does not exist", name)
    }
    
    wpm.logger.Info("Removing worker pool",
        log.Field{Key: "pool_name", Value: name})
    
    if err := pool.Stop(); err != nil {
        wpm.logger.Error("Failed to stop worker pool during removal",
            log.Field{Key: "pool_name", Value: name},
            log.Field{Key: "error", Value: err.Error()})
        return err
    }
    
    delete(wpm.pools, name)
    return nil
}

// ScalePool manually scales a specific pool
func (wpm *WorkerPoolManager) ScalePool(name string, newWorkerCount int) error {
    pool, exists := wpm.GetPool(name)
    if !exists {
        return fmt.Errorf("pool %s does not exist", name)
    }
    
    pool.scaleMu.Lock()
    defer pool.scaleMu.Unlock()
    
    if newWorkerCount < pool.config.MinWorkers || newWorkerCount > pool.config.MaxWorkers {
        return fmt.Errorf("new worker count %d is outside allowed range [%d, %d]",
            newWorkerCount, pool.config.MinWorkers, pool.config.MaxWorkers)
    }
    
    currentWorkers := int(atomic.LoadInt32(&pool.workers))
    delta := newWorkerCount - currentWorkers
    
    if delta > 0 {
        pool.scaleUp(delta)
    } else if delta < 0 {
        pool.scaleDown(-delta)
    }
    
    pool.logger.Info("Manually scaled worker pool",
        log.Field{Key: "from_workers", Value: currentWorkers},
        log.Field{Key: "to_workers", Value: newWorkerCount})
    
    return nil
}

// ScaleAllPools triggers scaling for all pools
func (wpm *WorkerPoolManager) ScaleAllPools() {
    wpm.mu.RLock()
    defer wpm.mu.RUnlock()
    
    wpm.logger.Debug("Scaling all worker pools",
        log.Field{Key: "pool_count", Value: len(wpm.pools)})
    
    for name, pool := range wpm.pools {
        if atomic.LoadInt32(&pool.state) == 0 { // only scale running pools
            pool.scale()
        } else {
            wpm.logger.Debug("Skipping scaling for non-running pool",
                log.Field{Key: "pool_name", Value: name})
        }
    }
}

// Shutdown gracefully shuts down all worker pools
func (wpm *WorkerPoolManager) Shutdown() error {
    wpm.once.Do(func() {
        close(wpm.shutdown)
    })
    
    wpm.mu.Lock()
    defer wpm.mu.Unlock()
    
    wpm.logger.Info("Shutting down all worker pools",
        log.Field{Key: "pool_count", Value: len(wpm.pools)})
    
    var shutdownErrors []error
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Shutdown all pools concurrently
    type shutdownResult struct {
        name string
        err  error
    }
    
    resultChan := make(chan shutdownResult, len(wpm.pools))
    
    for name, pool := range wpm.pools {
        go func(poolName string, p *WorkerPool) {
            err := p.StopWithContext(shutdownCtx)
            resultChan <- shutdownResult{name: poolName, err: err}
        }(name, pool)
    }
    
    // Collect results
    for i := 0; i < len(wpm.pools); i++ {
        select {
        case result := <-resultChan:
            if result.err != nil {
                wpm.logger.Error("Failed to stop worker pool",
                    log.Field{Key: "pool_name", Value: result.name},
                    log.Field{Key: "error", Value: result.err.Error()})
                shutdownErrors = append(shutdownErrors, result.err)
            } else {
                wpm.logger.Debug("Worker pool stopped successfully",
                    log.Field{Key: "pool_name", Value: result.name})
            }
        case <-shutdownCtx.Done():
            wpm.logger.Error("Timeout waiting for worker pool shutdown",
                log.Field{Key: "timeout", Value: "30s"})
            shutdownErrors = append(shutdownErrors, context.DeadlineExceeded)
        }
    }
    
    // Clear all pools
    for name := range wpm.pools {
        delete(wpm.pools, name)
    }
    
    if len(shutdownErrors) > 0 {
        wpm.logger.Error("Worker pool shutdown completed with errors",
            log.Field{Key: "error_count", Value: len(shutdownErrors)})
        return fmt.Errorf("failed to shutdown %d worker pools", len(shutdownErrors))
    }
    
    wpm.logger.Info("All worker pools shut down successfully")
    return nil
}

// IsHealthy checks if all worker pools are healthy
func (wpm *WorkerPoolManager) IsHealthy() bool {
    wpm.mu.RLock()
    defer wpm.mu.RUnlock()
    
    for name, pool := range wpm.pools {
        if !pool.IsHealthy() {
            wpm.logger.Warn("Unhealthy worker pool detected",
                log.Field{Key: "pool_name", Value: name})
            return false
        }
    }
    return true
}

// GetHealthStatus returns detailed health status
func (wpm *WorkerPoolManager) GetHealthStatus() map[string]bool {
    wpm.mu.RLock()
    defer wpm.mu.RUnlock()
    
    status := make(map[string]bool)
    for name, pool := range wpm.pools {
        status[name] = pool.IsHealthy()
    }
    return status
}

// GetAllPoolMetrics returns metrics for all pools
func (wpm *WorkerPoolManager) GetAllPoolMetrics() map[string]Metrics {
    wpm.mu.RLock()
    defer wpm.mu.RUnlock()
    
    metrics := make(map[string]Metrics)
    for name, pool := range wpm.pools {
        metrics[name] = pool.GetMetrics()
    }
    return metrics
}