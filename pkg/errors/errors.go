// Package errors provides custom error types for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes
const (
	CodeInternal          = "INTERNAL_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeBadRequest        = "BAD_REQUEST"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeConflict          = "CONFLICT"
	CodeValidation        = "VALIDATION_ERROR"
	CodeDatabase          = "DATABASE_ERROR"
	CodeKubernetes        = "KUBERNETES_ERROR"
	CodeGitOps            = "GITOPS_ERROR"
	CodePipeline          = "PIPELINE_ERROR"
	CodeCluster           = "CLUSTER_ERROR"
	CodeHelm              = "HELM_ERROR"
	CodeAuth              = "AUTH_ERROR"
	CodeSecurity          = "SECURITY_ERROR"
	CodeRateLimited       = "RATE_LIMITED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// AppError represents an application error with code and context
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Err        error             `json:"-"`
	Meta       map[string]string `json:"meta,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithMeta adds metadata to the error
func (e *AppError) WithMeta(key, value string) *AppError {
	if e.Meta == nil {
		e.Meta = make(map[string]string)
	}
	e.Meta[key] = value
	return e
}

// New creates a new AppError
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Common error constructors

// Internal creates an internal server error
func Internal(message string) *AppError {
	return New(CodeInternal, message, http.StatusInternalServerError)
}

// InternalWrap wraps an error as internal server error
func InternalWrap(err error, message string) *AppError {
	return Wrap(err, CodeInternal, message, http.StatusInternalServerError)
}

// NotFound creates a not found error
func NotFound(resource, id string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s with id '%s' not found", resource, id), http.StatusNotFound)
}

// NotFoundMsg creates a not found error with custom message
func NotFoundMsg(message string) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message, http.StatusBadRequest)
}

// BadRequestWrap wraps an error as bad request error
func BadRequestWrap(err error, message string) *AppError {
	return Wrap(err, CodeBadRequest, message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	return New(CodeConflict, message, http.StatusConflict)
}

// Validation creates a validation error
func Validation(message string) *AppError {
	return New(CodeValidation, message, http.StatusBadRequest)
}

// ValidationWrap wraps an error as validation error
func ValidationWrap(err error, message string) *AppError {
	return Wrap(err, CodeValidation, message, http.StatusBadRequest)
}

// Database creates a database error
func Database(message string) *AppError {
	return New(CodeDatabase, message, http.StatusInternalServerError)
}

// DatabaseWrap wraps an error as database error
func DatabaseWrap(err error, message string) *AppError {
	return Wrap(err, CodeDatabase, message, http.StatusInternalServerError)
}

// Kubernetes creates a Kubernetes error
func Kubernetes(message string) *AppError {
	return New(CodeKubernetes, message, http.StatusInternalServerError)
}

// KubernetesWrap wraps an error as Kubernetes error
func KubernetesWrap(err error, message string) *AppError {
	return Wrap(err, CodeKubernetes, message, http.StatusInternalServerError)
}

// GitOps creates a GitOps error
func GitOps(message string) *AppError {
	return New(CodeGitOps, message, http.StatusInternalServerError)
}

// GitOpsWrap wraps an error as GitOps error
func GitOpsWrap(err error, message string) *AppError {
	return Wrap(err, CodeGitOps, message, http.StatusInternalServerError)
}

// Pipeline creates a pipeline error
func Pipeline(message string) *AppError {
	return New(CodePipeline, message, http.StatusInternalServerError)
}

// PipelineWrap wraps an error as pipeline error
func PipelineWrap(err error, message string) *AppError {
	return Wrap(err, CodePipeline, message, http.StatusInternalServerError)
}

// Cluster creates a cluster error
func Cluster(message string) *AppError {
	return New(CodeCluster, message, http.StatusInternalServerError)
}

// ClusterWrap wraps an error as cluster error
func ClusterWrap(err error, message string) *AppError {
	return Wrap(err, CodeCluster, message, http.StatusInternalServerError)
}

// Helm creates a Helm error
func Helm(message string) *AppError {
	return New(CodeHelm, message, http.StatusInternalServerError)
}

// HelmWrap wraps an error as Helm error
func HelmWrap(err error, message string) *AppError {
	return Wrap(err, CodeHelm, message, http.StatusInternalServerError)
}

// Auth creates an auth error
func Auth(message string) *AppError {
	return New(CodeAuth, message, http.StatusUnauthorized)
}

// AuthWrap wraps an error as auth error
func AuthWrap(err error, message string) *AppError {
	return Wrap(err, CodeAuth, message, http.StatusUnauthorized)
}

// Security creates a security error
func Security(message string) *AppError {
	return New(CodeSecurity, message, http.StatusForbidden)
}

// SecurityWrap wraps an error as security error
func SecurityWrap(err error, message string) *AppError {
	return Wrap(err, CodeSecurity, message, http.StatusForbidden)
}

// RateLimited creates a rate limited error
func RateLimited(message string) *AppError {
	return New(CodeRateLimited, message, http.StatusTooManyRequests)
}

// ServiceUnavailable creates a service unavailable error
func ServiceUnavailable(message string) *AppError {
	return New(CodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

// Is checks if an error is of a specific type
func Is(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetCode returns the error code
func GetCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}

// ToAppError converts any error to AppError
func ToAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return InternalWrap(err, err.Error())
}

// ErrorResponse is the JSON response for errors
type ErrorResponse struct {
	Error   ErrorBody `json:"error"`
	TraceID string    `json:"trace_id,omitempty"`
}

// ErrorBody is the error body in the response
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details string            `json:"details,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// ToResponse converts AppError to ErrorResponse
func (e *AppError) ToResponse(traceID string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
			Meta:    e.Meta,
		},
		TraceID: traceID,
	}
}
