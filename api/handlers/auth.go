// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anubhavg-icpl/krustron/internal/auth"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// Login handles user login
func Login(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		resp, err := svc.Login(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resp})
	}
}

// Register handles user registration
func Register(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		resp, err := svc.Register(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": resp})
	}
}

// RefreshToken refreshes an access token
func RefreshToken(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		resp, err := svc.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resp})
	}
}

// OIDCLogin initiates OIDC login flow
func OIDCLogin(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		url, state, err := svc.GetOIDCAuthURL()
		if err != nil {
			handleError(c, err)
			return
		}

		// Store state in session/cookie
		c.SetCookie("oauth_state", state, 600, "/", "", true, true)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// OIDCCallback handles OIDC callback
func OIDCCallback(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		// Verify state
		savedState, err := c.Cookie("oauth_state")
		if err != nil || savedState != state {
			c.JSON(http.StatusBadRequest, errors.BadRequest("invalid state").ToResponse(getRequestID(c)))
			return
		}

		resp, err := svc.HandleOIDCCallback(c.Request.Context(), code)
		if err != nil {
			handleError(c, err)
			return
		}

		// Clear state cookie
		c.SetCookie("oauth_state", "", -1, "/", "", true, true)

		c.JSON(http.StatusOK, gin.H{"data": resp})
	}
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		user, err := svc.GetUser(c.Request.Context(), userID.(string))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": user})
	}
}

// UpdateCurrentUser updates the current user's profile
func UpdateCurrentUser(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req auth.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		user, err := svc.UpdateUser(c.Request.Context(), userID.(string), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": user})
	}
}

// Logout handles user logout
func Logout(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		if err := svc.Logout(c.Request.Context(), userID.(string)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	}
}

// ChangePassword changes the user's password
func ChangePassword(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req auth.ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		if err := svc.ChangePassword(c.Request.Context(), userID.(string), &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
	}
}

// ListUsers returns all users (admin only)
func ListUsers(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		users, total, err := svc.ListUsers(c.Request.Context(), page, limit)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  users,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetUser returns a single user
func GetUser(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		user, err := svc.GetUser(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": user})
	}
}

// CreateUser creates a new user (admin only)
func CreateUser(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		user, err := svc.CreateUser(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": user})
	}
}

// UpdateUser updates a user (admin only)
func UpdateUser(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req auth.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		user, err := svc.UpdateUser(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": user})
	}
}

// DeleteUser deletes a user (admin only)
func DeleteUser(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := svc.DeleteUser(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
	}
}

// AssignUserRoles assigns roles to a user
func AssignUserRoles(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req auth.AssignRolesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		if err := svc.AssignRoles(c.Request.Context(), id, &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "roles assigned successfully"})
	}
}

// ListRoles returns all roles
func ListRoles(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := svc.ListRoles(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": roles})
	}
}

// GetRole returns a single role
func GetRole(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		role, err := svc.GetRole(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": role})
	}
}

// CreateRole creates a new role
func CreateRole(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		role, err := svc.CreateRole(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": role})
	}
}

// UpdateRole updates a role
func UpdateRole(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req auth.UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		role, err := svc.UpdateRole(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": role})
	}
}

// DeleteRole deletes a role
func DeleteRole(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := svc.DeleteRole(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "role deleted successfully"})
	}
}

// ListPermissions returns all available permissions
func ListPermissions(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions := svc.ListPermissions()
		c.JSON(http.StatusOK, gin.H{"data": permissions})
	}
}

// ListAuditLogs returns audit logs
func ListAuditLogs(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		logs, total, err := svc.ListAuditLogs(c.Request.Context(), page, limit)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  logs,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetAuditLog returns a single audit log entry
func GetAuditLog(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		log, err := svc.GetAuditLog(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": log})
	}
}

// ExportAuditLogs exports audit logs
func ExportAuditLogs(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		format := c.DefaultQuery("format", "csv")

		data, contentType, err := svc.ExportAuditLogs(c.Request.Context(), format)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Header("Content-Disposition", "attachment; filename=audit_logs."+format)
		c.Data(http.StatusOK, contentType, data)
	}
}

// GetSettings returns system settings
func GetSettings(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := svc.GetSettings(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": settings})
	}
}

// UpdateSettings updates system settings
func UpdateSettings(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		if err := svc.UpdateSettings(c.Request.Context(), req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "settings updated successfully"})
	}
}

// GetNotificationSettings returns notification settings
func GetNotificationSettings(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := svc.GetNotificationSettings(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": settings})
	}
}

// UpdateNotificationSettings updates notification settings
func UpdateNotificationSettings(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		if err := svc.UpdateNotificationSettings(c.Request.Context(), req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "notification settings updated successfully"})
	}
}
