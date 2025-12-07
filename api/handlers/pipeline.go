// Package handlers provides HTTP handlers for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anubhavg-icpl/krustron/internal/pipeline"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ListPipelines returns all pipelines
func ListPipelines(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		appID := c.Query("application")

		filters := &pipeline.ListFilters{
			Page:          page,
			Limit:         limit,
			ApplicationID: appID,
		}

		pipelines, total, err := svc.List(c.Request.Context(), filters)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  pipelines,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetPipeline returns a single pipeline
func GetPipeline(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		p, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": p})
	}
}

// CreatePipeline creates a new pipeline
func CreatePipeline(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pipeline.CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		userID, _ := c.Get("user_id")
		req.CreatedBy = userID.(string)

		p, err := svc.Create(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": p})
	}
}

// UpdatePipeline updates a pipeline
func UpdatePipeline(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req pipeline.UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest(err.Error()).ToResponse(getRequestID(c)))
			return
		}

		p, err := svc.Update(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": p})
	}
}

// DeletePipeline deletes a pipeline
func DeletePipeline(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := svc.Delete(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "pipeline deleted successfully"})
	}
}

// TriggerPipeline triggers a pipeline run
func TriggerPipeline(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req pipeline.TriggerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// Allow empty body
			req = pipeline.TriggerRequest{}
		}

		userID, _ := c.Get("user_id")
		req.TriggeredBy = userID.(string)

		run, err := svc.Trigger(c.Request.Context(), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": run})
	}
}

// ListPipelineRuns returns runs for a pipeline
func ListPipelineRuns(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		status := c.Query("status")

		runs, total, err := svc.ListRuns(c.Request.Context(), id, page, limit, status)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  runs,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// GetPipelineRun returns a single pipeline run
func GetPipelineRun(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		runID := c.Param("runId")

		run, err := svc.GetRun(c.Request.Context(), id, runID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": run})
	}
}

// CancelPipelineRun cancels a pipeline run
func CancelPipelineRun(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		runID := c.Param("runId")

		if err := svc.CancelRun(c.Request.Context(), id, runID); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "pipeline run cancelled"})
	}
}

// RetryPipelineRun retries a failed pipeline run
func RetryPipelineRun(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		runID := c.Param("runId")

		userID, _ := c.Get("user_id")

		run, err := svc.RetryRun(c.Request.Context(), id, runID, userID.(string))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": run})
	}
}

// GetPipelineRunLogs returns logs for a pipeline run
func GetPipelineRunLogs(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		runID := c.Param("runId")
		stage := c.Query("stage")

		logs, err := svc.GetRunLogs(c.Request.Context(), id, runID, stage)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": logs})
	}
}

// PipelineLogsWS streams pipeline logs via WebSocket
func PipelineLogsWS(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebSocket handling will be implemented
		c.JSON(http.StatusNotImplemented, gin.H{"message": "WebSocket not yet implemented"})
	}
}

// GitHubWebhook handles GitHub webhook events
func GitHubWebhook(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		event := c.GetHeader("X-GitHub-Event")
		signature := c.GetHeader("X-Hub-Signature-256")

		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest("failed to read body").ToResponse(getRequestID(c)))
			return
		}

		if err := svc.HandleGitHubWebhook(c.Request.Context(), event, signature, body); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
	}
}

// GitLabWebhook handles GitLab webhook events
func GitLabWebhook(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		event := c.GetHeader("X-Gitlab-Event")
		token := c.GetHeader("X-Gitlab-Token")

		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest("failed to read body").ToResponse(getRequestID(c)))
			return
		}

		if err := svc.HandleGitLabWebhook(c.Request.Context(), event, token, body); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
	}
}

// BitbucketWebhook handles Bitbucket webhook events
func BitbucketWebhook(svc *pipeline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		event := c.GetHeader("X-Event-Key")

		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.BadRequest("failed to read body").ToResponse(getRequestID(c)))
			return
		}

		if err := svc.HandleBitbucketWebhook(c.Request.Context(), event, body); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
	}
}
