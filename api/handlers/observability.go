// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anubhavg-icpl/krustron/internal/observability"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// GetMetrics returns platform metrics
func GetMetrics(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics, err := svc.GetPlatformMetrics(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": metrics})
	}
}

// GetClusterMetrics returns metrics for a specific cluster
func GetClusterMetrics(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		start := c.Query("start")
		end := c.Query("end")
		step := c.Query("step")

		query := &observability.MetricsQuery{
			ClusterID: id,
			Start:     start,
			End:       end,
			Step:      step,
		}

		metrics, err := svc.GetClusterMetrics(c.Request.Context(), query)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": metrics})
	}
}

// GetApplicationMetrics returns metrics for a specific application
func GetApplicationMetrics(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		start := c.Query("start")
		end := c.Query("end")
		step := c.Query("step")

		query := &observability.MetricsQuery{
			ApplicationID: id,
			Start:         start,
			End:           end,
			Step:          step,
		}

		metrics, err := svc.GetApplicationMetrics(c.Request.Context(), query)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": metrics})
	}
}

// QueryLogs queries logs from the logging backend
func QueryLogs(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		start := c.Query("start")
		end := c.Query("end")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		clusterID := c.Query("cluster")
		namespace := c.Query("namespace")
		pod := c.Query("pod")

		logQuery := &observability.LogQuery{
			Query:     query,
			Start:     start,
			End:       end,
			Limit:     limit,
			ClusterID: clusterID,
			Namespace: namespace,
			Pod:       pod,
		}

		logs, err := svc.QueryLogs(c.Request.Context(), logQuery)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": logs})
	}
}

// QueryTraces queries traces from the tracing backend
func QueryTraces(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Query("service")
		operation := c.Query("operation")
		start := c.Query("start")
		end := c.Query("end")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		minDuration := c.Query("min_duration")

		traceQuery := &observability.TraceQuery{
			Service:     service,
			Operation:   operation,
			Start:       start,
			End:         end,
			Limit:       limit,
			MinDuration: minDuration,
		}

		traces, err := svc.QueryTraces(c.Request.Context(), traceQuery)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": traces})
	}
}

// ListAlerts returns all alerts
func ListAlerts(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		severity := c.Query("severity")
		clusterID := c.Query("cluster")

		filters := &observability.AlertFilters{
			State:     state,
			Severity:  severity,
			ClusterID: clusterID,
		}

		alerts, err := svc.ListAlerts(c.Request.Context(), filters)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": alerts})
	}
}

// ListDashboards returns available Grafana dashboards
func ListDashboards(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		dashboards, err := svc.ListDashboards(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": dashboards})
	}
}

// GetDORAMetrics returns DORA metrics
func GetDORAMetrics(svc *observability.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := c.Query("start")
		end := c.Query("end")
		appID := c.Query("application")
		clusterID := c.Query("cluster")

		query := &observability.DORAQuery{
			Start:         start,
			End:           end,
			ApplicationID: appID,
			ClusterID:     clusterID,
		}

		metrics, err := svc.GetDORAMetrics(c.Request.Context(), query)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": metrics})
	}
}
