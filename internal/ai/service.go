// Package ai provides AI-powered operations for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Provider represents an AI provider
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderOllama    Provider = "ollama"
	ProviderAzure     Provider = "azure"
	ProviderBedrock   Provider = "bedrock"
)

// Intent represents the detected intent of a user query
type Intent string

const (
	IntentDiagnose     Intent = "diagnose"
	IntentOptimize     Intent = "optimize"
	IntentExplain      Intent = "explain"
	IntentRecommend    Intent = "recommend"
	IntentTroubleshoot Intent = "troubleshoot"
	IntentGenerate     Intent = "generate"
	IntentAnalyze      Intent = "analyze"
	IntentChat         Intent = "chat"
)

// Config holds AI service configuration
type Config struct {
	Provider         Provider
	APIKey           string
	Endpoint         string
	Model            string
	MaxTokens        int
	Temperature      float64
	TopP             float64
	EnableCache      bool
	CacheTTL         time.Duration
	RateLimitRPM     int
	EnableStreaming  bool
	SystemPrompt     string
}

// Service provides AI operations
type Service struct {
	db          *gorm.DB
	logger      *zap.Logger
	config      *Config
	httpClient  *http.Client
	cache       sync.Map
	rateLimiter *rateLimiter
}

// Query represents an AI query
type Query struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	UserID       string                 `json:"user_id" gorm:"index"`
	Query        string                 `json:"query"`
	Intent       Intent                 `json:"intent"`
	Context      map[string]interface{} `json:"context" gorm:"serializer:json"`
	Response     string                 `json:"response"`
	Model        string                 `json:"model"`
	TokensUsed   int                    `json:"tokens_used"`
	Latency      time.Duration          `json:"latency"`
	Feedback     *QueryFeedback         `json:"feedback" gorm:"foreignKey:QueryID"`
	CreatedAt    time.Time              `json:"created_at"`
}

// QueryFeedback represents user feedback on AI responses
type QueryFeedback struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	QueryID   string    `json:"query_id" gorm:"index"`
	Rating    int       `json:"rating"` // 1-5
	Helpful   bool      `json:"helpful"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// Recommendation represents an AI-generated recommendation
type Recommendation struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	Type         string                 `json:"type"` // cost, performance, security, reliability
	Category     string                 `json:"category"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Impact       string                 `json:"impact"` // high, medium, low
	Effort       string                 `json:"effort"` // high, medium, low
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id"`
	CurrentState map[string]interface{} `json:"current_state" gorm:"serializer:json"`
	Suggested    map[string]interface{} `json:"suggested" gorm:"serializer:json"`
	Savings      float64                `json:"savings"` // cost savings if applicable
	Status       string                 `json:"status"`  // pending, applied, dismissed
	AppliedAt    *time.Time             `json:"applied_at"`
	AppliedBy    string                 `json:"applied_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Insight represents an AI-generated insight
type Insight struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"` // critical, warning, info
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Evidence    map[string]interface{} `json:"evidence" gorm:"serializer:json"`
	Actions     []string               `json:"actions" gorm:"serializer:json"`
	Resources   []string               `json:"resources" gorm:"serializer:json"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	AckedAt     *time.Time             `json:"acked_at"`
	AckedBy     string                 `json:"acked_by"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role      string `json:"role"` // system, user, assistant
	Content   string `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatSession represents a chat conversation
type ChatSession struct {
	ID        string                 `json:"id" gorm:"primaryKey"`
	UserID    string                 `json:"user_id" gorm:"index"`
	Title     string                 `json:"title"`
	Messages  []ChatMessage          `json:"messages" gorm:"serializer:json"`
	Context   map[string]interface{} `json:"context" gorm:"serializer:json"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewService creates a new AI service
func NewService(db *gorm.DB, logger *zap.Logger, config *Config) (*Service, error) {
	// Auto-migrate tables
	if err := db.AutoMigrate(
		&Query{},
		&QueryFeedback{},
		&Recommendation{},
		&Insight{},
		&ChatSession{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate AI tables: %w", err)
	}

	// Set defaults
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.TopP == 0 {
		config.TopP = 0.9
	}
	if config.RateLimitRPM == 0 {
		config.RateLimitRPM = 60
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour
	}
	if config.SystemPrompt == "" {
		config.SystemPrompt = getDefaultSystemPrompt()
	}

	svc := &Service{
		db:     db,
		logger: logger,
		config: config,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		rateLimiter: newRateLimiter(config.RateLimitRPM),
	}

	return svc, nil
}

func getDefaultSystemPrompt() string {
	return `You are Krustron AI, an intelligent assistant for Kubernetes operations.
You help DevOps engineers and developers with:
- Diagnosing issues in Kubernetes clusters
- Optimizing resource allocation and costs
- Explaining complex Kubernetes concepts
- Troubleshooting deployment failures
- Recommending best practices
- Generating Kubernetes manifests and configurations

Always provide actionable, specific advice based on the context provided.
When analyzing errors, provide step-by-step troubleshooting guides.
For cost optimization, provide specific resource recommendations.
Use markdown formatting for better readability.`
}

// AskQuestion processes a natural language question about Kubernetes
func (s *Service) AskQuestion(ctx context.Context, userID, question string, context map[string]interface{}) (*Query, error) {
	// Rate limiting
	if !s.rateLimiter.allow() {
		return nil, fmt.Errorf("rate limit exceeded, please try again later")
	}

	// Check cache
	if s.config.EnableCache {
		if cached, ok := s.getFromCache(question); ok {
			return cached, nil
		}
	}

	// Detect intent
	intent := s.detectIntent(question)

	// Build prompt with context
	prompt := s.buildPrompt(question, intent, context)

	// Call AI provider
	startTime := time.Now()
	response, tokensUsed, err := s.callProvider(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}
	latency := time.Since(startTime)

	// Create query record
	query := &Query{
		ID:         uuid.New().String(),
		UserID:     userID,
		Query:      question,
		Intent:     intent,
		Context:    context,
		Response:   response,
		Model:      s.config.Model,
		TokensUsed: tokensUsed,
		Latency:    latency,
		CreatedAt:  time.Now(),
	}

	// Save to database
	if err := s.db.Create(query).Error; err != nil {
		s.logger.Warn("Failed to save query", zap.Error(err))
	}

	// Cache the result
	if s.config.EnableCache {
		s.saveToCache(question, query)
	}

	return query, nil
}

// detectIntent analyzes the query to determine user intent
func (s *Service) detectIntent(query string) Intent {
	query = strings.ToLower(query)

	patterns := map[Intent][]string{
		IntentDiagnose: {
			"why is", "what's wrong", "not working", "failing", "error",
			"crashing", "down", "unhealthy", "issue", "problem",
		},
		IntentOptimize: {
			"optimize", "improve", "reduce cost", "save money", "efficiency",
			"rightsize", "scale", "resource", "cpu", "memory",
		},
		IntentExplain: {
			"what is", "explain", "how does", "what does", "meaning of",
			"definition", "describe", "tell me about",
		},
		IntentRecommend: {
			"recommend", "suggest", "best practice", "should i", "which",
			"better", "comparison", "advice",
		},
		IntentTroubleshoot: {
			"troubleshoot", "debug", "fix", "solve", "resolve",
			"cannot", "can't", "unable", "help with",
		},
		IntentGenerate: {
			"generate", "create", "write", "make", "build",
			"yaml", "manifest", "config", "template",
		},
		IntentAnalyze: {
			"analyze", "review", "audit", "check", "assess",
			"evaluate", "inspect", "scan",
		},
	}

	for intent, keywords := range patterns {
		for _, keyword := range keywords {
			if strings.Contains(query, keyword) {
				return intent
			}
		}
	}

	return IntentChat
}

// buildPrompt constructs the prompt with context
func (s *Service) buildPrompt(question string, intent Intent, context map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString(s.config.SystemPrompt)
	sb.WriteString("\n\n")

	// Add intent-specific instructions
	switch intent {
	case IntentDiagnose:
		sb.WriteString("Focus on identifying the root cause and providing diagnostic steps.\n")
	case IntentOptimize:
		sb.WriteString("Provide specific optimization recommendations with expected impact.\n")
	case IntentTroubleshoot:
		sb.WriteString("Provide a step-by-step troubleshooting guide.\n")
	case IntentGenerate:
		sb.WriteString("Generate production-ready YAML/configuration with best practices.\n")
	}

	// Add context
	if len(context) > 0 {
		sb.WriteString("\n### Context ###\n")
		for key, value := range context {
			if jsonVal, err := json.MarshalIndent(value, "", "  "); err == nil {
				sb.WriteString(fmt.Sprintf("%s:\n```json\n%s\n```\n", key, string(jsonVal)))
			}
		}
	}

	sb.WriteString("\n### Question ###\n")
	sb.WriteString(question)

	return sb.String()
}

// callProvider calls the configured AI provider
func (s *Service) callProvider(ctx context.Context, prompt string) (string, int, error) {
	switch s.config.Provider {
	case ProviderOpenAI:
		return s.callOpenAI(ctx, prompt)
	case ProviderAnthropic:
		return s.callAnthropic(ctx, prompt)
	case ProviderOllama:
		return s.callOllama(ctx, prompt)
	default:
		return s.callOpenAI(ctx, prompt)
	}
}

// callOpenAI calls the OpenAI API
func (s *Service) callOpenAI(ctx context.Context, prompt string) (string, int, error) {
	endpoint := s.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	requestBody := map[string]interface{}{
		"model": s.config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  s.config.MaxTokens,
		"temperature": s.config.Temperature,
		"top_p":       s.config.TopP,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, result.Usage.TotalTokens, nil
}

// callAnthropic calls the Anthropic API
func (s *Service) callAnthropic(ctx context.Context, prompt string) (string, int, error) {
	endpoint := s.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1/messages"
	}

	requestBody := map[string]interface{}{
		"model":      s.config.Model,
		"max_tokens": s.config.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.config.APIKey)
	req.Header.Set("anthropic-version", "2024-01-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	if len(result.Content) == 0 {
		return "", 0, fmt.Errorf("no response from AI")
	}

	totalTokens := result.Usage.InputTokens + result.Usage.OutputTokens
	return result.Content[0].Text, totalTokens, nil
}

// callOllama calls a local Ollama instance
func (s *Service) callOllama(ctx context.Context, prompt string) (string, int, error) {
	endpoint := s.config.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate"
	}

	requestBody := map[string]interface{}{
		"model":  s.config.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": s.config.Temperature,
			"top_p":       s.config.TopP,
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	return result.Response, 0, nil
}

// DiagnoseIssue analyzes cluster/pod issues and provides diagnosis
func (s *Service) DiagnoseIssue(ctx context.Context, userID string, issue DiagnosisRequest) (*DiagnosisResult, error) {
	context := map[string]interface{}{
		"resource_type": issue.ResourceType,
		"resource_name": issue.ResourceName,
		"namespace":     issue.Namespace,
		"cluster":       issue.Cluster,
		"events":        issue.Events,
		"logs":          issue.Logs,
		"status":        issue.Status,
		"describe":      issue.Describe,
	}

	question := fmt.Sprintf("Diagnose why %s '%s' in namespace '%s' is having issues: %s",
		issue.ResourceType, issue.ResourceName, issue.Namespace, issue.Description)

	query, err := s.AskQuestion(ctx, userID, question, context)
	if err != nil {
		return nil, err
	}

	// Parse the response to extract structured diagnosis
	result := &DiagnosisResult{
		QueryID:     query.ID,
		Analysis:    query.Response,
		RootCause:   s.extractRootCause(query.Response),
		Steps:       s.extractSteps(query.Response),
		Severity:    s.determineSeverity(issue, query.Response),
		Confidence:  s.calculateConfidence(query.Response),
		RelatedDocs: s.findRelatedDocs(query.Response),
	}

	return result, nil
}

// DiagnosisRequest represents a request to diagnose an issue
type DiagnosisRequest struct {
	ResourceType string                 `json:"resource_type"`
	ResourceName string                 `json:"resource_name"`
	Namespace    string                 `json:"namespace"`
	Cluster      string                 `json:"cluster"`
	Description  string                 `json:"description"`
	Events       []map[string]interface{} `json:"events"`
	Logs         string                 `json:"logs"`
	Status       map[string]interface{} `json:"status"`
	Describe     string                 `json:"describe"`
}

// DiagnosisResult represents the result of a diagnosis
type DiagnosisResult struct {
	QueryID     string   `json:"query_id"`
	Analysis    string   `json:"analysis"`
	RootCause   string   `json:"root_cause"`
	Steps       []string `json:"steps"`
	Severity    string   `json:"severity"`
	Confidence  float64  `json:"confidence"`
	RelatedDocs []string `json:"related_docs"`
}

func (s *Service) extractRootCause(response string) string {
	// Look for common root cause indicators
	patterns := []string{
		`(?i)root cause[:\s]+([^\n]+)`,
		`(?i)the issue is[:\s]+([^\n]+)`,
		`(?i)caused by[:\s]+([^\n]+)`,
		`(?i)problem[:\s]+([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(response); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// Return first meaningful sentence
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 20 && !strings.HasPrefix(line, "#") {
			return line
		}
	}

	return "Unable to determine specific root cause"
}

func (s *Service) extractSteps(response string) []string {
	var steps []string

	// Look for numbered lists
	re := regexp.MustCompile(`(?m)^\s*(\d+)[.\)]\s*(.+)$`)
	matches := re.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) > 2 {
			steps = append(steps, strings.TrimSpace(match[2]))
		}
	}

	// Look for bullet points if no numbered list
	if len(steps) == 0 {
		re = regexp.MustCompile(`(?m)^\s*[-*]\s*(.+)$`)
		matches = re.FindAllStringSubmatch(response, -1)
		for _, match := range matches {
			if len(match) > 1 {
				steps = append(steps, strings.TrimSpace(match[1]))
			}
		}
	}

	return steps
}

func (s *Service) determineSeverity(issue DiagnosisRequest, response string) string {
	response = strings.ToLower(response)

	if strings.Contains(response, "critical") || strings.Contains(response, "urgent") ||
		strings.Contains(response, "immediately") || strings.Contains(response, "data loss") {
		return "critical"
	}

	if strings.Contains(response, "important") || strings.Contains(response, "should") ||
		strings.Contains(response, "recommended") {
		return "warning"
	}

	return "info"
}

func (s *Service) calculateConfidence(response string) float64 {
	// Base confidence
	confidence := 0.5

	// Increase confidence based on specificity
	if len(response) > 500 {
		confidence += 0.1
	}
	if strings.Contains(response, "kubectl") || strings.Contains(response, "command") {
		confidence += 0.1
	}
	if regexp.MustCompile(`\d+`).MatchString(response) {
		confidence += 0.05
	}

	// Cap at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

func (s *Service) findRelatedDocs(response string) []string {
	var docs []string

	// Common Kubernetes documentation links based on keywords
	docMap := map[string]string{
		"crashloopbackoff":  "https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/",
		"imagepullbackoff":  "https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/",
		"oom":               "https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource/",
		"pending":           "https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/",
		"evicted":           "https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/",
		"probe":             "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
		"pvc":               "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"ingress":           "https://kubernetes.io/docs/concepts/services-networking/ingress/",
		"networkpolicy":     "https://kubernetes.io/docs/concepts/services-networking/network-policies/",
		"rbac":              "https://kubernetes.io/docs/reference/access-authn-authz/rbac/",
	}

	responseLower := strings.ToLower(response)
	for keyword, url := range docMap {
		if strings.Contains(responseLower, keyword) {
			docs = append(docs, url)
		}
	}

	return docs
}

// GenerateOptimizationRecommendations generates cost/resource optimization recommendations
func (s *Service) GenerateOptimizationRecommendations(ctx context.Context, metrics map[string]interface{}) ([]Recommendation, error) {
	question := "Based on the following resource utilization metrics, provide specific optimization recommendations:"

	query, err := s.AskQuestion(ctx, "system", question, metrics)
	if err != nil {
		return nil, err
	}

	// Parse recommendations from response
	recommendations := s.parseRecommendations(query.Response, metrics)

	// Save recommendations
	for i := range recommendations {
		recommendations[i].ID = uuid.New().String()
		recommendations[i].Status = "pending"
		recommendations[i].CreatedAt = time.Now()
		recommendations[i].UpdatedAt = time.Now()

		if err := s.db.Create(&recommendations[i]).Error; err != nil {
			s.logger.Warn("Failed to save recommendation", zap.Error(err))
		}
	}

	return recommendations, nil
}

func (s *Service) parseRecommendations(response string, metrics map[string]interface{}) []Recommendation {
	var recommendations []Recommendation

	// Parse structured recommendations from AI response
	// This is a simplified implementation
	lines := strings.Split(response, "\n")
	var currentRec *Recommendation

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "**") {
			if currentRec != nil {
				recommendations = append(recommendations, *currentRec)
			}
			currentRec = &Recommendation{
				Title: strings.Trim(line, "#* "),
				Type:  "performance",
			}
		} else if currentRec != nil && line != "" {
			currentRec.Description += line + " "
		}
	}

	if currentRec != nil {
		recommendations = append(recommendations, *currentRec)
	}

	return recommendations
}

// CreateChatSession creates a new chat session
func (s *Service) CreateChatSession(ctx context.Context, userID, title string) (*ChatSession, error) {
	session := &ChatSession{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     title,
		Messages:  []ChatMessage{},
		Context:   make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	return session, nil
}

// SendChatMessage sends a message in a chat session
func (s *Service) SendChatMessage(ctx context.Context, sessionID, message string) (*ChatMessage, error) {
	var session ChatSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Add user message
	userMsg := ChatMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMsg)

	// Build conversation history
	var conversationHistory strings.Builder
	for _, msg := range session.Messages {
		conversationHistory.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	// Get AI response
	response, _, err := s.callProvider(ctx, s.config.SystemPrompt+"\n\n"+conversationHistory.String())
	if err != nil {
		return nil, err
	}

	// Add assistant response
	assistantMsg := ChatMessage{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, assistantMsg)
	session.UpdatedAt = time.Now()

	// Update session
	if err := s.db.Save(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &assistantMsg, nil
}

// GetChatSession retrieves a chat session
func (s *Service) GetChatSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	var session ChatSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	return &session, nil
}

// ListChatSessions lists chat sessions for a user
func (s *Service) ListChatSessions(ctx context.Context, userID string, limit, offset int) ([]ChatSession, error) {
	var sessions []ChatSession
	if err := s.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, nil
}

// SubmitFeedback submits feedback for a query
func (s *Service) SubmitFeedback(ctx context.Context, queryID string, rating int, helpful bool, comment string) error {
	feedback := &QueryFeedback{
		ID:        uuid.New().String(),
		QueryID:   queryID,
		Rating:    rating,
		Helpful:   helpful,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(feedback).Error; err != nil {
		return fmt.Errorf("failed to submit feedback: %w", err)
	}

	return nil
}

// GetRecommendations retrieves recommendations
func (s *Service) GetRecommendations(ctx context.Context, filter map[string]interface{}) ([]Recommendation, error) {
	var recommendations []Recommendation
	query := s.db.Model(&Recommendation{})

	if recType, ok := filter["type"]; ok {
		query = query.Where("type = ?", recType)
	}
	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&recommendations).Error; err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	return recommendations, nil
}

// ApplyRecommendation marks a recommendation as applied
func (s *Service) ApplyRecommendation(ctx context.Context, recommendationID, userID string) error {
	now := time.Now()
	if err := s.db.Model(&Recommendation{}).
		Where("id = ?", recommendationID).
		Updates(map[string]interface{}{
			"status":     "applied",
			"applied_at": now,
			"applied_by": userID,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to apply recommendation: %w", err)
	}
	return nil
}

// DismissRecommendation marks a recommendation as dismissed
func (s *Service) DismissRecommendation(ctx context.Context, recommendationID string) error {
	if err := s.db.Model(&Recommendation{}).
		Where("id = ?", recommendationID).
		Updates(map[string]interface{}{
			"status":     "dismissed",
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to dismiss recommendation: %w", err)
	}
	return nil
}

// Cache operations
func (s *Service) getFromCache(key string) (*Query, bool) {
	if val, ok := s.cache.Load(key); ok {
		if entry, ok := val.(*cacheEntry); ok {
			if time.Now().Before(entry.expiry) {
				return entry.query, true
			}
			s.cache.Delete(key)
		}
	}
	return nil, false
}

func (s *Service) saveToCache(key string, query *Query) {
	s.cache.Store(key, &cacheEntry{
		query:  query,
		expiry: time.Now().Add(s.config.CacheTTL),
	})
}

type cacheEntry struct {
	query  *Query
	expiry time.Time
}

// Rate limiter
type rateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

func newRateLimiter(rpm int) *rateLimiter {
	return &rateLimiter{
		tokens:     rpm,
		maxTokens:  rpm,
		refillRate: time.Minute / time.Duration(rpm),
		lastRefill: time.Now(),
	}
}

func (r *rateLimiter) allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.refillRate)

	if tokensToAdd > 0 {
		r.tokens = min(r.maxTokens, r.tokens+tokensToAdd)
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
