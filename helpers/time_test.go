package helpers

import (
	"testing"
	"time"
)

func TestGetEpochTime(t *testing.T) {
	// Test that GetEpochTime returns a value close to the current time
	before := time.Now().Unix()
	result := GetEpochTime()
	after := time.Now().Unix()

	if result < before || result > after {
		t.Errorf("GetEpochTime() = %d, expected value between %d and %d", result, before, after)
	}
}

func TestConvertToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		// Minutes format
		{
			name:     "1 minute",
			input:    "1m",
			expected: 60,
			wantErr:  false,
		},
		{
			name:     "30 minutes",
			input:    "30m",
			expected: 1800,
			wantErr:  false,
		},
		{
			name:     "100 minutes",
			input:    "100m",
			expected: 6000,
			wantErr:  false,
		},
		// Hours format
		{
			name:     "1 hour",
			input:    "1h",
			expected: 3600,
			wantErr:  false,
		},
		{
			name:     "2 hours",
			input:    "2h",
			expected: 7200,
			wantErr:  false,
		},
		{
			name:     "24 hours",
			input:    "24h",
			expected: 86400,
			wantErr:  false,
		},
		// Days format
		{
			name:     "1 day",
			input:    "1d",
			expected: 86400,
			wantErr:  false,
		},
		{
			name:     "7 days",
			input:    "7d",
			expected: 604800,
			wantErr:  false,
		},
		// MM:SS format
		{
			name:     "30 seconds in MM:SS",
			input:    "0:30",
			expected: 30,
			wantErr:  false,
		},
		{
			name:     "1 minute 30 seconds",
			input:    "1:30",
			expected: 90,
			wantErr:  false,
		},
		{
			name:     "45 minutes 30 seconds",
			input:    "45:30",
			expected: 2730,
			wantErr:  false,
		},
		// HH:MM:SS format
		{
			name:     "1 hour in HH:MM:SS",
			input:    "1:00:00",
			expected: 3600,
			wantErr:  false,
		},
		{
			name:     "1 hour 30 minutes 15 seconds",
			input:    "1:30:15",
			expected: 5415,
			wantErr:  false,
		},
		{
			name:     "10 hours 5 minutes 3 seconds",
			input:    "10:05:03",
			expected: 36303,
			wantErr:  false,
		},
		// Plain seconds
		{
			name:     "45 seconds",
			input:    "45",
			expected: 45,
			wantErr:  false,
		},
		{
			name:     "120 seconds",
			input:    "120",
			expected: 120,
			wantErr:  false,
		},
		{
			name:     "0 seconds",
			input:    "0",
			expected: 0,
			wantErr:  false,
		},
		// With whitespace
		{
			name:     "with leading whitespace",
			input:    "  5m",
			expected: 300,
			wantErr:  false,
		},
		{
			name:     "with trailing whitespace",
			input:    "10h  ",
			expected: 36000,
			wantErr:  false,
		},
		{
			name:     "with surrounding whitespace",
			input:    "  2d  ",
			expected: 172800,
			wantErr:  false,
		},
		// Error cases
		{
			name:     "invalid format - just letter",
			input:    "m",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - unsupported suffix",
			input:    "5w",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "negative number",
			input:    "-5m",
			expected: -300,
			wantErr:  false,
		},
		{
			name:     "invalid format - letters in number",
			input:    "abc",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - too many colons",
			input:    "1:2:3:4",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - non-numeric with colons",
			input:    "a:b:c",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToSeconds(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ConvertToSeconds(%q) expected error, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ConvertToSeconds(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ConvertToSeconds(%q) = %d, expected %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}
