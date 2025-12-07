// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anubhavg-icpl/krustron/internal/gitops"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ListApplications returns all applications
func ListApplications(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		clusterID := c.Query("cluster")
		namespace := c.Query("namespace")
		status := c.Query("status")

		filters := &gitops.ListFilters{
			Page:      page,
			Limit:     limit,
			ClusterID: clusterID,
			Namespace: namespace,
			Status:    status,
		}

		apps, total, err := svc.List(c.Request.Context(), filters)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  apps,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetApplication returns a single application
func GetApplication(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		app, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": app})
	}
}

// CreateApplication creates a new application
func CreateApplication(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req gitops.CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		userID, _ := c.Get("user_id")
		req.CreatedBy = userID.(string)

		app, err := svc.Create(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": app})
	}
}

// UpdateApplication updates an application
func UpdateApplication(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req gitops.UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		app, err := svc.Update(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": app})
	}
}

// DeleteApplication deletes an application
func DeleteApplication(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		cascade := c.Query("cascade") == "true"

		if err := svc.Delete(c.Request.Context(), id, cascade); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "application deleted successfully"})
	}
}

// SyncApplication triggers a sync for an application
func SyncApplication(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req gitops.SyncRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// Allow empty body for simple sync
			req = gitops.SyncRequest{}
		}

		status, err := svc.Sync(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": status})
	}
}

// GetApplicationStatus returns the status of an application
func GetApplicationStatus(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		status, err := svc.GetStatus(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": status})
	}
}

// GetApplicationResources returns resources managed by an application
func GetApplicationResources(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		resources, err := svc.GetResources(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resources})
	}
}

// GetApplicationEvents returns events for an application
func GetApplicationEvents(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		events, err := svc.GetEvents(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": events})
	}
}

// GetApplicationManifests returns the manifests for an application
func GetApplicationManifests(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		revision := c.Query("revision")

		manifests, err := svc.GetManifests(c.Request.Context(), id, revision)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": manifests})
	}
}

// GetApplicationDiff returns the diff between live and desired state
func GetApplicationDiff(svc *gitops.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		diff, err := svc.GetDiff(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": diff})
	}
}
