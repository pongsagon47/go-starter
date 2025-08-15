package queue

import (
	"context"
	"time"
)

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Priority    int                    `json:"priority"` // Higher number = higher priority
	Delay       time.Duration          `json:"delay"`    // Delay before processing
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
	StatusRetrying   JobStatus = "retrying"
	StatusCancelled  JobStatus = "cancelled"
)

// JobResult represents the result of job processing
type JobResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Handler defines the interface for job handlers
type Handler interface {
	Handle(ctx context.Context, job *Job) *JobResult
}

// HandlerFunc allows using functions as handlers
type HandlerFunc func(ctx context.Context, job *Job) *JobResult

// Handle implements the Handler interface
func (fn HandlerFunc) Handle(ctx context.Context, job *Job) *JobResult {
	return fn(ctx, job)
}

// Queue defines the interface for job queues
type Queue interface {
	// Push adds a job to the queue
	Push(job *Job) error

	// PushDelayed adds a job to be processed after a delay
	PushDelayed(job *Job, delay time.Duration) error

	// Pop retrieves the next job from the queue
	Pop() (*Job, error)

	// Ack acknowledges successful job completion
	Ack(jobID string) error

	// Nack marks a job as failed and potentially retries it
	Nack(jobID string, err error) error

	// GetJob retrieves a job by ID
	GetJob(jobID string) (*Job, error)

	// GetJobStatus returns the current status of a job
	GetJobStatus(jobID string) (JobStatus, error)

	// CancelJob cancels a pending job
	CancelJob(jobID string) error

	// GetQueueSize returns the number of pending jobs
	GetQueueSize() (int64, error)

	// GetStats returns queue statistics
	GetStats() (*QueueStats, error)

	// Close closes the queue connection
	Close() error
}

// Worker defines the interface for job workers
type Worker interface {
	// RegisterHandler registers a handler for a specific job type
	RegisterHandler(jobType string, handler Handler)

	// Start starts the worker
	Start(ctx context.Context) error

	// Stop stops the worker gracefully
	Stop() error

	// IsRunning returns whether the worker is currently running
	IsRunning() bool
}

// QueueStats represents queue statistics
type QueueStats struct {
	PendingJobs    int64            `json:"pending_jobs"`
	ProcessingJobs int64            `json:"processing_jobs"`
	CompletedJobs  int64            `json:"completed_jobs"`
	FailedJobs     int64            `json:"failed_jobs"`
	TotalJobs      int64            `json:"total_jobs"`
	QueueSizes     map[string]int64 `json:"queue_sizes"` // Per queue/priority
	Workers        int              `json:"workers"`
	Uptime         time.Duration    `json:"uptime"`
}

// JobOptions represents options for job creation
type JobOptions struct {
	MaxAttempts int           `json:"max_attempts"`
	Priority    int           `json:"priority"`
	Delay       time.Duration `json:"delay"`
	Queue       string        `json:"queue"` // Queue name for multiple queues
}

// Manager defines the interface for queue management
type Manager interface {
	// CreateQueue creates a new queue
	CreateQueue(name string) (Queue, error)

	// GetQueue returns an existing queue
	GetQueue(name string) (Queue, error)

	// DeleteQueue deletes a queue
	DeleteQueue(name string) error

	// ListQueues returns all queue names
	ListQueues() ([]string, error)

	// GetGlobalStats returns stats for all queues
	GetGlobalStats() (*QueueStats, error)
}
