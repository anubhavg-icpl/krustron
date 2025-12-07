// Package utils provides utility functions for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// String utilities

// TruncateString truncates a string to max length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// SanitizeString removes non-printable characters
func SanitizeString(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}

// SlugifyString converts a string to a URL-friendly slug
func SlugifyString(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`^-+|-+$`).ReplaceAllString(s, "")
	return s
}

// PadLeft pads a string on the left with a character
func PadLeft(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(string(pad), length-len(s)) + s
}

// PadRight pads a string on the right with a character
func PadRight(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(pad), length-len(s))
}

// CamelToSnake converts CamelCase to snake_case
func CamelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// SnakeToCamel converts snake_case to CamelCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// ContainsString checks if a slice contains a string
func ContainsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// FilterStrings filters a slice based on a predicate
func FilterStrings(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if predicate(s) {
			result = append(result, s)
		}
	}
	return result
}

// ID and Token utilities

// GenerateID generates a new UUID
func GenerateID() string {
	return uuid.New().String()
}

// GenerateShortID generates a short ID (8 characters)
func GenerateShortID() string {
	id := uuid.New().String()
	return strings.Replace(id[:8], "-", "", -1)
}

// GenerateToken generates a secure random token
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateAPIKey generates an API key
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "krn_" + hex.EncodeToString(bytes), nil
}

// Hash utilities

// HashSHA256 returns SHA256 hash of a string
func HashSHA256(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// HashPassword hashes a password (simple implementation, use bcrypt in production)
func HashPassword(password string) string {
	salt := GenerateShortID()
	hash := sha256.Sum256([]byte(password + salt))
	return salt + "." + hex.EncodeToString(hash[:])
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	parts := strings.SplitN(hash, ".", 2)
	if len(parts) != 2 {
		return false
	}
	salt := parts[0]
	expectedHash := sha256.Sum256([]byte(password + salt))
	return parts[1] == hex.EncodeToString(expectedHash[:])
}

// Time utilities

// FormatDuration formats a duration in human-readable form
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// ParseDuration parses a duration string with extended support
func ParseDuration(s string) (time.Duration, error) {
	// Support days
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// TimeAgo returns a human-readable time ago string
func TimeAgo(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	if d < 7*24*time.Hour {
		days := int(d.Hours()) / 24
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if d < 30*24*time.Hour {
		weeks := int(d.Hours()) / (24 * 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
	if d < 365*24*time.Hour {
		months := int(d.Hours()) / (24 * 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	years := int(d.Hours()) / (24 * 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// StartOfDay returns the start of the day for a given time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for a given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// File utilities

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if !DirExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// GetFileExtension returns the file extension without the dot
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	return ext[1:]
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatBytes formats bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Network utilities

// IsValidIP checks if a string is a valid IP address
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// GetLocalIP returns the local IP address
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no local IP found")
}

// GetHostname returns the hostname
func GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

// Map utilities

// MergeMaps merges multiple maps into one
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// MapKeys returns the keys of a map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapValues returns the values of a map
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// JSON utilities

// ToJSON converts a value to JSON string
func ToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// ToJSONPretty converts a value to pretty JSON string
func ToJSONPretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// FromJSON parses JSON into a value
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// DeepCopy creates a deep copy using JSON marshaling
func DeepCopy(src, dst interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// Environment utilities

// GetEnv returns an environment variable or default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt returns an int environment variable or default value
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvBool returns a bool environment variable or default value
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// GetEnvDuration returns a duration environment variable or default value
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

// Reflection utilities

// IsNil checks if a value is nil
func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// GetTypeName returns the type name of a value
func GetTypeName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

// Concurrent utilities

// SafeGo runs a function in a goroutine with panic recovery
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				fmt.Printf("Panic recovered: %v\n%s\n", r, buf[:n])
			}
		}()
		fn()
	}()
}

// Semaphore is a simple semaphore implementation
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Acquire acquires a semaphore slot
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// Release releases a semaphore slot
func (s *Semaphore) Release() {
	<-s.ch
}

// TryAcquire tries to acquire a semaphore slot without blocking
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// Once is a wrapper around sync.Once with reset capability
type Once struct {
	mu   sync.Mutex
	done bool
}

// Do executes the function only once
func (o *Once) Do(fn func()) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.done {
		fn()
		o.done = true
	}
}

// Reset resets the Once
func (o *Once) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.done = false
}

// Validation utilities

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// ValidateKubernetesName validates a Kubernetes resource name
func ValidateKubernetesName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// ValidateLabelKey validates a Kubernetes label key
func ValidateLabelKey(key string) bool {
	if len(key) == 0 || len(key) > 63 {
		return false
	}
	pattern := `^([a-z0-9A-Z]([a-z0-9A-Z\-_.]*[a-z0-9A-Z])?)?$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

// ValidateLabelValue validates a Kubernetes label value
func ValidateLabelValue(value string) bool {
	if len(value) > 63 {
		return false
	}
	if value == "" {
		return true
	}
	pattern := `^([a-z0-9A-Z]([a-z0-9A-Z\-_.]*[a-z0-9A-Z])?)?$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// Retry utilities

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     10 * time.Second,
		Multiplier:  2.0,
	}
}

// Retry retries a function with exponential backoff
func Retry(config RetryConfig, fn func() error) error {
	var err error
	wait := config.InitialWait

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt < config.MaxAttempts-1 {
			time.Sleep(wait)
			wait = time.Duration(float64(wait) * config.Multiplier)
			if wait > config.MaxWait {
				wait = config.MaxWait
			}
		}
	}

	return fmt.Errorf("after %d attempts: %w", config.MaxAttempts, err)
}

// Math utilities

// Min returns the minimum of two values
func Min[T int | int64 | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two values
func Max[T int | int64 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps a value between min and max
func Clamp[T int | int64 | float64](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Abs returns the absolute value
func Abs[T int | int64 | float64](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

// Round rounds a float to n decimal places
func Round(f float64, places int) float64 {
	shift := 1.0
	for i := 0; i < places; i++ {
		shift *= 10
	}
	return float64(int(f*shift+0.5)) / shift
}

// Percentage calculates percentage
func Percentage(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}
