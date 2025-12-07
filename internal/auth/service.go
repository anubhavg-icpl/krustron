// Package auth provides authentication and authorization functionality
// Author: Anubhav Gain <anubhavg@infopercept.com>
package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anubhavg-icpl/krustron/pkg/cache"
	"github.com/anubhavg-icpl/krustron/pkg/config"
	"github.com/anubhavg-icpl/krustron/pkg/database"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

// Service provides authentication and authorization functionality
type Service struct {
	db           *database.PostgresDB
	cache        *cache.RedisCache
	config       *config.AuthConfig
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
}

// NewService creates a new auth service
func NewService(db *database.PostgresDB, cache *cache.RedisCache, cfg *config.AuthConfig) *Service {
	svc := &Service{
		db:     db,
		cache:  cache,
		config: cfg,
	}

	// Initialize OIDC if enabled
	if cfg.OIDCEnabled && cfg.OIDCIssuer != "" {
		ctx := context.Background()
		provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
		if err != nil {
			logger.Warn("Failed to initialize OIDC provider", zap.Error(err))
		} else {
			svc.oidcProvider = provider
			svc.oauth2Config = &oauth2.Config{
				ClientID:     cfg.OIDCClientID,
				ClientSecret: cfg.OIDCClientSecret,
				RedirectURL:  cfg.OIDCRedirectURL,
				Endpoint:     provider.Endpoint(),
				Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			}
		}
	}

	return svc
}

// User represents a user
type User struct {
	ID          string     `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	Name        string     `json:"name" db:"name"`
	AvatarURL   string     `json:"avatar_url" db:"avatar_url"`
	Provider    string     `json:"provider" db:"provider"`
	Role        string     `json:"role" db:"role"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Claims represents JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// LoginRequest contains login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse contains login result
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

// RegisterRequest contains registration data
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// RefreshRequest contains refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	var user User
	var passwordHash string

	query := `
		SELECT id, email, password_hash, name, avatar_url, provider, role, is_active,
		       last_login_at, created_at, updated_at
		FROM users WHERE email = $1 AND provider = 'local'
	`

	if err := s.db.QueryRowContext(ctx, query, req.Email).Scan(
		&user.ID, &user.Email, &passwordHash, &user.Name, &user.AvatarURL,
		&user.Provider, &user.Role, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Unauthorized("invalid email or password")
		}
		return nil, errors.DatabaseWrap(err, "failed to query user")
	}

	if !user.IsActive {
		return nil, errors.Unauthorized("account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, errors.Unauthorized("invalid email or password")
	}

	return s.generateTokens(ctx, &user)
}

// Register creates a new user
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	// Check if email exists
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to check email")
	}
	if exists {
		return nil, errors.Conflict("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BCryptCost)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to hash password")
	}

	// Create user
	var user User
	query := `
		INSERT INTO users (email, password_hash, name, provider, role)
		VALUES ($1, $2, $3, 'local', 'user')
		RETURNING id, email, name, avatar_url, provider, role, is_active, created_at, updated_at
	`

	if err := s.db.QueryRowContext(ctx, query, req.Email, string(hashedPassword), req.Name).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Provider,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create user")
	}

	logger.Info("User registered", zap.String("user_id", user.ID), zap.String("email", user.Email))

	return s.generateTokens(ctx, &user)
}

// RefreshToken refreshes an access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Validate refresh token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.Unauthorized("invalid refresh token")
	}

	// Get user
	user, err := s.GetUser(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.Unauthorized("account is disabled")
	}

	return s.generateTokens(ctx, user)
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, errors.AuthWrap(err, "invalid token")
	}

	if !token.Valid {
		return nil, errors.Unauthorized("invalid token")
	}

	return claims, nil
}

// GetOIDCAuthURL returns the OIDC authorization URL
func (s *Service) GetOIDCAuthURL() (string, string, error) {
	if s.oauth2Config == nil {
		return "", "", errors.BadRequest("OIDC not configured")
	}

	state := generateState()
	url := s.oauth2Config.AuthCodeURL(state)

	return url, state, nil
}

// HandleOIDCCallback handles the OIDC callback
func (s *Service) HandleOIDCCallback(ctx context.Context, code string) (*LoginResponse, error) {
	if s.oauth2Config == nil {
		return nil, errors.BadRequest("OIDC not configured")
	}

	oauth2Token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.AuthWrap(err, "failed to exchange code")
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, errors.Auth("no id_token in response")
	}

	verifier := s.oidcProvider.Verifier(&oidc.Config{ClientID: s.config.OIDCClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, errors.AuthWrap(err, "failed to verify id_token")
	}

	var claims struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Sub     string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, errors.AuthWrap(err, "failed to parse claims")
	}

	// Find or create user
	user, err := s.findOrCreateOIDCUser(ctx, &claims)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, user)
}

func (s *Service) findOrCreateOIDCUser(ctx context.Context, claims *struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Sub     string `json:"sub"`
}) (*User, error) {
	var user User

	query := `
		SELECT id, email, name, avatar_url, provider, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`

	err := s.db.QueryRowContext(ctx, query, claims.Email).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL,
		&user.Provider, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new user
		insertQuery := `
			INSERT INTO users (email, name, avatar_url, provider, provider_id, role)
			VALUES ($1, $2, $3, 'oidc', $4, 'user')
			RETURNING id, email, name, avatar_url, provider, role, is_active, created_at, updated_at
		`

		if err := s.db.QueryRowContext(ctx, insertQuery, claims.Email, claims.Name, claims.Picture, claims.Sub).Scan(
			&user.ID, &user.Email, &user.Name, &user.AvatarURL,
			&user.Provider, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to create user")
		}

		logger.Info("OIDC user created", zap.String("user_id", user.ID), zap.String("email", user.Email))
	} else if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to query user")
	}

	return &user, nil
}

func (s *Service) generateTokens(ctx context.Context, user *User) (*LoginResponse, error) {
	// Get user permissions
	permissions := s.getUserPermissions(ctx, user.ID, user.Role)

	now := time.Now()
	accessExpiry := now.Add(s.config.JWTExpiration)
	refreshExpiry := now.Add(s.config.RefreshExpiration)

	// Generate access token
	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "krustron",
			Subject:   user.ID,
		},
		UserID:      user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Role:        user.Role,
		Permissions: permissions,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to sign access token")
	}

	// Generate refresh token
	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "krustron",
			Subject:   user.ID,
		},
		UserID: user.ID,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to sign refresh token")
	}

	// Update last login
	s.db.ExecContext(ctx, "UPDATE users SET last_login_at = NOW() WHERE id = $1", user.ID)

	return &LoginResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
		User:         user,
	}, nil
}

func (s *Service) getUserPermissions(ctx context.Context, userID, role string) []string {
	if role == "admin" {
		return []string{"*"}
	}

	query := `
		SELECT r.permissions
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var allPermissions []string
	for rows.Next() {
		var perms []byte
		if err := rows.Scan(&perms); err != nil {
			continue
		}
		var p []string
		json.Unmarshal(perms, &p)
		allPermissions = append(allPermissions, p...)
	}

	return allPermissions
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetUser returns a user by ID
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	var user User

	query := `
		SELECT id, email, name, avatar_url, provider, role, is_active,
		       last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`

	var lastLoginAt sql.NullTime
	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Provider,
		&user.Role, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("user", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get user")
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// UpdateUserRequest contains user update data
type UpdateUserRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error) {
	query := `
		UPDATE users
		SET name = COALESCE(NULLIF($2, ''), name),
		    avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, id, req.Name, req.AvatarURL)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to update user")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.NotFound("user", id)
	}

	return s.GetUser(ctx, id)
}

// Logout invalidates user session
func (s *Service) Logout(ctx context.Context, userID string) error {
	// In a stateless JWT setup, we just log the logout
	// With Redis, we could blacklist the token
	logger.Info("User logged out", zap.String("user_id", userID))
	return nil
}

// ChangePasswordRequest contains password change data
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword changes user password
func (s *Service) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	var currentHash string
	if err := s.db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentHash); err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("user", userID)
		}
		return errors.DatabaseWrap(err, "failed to get user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		return errors.Unauthorized("current password is incorrect")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.BCryptCost)
	if err != nil {
		return errors.InternalWrap(err, "failed to hash password")
	}

	if _, err := s.db.ExecContext(ctx, "UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1", userID, string(newHash)); err != nil {
		return errors.DatabaseWrap(err, "failed to update password")
	}

	logger.Info("Password changed", zap.String("user_id", userID))
	return nil
}

// ListUsers returns all users with pagination
func (s *Service) ListUsers(ctx context.Context, page, limit int) ([]User, int, error) {
	offset := (page - 1) * limit

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count users")
	}

	query := `
		SELECT id, email, name, avatar_url, provider, role, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query users")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var lastLoginAt sql.NullTime

		if err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Provider,
			&user.Role, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan user")
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	return users, total, nil
}

// CreateUserRequest contains data for creating a user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
}

// CreateUser creates a new user (admin only)
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BCryptCost)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to hash password")
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	var user User
	query := `
		INSERT INTO users (email, password_hash, name, provider, role)
		VALUES ($1, $2, $3, 'local', $4)
		RETURNING id, email, name, avatar_url, provider, role, is_active, created_at, updated_at
	`

	if err := s.db.QueryRowContext(ctx, query, req.Email, string(hashedPassword), req.Name, role).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Provider,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create user")
	}

	return &user, nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete user")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("user", id)
	}

	return nil
}

// AssignRolesRequest contains role assignment data
type AssignRolesRequest struct {
	RoleIDs   []string `json:"role_ids"`
	ClusterID string   `json:"cluster_id"`
	Namespace string   `json:"namespace"`
}

// AssignRoles assigns roles to a user
func (s *Service) AssignRoles(ctx context.Context, userID string, req *AssignRolesRequest) error {
	// Remove existing roles for this cluster/namespace
	if _, err := s.db.ExecContext(ctx,
		"DELETE FROM user_roles WHERE user_id = $1 AND cluster_id = $2",
		userID, req.ClusterID); err != nil {
		return errors.DatabaseWrap(err, "failed to remove existing roles")
	}

	// Add new roles
	for _, roleID := range req.RoleIDs {
		if _, err := s.db.ExecContext(ctx,
			"INSERT INTO user_roles (user_id, role_id, cluster_id, namespace) VALUES ($1, $2, $3, $4)",
			userID, roleID, req.ClusterID, req.Namespace); err != nil {
			return errors.DatabaseWrap(err, "failed to assign role")
		}
	}

	return nil
}

// Role represents a role
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListRoles returns all roles
func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	query := `
		SELECT id, name, display_name, description, permissions, is_system, created_at, updated_at
		FROM roles
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to query roles")
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var perms []byte

		if err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&perms, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan role")
		}

		json.Unmarshal(perms, &role.Permissions)
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRole returns a role by ID
func (s *Service) GetRole(ctx context.Context, id string) (*Role, error) {
	var role Role
	var perms []byte

	query := `
		SELECT id, name, display_name, description, permissions, is_system, created_at, updated_at
		FROM roles WHERE id = $1
	`

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&perms, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("role", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get role")
	}

	json.Unmarshal(perms, &role.Permissions)
	return &role, nil
}

// CreateRoleRequest contains role creation data
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
}

// CreateRole creates a new role
func (s *Service) CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error) {
	perms, _ := json.Marshal(req.Permissions)

	var role Role
	query := `
		INSERT INTO roles (name, display_name, description, permissions)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, display_name, description, permissions, is_system, created_at, updated_at
	`

	var permsOut []byte
	if err := s.db.QueryRowContext(ctx, query, req.Name, req.DisplayName, req.Description, perms).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&permsOut, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create role")
	}

	json.Unmarshal(permsOut, &role.Permissions)
	return &role, nil
}

// UpdateRoleRequest contains role update data
type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// UpdateRole updates a role
func (s *Service) UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest) (*Role, error) {
	perms, _ := json.Marshal(req.Permissions)

	query := `
		UPDATE roles
		SET display_name = COALESCE(NULLIF($2, ''), display_name),
		    description = COALESCE(NULLIF($3, ''), description),
		    permissions = COALESCE($4, permissions),
		    updated_at = NOW()
		WHERE id = $1 AND is_system = false
	`

	result, err := s.db.ExecContext(ctx, query, id, req.DisplayName, req.Description, perms)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to update role")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.BadRequest("cannot update system role")
	}

	return s.GetRole(ctx, id)
}

// DeleteRole deletes a role
func (s *Service) DeleteRole(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM roles WHERE id = $1 AND is_system = false", id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete role")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.BadRequest("cannot delete system role")
	}

	return nil
}

// ListPermissions returns all available permissions
func (s *Service) ListPermissions() []string {
	return []string{
		"clusters:read", "clusters:write", "clusters:delete",
		"applications:read", "applications:write", "applications:delete", "applications:sync",
		"pipelines:read", "pipelines:write", "pipelines:delete", "pipelines:trigger",
		"helm:read", "helm:write", "helm:delete",
		"security:read", "security:write",
		"audit:read", "audit:export",
		"users:read", "users:write", "users:delete",
		"roles:read", "roles:write", "roles:delete",
		"settings:read", "settings:write",
	}
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	UserEmail    string                 `json:"user_email"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceName string                 `json:"resource_name"`
	ClusterID    string                 `json:"cluster_id"`
	ClusterName  string                 `json:"cluster_name"`
	OldValue     map[string]interface{} `json:"old_value,omitempty"`
	NewValue     map[string]interface{} `json:"new_value,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	IPAddress    string                 `json:"ip_address"`
	Status       string                 `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ListAuditLogs returns audit logs with pagination
func (s *Service) ListAuditLogs(ctx context.Context, page, limit int) ([]AuditLog, int, error) {
	offset := (page - 1) * limit

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count audit logs")
	}

	query := `
		SELECT id, user_id, user_email, action, resource_type, resource_id,
		       resource_name, cluster_id, cluster_name, metadata, ip_address,
		       status, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query audit logs")
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var metadata []byte
		var userID, clusterID sql.NullString

		if err := rows.Scan(
			&log.ID, &userID, &log.UserEmail, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.ResourceName, &clusterID, &log.ClusterName,
			&metadata, &log.IPAddress, &log.Status, &log.CreatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan audit log")
		}

		if userID.Valid {
			log.UserID = userID.String
		}
		if clusterID.Valid {
			log.ClusterID = clusterID.String
		}
		json.Unmarshal(metadata, &log.Metadata)

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetAuditLog returns a single audit log
func (s *Service) GetAuditLog(ctx context.Context, id string) (*AuditLog, error) {
	var log AuditLog
	var metadata, oldValue, newValue []byte
	var userID, clusterID sql.NullString

	query := `
		SELECT id, user_id, user_email, action, resource_type, resource_id,
		       resource_name, cluster_id, cluster_name, old_value, new_value,
		       metadata, ip_address, status, created_at
		FROM audit_logs WHERE id = $1
	`

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &userID, &log.UserEmail, &log.Action, &log.ResourceType,
		&log.ResourceID, &log.ResourceName, &clusterID, &log.ClusterName,
		&oldValue, &newValue, &metadata, &log.IPAddress, &log.Status, &log.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("audit log", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get audit log")
	}

	if userID.Valid {
		log.UserID = userID.String
	}
	if clusterID.Valid {
		log.ClusterID = clusterID.String
	}
	json.Unmarshal(metadata, &log.Metadata)
	json.Unmarshal(oldValue, &log.OldValue)
	json.Unmarshal(newValue, &log.NewValue)

	return &log, nil
}

// ExportAuditLogs exports audit logs in the specified format
func (s *Service) ExportAuditLogs(ctx context.Context, format string) ([]byte, string, error) {
	logs, _, err := s.ListAuditLogs(ctx, 1, 10000)
	if err != nil {
		return nil, "", err
	}

	switch format {
	case "json":
		data, _ := json.Marshal(logs)
		return data, "application/json", nil
	case "csv":
		// Simple CSV export
		csv := "id,user_email,action,resource_type,resource_id,created_at\n"
		for _, log := range logs {
			csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
				log.ID, log.UserEmail, log.Action, log.ResourceType, log.ResourceID, log.CreatedAt.Format(time.RFC3339))
		}
		return []byte(csv), "text/csv", nil
	default:
		return nil, "", errors.BadRequest("unsupported format")
	}
}

// GetSettings returns system settings
func (s *Service) GetSettings(ctx context.Context) (map[string]interface{}, error) {
	// Return default settings for now
	return map[string]interface{}{
		"auth": map[string]interface{}{
			"oidc_enabled":    s.config.OIDCEnabled,
			"session_timeout": s.config.JWTExpiration.String(),
		},
	}, nil
}

// UpdateSettings updates system settings
func (s *Service) UpdateSettings(ctx context.Context, settings map[string]interface{}) error {
	// Settings update logic
	return nil
}

// GetNotificationSettings returns notification settings
func (s *Service) GetNotificationSettings(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"email":   map[string]interface{}{"enabled": false},
		"slack":   map[string]interface{}{"enabled": false},
		"webhook": map[string]interface{}{"enabled": false},
	}, nil
}

// UpdateNotificationSettings updates notification settings
func (s *Service) UpdateNotificationSettings(ctx context.Context, settings map[string]interface{}) error {
	return nil
}
