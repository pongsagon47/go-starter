package queue

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"go-starter/pkg/logger"

	"go.uber.org/zap"
)

// RedisWorker implements Worker interface
type RedisWorker struct {
	queue      Queue
	handlers   map[string]Handler
	numWorkers int
	pollTime   time.Duration
	running    bool
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	logger     *zap.Logger
}

// WorkerConfig holds configuration for Redis worker
type WorkerConfig struct {
	NumWorkers int           // Number of concurrent workers
	PollTime   time.Duration // How often to poll for jobs
	Logger     *zap.Logger   // Logger instance
}

// NewRedisWorker creates a new Redis-based worker
func NewRedisWorker(queue Queue, config *WorkerConfig) *RedisWorker {
	if config == nil {
		config = &WorkerConfig{}
	}

	numWorkers := config.NumWorkers
	if numWorkers <= 0 {
		numWorkers = 4 // Default number of workers
	}

	pollTime := config.PollTime
	if pollTime <= 0 {
		pollTime = 1 * time.Second // Default poll time
	}

	workerLogger := config.Logger
	if workerLogger == nil {
		workerLogger = logger.Logger // Use default logger
	}

	return &RedisWorker{
		queue:      queue,
		handlers:   make(map[string]Handler),
		numWorkers: numWorkers,
		pollTime:   pollTime,
		logger:     workerLogger,
	}
}

// RegisterHandler registers a handler for a specific job type
func (w *RedisWorker) RegisterHandler(jobType string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.handlers[jobType] = handler
	w.logger.Info("Job handler registered",
		zap.String("job_type", jobType),
	)
}

// Start starts the worker
func (w *RedisWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("worker is already running")
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.running = true

	w.logger.Info("Starting worker",
		zap.Int("num_workers", w.numWorkers),
		zap.Duration("poll_time", w.pollTime),
	)

	// Start worker goroutines
	for i := 0; i < w.numWorkers; i++ {
		w.wg.Add(1)
		go w.workerLoop(i)
	}

	return nil
}

// Stop stops the worker gracefully
func (w *RedisWorker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.logger.Info("Stopping worker...")

	w.cancel()
	w.running = false

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("Worker stopped successfully")
	case <-time.After(30 * time.Second):
		w.logger.Warn("Worker stop timeout, some jobs may still be processing")
	}

	return nil
}

// IsRunning returns whether the worker is currently running
func (w *RedisWorker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// workerLoop is the main loop for each worker goroutine
func (w *RedisWorker) workerLoop(workerID int) {
	defer w.wg.Done()

	workerLogger := w.logger.With(zap.Int("worker_id", workerID))
	workerLogger.Info("Worker started")

	ticker := time.NewTicker(w.pollTime)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			workerLogger.Info("Worker shutting down")
			return
		case <-ticker.C:
			w.processNextJob(workerLogger)
		}
	}
}

// processNextJob processes the next available job
func (w *RedisWorker) processNextJob(workerLogger *zap.Logger) {
	// Get next job
	job, err := w.queue.Pop()
	if err != nil {
		workerLogger.Error("Failed to pop job from queue", zap.Error(err))
		return
	}

	if job == nil {
		// No jobs available
		return
	}

	// Process the job
	w.processJob(job, workerLogger)
}

// processJob processes a single job
func (w *RedisWorker) processJob(job *Job, workerLogger *zap.Logger) {
	jobLogger := workerLogger.With(
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
		zap.Int("attempt", job.Attempts+1),
		zap.Int("max_attempts", job.MaxAttempts),
	)

	jobLogger.Info("Processing job")
	startTime := time.Now()

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("job panicked: %v\nstack trace: %s", r, debug.Stack())
			jobLogger.Error("Job panicked", zap.Error(err))

			if nackErr := w.queue.Nack(job.ID, err); nackErr != nil {
				jobLogger.Error("Failed to nack job after panic", zap.Error(nackErr))
			}
		}
	}()

	// Get handler
	w.mu.RLock()
	handler, exists := w.handlers[job.Type]
	w.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no handler registered for job type: %s", job.Type)
		jobLogger.Error("No handler found", zap.Error(err))

		if nackErr := w.queue.Nack(job.ID, err); nackErr != nil {
			jobLogger.Error("Failed to nack job", zap.Error(nackErr))
		}
		return
	}

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	// Process the job
	result := handler.Handle(jobCtx, job)

	duration := time.Since(startTime)

	if result.Success {
		jobLogger.Info("Job completed successfully",
			zap.Duration("duration", duration),
			zap.Any("result_data", result.Data),
		)

		if err := w.queue.Ack(job.ID); err != nil {
			jobLogger.Error("Failed to ack job", zap.Error(err))
		}
	} else {
		err := fmt.Errorf(result.Error)
		jobLogger.Error("Job failed",
			zap.Duration("duration", duration),
			zap.Error(err),
		)

		if nackErr := w.queue.Nack(job.ID, err); nackErr != nil {
			jobLogger.Error("Failed to nack job", zap.Error(nackErr))
		}
	}
}

// JobDispatcher helps with job creation and dispatching
type JobDispatcher struct {
	queue Queue
}

// NewJobDispatcher creates a new job dispatcher
func NewJobDispatcher(queue Queue) *JobDispatcher {
	return &JobDispatcher{queue: queue}
}

// Dispatch creates and dispatches a job
func (jd *JobDispatcher) Dispatch(jobType string, payload map[string]interface{}, options ...*JobOptions) error {
	job := &Job{
		ID:          generateJobID(),
		Type:        jobType,
		Payload:     payload,
		CreatedAt:   time.Now(),
		MaxAttempts: 3,
		Priority:    0,
	}

	// Apply options
	if len(options) > 0 && options[0] != nil {
		opt := options[0]
		if opt.MaxAttempts > 0 {
			job.MaxAttempts = opt.MaxAttempts
		}
		job.Priority = opt.Priority
		job.Delay = opt.Delay
	}

	if job.Delay > 0 {
		return jd.queue.PushDelayed(job, job.Delay)
	}

	return jd.queue.Push(job)
}

// DispatchDelayed creates and dispatches a delayed job
func (jd *JobDispatcher) DispatchDelayed(jobType string, payload map[string]interface{}, delay time.Duration, options ...*JobOptions) error {
	job := &Job{
		ID:          generateJobID(),
		Type:        jobType,
		Payload:     payload,
		CreatedAt:   time.Now(),
		MaxAttempts: 3,
		Priority:    0,
		Delay:       delay,
	}

	// Apply options
	if len(options) > 0 && options[0] != nil {
		opt := options[0]
		if opt.MaxAttempts > 0 {
			job.MaxAttempts = opt.MaxAttempts
		}
		job.Priority = opt.Priority
	}

	return jd.queue.PushDelayed(job, delay)
}

// Common job types
const (
	JobTypeEmail             = "email"
	JobTypeEmailNotification = "email_notification"
	JobTypeImageProcessing   = "image_processing"
	JobTypeDataExport        = "data_export"
	JobTypeDataImport        = "data_import"
	JobTypeReportGeneration  = "report_generation"
	JobTypeCleanup           = "cleanup"
	JobTypeWebhook           = "webhook"
	JobTypeBackup            = "backup"
)

// Helper function to create job handlers

// EmailJobHandler creates a handler for email jobs
func EmailJobHandler(mailer interface{}) Handler {
	return HandlerFunc(func(ctx context.Context, job *Job) *JobResult {
		// Extract email data from job payload
		to, _ := job.Payload["to"].(string)
		subject, _ := job.Payload["subject"].(string)
		body, _ := job.Payload["body"].(string)

		if to == "" || subject == "" || body == "" {
			return &JobResult{
				Success: false,
				Error:   "missing required email fields: to, subject, body",
			}
		}

		// Here you would use your actual mailer
		// For example: mailer.SendEmail([]string{to}, subject, body, nil)

		logger.Info("Email job processed",
			zap.String("job_id", job.ID),
			zap.String("to", to),
			zap.String("subject", subject),
		)

		return &JobResult{
			Success: true,
			Data: map[string]interface{}{
				"to":      to,
				"subject": subject,
				"sent_at": time.Now(),
			},
		}
	})
}

// WebhookJobHandler creates a handler for webhook jobs
func WebhookJobHandler() Handler {
	return HandlerFunc(func(ctx context.Context, job *Job) *JobResult {
		url, _ := job.Payload["url"].(string)
		method, _ := job.Payload["method"].(string)
		data := job.Payload["data"]

		if url == "" {
			return &JobResult{
				Success: false,
				Error:   "missing webhook URL",
			}
		}

		if method == "" {
			method = "POST"
		}

		logger.Info("Webhook job processed",
			zap.String("job_id", job.ID),
			zap.String("url", url),
			zap.String("method", method),
		)

		// Here you would make the actual HTTP request
		// For example: http.Post(url, "application/json", bytes.NewBuffer(jsonData))

		return &JobResult{
			Success: true,
			Data: map[string]interface{}{
				"url":     url,
				"method":  method,
				"sent_at": time.Now(),
				"payload": data,
			},
		}
	})
}
