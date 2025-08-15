package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue implements Queue using Redis
type RedisQueue struct {
	client      *redis.Client
	name        string
	prefix      string
	retryDelays []time.Duration
	maxRetries  int
}

// RedisQueueConfig holds configuration for Redis queue
type RedisQueueConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	RetryDelays  []time.Duration
	Prefix       string
}

// NewRedisQueue creates a new Redis-based queue
func NewRedisQueue(name string, config *RedisQueueConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	prefix := config.Prefix
	if prefix == "" {
		prefix = "queue"
	}

	retryDelays := config.RetryDelays
	if len(retryDelays) == 0 {
		retryDelays = []time.Duration{
			1 * time.Second,
			5 * time.Second,
			30 * time.Second,
			5 * time.Minute,
			30 * time.Minute,
		}
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &RedisQueue{
		client:      client,
		name:        name,
		prefix:      prefix,
		retryDelays: retryDelays,
		maxRetries:  maxRetries,
	}, nil
}

// Push adds a job to the queue
func (rq *RedisQueue) Push(job *Job) error {
	return rq.PushDelayed(job, 0)
}

// PushDelayed adds a job to be processed after a delay
func (rq *RedisQueue) PushDelayed(job *Job, delay time.Duration) error {
	ctx := context.Background()

	// Set default values
	if job.ID == "" {
		job.ID = generateJobID()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.MaxAttempts == 0 {
		job.MaxAttempts = rq.maxRetries
	}

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job data
	jobKey := rq.jobKey(job.ID)
	if err := rq.client.Set(ctx, jobKey, jobData, 0).Err(); err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}

	// Add to appropriate queue
	if delay > 0 || job.Delay > 0 {
		totalDelay := delay + job.Delay
		score := float64(time.Now().Add(totalDelay).Unix())
		delayedKey := rq.delayedKey()

		if err := rq.client.ZAdd(ctx, delayedKey, redis.Z{
			Score:  score,
			Member: job.ID,
		}).Err(); err != nil {
			return fmt.Errorf("failed to add job to delayed queue: %w", err)
		}
	} else {
		// Add to priority queue (higher priority = higher score)
		queueKey := rq.queueKey()
		score := float64(job.Priority)

		if err := rq.client.ZAdd(ctx, queueKey, redis.Z{
			Score:  score,
			Member: job.ID,
		}).Err(); err != nil {
			return fmt.Errorf("failed to add job to queue: %w", err)
		}
	}

	return nil
}

// Pop retrieves the next job from the queue
func (rq *RedisQueue) Pop() (*Job, error) {
	ctx := context.Background()

	// First check delayed jobs
	if err := rq.moveDelayedJobs(); err != nil {
		return nil, fmt.Errorf("failed to move delayed jobs: %w", err)
	}

	// Get highest priority job
	queueKey := rq.queueKey()
	result := rq.client.ZPopMax(ctx, queueKey)

	if result.Err() == redis.Nil {
		return nil, nil // No jobs available
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to pop job: %w", result.Err())
	}

	if len(result.Val()) == 0 {
		return nil, nil // No jobs available
	}

	jobID := result.Val()[0].Member.(string)

	// Get job data
	job, err := rq.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	// Move to processing
	processingKey := rq.processingKey()
	if err := rq.client.SAdd(ctx, processingKey, jobID).Err(); err != nil {
		// Put job back in queue if we can't mark it as processing
		rq.client.ZAdd(ctx, queueKey, redis.Z{
			Score:  float64(job.Priority),
			Member: jobID,
		})
		return nil, fmt.Errorf("failed to mark job as processing: %w", err)
	}

	return job, nil
}

// Ack acknowledges successful job completion
func (rq *RedisQueue) Ack(jobID string) error {
	ctx := context.Background()

	// Remove from processing set
	processingKey := rq.processingKey()
	if err := rq.client.SRem(ctx, processingKey, jobID).Err(); err != nil {
		return fmt.Errorf("failed to remove job from processing: %w", err)
	}

	// Update job status
	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.ProcessedAt = &now

	return rq.updateJob(job)
}

// Nack marks a job as failed and potentially retries it
func (rq *RedisQueue) Nack(jobID string, jobErr error) error {
	ctx := context.Background()

	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	job.Attempts++
	job.Error = jobErr.Error()

	// Remove from processing
	processingKey := rq.processingKey()
	rq.client.SRem(ctx, processingKey, jobID)

	// Check if should retry
	if job.Attempts < job.MaxAttempts {
		// Calculate retry delay
		var delay time.Duration
		if job.Attempts-1 < len(rq.retryDelays) {
			delay = rq.retryDelays[job.Attempts-1]
		} else {
			delay = rq.retryDelays[len(rq.retryDelays)-1]
		}

		// Add to delayed queue for retry
		delayedKey := rq.delayedKey()
		score := float64(time.Now().Add(delay).Unix())

		if err := rq.client.ZAdd(ctx, delayedKey, redis.Z{
			Score:  score,
			Member: jobID,
		}).Err(); err != nil {
			return fmt.Errorf("failed to schedule retry: %w", err)
		}
	} else {
		// Mark as permanently failed
		now := time.Now()
		job.FailedAt = &now

		failedKey := rq.failedKey()
		if err := rq.client.SAdd(ctx, failedKey, jobID).Err(); err != nil {
			return fmt.Errorf("failed to add job to failed set: %w", err)
		}
	}

	return rq.updateJob(job)
}

// GetJob retrieves a job by ID
func (rq *RedisQueue) GetJob(jobID string) (*Job, error) {
	ctx := context.Background()

	jobKey := rq.jobKey(jobID)
	result := rq.client.Get(ctx, jobKey)

	if result.Err() == redis.Nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to get job: %w", result.Err())
	}

	var job Job
	if err := json.Unmarshal([]byte(result.Val()), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// GetJobStatus returns the current status of a job
func (rq *RedisQueue) GetJobStatus(jobID string) (JobStatus, error) {
	ctx := context.Background()

	// Check if job exists
	job, err := rq.GetJob(jobID)
	if err != nil {
		return "", err
	}

	// Check different sets/queues to determine status
	processingKey := rq.processingKey()
	if rq.client.SIsMember(ctx, processingKey, jobID).Val() {
		return StatusProcessing, nil
	}

	failedKey := rq.failedKey()
	if rq.client.SIsMember(ctx, failedKey, jobID).Val() {
		return StatusFailed, nil
	}

	if job.ProcessedAt != nil {
		return StatusCompleted, nil
	}

	if job.FailedAt != nil {
		return StatusFailed, nil
	}

	// Check if in delayed queue
	delayedKey := rq.delayedKey()
	score := rq.client.ZScore(ctx, delayedKey, jobID)
	if score.Err() == nil {
		if job.Attempts > 0 {
			return StatusRetrying, nil
		}
		return StatusPending, nil
	}

	// Check if in main queue
	queueKey := rq.queueKey()
	score = rq.client.ZScore(ctx, queueKey, jobID)
	if score.Err() == nil {
		return StatusPending, nil
	}

	return StatusCompleted, nil
}

// CancelJob cancels a pending job
func (rq *RedisQueue) CancelJob(jobID string) error {
	ctx := context.Background()

	// Remove from all possible queues
	queueKey := rq.queueKey()
	delayedKey := rq.delayedKey()
	processingKey := rq.processingKey()

	rq.client.ZRem(ctx, queueKey, jobID)
	rq.client.ZRem(ctx, delayedKey, jobID)
	rq.client.SRem(ctx, processingKey, jobID)

	// Update job status
	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	job.Error = "Job cancelled"
	return rq.updateJob(job)
}

// GetQueueSize returns the number of pending jobs
func (rq *RedisQueue) GetQueueSize() (int64, error) {
	ctx := context.Background()

	queueKey := rq.queueKey()
	return rq.client.ZCard(ctx, queueKey).Result()
}

// GetStats returns queue statistics
func (rq *RedisQueue) GetStats() (*QueueStats, error) {
	ctx := context.Background()

	queueKey := rq.queueKey()
	delayedKey := rq.delayedKey()
	processingKey := rq.processingKey()
	failedKey := rq.failedKey()

	pending := rq.client.ZCard(ctx, queueKey).Val()
	delayed := rq.client.ZCard(ctx, delayedKey).Val()
	processing := rq.client.SCard(ctx, processingKey).Val()
	failed := rq.client.SCard(ctx, failedKey).Val()

	return &QueueStats{
		PendingJobs:    pending + delayed,
		ProcessingJobs: processing,
		FailedJobs:     failed,
		TotalJobs:      pending + delayed + processing + failed,
		QueueSizes: map[string]int64{
			"pending":    pending,
			"delayed":    delayed,
			"processing": processing,
			"failed":     failed,
		},
	}, nil
}

// Close closes the queue connection
func (rq *RedisQueue) Close() error {
	return rq.client.Close()
}

// Helper methods

func (rq *RedisQueue) moveDelayedJobs() error {
	ctx := context.Background()

	delayedKey := rq.delayedKey()
	queueKey := rq.queueKey()

	now := float64(time.Now().Unix())

	// Get jobs ready to be processed
	result := rq.client.ZRangeByScoreWithScores(ctx, delayedKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatFloat(now, 'f', 0, 64),
	})

	if result.Err() != nil {
		return result.Err()
	}

	for _, z := range result.Val() {
		jobID := z.Member.(string)

		// Get job to check priority
		job, err := rq.GetJob(jobID)
		if err != nil {
			continue
		}

		// Move to main queue
		pipe := rq.client.Pipeline()
		pipe.ZRem(ctx, delayedKey, jobID)
		pipe.ZAdd(ctx, queueKey, redis.Z{
			Score:  float64(job.Priority),
			Member: jobID,
		})

		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (rq *RedisQueue) updateJob(job *Job) error {
	ctx := context.Background()

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	jobKey := rq.jobKey(job.ID)
	return rq.client.Set(ctx, jobKey, jobData, 0).Err()
}

// Key generation methods
func (rq *RedisQueue) queueKey() string {
	return fmt.Sprintf("%s:%s:queue", rq.prefix, rq.name)
}

func (rq *RedisQueue) delayedKey() string {
	return fmt.Sprintf("%s:%s:delayed", rq.prefix, rq.name)
}

func (rq *RedisQueue) processingKey() string {
	return fmt.Sprintf("%s:%s:processing", rq.prefix, rq.name)
}

func (rq *RedisQueue) failedKey() string {
	return fmt.Sprintf("%s:%s:failed", rq.prefix, rq.name)
}

func (rq *RedisQueue) jobKey(jobID string) string {
	return fmt.Sprintf("%s:%s:job:%s", rq.prefix, rq.name, jobID)
}
