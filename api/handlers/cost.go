// Package handlers - Cost management handlers
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anubhavg-icpl/krustron/internal/cost"
	"github.com/gin-gonic/gin"
)

// GetCostSummary returns the platform cost summary
func GetCostSummary(svc *cost.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		summary, err := svc.GetCostSummary(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": summary})
	}
}

// ListCostAllocations returns cost allocations filtered by cluster/namespace
func ListCostAllocations(svc *cost.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		filter := cost.CostAllocationFilter{
			ClusterID:    c.Query("cluster"),
			Namespace:    c.Query("namespace"),
			WorkloadType: c.Query("workload_type"),
			Limit:        limit,
		}
		allocations, err := svc.GetCostAllocation(c.Request.Context(), filter)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": allocations})
	}
}

// ListBudgets returns all configured budgets
func ListBudgets(svc *cost.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		budgets, err := svc.ListBudgets(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": budgets})
	}
}

// CreateBudget creates a new budget
func CreateBudget(svc *cost.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var budget cost.Budget
		if err := c.ShouldBindJSON(&budget); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		budget.ID = "" // let the DB assign
		if budget.PeriodStart.IsZero() {
			budget.PeriodStart = time.Now()
		}
		if err := svc.CreateBudget(c.Request.Context(), &budget); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": budget})
	}
}

// GenerateCostReport generates a cost report
func GenerateCostReport(svc *cost.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		req := cost.ReportRequest{
			Name:      c.Query("name"),
			Type:      c.DefaultQuery("type", "monthly"),
			ClusterID: c.Query("cluster"),
			Namespace: c.Query("namespace"),
			StartTime: now.AddDate(0, -1, 0),
			EndTime:   now,
		}
		report, err := svc.GenerateReport(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": report})
	}
}
