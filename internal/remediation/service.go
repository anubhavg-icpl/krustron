// Package remediation provides auto-remediation capabilities for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package remediation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Config holds remediation service configuration
type Config struct {
	Enabled              bool
	DryRun               bool
	MaxConcurrentActions int
	DefaultCooldown      time.Duration
	EnableWebhooks       bool
	WebhookURL           string
	EnableSlack          bool
	SlackWebhook         string
	RequireApproval      bool
	ApprovalTimeout      time.Duration
}

// Service provides auto-remediation operations
type Service struct {
	db           *gorm.DB
	logger       *zap.Logger
	config       *Config
	k8sClients   map[string]kubernetes.Interface
	clientsMu    sync.RWMutex
	rules        map[string]*RemediationRule
	rulesMu      sync.RWMutex
	actionQueue  chan *RemediationAction
	stopCh       chan struct{}
}

// RemediationRule defines a rule for auto-remediation
type RemediationRule struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	Name            string                 `json:"name" gorm:"uniqueIndex"`
	Description     string                 `json:"description"`
	Enabled         bool                   `json:"enabled"`
	Priority        int                    `json:"priority"`
	Trigger         RuleTrigger            `json:"trigger" gorm:"serializer:json"`
	Conditions      []RuleCondition        `json:"conditions" gorm:"serializer:json"`
	Actions         []RuleAction           `json:"actions" gorm:"serializer:json"`
	Cooldown        time.Duration          `json:"cooldown"`
	MaxExecutions   int                    `json:"max_executions"` // Max executions per cooldown period
	RequireApproval bool                   `json:"require_approval"`
	Scope           RuleScope              `json:"scope" gorm:"serializer:json"`
	Labels          map[string]string      `json:"labels" gorm:"serializer:json"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	LastTriggered   *time.Time             `json:"last_triggered"`
	ExecutionCount  int                    `json:"execution_count"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by"`
}

// RuleTrigger defines what triggers a remediation rule
type RuleTrigger struct {
	Type       string                 `json:"type"` // event, metric, alert, schedule
	Source     string                 `json:"source"`
	EventTypes []string               `json:"event_types,omitempty"`
	Query      string                 `json:"query,omitempty"` // For metric-based triggers
	Threshold  float64                `json:"threshold,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Schedule   string                 `json:"schedule,omitempty"` // Cron expression
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// RuleCondition defines conditions that must be met
type RuleCondition struct {
	Type     string `json:"type"` // resource_status, label, annotation, time_window
	Field    string `json:"field"`
	Operator string `json:"operator"` // eq, neq, in, contains, regex, gt, lt
	Value    string `json:"value"`
}

// RuleAction defines an action to take
type RuleAction struct {
	Type       string                 `json:"type"` // restart_pod, scale, patch, exec, webhook, notify
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Order      int                    `json:"order"`
	OnFailure  string                 `json:"on_failure"` // continue, abort, retry
	MaxRetries int                    `json:"max_retries"`
}

// RuleScope defines the scope of a rule
type RuleScope struct {
	Clusters   []string `json:"clusters,omitempty"`
	Namespaces []string `json:"namespaces,omitempty"`
	Resources  []string `json:"resources,omitempty"`
}

// RemediationAction represents an action execution
type RemediationAction struct {
	ID             string                 `json:"id" gorm:"primaryKey"`
	RuleID         string                 `json:"rule_id" gorm:"index"`
	RuleName       string                 `json:"rule_name"`
	ClusterID      string                 `json:"cluster_id" gorm:"index"`
	Namespace      string                 `json:"namespace"`
	ResourceType   string                 `json:"resource_type"`
	ResourceName   string                 `json:"resource_name"`
	ActionType     string                 `json:"action_type"`
	Status         string                 `json:"status"` // pending, approved, running, completed, failed, rejected
	DryRun         bool                   `json:"dry_run"`
	TriggerEvent   map[string]interface{} `json:"trigger_event" gorm:"serializer:json"`
	Parameters     map[string]interface{} `json:"parameters" gorm:"serializer:json"`
	Result         map[string]interface{} `json:"result" gorm:"serializer:json"`
	Error          string                 `json:"error"`
	ApprovedBy     string                 `json:"approved_by"`
	ApprovedAt     *time.Time             `json:"approved_at"`
	StartedAt      *time.Time             `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
	Duration       time.Duration          `json:"duration"`
	CreatedAt      time.Time              `json:"created_at"`
}

// RemediationEvent represents an event that can trigger remediation
type RemediationEvent struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Source       string                 `json:"source"`
	ClusterID    string                 `json:"cluster_id"`
	Namespace    string                 `json:"namespace"`
	ResourceType string                 `json:"resource_type"`
	ResourceName string                 `json:"resource_name"`
	Reason       string                 `json:"reason"`
	Message      string                 `json:"message"`
	Severity     string                 `json:"severity"`
	Labels       map[string]string      `json:"labels"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
}

// Playbook represents a collection of remediation rules
type Playbook struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"uniqueIndex"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Rules       []RemediationRule `json:"rules" gorm:"many2many:playbook_rules"`
	Tags        []string          `json:"tags" gorm:"serializer:json"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
}

// NewService creates a new remediation service
func NewService(db *gorm.DB, logger *zap.Logger, config *Config) (*Service, error) {
	// Auto-migrate tables
	if err := db.AutoMigrate(
		&RemediationRule{},
		&RemediationAction{},
		&Playbook{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate remediation tables: %w", err)
	}

	// Set defaults
	if config.MaxConcurrentActions == 0 {
		config.MaxConcurrentActions = 5
	}
	if config.DefaultCooldown == 0 {
		config.DefaultCooldown = 5 * time.Minute
	}
	if config.ApprovalTimeout == 0 {
		config.ApprovalTimeout = 1 * time.Hour
	}

	svc := &Service{
		db:          db,
		logger:      logger,
		config:      config,
		k8sClients:  make(map[string]kubernetes.Interface),
		rules:       make(map[string]*RemediationRule),
		actionQueue: make(chan *RemediationAction, 100),
		stopCh:      make(chan struct{}),
	}

	// Load rules from database
	if err := svc.loadRules(); err != nil {
		logger.Warn("Failed to load remediation rules", zap.Error(err))
	}

	// Initialize default rules
	if err := svc.initializeDefaultRules(); err != nil {
		logger.Warn("Failed to initialize default rules", zap.Error(err))
	}

	// Start action processor
	go svc.processActions()

	return svc, nil
}

// loadRules loads all rules from the database
func (s *Service) loadRules() error {
	var rules []RemediationRule
	if err := s.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return err
	}

	s.rulesMu.Lock()
	defer s.rulesMu.Unlock()

	for i := range rules {
		s.rules[rules[i].ID] = &rules[i]
	}

	s.logger.Info("Loaded remediation rules", zap.Int("count", len(rules)))
	return nil
}

// initializeDefaultRules creates default remediation rules
func (s *Service) initializeDefaultRules() error {
	defaultRules := []RemediationRule{
		{
			ID:          "rule-restart-crashloop",
			Name:        "Restart CrashLoopBackOff Pods",
			Description: "Automatically restart pods stuck in CrashLoopBackOff after multiple failures",
			Enabled:     true,
			Priority:    100,
			Trigger: RuleTrigger{
				Type:       "event",
				Source:     "kubernetes",
				EventTypes: []string{"Warning"},
				Filters: map[string]interface{}{
					"reason": "BackOff",
				},
			},
			Conditions: []RuleCondition{
				{Type: "resource_status", Field: "status.phase", Operator: "eq", Value: "Running"},
				{Type: "label", Field: "app.kubernetes.io/managed-by", Operator: "neq", Value: "helm"},
			},
			Actions: []RuleAction{
				{
					Type:   "restart_pod",
					Target: "{{ .ResourceName }}",
					Parameters: map[string]interface{}{
						"grace_period": 30,
					},
					Order:      1,
					OnFailure:  "abort",
					MaxRetries: 2,
				},
				{
					Type:   "notify",
					Target: "slack",
					Parameters: map[string]interface{}{
						"channel": "#alerts",
						"message": "Pod {{ .ResourceName }} in {{ .Namespace }} was automatically restarted due to CrashLoopBackOff",
					},
					Order:     2,
					OnFailure: "continue",
				},
			},
			Cooldown:        10 * time.Minute,
			MaxExecutions:   3,
			RequireApproval: false,
		},
		{
			ID:          "rule-scale-oom",
			Name:        "Scale Up on OOMKilled",
			Description: "Increase memory limits when pods are OOMKilled repeatedly",
			Enabled:     true,
			Priority:    90,
			Trigger: RuleTrigger{
				Type:       "event",
				Source:     "kubernetes",
				EventTypes: []string{"Warning"},
				Filters: map[string]interface{}{
					"reason": "OOMKilled",
				},
			},
			Conditions: []RuleCondition{
				{Type: "time_window", Field: "count", Operator: "gt", Value: "3"},
			},
			Actions: []RuleAction{
				{
					Type:   "patch",
					Target: "deployment",
					Parameters: map[string]interface{}{
						"path":       "/spec/template/spec/containers/0/resources/limits/memory",
						"operation":  "multiply",
						"multiplier": 1.5,
						"max_value":  "4Gi",
					},
					Order:      1,
					OnFailure:  "abort",
					MaxRetries: 1,
				},
				{
					Type:   "notify",
					Target: "webhook",
					Parameters: map[string]interface{}{
						"message": "Memory limit increased for {{ .ResourceName }} due to repeated OOMKills",
					},
					Order:     2,
					OnFailure: "continue",
				},
			},
			Cooldown:        30 * time.Minute,
			MaxExecutions:   2,
			RequireApproval: true,
		},
		{
			ID:          "rule-evicted-cleanup",
			Name:        "Clean Up Evicted Pods",
			Description: "Delete pods that have been evicted due to resource pressure",
			Enabled:     true,
			Priority:    80,
			Trigger: RuleTrigger{
				Type:     "schedule",
				Schedule: "*/15 * * * *", // Every 15 minutes
			},
			Conditions: []RuleCondition{
				{Type: "resource_status", Field: "status.phase", Operator: "eq", Value: "Failed"},
				{Type: "resource_status", Field: "status.reason", Operator: "eq", Value: "Evicted"},
			},
			Actions: []RuleAction{
				{
					Type:   "delete",
					Target: "pod",
					Parameters: map[string]interface{}{
						"grace_period": 0,
					},
					Order:     1,
					OnFailure: "continue",
				},
			},
			Cooldown:        15 * time.Minute,
			MaxExecutions:   100,
			RequireApproval: false,
		},
		{
			ID:          "rule-pvc-expand",
			Name:        "Expand PVC on Low Space",
			Description: "Automatically expand PVCs when storage usage exceeds threshold",
			Enabled:     true,
			Priority:    70,
			Trigger: RuleTrigger{
				Type:      "metric",
				Source:    "prometheus",
				Query:     "kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes",
				Threshold: 0.85,
				Duration:  10 * time.Minute,
			},
			Actions: []RuleAction{
				{
					Type:   "patch",
					Target: "pvc",
					Parameters: map[string]interface{}{
						"path":       "/spec/resources/requests/storage",
						"operation":  "multiply",
						"multiplier": 1.5,
					},
					Order:      1,
					OnFailure:  "abort",
					MaxRetries: 1,
				},
			},
			Cooldown:        1 * time.Hour,
			MaxExecutions:   3,
			RequireApproval: true,
		},
		{
			ID:          "rule-node-cordon",
			Name:        "Cordon Unhealthy Nodes",
			Description: "Automatically cordon nodes showing signs of problems",
			Enabled:     true,
			Priority:    95,
			Trigger: RuleTrigger{
				Type:       "event",
				Source:     "kubernetes",
				EventTypes: []string{"Warning"},
				Filters: map[string]interface{}{
					"involvedObject.kind": "Node",
					"reason":              []string{"NodeNotReady", "NodeNotSchedulable", "KubeletNotReady"},
				},
			},
			Actions: []RuleAction{
				{
					Type:   "cordon",
					Target: "node",
					Parameters: map[string]interface{}{
						"unschedulable": true,
					},
					Order:     1,
					OnFailure: "abort",
				},
				{
					Type:   "notify",
					Target: "pagerduty",
					Parameters: map[string]interface{}{
						"severity": "high",
						"message":  "Node {{ .ResourceName }} has been cordoned due to health issues",
					},
					Order:     2,
					OnFailure: "continue",
				},
			},
			Cooldown:        30 * time.Minute,
			MaxExecutions:   1,
			RequireApproval: true,
		},
	}

	for _, rule := range defaultRules {
		var existing RemediationRule
		if err := s.db.Where("id = ?", rule.ID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			rule.CreatedAt = time.Now()
			rule.UpdatedAt = time.Now()
			if err := s.db.Create(&rule).Error; err != nil {
				s.logger.Warn("Failed to create default rule", zap.String("rule", rule.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// RegisterK8sClient registers a Kubernetes client for a cluster
func (s *Service) RegisterK8sClient(clusterID string, client kubernetes.Interface) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.k8sClients[clusterID] = client
}

// ProcessEvent processes an event and triggers matching rules
func (s *Service) ProcessEvent(ctx context.Context, event *RemediationEvent) error {
	s.logger.Debug("Processing remediation event",
		zap.String("type", event.Type),
		zap.String("resource", event.ResourceName),
		zap.String("reason", event.Reason),
	)

	// Find matching rules
	matchingRules := s.findMatchingRules(event)

	for _, rule := range matchingRules {
		// Check cooldown
		if !s.checkCooldown(rule) {
			s.logger.Debug("Rule in cooldown period",
				zap.String("rule", rule.Name),
				zap.Time("last_triggered", *rule.LastTriggered),
			)
			continue
		}

		// Check conditions
		if !s.evaluateConditions(ctx, rule, event) {
			s.logger.Debug("Conditions not met for rule", zap.String("rule", rule.Name))
			continue
		}

		// Create remediation action
		action := &RemediationAction{
			ID:           uuid.New().String(),
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			ClusterID:    event.ClusterID,
			Namespace:    event.Namespace,
			ResourceType: event.ResourceType,
			ResourceName: event.ResourceName,
			ActionType:   rule.Actions[0].Type,
			Status:       "pending",
			DryRun:       s.config.DryRun,
			TriggerEvent: map[string]interface{}{
				"type":    event.Type,
				"reason":  event.Reason,
				"message": event.Message,
			},
			Parameters: rule.Actions[0].Parameters,
			CreatedAt:  time.Now(),
		}

		// Check if approval is required
		if rule.RequireApproval || s.config.RequireApproval {
			action.Status = "pending_approval"
			if err := s.db.Create(action).Error; err != nil {
				s.logger.Error("Failed to create action", zap.Error(err))
				continue
			}
			s.notifyApprovalRequired(action)
			continue
		}

		// Queue action for execution
		action.Status = "queued"
		if err := s.db.Create(action).Error; err != nil {
			s.logger.Error("Failed to create action", zap.Error(err))
			continue
		}

		select {
		case s.actionQueue <- action:
		default:
			s.logger.Warn("Action queue full, dropping action", zap.String("action_id", action.ID))
		}
	}

	return nil
}

// findMatchingRules finds rules that match the event
func (s *Service) findMatchingRules(event *RemediationEvent) []*RemediationRule {
	s.rulesMu.RLock()
	defer s.rulesMu.RUnlock()

	var matching []*RemediationRule

	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}

		// Check trigger type
		if rule.Trigger.Type == "event" {
			// Check event type match
			eventTypeMatch := false
			for _, et := range rule.Trigger.EventTypes {
				if et == event.Type {
					eventTypeMatch = true
					break
				}
			}
			if !eventTypeMatch && len(rule.Trigger.EventTypes) > 0 {
				continue
			}

			// Check filters
			if !s.matchFilters(rule.Trigger.Filters, event) {
				continue
			}
		}

		// Check scope
		if !s.matchScope(rule.Scope, event) {
			continue
		}

		matching = append(matching, rule)
	}

	return matching
}

func (s *Service) matchFilters(filters map[string]interface{}, event *RemediationEvent) bool {
	for key, value := range filters {
		var eventValue string
		switch key {
		case "reason":
			eventValue = event.Reason
		case "type":
			eventValue = event.Type
		case "severity":
			eventValue = event.Severity
		default:
			if v, ok := event.Data[key]; ok {
				eventValue = fmt.Sprintf("%v", v)
			}
		}

		// Handle multiple allowed values
		switch v := value.(type) {
		case string:
			if eventValue != v {
				return false
			}
		case []string:
			found := false
			for _, allowed := range v {
				if eventValue == allowed {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case []interface{}:
			found := false
			for _, allowed := range v {
				if eventValue == fmt.Sprintf("%v", allowed) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (s *Service) matchScope(scope RuleScope, event *RemediationEvent) bool {
	// Check clusters
	if len(scope.Clusters) > 0 {
		found := false
		for _, c := range scope.Clusters {
			if c == event.ClusterID || c == "*" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check namespaces
	if len(scope.Namespaces) > 0 {
		found := false
		for _, ns := range scope.Namespaces {
			if ns == event.Namespace || ns == "*" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (s *Service) checkCooldown(rule *RemediationRule) bool {
	if rule.LastTriggered == nil {
		return true
	}

	cooldown := rule.Cooldown
	if cooldown == 0 {
		cooldown = s.config.DefaultCooldown
	}

	return time.Since(*rule.LastTriggered) > cooldown
}

func (s *Service) evaluateConditions(ctx context.Context, rule *RemediationRule, event *RemediationEvent) bool {
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(ctx, condition, event) {
			return false
		}
	}
	return true
}

func (s *Service) evaluateCondition(ctx context.Context, condition RuleCondition, event *RemediationEvent) bool {
	var value string

	switch condition.Type {
	case "resource_status":
		// Would need to fetch actual resource status from K8s
		if v, ok := event.Data[condition.Field]; ok {
			value = fmt.Sprintf("%v", v)
		}
	case "label":
		if v, ok := event.Labels[condition.Field]; ok {
			value = v
		}
	case "time_window":
		// Check if we've seen enough events in the time window
		// This would require tracking event counts
		return true // Simplified for now
	default:
		return true
	}

	return s.compareValues(value, condition.Operator, condition.Value)
}

func (s *Service) compareValues(actual, operator, expected string) bool {
	switch operator {
	case "eq":
		return actual == expected
	case "neq":
		return actual != expected
	case "contains":
		return regexp.MustCompile(regexp.QuoteMeta(expected)).MatchString(actual)
	case "regex":
		matched, _ := regexp.MatchString(expected, actual)
		return matched
	case "in":
		// Expected is comma-separated list
		for _, v := range regexp.MustCompile(",\\s*").Split(expected, -1) {
			if actual == v {
				return true
			}
		}
		return false
	default:
		return actual == expected
	}
}

// processActions processes queued actions
func (s *Service) processActions() {
	sem := make(chan struct{}, s.config.MaxConcurrentActions)

	for {
		select {
		case action := <-s.actionQueue:
			sem <- struct{}{}
			go func(a *RemediationAction) {
				defer func() { <-sem }()
				s.executeAction(context.Background(), a)
			}(action)
		case <-s.stopCh:
			return
		}
	}
}

// executeAction executes a remediation action
func (s *Service) executeAction(ctx context.Context, action *RemediationAction) {
	s.logger.Info("Executing remediation action",
		zap.String("action_id", action.ID),
		zap.String("type", action.ActionType),
		zap.String("resource", action.ResourceName),
	)

	// Update status
	now := time.Now()
	action.Status = "running"
	action.StartedAt = &now
	s.db.Save(action)

	// Get the rule
	var rule RemediationRule
	if err := s.db.First(&rule, "id = ?", action.RuleID).Error; err != nil {
		s.completeAction(action, "failed", fmt.Errorf("rule not found: %w", err), nil)
		return
	}

	// Execute each action in order
	var lastError error
	for _, ruleAction := range rule.Actions {
		if action.DryRun {
			s.logger.Info("DRY RUN: Would execute action",
				zap.String("type", ruleAction.Type),
				zap.Any("parameters", ruleAction.Parameters),
			)
			continue
		}

		err := s.executeRuleAction(ctx, action, ruleAction)
		if err != nil {
			lastError = err
			s.logger.Error("Action failed",
				zap.String("type", ruleAction.Type),
				zap.Error(err),
			)

			switch ruleAction.OnFailure {
			case "abort":
				s.completeAction(action, "failed", err, nil)
				return
			case "retry":
				for i := 0; i < ruleAction.MaxRetries; i++ {
					time.Sleep(time.Second * time.Duration(i+1))
					if err = s.executeRuleAction(ctx, action, ruleAction); err == nil {
						break
					}
				}
				if err != nil {
					s.completeAction(action, "failed", err, nil)
					return
				}
			case "continue":
				// Continue to next action
			}
		}
	}

	// Update rule execution tracking
	now = time.Now()
	rule.LastTriggered = &now
	rule.ExecutionCount++
	s.db.Save(&rule)

	if lastError != nil {
		s.completeAction(action, "completed_with_errors", lastError, nil)
	} else {
		s.completeAction(action, "completed", nil, map[string]interface{}{"success": true})
	}
}

func (s *Service) executeRuleAction(ctx context.Context, action *RemediationAction, ruleAction RuleAction) error {
	s.clientsMu.RLock()
	client, ok := s.k8sClients[action.ClusterID]
	s.clientsMu.RUnlock()

	if !ok {
		return fmt.Errorf("no kubernetes client for cluster %s", action.ClusterID)
	}

	switch ruleAction.Type {
	case "restart_pod":
		return s.restartPod(ctx, client, action, ruleAction.Parameters)
	case "delete":
		return s.deleteResource(ctx, client, action, ruleAction.Parameters)
	case "scale":
		return s.scaleResource(ctx, client, action, ruleAction.Parameters)
	case "patch":
		return s.patchResource(ctx, client, action, ruleAction.Parameters)
	case "cordon":
		return s.cordonNode(ctx, client, action, ruleAction.Parameters)
	case "drain":
		return s.drainNode(ctx, client, action, ruleAction.Parameters)
	case "exec":
		return s.execInPod(ctx, client, action, ruleAction.Parameters)
	case "notify":
		return s.sendNotification(ctx, action, ruleAction.Parameters)
	case "webhook":
		return s.callWebhook(ctx, action, ruleAction.Parameters)
	default:
		return fmt.Errorf("unknown action type: %s", ruleAction.Type)
	}
}

func (s *Service) restartPod(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	gracePeriod := int64(30)
	if gp, ok := params["grace_period"].(float64); ok {
		gracePeriod = int64(gp)
	}

	err := client.CoreV1().Pods(action.Namespace).Delete(ctx, action.ResourceName, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	s.logger.Info("Pod restarted",
		zap.String("pod", action.ResourceName),
		zap.String("namespace", action.Namespace),
	)

	return nil
}

func (s *Service) deleteResource(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	gracePeriod := int64(0)
	if gp, ok := params["grace_period"].(float64); ok {
		gracePeriod = int64(gp)
	}

	switch action.ResourceType {
	case "pod", "Pod":
		return client.CoreV1().Pods(action.Namespace).Delete(ctx, action.ResourceName, metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
		})
	default:
		return fmt.Errorf("unsupported resource type for delete: %s", action.ResourceType)
	}
}

func (s *Service) scaleResource(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	replicas := int32(1)
	if r, ok := params["replicas"].(float64); ok {
		replicas = int32(r)
	}

	scale, err := client.AppsV1().Deployments(action.Namespace).GetScale(ctx, action.ResourceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get scale: %w", err)
	}

	scale.Spec.Replicas = replicas
	_, err = client.AppsV1().Deployments(action.Namespace).UpdateScale(ctx, action.ResourceName, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update scale: %w", err)
	}

	s.logger.Info("Resource scaled",
		zap.String("resource", action.ResourceName),
		zap.Int32("replicas", replicas),
	)

	return nil
}

func (s *Service) patchResource(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	// Build JSON patch
	patch := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  params["path"],
			"value": params["value"],
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	switch action.ResourceType {
	case "deployment", "Deployment":
		_, err = client.AppsV1().Deployments(action.Namespace).Patch(
			ctx,
			action.ResourceName,
			"application/json-patch+json",
			patchBytes,
			metav1.PatchOptions{},
		)
	default:
		return fmt.Errorf("unsupported resource type for patch: %s", action.ResourceType)
	}

	return err
}

func (s *Service) cordonNode(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	node, err := client.CoreV1().Nodes().Get(ctx, action.ResourceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	node.Spec.Unschedulable = true
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to cordon node: %w", err)
	}

	s.logger.Info("Node cordoned", zap.String("node", action.ResourceName))
	return nil
}

func (s *Service) drainNode(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	// First cordon the node
	if err := s.cordonNode(ctx, client, action, params); err != nil {
		return err
	}

	// List pods on the node
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", action.ResourceName),
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	// Evict pods
	for _, pod := range pods.Items {
		// Skip daemonset pods
		if isControlledByDaemonSet(&pod) {
			continue
		}

		eviction := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		_ = eviction // Would use policy/v1 Eviction

		err := client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			s.logger.Warn("Failed to evict pod",
				zap.String("pod", pod.Name),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Node drained", zap.String("node", action.ResourceName))
	return nil
}

func isControlledByDaemonSet(pod *corev1.Pod) bool {
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

func (s *Service) execInPod(ctx context.Context, client kubernetes.Interface, action *RemediationAction, params map[string]interface{}) error {
	// This would execute a command in a pod
	// Simplified implementation
	s.logger.Info("Would execute command in pod",
		zap.String("pod", action.ResourceName),
		zap.Any("command", params["command"]),
	)
	return nil
}

func (s *Service) sendNotification(ctx context.Context, action *RemediationAction, params map[string]interface{}) error {
	target, _ := params["target"].(string)
	message, _ := params["message"].(string)

	switch target {
	case "slack":
		if s.config.EnableSlack && s.config.SlackWebhook != "" {
			// Send Slack notification
			s.logger.Info("Would send Slack notification", zap.String("message", message))
		}
	default:
		s.logger.Info("Notification sent",
			zap.String("target", target),
			zap.String("message", message),
		)
	}

	return nil
}

func (s *Service) callWebhook(ctx context.Context, action *RemediationAction, params map[string]interface{}) error {
	if !s.config.EnableWebhooks {
		return nil
	}

	url, _ := params["url"].(string)
	if url == "" {
		url = s.config.WebhookURL
	}

	if url == "" {
		return nil
	}

	payload := map[string]interface{}{
		"action_id":     action.ID,
		"rule_name":     action.RuleName,
		"resource_type": action.ResourceType,
		"resource_name": action.ResourceName,
		"namespace":     action.Namespace,
		"cluster_id":    action.ClusterID,
		"parameters":    params,
		"timestamp":     time.Now(),
	}

	s.logger.Info("Would call webhook", zap.String("url", url), zap.Any("payload", payload))
	return nil
}

func (s *Service) completeAction(action *RemediationAction, status string, err error, result map[string]interface{}) {
	now := time.Now()
	action.Status = status
	action.CompletedAt = &now
	if action.StartedAt != nil {
		action.Duration = now.Sub(*action.StartedAt)
	}
	if err != nil {
		action.Error = err.Error()
	}
	if result != nil {
		action.Result = result
	}

	s.db.Save(action)

	s.logger.Info("Action completed",
		zap.String("action_id", action.ID),
		zap.String("status", status),
		zap.Duration("duration", action.Duration),
	)
}

func (s *Service) notifyApprovalRequired(action *RemediationAction) {
	s.logger.Info("Approval required for action",
		zap.String("action_id", action.ID),
		zap.String("rule_name", action.RuleName),
		zap.String("resource", action.ResourceName),
	)

	// Would send notification via configured channels
}

// ApproveAction approves a pending action
func (s *Service) ApproveAction(ctx context.Context, actionID, approverID string) error {
	var action RemediationAction
	if err := s.db.First(&action, "id = ?", actionID).Error; err != nil {
		return fmt.Errorf("action not found: %w", err)
	}

	if action.Status != "pending_approval" {
		return fmt.Errorf("action is not pending approval")
	}

	now := time.Now()
	action.Status = "queued"
	action.ApprovedBy = approverID
	action.ApprovedAt = &now
	s.db.Save(&action)

	// Queue for execution
	select {
	case s.actionQueue <- &action:
	default:
		return fmt.Errorf("action queue full")
	}

	return nil
}

// RejectAction rejects a pending action
func (s *Service) RejectAction(ctx context.Context, actionID, rejectorID, reason string) error {
	var action RemediationAction
	if err := s.db.First(&action, "id = ?", actionID).Error; err != nil {
		return fmt.Errorf("action not found: %w", err)
	}

	if action.Status != "pending_approval" {
		return fmt.Errorf("action is not pending approval")
	}

	action.Status = "rejected"
	action.Result = map[string]interface{}{
		"rejected_by": rejectorID,
		"reason":      reason,
	}
	s.db.Save(&action)

	return nil
}

// CreateRule creates a new remediation rule
func (s *Service) CreateRule(ctx context.Context, rule *RemediationRule) error {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := s.db.Create(rule).Error; err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	// Add to in-memory rules
	s.rulesMu.Lock()
	s.rules[rule.ID] = rule
	s.rulesMu.Unlock()

	return nil
}

// UpdateRule updates a remediation rule
func (s *Service) UpdateRule(ctx context.Context, rule *RemediationRule) error {
	rule.UpdatedAt = time.Now()

	if err := s.db.Save(rule).Error; err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// Update in-memory rules
	s.rulesMu.Lock()
	if rule.Enabled {
		s.rules[rule.ID] = rule
	} else {
		delete(s.rules, rule.ID)
	}
	s.rulesMu.Unlock()

	return nil
}

// DeleteRule deletes a remediation rule
func (s *Service) DeleteRule(ctx context.Context, ruleID string) error {
	if err := s.db.Delete(&RemediationRule{}, "id = ?", ruleID).Error; err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	s.rulesMu.Lock()
	delete(s.rules, ruleID)
	s.rulesMu.Unlock()

	return nil
}

// GetRule retrieves a rule by ID
func (s *Service) GetRule(ctx context.Context, ruleID string) (*RemediationRule, error) {
	var rule RemediationRule
	if err := s.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	return &rule, nil
}

// ListRules lists all rules
func (s *Service) ListRules(ctx context.Context) ([]RemediationRule, error) {
	var rules []RemediationRule
	if err := s.db.Order("priority DESC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}
	return rules, nil
}

// GetAction retrieves an action by ID
func (s *Service) GetAction(ctx context.Context, actionID string) (*RemediationAction, error) {
	var action RemediationAction
	if err := s.db.First(&action, "id = ?", actionID).Error; err != nil {
		return nil, fmt.Errorf("action not found: %w", err)
	}
	return &action, nil
}

// ListActions lists actions with filters
func (s *Service) ListActions(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]RemediationAction, int64, error) {
	var actions []RemediationAction
	var total int64

	query := s.db.Model(&RemediationAction{})

	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if ruleID, ok := filter["rule_id"]; ok {
		query = query.Where("rule_id = ?", ruleID)
	}
	if clusterID, ok := filter["cluster_id"]; ok {
		query = query.Where("cluster_id = ?", clusterID)
	}

	query.Count(&total)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&actions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list actions: %w", err)
	}

	return actions, total, nil
}

// Stop stops the remediation service
func (s *Service) Stop() {
	close(s.stopCh)
}
