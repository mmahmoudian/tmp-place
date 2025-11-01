package server

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GetEpochTime returns the current Unix epoch time in seconds.
func GetEpochTime() int64 {
	return time.Now().Unix()
}

// ConvertToSeconds converts a time string to seconds.
// Supported formats:
//   - "1m", "30m" - minutes
//   - "1h", "2h" - hours
//   - "1d", "7d" - days
//   - "1:00:15" - HH:MM:SS format
//   - "30:45" - MM:SS format
//   - "45" - seconds only
//
// Returns the number of seconds and an error if the format is invalid.
func ConvertToSeconds(timeStr string) (int64, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Check for time format with colons (HH:MM:SS or MM:SS)
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		var hours, minutes, seconds int64
		var err error

		switch len(parts) {
		case 2: // MM:SS format
			minutes, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes: %v", err)
			}
			seconds, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid seconds: %v", err)
			}
			return minutes*60 + seconds, nil

		case 3: // HH:MM:SS format
			hours, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid hours: %v", err)
			}
			minutes, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes: %v", err)
			}
			seconds, err = strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid seconds: %v", err)
			}
			return hours*3600 + minutes*60 + seconds, nil

		default:
			return 0, fmt.Errorf("invalid time format: %s", timeStr)
		}
	}

	// Check for suffix-based formats (m, h, d)
	if len(timeStr) >= 2 {
		suffix := timeStr[len(timeStr)-1:]
		valueStr := timeStr[:len(timeStr)-1]
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err == nil {
			switch suffix {
			case "m":
				return value * 60, nil
			case "h":
				return value * 3600, nil
			case "d":
				return value * 86400, nil
			}
		}
	}

	// Try parsing as plain seconds
	seconds, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}
	return seconds, nil
}

// SecondsToHumanReadable converts seconds to a human-readable format.
// Returns a string in the format "%d Days, %d Hours, %d Minutes, and %d Seconds"
func SecondsToHumanReadable(totalSeconds int64) string {
	days := totalSeconds / 86400
	remainingAfterDays := totalSeconds % 86400

	hours := remainingAfterDays / 3600
	remainingAfterHours := remainingAfterDays % 3600

	minutes := remainingAfterHours / 60
	seconds := remainingAfterHours % 60

	return fmt.Sprintf("%d Days, %d Hours, %d Minutes, and %d Seconds",
		days, hours, minutes, seconds)
}
