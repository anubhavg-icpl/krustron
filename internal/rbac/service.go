// Package rbac provides advanced Role-Based Access Control for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Permission levels for fine-grained access control
const (
	PermissionNone    = "none"
	PermissionView    = "view"
	PermissionEdit    = "edit"
	PermissionAdmin   = "admin"
	PermissionSuper   = "super"
)

// Resource types for RBAC
const (
	ResourceCluster     = "cluster"
	ResourceNamespace   = "namespace"
	ResourceApplication = "application"
	ResourcePipeline    = "pipeline"
	ResourceHelm        = "helm"
	ResourceSecret      = "secret"
	ResourceConfigMap   = "configmap"
	ResourceUser        = "user"
	ResourceRole        = "role"
	ResourceTeam        = "team"
	ResourceEnvironment = "environment"
	ResourceProject     = "project"
)

// Actions for RBAC
const (
	ActionCreate  = "create"
	ActionRead    = "read"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionExecute = "execute"
	ActionApprove = "approve"
	ActionDeploy  = "deploy"
	ActionRollback = "rollback"
)

// Role represents a role with permissions
type Role struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"uniqueIndex"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // system, custom
	Permissions []Permission           `json:"permissions" gorm:"foreignKey:RoleID"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// Permission represents a specific permission
type Permission struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	RoleID     string    `json:"role_id" gorm:"index"`
	Resource   string    `json:"resource"`
	Action     string    `json:"action"`
	Scope      string    `json:"scope"` // global, cluster, namespace, project
	ScopeID    string    `json:"scope_id"`
	Conditions string    `json:"conditions"` // JSON conditions for attribute-based access
	Effect     string    `json:"effect"`     // allow, deny
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

// Team represents a group of users
type Team struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"uniqueIndex"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Members     []TeamMember           `json:"members" gorm:"foreignKey:TeamID"`
	Roles       []TeamRole             `json:"roles" gorm:"foreignKey:TeamID"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TeamID    string    `json:"team_id" gorm:"index"`
	UserID    string    `json:"user_id" gorm:"index"`
	Role      string    `json:"role"` // owner, admin, member
	JoinedAt  time.Time `json:"joined_at"`
	InvitedBy string    `json:"invited_by"`
}

// TeamRole represents a role assigned to a team
type TeamRole struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	TeamID     string    `json:"team_id" gorm:"index"`
	RoleID     string    `json:"role_id" gorm:"index"`
	Scope      string    `json:"scope"`
	ScopeID    string    `json:"scope_id"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy string    `json:"assigned_by"`
}

// Project represents a project for organizing resources
type Project struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"uniqueIndex"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Teams       []string               `json:"teams" gorm:"serializer:json"`
	Clusters    []string               `json:"clusters" gorm:"serializer:json"`
	Namespaces  []string               `json:"namespaces" gorm:"serializer:json"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// AccessRequest represents a request for elevated access
type AccessRequest struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	UserID       string                 `json:"user_id" gorm:"index"`
	RoleID       string                 `json:"role_id"`
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id"`
	Reason       string                 `json:"reason"`
	Duration     time.Duration          `json:"duration"`
	Status       string                 `json:"status"` // pending, approved, denied, expired
	ApprovedBy   string                 `json:"approved_by"`
	ApprovedAt   *time.Time             `json:"approved_at"`
	ExpiresAt    *time.Time             `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string                 `json:"id" gorm:"primaryKey"`
	UserID     string                 `json:"user_id" gorm:"index"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	OldValue   string                 `json:"old_value"`
	NewValue   string                 `json:"new_value"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Result     string                 `json:"result"` // success, failure, denied
	Reason     string                 `json:"reason"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt  time.Time              `json:"created_at"`
}

// PolicyCondition represents a condition for attribute-based access control
type PolicyCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, neq, in, notin, contains, regex
	Value    interface{} `json:"value"`
}

// Service provides RBAC operations
type Service struct {
	db            *gorm.DB
	enforcer      *casbin.Enforcer
	logger        *zap.Logger
	cache         sync.Map
	cacheTTL      time.Duration
	auditEnabled  bool
	webhookURL    string
}

// Config holds RBAC service configuration
type Config struct {
	ModelPath    string
	PolicyPath   string
	CacheTTL     time.Duration
	AuditEnabled bool
	WebhookURL   string
}

// NewService creates a new RBAC service
func NewService(db *gorm.DB, logger *zap.Logger, cfg *Config) (*Service, error) {
	// Auto-migrate tables
	if err := db.AutoMigrate(
		&Role{},
		&Permission{},
		&Team{},
		&TeamMember{},
		&TeamRole{},
		&Project{},
		&AccessRequest{},
		&AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate RBAC tables: %w", err)
	}

	// Create Casbin adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Define the RBAC model with domain support and priority
	modelText := `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft, priority

[role_definition]
g = _, _, _
g2 = _, _

[policy_effect]
e = priority(p.eft) || deny

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin model: %w", err)
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load policies
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	cacheTTL := cfg.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}

	svc := &Service{
		db:           db,
		enforcer:     enforcer,
		logger:       logger,
		cacheTTL:     cacheTTL,
		auditEnabled: cfg.AuditEnabled,
		webhookURL:   cfg.WebhookURL,
	}

	// Initialize default roles
	if err := svc.initializeDefaultRoles(); err != nil {
		logger.Warn("Failed to initialize default roles", zap.Error(err))
	}

	return svc, nil
}

// initializeDefaultRoles creates system default roles
func (s *Service) initializeDefaultRoles() error {
	defaultRoles := []Role{
		{
			ID:          "role-super-admin",
			Name:        "super-admin",
			DisplayName: "Super Administrator",
			Description: "Full access to all resources",
			Type:        "system",
			Permissions: []Permission{
				{ID: uuid.New().String(), Resource: "*", Action: "*", Scope: "global", Effect: "allow", Priority: 1000},
			},
		},
		{
			ID:          "role-cluster-admin",
			Name:        "cluster-admin",
			DisplayName: "Cluster Administrator",
			Description: "Full access to cluster resources",
			Type:        "system",
			Permissions: []Permission{
				{ID: uuid.New().String(), Resource: ResourceCluster, Action: "*", Scope: "cluster", Effect: "allow", Priority: 900},
				{ID: uuid.New().String(), Resource: ResourceNamespace, Action: "*", Scope: "cluster", Effect: "allow", Priority: 900},
			},
		},
		{
			ID:          "role-developer",
			Name:        "developer",
			DisplayName: "Developer",
			Description: "Standard developer access",
			Type:        "system",
			Permissions: []Permission{
				{ID: uuid.New().String(), Resource: ResourceApplication, Action: ActionRead, Scope: "project", Effect: "allow", Priority: 500},
				{ID: uuid.New().String(), Resource: ResourceApplication, Action: ActionUpdate, Scope: "project", Effect: "allow", Priority: 500},
				{ID: uuid.New().String(), Resource: ResourcePipeline, Action: ActionRead, Scope: "project", Effect: "allow", Priority: 500},
				{ID: uuid.New().String(), Resource: ResourcePipeline, Action: ActionExecute, Scope: "project", Effect: "allow", Priority: 500},
			},
		},
		{
			ID:          "role-viewer",
			Name:        "viewer",
			DisplayName: "Viewer",
			Description: "Read-only access",
			Type:        "system",
			Permissions: []Permission{
				{ID: uuid.New().String(), Resource: "*", Action: ActionRead, Scope: "project", Effect: "allow", Priority: 100},
			},
		},
		{
			ID:          "role-release-manager",
			Name:        "release-manager",
			DisplayName: "Release Manager",
			Description: "Can approve and deploy releases",
			Type:        "system",
			Permissions: []Permission{
				{ID: uuid.New().String(), Resource: ResourceApplication, Action: "*", Scope: "project", Effect: "allow", Priority: 700},
				{ID: uuid.New().String(), Resource: ResourcePipeline, Action: "*", Scope: "project", Effect: "allow", Priority: 700},
				{ID: uuid.New().String(), Resource: ResourceHelm, Action: "*", Scope: "project", Effect: "allow", Priority: 700},
			},
		},
		{
			ID:          "role-security-auditor",
			Name:        "security-auditor",
			DisplayName: "Security Auditor",
			Description: "Security scanning and audit access",
			Type:        "system",
			Permissions: []Permission{
				{ID: uuid.New().String(), Resource: "*", Action: ActionRead, Scope: "global", Effect: "allow", Priority: 600},
				{ID: uuid.New().String(), Resource: "security", Action: "*", Scope: "global", Effect: "allow", Priority: 600},
				{ID: uuid.New().String(), Resource: "audit", Action: "*", Scope: "global", Effect: "allow", Priority: 600},
			},
		},
	}

	for _, role := range defaultRoles {
		var existing Role
		if err := s.db.Where("name = ?", role.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			role.CreatedAt = time.Now()
			role.UpdatedAt = time.Now()
			for i := range role.Permissions {
				role.Permissions[i].RoleID = role.ID
				role.Permissions[i].CreatedAt = time.Now()
			}
			if err := s.db.Create(&role).Error; err != nil {
				s.logger.Warn("Failed to create default role", zap.String("role", role.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// Authorize checks if a subject can perform an action on a resource
func (s *Service) Authorize(ctx context.Context, userID, domain, resource, action string) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", userID, domain, resource, action)
	if cached, ok := s.cache.Load(cacheKey); ok {
		if entry, ok := cached.(*cacheEntry); ok {
			if time.Now().Before(entry.expiry) {
				return entry.allowed, nil
			}
			s.cache.Delete(cacheKey)
		}
	}

	// Check with Casbin enforcer
	allowed, err := s.enforcer.Enforce(userID, domain, resource, action)
	if err != nil {
		s.logger.Error("Authorization check failed",
			zap.String("user_id", userID),
			zap.String("resource", resource),
			zap.String("action", action),
			zap.Error(err),
		)
		return false, err
	}

	// Cache the result
	s.cache.Store(cacheKey, &cacheEntry{
		allowed: allowed,
		expiry:  time.Now().Add(s.cacheTTL),
	})

	// Audit log
	if s.auditEnabled {
		result := "denied"
		if allowed {
			result = "success"
		}
		s.logAudit(ctx, userID, action, resource, "", result, "")
	}

	return allowed, nil
}

type cacheEntry struct {
	allowed bool
	expiry  time.Time
}

// AuthorizeWithConditions checks authorization with attribute-based conditions
func (s *Service) AuthorizeWithConditions(ctx context.Context, userID, domain, resource, action string, attributes map[string]interface{}) (bool, error) {
	// First check basic authorization
	allowed, err := s.Authorize(ctx, userID, domain, resource, action)
	if err != nil || !allowed {
		return false, err
	}

	// Get user's permissions with conditions
	var permissions []Permission
	if err := s.db.Joins("JOIN roles ON permissions.role_id = roles.id").
		Joins("JOIN team_roles ON roles.id = team_roles.role_id").
		Joins("JOIN team_members ON team_roles.team_id = team_members.team_id").
		Where("team_members.user_id = ? AND permissions.resource = ?", userID, resource).
		Find(&permissions).Error; err != nil {
		return false, err
	}

	// Check conditions
	for _, perm := range permissions {
		if perm.Conditions != "" {
			var conditions []PolicyCondition
			if err := json.Unmarshal([]byte(perm.Conditions), &conditions); err != nil {
				continue
			}

			if !s.evaluateConditions(conditions, attributes) {
				return false, nil
			}
		}
	}

	return true, nil
}

// evaluateConditions evaluates policy conditions against attributes
func (s *Service) evaluateConditions(conditions []PolicyCondition, attributes map[string]interface{}) bool {
	for _, cond := range conditions {
		attrValue, ok := attributes[cond.Field]
		if !ok {
			return false
		}

		switch cond.Operator {
		case "eq":
			if attrValue != cond.Value {
				return false
			}
		case "neq":
			if attrValue == cond.Value {
				return false
			}
		case "in":
			if values, ok := cond.Value.([]interface{}); ok {
				found := false
				for _, v := range values {
					if v == attrValue {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		case "contains":
			if str, ok := attrValue.(string); ok {
				if condStr, ok := cond.Value.(string); ok {
					if !containsString(str, condStr) {
						return false
					}
				}
			}
		}
	}
	return true
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// CreateRole creates a new role
func (s *Service) CreateRole(ctx context.Context, role *Role) error {
	role.ID = uuid.New().String()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	for i := range role.Permissions {
		role.Permissions[i].ID = uuid.New().String()
		role.Permissions[i].RoleID = role.ID
		role.Permissions[i].CreatedAt = time.Now()
	}

	if err := s.db.Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	// Add policies to Casbin
	for _, perm := range role.Permissions {
		_, err := s.enforcer.AddPolicy(role.Name, perm.Scope, perm.Resource, perm.Action, perm.Effect, fmt.Sprintf("%d", perm.Priority))
		if err != nil {
			s.logger.Warn("Failed to add policy to Casbin", zap.Error(err))
		}
	}

	s.enforcer.SavePolicy()
	s.invalidateCache()

	return nil
}

// UpdateRole updates an existing role
func (s *Service) UpdateRole(ctx context.Context, role *Role) error {
	role.UpdatedAt = time.Now()

	// Delete existing permissions
	if err := s.db.Where("role_id = ?", role.ID).Delete(&Permission{}).Error; err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// Update role and create new permissions
	for i := range role.Permissions {
		role.Permissions[i].ID = uuid.New().String()
		role.Permissions[i].RoleID = role.ID
		role.Permissions[i].CreatedAt = time.Now()
	}

	if err := s.db.Save(role).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Reload policies
	s.enforcer.LoadPolicy()
	s.invalidateCache()

	return nil
}

// DeleteRole deletes a role
func (s *Service) DeleteRole(ctx context.Context, roleID string) error {
	var role Role
	if err := s.db.First(&role, "id = ?", roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.Type == "system" {
		return fmt.Errorf("cannot delete system role")
	}

	if err := s.db.Delete(&role).Error; err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.enforcer.LoadPolicy()
	s.invalidateCache()

	return nil
}

// GetRole retrieves a role by ID
func (s *Service) GetRole(ctx context.Context, roleID string) (*Role, error) {
	var role Role
	if err := s.db.Preload("Permissions").First(&role, "id = ?", roleID).Error; err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}
	return &role, nil
}

// ListRoles lists all roles
func (s *Service) ListRoles(ctx context.Context, filter map[string]interface{}) ([]Role, error) {
	var roles []Role
	query := s.db.Preload("Permissions")

	if roleType, ok := filter["type"]; ok {
		query = query.Where("type = ?", roleType)
	}

	if err := query.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// CreateTeam creates a new team
func (s *Service) CreateTeam(ctx context.Context, team *Team) error {
	team.ID = uuid.New().String()
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()

	if err := s.db.Create(team).Error; err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

// AddTeamMember adds a user to a team
func (s *Service) AddTeamMember(ctx context.Context, teamID, userID, role, invitedBy string) error {
	member := &TeamMember{
		ID:        uuid.New().String(),
		TeamID:    teamID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
		InvitedBy: invitedBy,
	}

	if err := s.db.Create(member).Error; err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	// Add group membership to Casbin
	s.enforcer.AddGroupingPolicy(userID, teamID, "team")
	s.enforcer.SavePolicy()
	s.invalidateCache()

	return nil
}

// RemoveTeamMember removes a user from a team
func (s *Service) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	if err := s.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&TeamMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	s.enforcer.RemoveGroupingPolicy(userID, teamID, "team")
	s.enforcer.SavePolicy()
	s.invalidateCache()

	return nil
}

// AssignRoleToTeam assigns a role to a team for a specific scope
func (s *Service) AssignRoleToTeam(ctx context.Context, teamID, roleID, scope, scopeID, assignedBy string) error {
	teamRole := &TeamRole{
		ID:         uuid.New().String(),
		TeamID:     teamID,
		RoleID:     roleID,
		Scope:      scope,
		ScopeID:    scopeID,
		AssignedAt: time.Now(),
		AssignedBy: assignedBy,
	}

	if err := s.db.Create(teamRole).Error; err != nil {
		return fmt.Errorf("failed to assign role to team: %w", err)
	}

	// Get role name
	var role Role
	if err := s.db.First(&role, "id = ?", roleID).Error; err == nil {
		s.enforcer.AddGroupingPolicy(teamID, role.Name, scope)
		s.enforcer.SavePolicy()
	}

	s.invalidateCache()

	return nil
}

// CreateProject creates a new project
func (s *Service) CreateProject(ctx context.Context, project *Project) error {
	project.ID = uuid.New().String()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	if err := s.db.Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var project Project
	if err := s.db.First(&project, "id = ?", projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	return &project, nil
}

// ListProjects lists projects accessible by a user
func (s *Service) ListProjects(ctx context.Context, userID string) ([]Project, error) {
	var projects []Project

	// Get user's teams
	var teamIDs []string
	if err := s.db.Model(&TeamMember{}).Where("user_id = ?", userID).Pluck("team_id", &teamIDs).Error; err != nil {
		return nil, err
	}

	// Get projects for those teams
	if err := s.db.Where("teams && ?", teamIDs).Find(&projects).Error; err != nil {
		// Fallback: get all projects if array operator not supported
		if err := s.db.Find(&projects).Error; err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}
	}

	return projects, nil
}

// CreateAccessRequest creates a request for elevated access
func (s *Service) CreateAccessRequest(ctx context.Context, req *AccessRequest) error {
	req.ID = uuid.New().String()
	req.Status = "pending"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := s.db.Create(req).Error; err != nil {
		return fmt.Errorf("failed to create access request: %w", err)
	}

	// Send webhook notification if configured
	if s.webhookURL != "" {
		go s.sendAccessRequestNotification(req)
	}

	return nil
}

// ApproveAccessRequest approves an access request
func (s *Service) ApproveAccessRequest(ctx context.Context, requestID, approverID string) error {
	var req AccessRequest
	if err := s.db.First(&req, "id = ?", requestID).Error; err != nil {
		return fmt.Errorf("access request not found: %w", err)
	}

	if req.Status != "pending" {
		return fmt.Errorf("access request is not pending")
	}

	now := time.Now()
	expiresAt := now.Add(req.Duration)
	req.Status = "approved"
	req.ApprovedBy = approverID
	req.ApprovedAt = &now
	req.ExpiresAt = &expiresAt
	req.UpdatedAt = now

	if err := s.db.Save(&req).Error; err != nil {
		return fmt.Errorf("failed to approve access request: %w", err)
	}

	// Grant temporary access
	// This would typically add a temporary policy to Casbin

	s.invalidateCache()

	return nil
}

// DenyAccessRequest denies an access request
func (s *Service) DenyAccessRequest(ctx context.Context, requestID, approverID, reason string) error {
	var req AccessRequest
	if err := s.db.First(&req, "id = ?", requestID).Error; err != nil {
		return fmt.Errorf("access request not found: %w", err)
	}

	now := time.Now()
	req.Status = "denied"
	req.ApprovedBy = approverID
	req.ApprovedAt = &now
	req.UpdatedAt = now
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["denial_reason"] = reason

	if err := s.db.Save(&req).Error; err != nil {
		return fmt.Errorf("failed to deny access request: %w", err)
	}

	return nil
}

// ListAccessRequests lists access requests
func (s *Service) ListAccessRequests(ctx context.Context, filter map[string]interface{}) ([]AccessRequest, error) {
	var requests []AccessRequest
	query := s.db.Model(&AccessRequest{})

	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filter["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("failed to list access requests: %w", err)
	}

	return requests, nil
}

// GetUserPermissions returns all permissions for a user
func (s *Service) GetUserPermissions(ctx context.Context, userID string) ([]Permission, error) {
	var permissions []Permission

	// Get permissions through team memberships
	if err := s.db.Raw(`
		SELECT DISTINCT p.* FROM permissions p
		JOIN roles r ON p.role_id = r.id
		JOIN team_roles tr ON r.id = tr.role_id
		JOIN team_members tm ON tr.team_id = tm.team_id
		WHERE tm.user_id = ?
	`, userID).Scan(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// logAudit creates an audit log entry
func (s *Service) logAudit(ctx context.Context, userID, action, resource, resourceID, result, reason string) {
	audit := &AuditLog{
		ID:         uuid.New().String(),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Result:     result,
		Reason:     reason,
		CreatedAt:  time.Now(),
	}

	// Extract IP and user agent from context if available
	if ipAddr, ok := ctx.Value("ip_address").(string); ok {
		audit.IPAddress = ipAddr
	}
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		audit.UserAgent = userAgent
	}

	if err := s.db.Create(audit).Error; err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}
}

// GetAuditLogs retrieves audit logs with filtering
func (s *Service) GetAuditLogs(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := s.db.Model(&AuditLog{})

	if userID, ok := filter["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filter["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filter["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}
	if startTime, ok := filter["start_time"].(time.Time); ok {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filter["end_time"].(time.Time); ok {
		query = query.Where("created_at <= ?", endTime)
	}

	query.Count(&total)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// invalidateCache clears the authorization cache
func (s *Service) invalidateCache() {
	s.cache = sync.Map{}
}

// sendAccessRequestNotification sends a webhook notification for access requests
func (s *Service) sendAccessRequestNotification(req *AccessRequest) {
	// Implementation would send HTTP POST to webhook URL
	s.logger.Info("Access request notification sent",
		zap.String("request_id", req.ID),
		zap.String("user_id", req.UserID),
	)
}

// SyncPolicies synchronizes policies from database to Casbin
func (s *Service) SyncPolicies(ctx context.Context) error {
	if err := s.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to sync policies: %w", err)
	}
	s.invalidateCache()
	return nil
}

// ExportPolicies exports all policies as JSON
func (s *Service) ExportPolicies(ctx context.Context) ([]byte, error) {
	policies, _ := s.enforcer.GetPolicy()
	groupingPolicies, _ := s.enforcer.GetGroupingPolicy()

	export := map[string]interface{}{
		"policies":         policies,
		"grouping_policies": groupingPolicies,
		"exported_at":      time.Now(),
	}

	return json.Marshal(export)
}

// ImportPolicies imports policies from JSON
func (s *Service) ImportPolicies(ctx context.Context, data []byte) error {
	var imported struct {
		Policies         [][]string `json:"policies"`
		GroupingPolicies [][]string `json:"grouping_policies"`
	}

	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("failed to parse policies: %w", err)
	}

	for _, p := range imported.Policies {
		if len(p) >= 4 {
			params := make([]interface{}, len(p))
			for i, v := range p {
				params[i] = v
			}
			s.enforcer.AddPolicy(params...)
		}
	}

	for _, g := range imported.GroupingPolicies {
		if len(g) >= 2 {
			params := make([]interface{}, len(g))
			for i, v := range g {
				params[i] = v
			}
			s.enforcer.AddGroupingPolicy(params...)
		}
	}

	if err := s.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save imported policies: %w", err)
	}

	s.invalidateCache()

	return nil
}
