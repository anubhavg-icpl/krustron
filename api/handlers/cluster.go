// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anubhavg-icpl/krustron/internal/cluster"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ListClusters returns all clusters
func ListClusters(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		environment := c.Query("environment")
		status := c.Query("status")

		filters := &cluster.ListFilters{
			Page:        page,
			Limit:       limit,
			Environment: environment,
			Status:      status,
		}

		clusters, total, err := svc.List(c.Request.Context(), filters)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  clusters,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetCluster returns a single cluster
func GetCluster(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		cluster, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": cluster})
	}
}

// CreateCluster creates a new cluster
func CreateCluster(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req cluster.CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		userID, _ := c.Get("user_id")
		req.CreatedBy = userID.(string)

		created, err := svc.Create(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": created})
	}
}

// UpdateCluster updates a cluster
func UpdateCluster(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req cluster.UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		updated, err := svc.Update(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": updated})
	}
}

// DeleteCluster deletes a cluster
func DeleteCluster(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := svc.Delete(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "cluster deleted successfully"})
	}
}

// GetClusterHealth returns cluster health status
func GetClusterHealth(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		health, err := svc.GetHealth(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": health})
	}
}

// GetClusterResources returns cluster resources summary
func GetClusterResources(svc *cluster.Service) gin.HandlerFunc {
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

// GetNamespaces returns namespaces in a cluster
func GetNamespaces(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		namespaces, err := svc.GetNamespaces(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": namespaces})
	}
}

// GetPods returns pods in a namespace
func GetPods(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.Param("id")
		namespace := c.Param("namespace")

		pods, err := svc.GetPods(c.Request.Context(), clusterID, namespace)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": pods})
	}
}

// GetPodLogs returns logs from a pod
func GetPodLogs(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.Param("id")
		namespace := c.Param("namespace")
		podName := c.Param("pod")
		container := c.Query("container")
		tailLines, _ := strconv.ParseInt(c.DefaultQuery("tail", "100"), 10, 64)

		logs, err := svc.GetPodLogs(c.Request.Context(), clusterID, namespace, podName, container, tailLines)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": logs})
	}
}

// GetServices returns services in a namespace
func GetServices(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.Param("id")
		namespace := c.Param("namespace")

		services, err := svc.GetServices(c.Request.Context(), clusterID, namespace)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": services})
	}
}

// GetDeployments returns deployments in a namespace
func GetDeployments(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.Param("id")
		namespace := c.Param("namespace")

		deployments, err := svc.GetDeployments(c.Request.Context(), clusterID, namespace)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": deployments})
	}
}

// GetEvents returns events in a namespace
func GetEvents(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.Param("id")
		namespace := c.Param("namespace")

		events, err := svc.GetEvents(c.Request.Context(), clusterID, namespace)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": events})
	}
}

// InstallAgent installs the Krustron agent on a cluster
func InstallAgent(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := svc.InstallAgent(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "agent installation initiated"})
	}
}

// ClusterEventsWS streams cluster events via WebSocket
func ClusterEventsWS(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebSocket handling will be implemented
		c.JSON(http.StatusNotImplemented, gin.H{"message": "WebSocket not yet implemented"})
	}
}

// PodLogsWS streams pod logs via WebSocket
func PodLogsWS(svc *cluster.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebSocket handling will be implemented
		c.JSON(http.StatusNotImplemented, gin.H{"message": "WebSocket not yet implemented"})
	}
}

// Helper function to handle errors
func handleError(c *gin.Context, err error) {
	appErr := errors.ToAppError(err)
	c.JSON(appErr.HTTPStatus, appErr.ToResponse(getRequestID(c)))
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}
