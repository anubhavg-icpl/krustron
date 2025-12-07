// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"

	"github.com/anubhavg-icpl/krustron/internal/helm"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ListHelmRepos returns all Helm repositories
func ListHelmRepos(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		repos, err := svc.ListRepositories(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": repos})
	}
}

// AddHelmRepo adds a new Helm repository
func AddHelmRepo(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req helm.AddRepoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		if err := svc.AddRepository(c.Request.Context(), &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "repository added successfully"})
	}
}

// RemoveHelmRepo removes a Helm repository
func RemoveHelmRepo(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := svc.RemoveRepository(c.Request.Context(), name); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "repository removed successfully"})
	}
}

// SyncHelmRepo synchronizes a Helm repository
func SyncHelmRepo(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := svc.SyncRepository(c.Request.Context(), name); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "repository synced successfully"})
	}
}

// SearchCharts searches for Helm charts
func SearchCharts(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyword := c.Query("q")
		repo := c.Query("repo")

		charts, err := svc.SearchCharts(c.Request.Context(), keyword, repo)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": charts})
	}
}

// GetChartDetails returns details of a specific chart
func GetChartDetails(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := c.Param("repo")
		chart := c.Param("chart")
		version := c.Query("version")

		details, err := svc.GetChartDetails(c.Request.Context(), repo, chart, version)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": details})
	}
}

// GetChartVersions returns available versions of a chart
func GetChartVersions(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := c.Param("repo")
		chart := c.Param("chart")

		versions, err := svc.GetChartVersions(c.Request.Context(), repo, chart)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": versions})
	}
}

// ListReleases returns all Helm releases
func ListReleases(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.Query("cluster")
		namespace := c.Query("namespace")

		releases, err := svc.ListReleases(c.Request.Context(), clusterID, namespace)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": releases})
	}
}

// GetRelease returns details of a specific release
func GetRelease(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Param("cluster")
		namespace := c.Param("namespace")
		name := c.Param("name")

		release, err := svc.GetRelease(c.Request.Context(), cluster, namespace, name)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": release})
	}
}

// InstallRelease installs a Helm release
func InstallRelease(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req helm.InstallRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		userID, _ := c.Get("user_id")
		req.CreatedBy = userID.(string)

		release, err := svc.Install(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": release})
	}
}

// UpgradeRelease upgrades a Helm release
func UpgradeRelease(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Param("cluster")
		namespace := c.Param("namespace")
		name := c.Param("name")

		var req helm.UpgradeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		req.ClusterID = cluster
		req.Namespace = namespace
		req.Name = name

		release, err := svc.Upgrade(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": release})
	}
}

// UninstallRelease uninstalls a Helm release
func UninstallRelease(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Param("cluster")
		namespace := c.Param("namespace")
		name := c.Param("name")

		if err := svc.Uninstall(c.Request.Context(), cluster, namespace, name); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "release uninstalled successfully"})
	}
}

// RollbackRelease rolls back a Helm release
func RollbackRelease(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Param("cluster")
		namespace := c.Param("namespace")
		name := c.Param("name")

		var req helm.RollbackRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		if err := svc.Rollback(c.Request.Context(), cluster, namespace, name, req.Revision); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "release rolled back successfully"})
	}
}

// GetReleaseHistory returns the history of a Helm release
func GetReleaseHistory(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Param("cluster")
		namespace := c.Param("namespace")
		name := c.Param("name")

		history, err := svc.GetHistory(c.Request.Context(), cluster, namespace, name)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": history})
	}
}

// GetReleaseValues returns the values of a Helm release
func GetReleaseValues(svc *helm.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Param("cluster")
		namespace := c.Param("namespace")
		name := c.Param("name")
		allValues := c.Query("all") == "true"

		values, err := svc.GetValues(c.Request.Context(), cluster, namespace, name, allValues)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": values})
	}
}
