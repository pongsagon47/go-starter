package queue

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// generateJobID generates a unique job ID
func generateJobID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("job_%d", time.Now().UnixNano())
	}
	return "job_" + hex.EncodeToString(bytes)
}
