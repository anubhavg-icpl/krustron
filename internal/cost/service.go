// Package cost provides cost management and optimization for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package cost

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Config holds cost service configuration
type Config struct {
	KubecostEndpoint    string
	PrometheusEndpoint  string
	CloudProvider       string // aws, gcp, azure
	AWSRegion           string
	GCPProject          string
	AzureSubscription   string
	DefaultCurrency     string
	EnableRightsizing   bool
	EnableForecasting   bool
	AlertThreshold      float64
	CacheEnabled        bool
	CacheTTL            time.Duration
}

// Service provides cost management operations
type Service struct {
	db          *gorm.DB
	logger      *zap.Logger
	config      *Config
	httpClient  *http.Client
	cache       sync.Map
	pricingData map[string]map[string]float64
}

// CostAllocation represents cost allocation for a resource
type CostAllocation struct {
	ID                 string                 `json:"id" gorm:"primaryKey"`
	ClusterID          string                 `json:"cluster_id" gorm:"index"`
	ClusterName        string                 `json:"cluster_name"`
	Namespace          string                 `json:"namespace" gorm:"index"`
	WorkloadType       string                 `json:"workload_type"`
	WorkloadName       string                 `json:"workload_name"`
	ContainerName      string                 `json:"container_name"`
	Labels             map[string]string      `json:"labels" gorm:"serializer:json"`
	CPUCoreHours       float64                `json:"cpu_core_hours"`
	CPUCost            float64                `json:"cpu_cost"`
	MemoryGBHours      float64                `json:"memory_gb_hours"`
	MemoryCost         float64                `json:"memory_cost"`
	StorageGBHours     float64                `json:"storage_gb_hours"`
	StorageCost        float64                `json:"storage_cost"`
	NetworkCost        float64                `json:"network_cost"`
	GPUCost            float64                `json:"gpu_cost"`
	TotalCost          float64                `json:"total_cost"`
	Efficiency         float64                `json:"efficiency"` // 0-100%
	Metadata           map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	PeriodStart        time.Time              `json:"period_start"`
	PeriodEnd          time.Time              `json:"period_end"`
	CreatedAt          time.Time              `json:"created_at"`
}

// CostReport represents a cost report
type CostReport struct {
	ID              string                   `json:"id" gorm:"primaryKey"`
	Name            string                   `json:"name"`
	Type            string                   `json:"type"` // daily, weekly, monthly, custom
	Grouping        []string                 `json:"grouping" gorm:"serializer:json"` // cluster, namespace, label, workload
	Filters         map[string]interface{}   `json:"filters" gorm:"serializer:json"`
	PeriodStart     time.Time                `json:"period_start"`
	PeriodEnd       time.Time                `json:"period_end"`
	TotalCost       float64                  `json:"total_cost"`
	CPUCost         float64                  `json:"cpu_cost"`
	MemoryCost      float64                  `json:"memory_cost"`
	StorageCost     float64                  `json:"storage_cost"`
	NetworkCost     float64                  `json:"network_cost"`
	Breakdown       []CostBreakdown          `json:"breakdown" gorm:"foreignKey:ReportID"`
	Trends          []CostTrend              `json:"trends" gorm:"serializer:json"`
	Recommendations []CostRecommendation     `json:"recommendations" gorm:"foreignKey:ReportID"`
	GeneratedAt     time.Time                `json:"generated_at"`
	CreatedBy       string                   `json:"created_by"`
}

// CostBreakdown represents cost breakdown by category
type CostBreakdown struct {
	ID        string  `json:"id" gorm:"primaryKey"`
	ReportID  string  `json:"report_id" gorm:"index"`
	Category  string  `json:"category"`
	Name      string  `json:"name"`
	Cost      float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
	Trend     float64 `json:"trend"` // % change from previous period
}

// CostTrend represents cost trend data
type CostTrend struct {
	Date   time.Time `json:"date"`
	Cost   float64   `json:"cost"`
	Change float64   `json:"change"` // % change
}

// CostRecommendation represents a cost optimization recommendation
type CostRecommendation struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	ReportID         string                 `json:"report_id" gorm:"index"`
	Type             string                 `json:"type"` // rightsize, idle, spot, reserved
	Category         string                 `json:"category"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	ResourceType     string                 `json:"resource_type"`
	ResourceName     string                 `json:"resource_name"`
	Namespace        string                 `json:"namespace"`
	ClusterID        string                 `json:"cluster_id"`
	CurrentCost      float64                `json:"current_cost"`
	ProjectedCost    float64                `json:"projected_cost"`
	MonthlySavings   float64                `json:"monthly_savings"`
	AnnualSavings    float64                `json:"annual_savings"`
	Effort           string                 `json:"effort"` // low, medium, high
	Risk             string                 `json:"risk"`   // low, medium, high
	CurrentState     map[string]interface{} `json:"current_state" gorm:"serializer:json"`
	RecommendedState map[string]interface{} `json:"recommended_state" gorm:"serializer:json"`
	Status           string                 `json:"status"` // pending, applied, dismissed
	AppliedAt        *time.Time             `json:"applied_at"`
	AppliedBy        string                 `json:"applied_by"`
	CreatedAt        time.Time              `json:"created_at"`
}

// Budget represents a cost budget
type Budget struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"` // monthly, quarterly, annual
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Scope         string                 `json:"scope"` // cluster, namespace, label
	ScopeValue    string                 `json:"scope_value"`
	Filters       map[string]interface{} `json:"filters" gorm:"serializer:json"`
	AlertThresholds []float64            `json:"alert_thresholds" gorm:"serializer:json"` // e.g., [50, 75, 90, 100]
	CurrentSpend  float64                `json:"current_spend"`
	ForecastSpend float64                `json:"forecast_spend"`
	Status        string                 `json:"status"` // on_track, warning, exceeded
	Alerts        []BudgetAlert          `json:"alerts" gorm:"foreignKey:BudgetID"`
	PeriodStart   time.Time              `json:"period_start"`
	PeriodEnd     time.Time              `json:"period_end"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CreatedBy     string                 `json:"created_by"`
}

// BudgetAlert represents a budget alert
type BudgetAlert struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	BudgetID    string    `json:"budget_id" gorm:"index"`
	Threshold   float64   `json:"threshold"`
	CurrentSpend float64  `json:"current_spend"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"` // warning, critical
	Notified    bool      `json:"notified"`
	NotifiedAt  *time.Time `json:"notified_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// CostForecast represents a cost forecast
type CostForecast struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	Scope        string                 `json:"scope"`
	ScopeValue   string                 `json:"scope_value"`
	ForecastDate time.Time              `json:"forecast_date"`
	CurrentCost  float64                `json:"current_cost"`
	ForecastCost float64                `json:"forecast_cost"`
	Confidence   float64                `json:"confidence"`
	Model        string                 `json:"model"` // linear, seasonal, ml
	Parameters   map[string]interface{} `json:"parameters" gorm:"serializer:json"`
	Predictions  []ForecastPrediction   `json:"predictions" gorm:"serializer:json"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ForecastPrediction represents a single forecast prediction
type ForecastPrediction struct {
	Date       time.Time `json:"date"`
	LowerBound float64   `json:"lower_bound"`
	Predicted  float64   `json:"predicted"`
	UpperBound float64   `json:"upper_bound"`
}

// RightsizingRecommendation represents a rightsizing recommendation
type RightsizingRecommendation struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	ClusterID         string    `json:"cluster_id" gorm:"index"`
	Namespace         string    `json:"namespace"`
	WorkloadType      string    `json:"workload_type"`
	WorkloadName      string    `json:"workload_name"`
	ContainerName     string    `json:"container_name"`
	CurrentCPURequest string    `json:"current_cpu_request"`
	CurrentCPULimit   string    `json:"current_cpu_limit"`
	CurrentMemRequest string    `json:"current_mem_request"`
	CurrentMemLimit   string    `json:"current_mem_limit"`
	RecommendedCPURequest string `json:"recommended_cpu_request"`
	RecommendedCPULimit   string `json:"recommended_cpu_limit"`
	RecommendedMemRequest string `json:"recommended_mem_request"`
	RecommendedMemLimit   string `json:"recommended_mem_limit"`
	CPUUsageP50       float64   `json:"cpu_usage_p50"`
	CPUUsageP95       float64   `json:"cpu_usage_p95"`
	CPUUsageP99       float64   `json:"cpu_usage_p99"`
	MemUsageP50       float64   `json:"mem_usage_p50"`
	MemUsageP95       float64   `json:"mem_usage_p95"`
	MemUsageP99       float64   `json:"mem_usage_p99"`
	MonthlySavings    float64   `json:"monthly_savings"`
	Confidence        float64   `json:"confidence"`
	Status            string    `json:"status"`
	AppliedAt         *time.Time `json:"applied_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// NewService creates a new cost service
func NewService(db *gorm.DB, logger *zap.Logger, config *Config) (*Service, error) {
	// Auto-migrate tables
	if err := db.AutoMigrate(
		&CostAllocation{},
		&CostReport{},
		&CostBreakdown{},
		&CostRecommendation{},
		&Budget{},
		&BudgetAlert{},
		&CostForecast{},
		&RightsizingRecommendation{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate cost tables: %w", err)
	}

	// Set defaults
	if config.DefaultCurrency == "" {
		config.DefaultCurrency = "USD"
	}
	if config.AlertThreshold == 0 {
		config.AlertThreshold = 80.0
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 15 * time.Minute
	}

	svc := &Service{
		db:          db,
		logger:      logger,
		config:      config,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		pricingData: initializePricingData(),
	}

	return svc, nil
}

// initializePricingData initializes default pricing data
func initializePricingData() map[string]map[string]float64 {
	return map[string]map[string]float64{
		"aws": {
			"cpu_per_hour":     0.0336,  // m5.large equivalent
			"memory_gb_hour":   0.00446,
			"storage_gb_month": 0.10,
			"network_gb":       0.09,
		},
		"gcp": {
			"cpu_per_hour":     0.0310,
			"memory_gb_hour":   0.00415,
			"storage_gb_month": 0.08,
			"network_gb":       0.08,
		},
		"azure": {
			"cpu_per_hour":     0.0340,
			"memory_gb_hour":   0.00450,
			"storage_gb_month": 0.10,
			"network_gb":       0.087,
		},
		"on-prem": {
			"cpu_per_hour":     0.025,
			"memory_gb_hour":   0.003,
			"storage_gb_month": 0.05,
			"network_gb":       0.01,
		},
	}
}

// GetCostAllocation retrieves cost allocation data
func (s *Service) GetCostAllocation(ctx context.Context, filter CostAllocationFilter) ([]CostAllocation, error) {
	var allocations []CostAllocation
	query := s.db.Model(&CostAllocation{})

	if filter.ClusterID != "" {
		query = query.Where("cluster_id = ?", filter.ClusterID)
	}
	if filter.Namespace != "" {
		query = query.Where("namespace = ?", filter.Namespace)
	}
	if filter.WorkloadType != "" {
		query = query.Where("workload_type = ?", filter.WorkloadType)
	}
	if !filter.StartTime.IsZero() {
		query = query.Where("period_start >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("period_end <= ?", filter.EndTime)
	}

	if err := query.Order("total_cost DESC").Limit(filter.Limit).Find(&allocations).Error; err != nil {
		return nil, fmt.Errorf("failed to get cost allocation: %w", err)
	}

	return allocations, nil
}

// CostAllocationFilter represents filters for cost allocation queries
type CostAllocationFilter struct {
	ClusterID    string
	Namespace    string
	WorkloadType string
	Labels       map[string]string
	StartTime    time.Time
	EndTime      time.Time
	Limit        int
}

// CalculateCost calculates cost for given resource usage
func (s *Service) CalculateCost(ctx context.Context, usage ResourceUsage) (*CostResult, error) {
	provider := s.config.CloudProvider
	if provider == "" {
		provider = "aws"
	}

	pricing, ok := s.pricingData[provider]
	if !ok {
		pricing = s.pricingData["aws"]
	}

	cpuCost := usage.CPUCoreHours * pricing["cpu_per_hour"]
	memoryCost := usage.MemoryGBHours * pricing["memory_gb_hour"]
	storageCost := usage.StorageGB * pricing["storage_gb_month"] / 720 * usage.Hours // Convert monthly to hourly
	networkCost := usage.NetworkGB * pricing["network_gb"]

	totalCost := cpuCost + memoryCost + storageCost + networkCost

	return &CostResult{
		CPUCost:     cpuCost,
		MemoryCost:  memoryCost,
		StorageCost: storageCost,
		NetworkCost: networkCost,
		TotalCost:   totalCost,
		Currency:    s.config.DefaultCurrency,
	}, nil
}

// ResourceUsage represents resource usage for cost calculation
type ResourceUsage struct {
	CPUCoreHours   float64
	MemoryGBHours  float64
	StorageGB      float64
	NetworkGB      float64
	Hours          float64
}

// CostResult represents cost calculation result
type CostResult struct {
	CPUCost     float64 `json:"cpu_cost"`
	MemoryCost  float64 `json:"memory_cost"`
	StorageCost float64 `json:"storage_cost"`
	NetworkCost float64 `json:"network_cost"`
	TotalCost   float64 `json:"total_cost"`
	Currency    string  `json:"currency"`
}

// GenerateReport generates a cost report
func (s *Service) GenerateReport(ctx context.Context, req ReportRequest) (*CostReport, error) {
	// Get cost allocations for the period
	filter := CostAllocationFilter{
		ClusterID: req.ClusterID,
		Namespace: req.Namespace,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     10000,
	}

	allocations, err := s.GetCostAllocation(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	var totalCost, cpuCost, memoryCost, storageCost, networkCost float64
	for _, alloc := range allocations {
		totalCost += alloc.TotalCost
		cpuCost += alloc.CPUCost
		memoryCost += alloc.MemoryCost
		storageCost += alloc.StorageCost
		networkCost += alloc.NetworkCost
	}

	// Generate breakdown
	breakdown := s.generateBreakdown(allocations, req.Grouping)

	// Generate recommendations
	recommendations := s.generateRecommendations(allocations)

	// Calculate trends
	trends, err := s.calculateTrends(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to calculate trends", zap.Error(err))
	}

	report := &CostReport{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Type:            req.Type,
		Grouping:        req.Grouping,
		Filters:         req.Filters,
		PeriodStart:     req.StartTime,
		PeriodEnd:       req.EndTime,
		TotalCost:       totalCost,
		CPUCost:         cpuCost,
		MemoryCost:      memoryCost,
		StorageCost:     storageCost,
		NetworkCost:     networkCost,
		Breakdown:       breakdown,
		Trends:          trends,
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
		CreatedBy:       req.UserID,
	}

	// Save report
	if err := s.db.Create(report).Error; err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// ReportRequest represents a request to generate a cost report
type ReportRequest struct {
	Name      string
	Type      string // daily, weekly, monthly, custom
	Grouping  []string
	Filters   map[string]interface{}
	ClusterID string
	Namespace string
	StartTime time.Time
	EndTime   time.Time
	UserID    string
}

func (s *Service) generateBreakdown(allocations []CostAllocation, grouping []string) []CostBreakdown {
	var breakdowns []CostBreakdown

	if len(grouping) == 0 {
		grouping = []string{"namespace"}
	}

	// Group by first grouping dimension
	groups := make(map[string]float64)
	for _, alloc := range allocations {
		var key string
		switch grouping[0] {
		case "namespace":
			key = alloc.Namespace
		case "cluster":
			key = alloc.ClusterName
		case "workload":
			key = alloc.WorkloadName
		default:
			key = "unknown"
		}
		groups[key] += alloc.TotalCost
	}

	// Calculate total for percentages
	var total float64
	for _, cost := range groups {
		total += cost
	}

	// Create breakdown entries
	for name, cost := range groups {
		percentage := 0.0
		if total > 0 {
			percentage = (cost / total) * 100
		}

		breakdowns = append(breakdowns, CostBreakdown{
			ID:         uuid.New().String(),
			Category:   grouping[0],
			Name:       name,
			Cost:       cost,
			Percentage: percentage,
		})
	}

	// Sort by cost descending
	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].Cost > breakdowns[j].Cost
	})

	return breakdowns
}

func (s *Service) generateRecommendations(allocations []CostAllocation) []CostRecommendation {
	var recommendations []CostRecommendation

	// Identify idle resources (efficiency < 10%)
	for _, alloc := range allocations {
		if alloc.Efficiency < 10 && alloc.TotalCost > 10 {
			recommendations = append(recommendations, CostRecommendation{
				ID:             uuid.New().String(),
				Type:           "idle",
				Category:       "waste",
				Title:          fmt.Sprintf("Idle workload: %s", alloc.WorkloadName),
				Description:    fmt.Sprintf("Workload %s in namespace %s has only %.1f%% efficiency", alloc.WorkloadName, alloc.Namespace, alloc.Efficiency),
				ResourceType:   alloc.WorkloadType,
				ResourceName:   alloc.WorkloadName,
				Namespace:      alloc.Namespace,
				ClusterID:      alloc.ClusterID,
				CurrentCost:    alloc.TotalCost,
				ProjectedCost:  0,
				MonthlySavings: alloc.TotalCost * 720 / 24, // Convert to monthly
				AnnualSavings:  alloc.TotalCost * 8760 / 24,
				Effort:         "low",
				Risk:           "medium",
				Status:         "pending",
				CreatedAt:      time.Now(),
			})
		}

		// Identify over-provisioned resources (efficiency < 30%)
		if alloc.Efficiency < 30 && alloc.Efficiency >= 10 && alloc.TotalCost > 5 {
			savings := alloc.TotalCost * (1 - alloc.Efficiency/100) * 0.5
			recommendations = append(recommendations, CostRecommendation{
				ID:             uuid.New().String(),
				Type:           "rightsize",
				Category:       "optimization",
				Title:          fmt.Sprintf("Rightsize workload: %s", alloc.WorkloadName),
				Description:    fmt.Sprintf("Workload %s is over-provisioned with only %.1f%% utilization", alloc.WorkloadName, alloc.Efficiency),
				ResourceType:   alloc.WorkloadType,
				ResourceName:   alloc.WorkloadName,
				Namespace:      alloc.Namespace,
				ClusterID:      alloc.ClusterID,
				CurrentCost:    alloc.TotalCost,
				ProjectedCost:  alloc.TotalCost - savings,
				MonthlySavings: savings * 720 / 24,
				AnnualSavings:  savings * 8760 / 24,
				Effort:         "medium",
				Risk:           "low",
				Status:         "pending",
				CreatedAt:      time.Now(),
			})
		}
	}

	return recommendations
}

func (s *Service) calculateTrends(ctx context.Context, req ReportRequest) ([]CostTrend, error) {
	var trends []CostTrend

	// Get historical data for trend calculation
	duration := req.EndTime.Sub(req.StartTime)
	intervals := 30 // Default to 30 data points
	intervalDuration := duration / time.Duration(intervals)

	for i := 0; i < intervals; i++ {
		start := req.StartTime.Add(time.Duration(i) * intervalDuration)
		end := start.Add(intervalDuration)

		var totalCost float64
		s.db.Model(&CostAllocation{}).
			Where("period_start >= ? AND period_end <= ?", start, end).
			Select("COALESCE(SUM(total_cost), 0)").
			Scan(&totalCost)

		var change float64
		if len(trends) > 0 && trends[len(trends)-1].Cost > 0 {
			change = ((totalCost - trends[len(trends)-1].Cost) / trends[len(trends)-1].Cost) * 100
		}

		trends = append(trends, CostTrend{
			Date:   start,
			Cost:   totalCost,
			Change: change,
		})
	}

	return trends, nil
}

// CreateBudget creates a new budget
func (s *Service) CreateBudget(ctx context.Context, budget *Budget) error {
	budget.ID = uuid.New().String()
	budget.Status = "on_track"
	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()

	if budget.AlertThresholds == nil {
		budget.AlertThresholds = []float64{50, 75, 90, 100}
	}

	if err := s.db.Create(budget).Error; err != nil {
		return fmt.Errorf("failed to create budget: %w", err)
	}

	return nil
}

// GetBudget retrieves a budget by ID
func (s *Service) GetBudget(ctx context.Context, budgetID string) (*Budget, error) {
	var budget Budget
	if err := s.db.Preload("Alerts").First(&budget, "id = ?", budgetID).Error; err != nil {
		return nil, fmt.Errorf("budget not found: %w", err)
	}

	// Update current spend
	budget.CurrentSpend = s.calculateCurrentSpend(ctx, &budget)
	budget.Status = s.determineBudgetStatus(&budget)

	return &budget, nil
}

// ListBudgets lists all budgets
func (s *Service) ListBudgets(ctx context.Context) ([]Budget, error) {
	var budgets []Budget
	if err := s.db.Preload("Alerts").Find(&budgets).Error; err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}

	// Update current spend for each budget
	for i := range budgets {
		budgets[i].CurrentSpend = s.calculateCurrentSpend(ctx, &budgets[i])
		budgets[i].Status = s.determineBudgetStatus(&budgets[i])
	}

	return budgets, nil
}

func (s *Service) calculateCurrentSpend(ctx context.Context, budget *Budget) float64 {
	var totalCost float64
	query := s.db.Model(&CostAllocation{}).
		Where("period_start >= ? AND period_end <= ?", budget.PeriodStart, budget.PeriodEnd)

	if budget.Scope == "cluster" {
		query = query.Where("cluster_id = ?", budget.ScopeValue)
	} else if budget.Scope == "namespace" {
		query = query.Where("namespace = ?", budget.ScopeValue)
	}

	query.Select("COALESCE(SUM(total_cost), 0)").Scan(&totalCost)
	return totalCost
}

func (s *Service) determineBudgetStatus(budget *Budget) string {
	percentage := (budget.CurrentSpend / budget.Amount) * 100

	if percentage >= 100 {
		return "exceeded"
	} else if percentage >= 75 {
		return "warning"
	}
	return "on_track"
}

// CheckBudgetAlerts checks and creates alerts for budgets
func (s *Service) CheckBudgetAlerts(ctx context.Context) error {
	budgets, err := s.ListBudgets(ctx)
	if err != nil {
		return err
	}

	for _, budget := range budgets {
		percentage := (budget.CurrentSpend / budget.Amount) * 100

		for _, threshold := range budget.AlertThresholds {
			if percentage >= threshold {
				// Check if alert already exists
				var existingAlert BudgetAlert
				err := s.db.Where("budget_id = ? AND threshold = ?", budget.ID, threshold).First(&existingAlert).Error
				if err == gorm.ErrRecordNotFound {
					// Create new alert
					severity := "warning"
					if threshold >= 100 {
						severity = "critical"
					}

					alert := &BudgetAlert{
						ID:           uuid.New().String(),
						BudgetID:     budget.ID,
						Threshold:    threshold,
						CurrentSpend: budget.CurrentSpend,
						Message:      fmt.Sprintf("Budget %s has reached %.0f%% (%.2f of %.2f %s)", budget.Name, percentage, budget.CurrentSpend, budget.Amount, budget.Currency),
						Severity:     severity,
						Notified:     false,
						CreatedAt:    time.Now(),
					}

					if err := s.db.Create(alert).Error; err != nil {
						s.logger.Error("Failed to create budget alert", zap.Error(err))
					}
				}
			}
		}
	}

	return nil
}

// GenerateForecast generates a cost forecast
func (s *Service) GenerateForecast(ctx context.Context, scope, scopeValue string, days int) (*CostForecast, error) {
	// Get historical cost data
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -90) // Use 90 days of history

	filter := CostAllocationFilter{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     10000,
	}

	if scope == "cluster" {
		filter.ClusterID = scopeValue
	} else if scope == "namespace" {
		filter.Namespace = scopeValue
	}

	allocations, err := s.GetCostAllocation(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Calculate daily costs
	dailyCosts := make(map[string]float64)
	for _, alloc := range allocations {
		day := alloc.PeriodStart.Format("2006-01-02")
		dailyCosts[day] += alloc.TotalCost
	}

	// Simple linear regression for forecasting
	var costs []float64
	for _, cost := range dailyCosts {
		costs = append(costs, cost)
	}

	if len(costs) < 7 {
		return nil, fmt.Errorf("insufficient data for forecasting")
	}

	// Calculate trend
	avgCost := average(costs)
	trend := s.calculateTrendSlope(costs)

	// Generate predictions
	var predictions []ForecastPrediction
	for i := 1; i <= days; i++ {
		date := endTime.AddDate(0, 0, i)
		predicted := avgCost + trend*float64(i)
		if predicted < 0 {
			predicted = 0
		}

		// Add confidence interval (±20%)
		predictions = append(predictions, ForecastPrediction{
			Date:       date,
			LowerBound: predicted * 0.8,
			Predicted:  predicted,
			UpperBound: predicted * 1.2,
		})
	}

	// Calculate total forecast cost
	var totalForecast float64
	for _, p := range predictions {
		totalForecast += p.Predicted
	}

	forecast := &CostForecast{
		ID:           uuid.New().String(),
		Scope:        scope,
		ScopeValue:   scopeValue,
		ForecastDate: endTime.AddDate(0, 0, days),
		CurrentCost:  sum(costs),
		ForecastCost: totalForecast,
		Confidence:   0.75, // 75% confidence for simple linear model
		Model:        "linear",
		Predictions:  predictions,
		CreatedAt:    time.Now(),
	}

	// Save forecast
	if err := s.db.Create(forecast).Error; err != nil {
		return nil, fmt.Errorf("failed to save forecast: %w", err)
	}

	return forecast, nil
}

func (s *Service) calculateTrendSlope(costs []float64) float64 {
	n := float64(len(costs))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, cost := range costs {
		x := float64(i)
		sumX += x
		sumY += cost
		sumXY += x * cost
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

// GenerateRightsizingRecommendations generates rightsizing recommendations
func (s *Service) GenerateRightsizingRecommendations(ctx context.Context, clusterID string, metrics map[string]interface{}) ([]RightsizingRecommendation, error) {
	var recommendations []RightsizingRecommendation

	// Get current allocations
	filter := CostAllocationFilter{
		ClusterID: clusterID,
		StartTime: time.Now().AddDate(0, 0, -7), // Last 7 days
		EndTime:   time.Now(),
		Limit:     1000,
	}

	allocations, err := s.GetCostAllocation(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, alloc := range allocations {
		// Skip if efficiency is already good
		if alloc.Efficiency > 60 {
			continue
		}

		// Calculate recommended resources based on usage
		cpuRecommended := alloc.CPUCoreHours / 168 * 1.2 // Add 20% buffer
		memRecommended := alloc.MemoryGBHours / 168 * 1.2

		// Calculate savings
		currentCost := alloc.TotalCost
		projectedCost := currentCost * (alloc.Efficiency / 100) * 1.3 // Target 70% efficiency with buffer
		savings := currentCost - projectedCost

		if savings > 1 { // Only recommend if savings > $1
			rec := RightsizingRecommendation{
				ID:                    uuid.New().String(),
				ClusterID:             clusterID,
				Namespace:             alloc.Namespace,
				WorkloadType:          alloc.WorkloadType,
				WorkloadName:          alloc.WorkloadName,
				ContainerName:         alloc.ContainerName,
				RecommendedCPURequest: fmt.Sprintf("%.0fm", cpuRecommended*1000),
				RecommendedMemRequest: fmt.Sprintf("%.0fMi", memRecommended*1024),
				MonthlySavings:        savings * 30,
				Confidence:            math.Min(alloc.Efficiency/100+0.5, 0.9),
				Status:                "pending",
				CreatedAt:             time.Now(),
			}

			recommendations = append(recommendations, rec)

			// Save recommendation
			if err := s.db.Create(&rec).Error; err != nil {
				s.logger.Warn("Failed to save rightsizing recommendation", zap.Error(err))
			}
		}
	}

	// Sort by savings
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].MonthlySavings > recommendations[j].MonthlySavings
	})

	return recommendations, nil
}

// GetCostSummary returns a cost summary dashboard
func (s *Service) GetCostSummary(ctx context.Context) (*CostSummary, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfPrevMonth := startOfMonth.AddDate(0, -1, 0)

	// Current month cost
	var currentMonthCost float64
	s.db.Model(&CostAllocation{}).
		Where("period_start >= ?", startOfMonth).
		Select("COALESCE(SUM(total_cost), 0)").
		Scan(&currentMonthCost)

	// Previous month cost
	var prevMonthCost float64
	s.db.Model(&CostAllocation{}).
		Where("period_start >= ? AND period_end < ?", startOfPrevMonth, startOfMonth).
		Select("COALESCE(SUM(total_cost), 0)").
		Scan(&prevMonthCost)

	// Potential savings
	var potentialSavings float64
	s.db.Model(&CostRecommendation{}).
		Where("status = ?", "pending").
		Select("COALESCE(SUM(monthly_savings), 0)").
		Scan(&potentialSavings)

	// Top namespaces
	var topNamespaces []CostBreakdown
	s.db.Model(&CostAllocation{}).
		Select("namespace as name, SUM(total_cost) as cost").
		Where("period_start >= ?", startOfMonth).
		Group("namespace").
		Order("cost DESC").
		Limit(5).
		Scan(&topNamespaces)

	// Calculate change
	var changePercent float64
	if prevMonthCost > 0 {
		changePercent = ((currentMonthCost - prevMonthCost) / prevMonthCost) * 100
	}

	summary := &CostSummary{
		CurrentMonthCost:  currentMonthCost,
		PreviousMonthCost: prevMonthCost,
		ChangePercent:     changePercent,
		PotentialSavings:  potentialSavings,
		TopNamespaces:     topNamespaces,
		Currency:          s.config.DefaultCurrency,
		GeneratedAt:       now,
	}

	return summary, nil
}

// CostSummary represents a cost dashboard summary
type CostSummary struct {
	CurrentMonthCost  float64         `json:"current_month_cost"`
	PreviousMonthCost float64         `json:"previous_month_cost"`
	ChangePercent     float64         `json:"change_percent"`
	PotentialSavings  float64         `json:"potential_savings"`
	TopNamespaces     []CostBreakdown `json:"top_namespaces"`
	Currency          string          `json:"currency"`
	GeneratedAt       time.Time       `json:"generated_at"`
}

// ExportReport exports a cost report in various formats
func (s *Service) ExportReport(ctx context.Context, reportID, format string) ([]byte, error) {
	var report CostReport
	if err := s.db.Preload("Breakdown").Preload("Recommendations").First(&report, "id = ?", reportID).Error; err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}

	switch format {
	case "json":
		return json.MarshalIndent(report, "", "  ")
	case "csv":
		return s.exportToCSV(&report)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (s *Service) exportToCSV(report *CostReport) ([]byte, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("Category,Name,Cost,Percentage\n")

	// Data
	for _, b := range report.Breakdown {
		sb.WriteString(fmt.Sprintf("%s,%s,%.2f,%.2f%%\n", b.Category, b.Name, b.Cost, b.Percentage))
	}

	sb.WriteString(fmt.Sprintf("\nTotal,,%%.2f,100%%\n", report.TotalCost))

	return []byte(sb.String()), nil
}

// Helper functions
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sum(values) / float64(len(values))
}

func sum(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}

