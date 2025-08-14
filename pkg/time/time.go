package time

import (
	"fmt"
	"sync"
	"time"
)

var (
	// Default timezone
	defaultLocation *time.Location
	once            sync.Once
	mu              sync.RWMutex
)

// InitTimezone initializes the default timezone for the application
func InitTimezone(timezone string) error {
	if timezone == "" {
		timezone = "Asia/Bangkok" // Default fallback
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("failed to load timezone %s: %w", timezone, err)
	}

	mu.Lock()
	defaultLocation = location
	mu.Unlock()

	return nil
}

// GetLocation returns the current default timezone location
func GetLocation() *time.Location {
	mu.RLock()
	defer mu.RUnlock()

	if defaultLocation == nil {
		// Fallback to UTC if not initialized
		return time.UTC
	}
	return defaultLocation
}

// Now returns current time in the configured timezone
func Now() time.Time {
	return time.Now().In(GetLocation())
}

// Parse parses a time string and converts it to the configured timezone
func Parse(layout, value string) (time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(GetLocation()), nil
}

// ParseInLocation parses time in a specific location then converts to default timezone
func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(GetLocation()), nil
}

// Unix creates time from unix timestamp in configured timezone
func Unix(sec int64, nsec int64) time.Time {
	return time.Unix(sec, nsec).In(GetLocation())
}

// Date creates time with specified date in configured timezone
func Date(year int, month time.Month, day, hour, min, sec, nsec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, nsec, GetLocation())
}

// ToLocal converts any time to the configured timezone
func ToLocal(t time.Time) time.Time {
	return t.In(GetLocation())
}

// Format formats time with timezone info
func Format(t time.Time, layout string) string {
	return ToLocal(t).Format(layout)
}

func FormatThaiDate(t time.Time) string {
	year := t.Year() + 543
	return fmt.Sprintf("%d/%02d/%02d", t.Day(), t.Month(), year)
}

func FormatThaiDateTime(t time.Time) string {
	year := t.Year() + 543
	return fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d", t.Day(), t.Month(), year, t.Hour(), t.Minute(), t.Second())
}

// FormatNow formats current time in configured timezone
func FormatNow(layout string) string {
	return Now().Format(layout)
}

// StartOfDay returns start of day (00:00:00) in configured timezone
func StartOfDay(t time.Time) time.Time {
	local := ToLocal(t)
	return Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0)
}

// EndOfDay returns end of day (23:59:59.999999999) in configured timezone
func EndOfDay(t time.Time) time.Time {
	local := ToLocal(t)
	return Date(local.Year(), local.Month(), local.Day(), 23, 59, 59, 999999999)
}

// Common layout constants for Thailand
const (
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04:05"
	DateTimeLayout = "2006-01-02 15:04:05"
	ThaiDateLayout = "02/01/2006"
	ISO8601Layout  = "2006-01-02T15:04:05Z07:00"
)
