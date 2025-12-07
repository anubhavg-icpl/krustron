// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anubhavg-icpl/krustron/internal/security"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ListSecurityScans returns all security scans
func ListSecurityScans(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		targetType := c.Query("target_type")
		status := c.Query("status")
		clusterID := c.Query("cluster")

		filters := &security.ScanFilters{
			Page:       page,
			Limit:      limit,
			TargetType: targetType,
			Status:     status,
			ClusterID:  clusterID,
		}

		scans, total, err := svc.ListScans(c.Request.Context(), filters)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  scans,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetSecurityScan returns a single security scan
func GetSecurityScan(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		scan, err := svc.GetScan(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": scan})
	}
}

// TriggerSecurityScan triggers a new security scan
func TriggerSecurityScan(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req security.ScanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		scan, err := svc.TriggerScan(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"data": scan})
	}
}

// ListVulnerabilities returns all vulnerabilities
func ListVulnerabilities(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		severity := c.Query("severity")
		clusterID := c.Query("cluster")
		namespace := c.Query("namespace")

		filters := &security.VulnFilters{
			Page:      page,
			Limit:     limit,
			Severity:  severity,
			ClusterID: clusterID,
			Namespace: namespace,
		}

		vulns, total, err := svc.ListVulnerabilities(c.Request.Context(), filters)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  vulns,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// ListPolicies returns all OPA policies
func ListPolicies(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		policies, err := svc.ListPolicies(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": policies})
	}
}

// CreatePolicy creates a new OPA policy
func CreatePolicy(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req security.PolicyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		policy, err := svc.CreatePolicy(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": policy})
	}
}

// UpdatePolicy updates an OPA policy
func UpdatePolicy(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req security.PolicyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		policy, err := svc.UpdatePolicy(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": policy})
	}
}

// DeletePolicy deletes an OPA policy
func DeletePolicy(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := svc.DeletePolicy(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "policy deleted successfully"})
	}
}

// ValidatePolicy validates resources against an OPA policy
func ValidatePolicy(svc *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req security.ValidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		result, err := svc.ValidatePolicy(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}
