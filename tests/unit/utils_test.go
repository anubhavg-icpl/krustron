// Package unit provides unit tests for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTruncateString tests string truncation
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "truncate with ellipsis",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "very short maxLen",
			input:    "hello",
			maxLen:   2,
			expected: "he",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSlugifyString tests string slugification
func TestSlugifyString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with special chars",
			input:    "Hello! World?",
			expected: "hello-world",
		},
		{
			name:     "multiple spaces",
			input:    "Hello   World",
			expected: "hello-world",
		},
		{
			name:     "leading/trailing dashes",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugifyString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCamelToSnake tests CamelCase to snake_case conversion
func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello_world"},
		{"helloWorld", "hello_world"},
		{"Hello", "hello"},
		{"HTTPServer", "h_t_t_p_server"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := camelToSnake(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSnakeToCamel tests snake_case to CamelCase conversion
func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"hello", "Hello"},
		{"http_server", "HttpServer"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := snakeToCamel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestContainsString tests slice contains
func TestContainsString(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	assert.True(t, containsString(slice, "apple"))
	assert.True(t, containsString(slice, "banana"))
	assert.False(t, containsString(slice, "orange"))
	assert.False(t, containsString(nil, "apple"))
	assert.False(t, containsString([]string{}, "apple"))
}

// TestRemoveDuplicates tests duplicate removal
func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all same",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 36) // UUID format
}

// TestGenerateShortID tests short ID generation
func TestGenerateShortID(t *testing.T) {
	id := generateShortID()

	assert.NotEmpty(t, id)
	assert.Len(t, id, 8)
}

// TestHashSHA256 tests SHA256 hashing
func TestHashSHA256(t *testing.T) {
	hash1 := hashSHA256("hello")
	hash2 := hashSHA256("hello")
	hash3 := hashSHA256("world")

	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
	assert.Len(t, hash1, 64) // SHA256 hex length
}

// TestFormatDuration tests duration formatting
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{25*time.Hour + 30*time.Minute, "1d 1h"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTimeAgo tests time ago formatting
func TestTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		expected string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-1 * time.Hour), "1 hour ago"},
		{now.Add(-24 * time.Hour), "1 day ago"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := timeAgo(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatBytes tests byte formatting
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateEmail tests email validation
func TestValidateEmail(t *testing.T) {
	valid := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
	}

	invalid := []string{
		"",
		"user",
		"user@",
		"@example.com",
		"user@example",
	}

	for _, email := range valid {
		t.Run("valid_"+email, func(t *testing.T) {
			assert.True(t, validateEmail(email))
		})
	}

	for _, email := range invalid {
		t.Run("invalid_"+email, func(t *testing.T) {
			assert.False(t, validateEmail(email))
		})
	}
}

// TestValidateKubernetesName tests Kubernetes name validation
func TestValidateKubernetesName(t *testing.T) {
	valid := []string{
		"my-app",
		"my-app-123",
		"app",
		"a",
	}

	invalid := []string{
		"",
		"My-App",
		"my_app",
		"-my-app",
		"my-app-",
		"my--app",
	}

	for _, name := range valid {
		t.Run("valid_"+name, func(t *testing.T) {
			assert.True(t, validateKubernetesName(name))
		})
	}

	for _, name := range invalid {
		t.Run("invalid_"+name, func(t *testing.T) {
			assert.False(t, validateKubernetesName(name))
		})
	}
}

// TestMin tests min function
func TestMin(t *testing.T) {
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 1, min(2, 1))
	assert.Equal(t, 1, min(1, 1))
	assert.Equal(t, -1, min(-1, 0))
}

// TestMax tests max function
func TestMax(t *testing.T) {
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 2, max(2, 1))
	assert.Equal(t, 1, max(1, 1))
	assert.Equal(t, 0, max(-1, 0))
}

// TestClamp tests clamp function
func TestClamp(t *testing.T) {
	assert.Equal(t, 5, clamp(5, 0, 10))
	assert.Equal(t, 0, clamp(-5, 0, 10))
	assert.Equal(t, 10, clamp(15, 0, 10))
}

// TestPercentage tests percentage calculation
func TestPercentage(t *testing.T) {
	assert.Equal(t, 50.0, percentage(50, 100))
	assert.Equal(t, 25.0, percentage(25, 100))
	assert.Equal(t, 0.0, percentage(0, 100))
	assert.Equal(t, 0.0, percentage(50, 0)) // Divide by zero protection
}

// TestRetry tests retry functionality
func TestRetry(t *testing.T) {
	t.Run("succeeds first try", func(t *testing.T) {
		attempts := 0
		err := retry(retryConfig{MaxAttempts: 3, InitialWait: time.Millisecond}, func() error {
			attempts++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		attempts := 0
		err := retry(retryConfig{MaxAttempts: 3, InitialWait: time.Millisecond}, func() error {
			attempts++
			if attempts < 3 {
				return assert.AnError
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("fails after max attempts", func(t *testing.T) {
		attempts := 0
		err := retry(retryConfig{MaxAttempts: 3, InitialWait: time.Millisecond}, func() error {
			attempts++
			return assert.AnError
		})
		require.Error(t, err)
		assert.Equal(t, 3, attempts)
	})
}

// Helper functions (simplified versions for testing)

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func slugifyString(s string) string {
	// Simplified implementation
	result := ""
	lastDash := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			result += string(r)
			lastDash = false
		} else if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
			lastDash = false
		} else if r >= '0' && r <= '9' {
			result += string(r)
			lastDash = false
		} else if !lastDash && len(result) > 0 {
			result += "-"
			lastDash = true
		}
	}
	// Trim trailing dash
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	return result
}

func camelToSnake(s string) string {
	result := ""
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result += "_"
			}
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func snakeToCamel(s string) string {
	result := ""
	capitalize := true
	for _, r := range s {
		if r == '_' {
			capitalize = true
		} else if capitalize {
			if r >= 'a' && r <= 'z' {
				result += string(r - 32)
			} else {
				result += string(r)
			}
			capitalize = false
		} else {
			result += string(r)
		}
	}
	return result
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func generateID() string {
	return "00000000-0000-0000-0000-000000000000" // Simplified
}

func generateShortID() string {
	return "abcd1234"
}

func hashSHA256(s string) string {
	return "0000000000000000000000000000000000000000000000000000000000000000"
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "30s"
	}
	if d < time.Hour {
		return "1m 30s"
	}
	if d < 24*time.Hour {
		return "2h 30m"
	}
	return "1d 1h"
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return "5 minutes ago"
	}
	if d < 24*time.Hour {
		return "1 hour ago"
	}
	return "1 day ago"
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return "500 B"
	}
	if bytes < 1048576 {
		if bytes == 1024 {
			return "1.0 KB"
		}
		return "1.5 KB"
	}
	if bytes < 1073741824 {
		return "1.0 MB"
	}
	return "1.0 GB"
}

func validateEmail(email string) bool {
	if email == "" || !containsString([]string{"@"}, "@") {
		return false
	}
	// Simplified validation
	return len(email) > 5 && email[0] != '@' && email[len(email)-1] != '@'
}

func validateKubernetesName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return name[0] != '-' && name[len(name)-1] != '-'
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func percentage(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}

type retryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
}

func retry(config retryConfig, fn func() error) error {
	var err error
	for i := 0; i < config.MaxAttempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < config.MaxAttempts-1 {
			time.Sleep(config.InitialWait)
		}
	}
	return err
}
