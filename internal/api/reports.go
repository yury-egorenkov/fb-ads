package api

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReportGenerator handles generation of various reports
type ReportGenerator struct {
	analyzer         *PerformanceAnalyzer
	metricsCollector *MetricsCollector
	outputDir        string
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(analyzer *PerformanceAnalyzer, metricsCollector *MetricsCollector, outputDir string) *ReportGenerator {
	return &ReportGenerator{
		analyzer:         analyzer,
		metricsCollector: metricsCollector,
		outputDir:        outputDir,
	}
}

// GenerateDailyReport generates a daily performance report
func (r *ReportGenerator) GenerateDailyReport() error {
	// Create time range for yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")

	timeRange := TimeRange{
		Since: yesterdayStr,
		Until: yesterdayStr,
	}

	// Generate analysis
	analysis, err := r.analyzer.AnalyzeCampaignPerformance(timeRange)
	if err != nil {
		return fmt.Errorf("error analyzing performance: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Generate report file name
	reportFileName := fmt.Sprintf("daily_report_%s.json", yesterdayStr)
	reportPath := filepath.Join(r.outputDir, reportFileName)

	// Save report
	return r.analyzer.GenerateReport(analysis, reportPath)
}

// GenerateWeeklyReport generates a weekly performance report
func (r *ReportGenerator) GenerateWeeklyReport() error {
	// Create time range for last week
	today := time.Now()
	endDate := today.AddDate(0, 0, -1)
	startDate := today.AddDate(0, 0, -7)

	timeRange := TimeRange{
		Since: startDate.Format("2006-01-02"),
		Until: endDate.Format("2006-01-02"),
	}

	// Generate analysis
	analysis, err := r.analyzer.AnalyzeCampaignPerformance(timeRange)
	if err != nil {
		return fmt.Errorf("error analyzing performance: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Generate report file name
	weekNum := int(today.Day()/7) + 1
	reportFileName := fmt.Sprintf("weekly_report_%s_week%d.json", today.Format("2006-01"), weekNum)
	reportPath := filepath.Join(r.outputDir, reportFileName)

	// Save report
	return r.analyzer.GenerateReport(analysis, reportPath)
}

// GenerateCustomReport generates a custom date range report
func (r *ReportGenerator) GenerateCustomReport(startDate, endDate time.Time) error {
	timeRange := TimeRange{
		Since: startDate.Format("2006-01-02"),
		Until: endDate.Format("2006-01-02"),
	}

	// Generate analysis
	analysis, err := r.analyzer.AnalyzeCampaignPerformance(timeRange)
	if err != nil {
		return fmt.Errorf("error analyzing performance: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Generate report file name
	reportFileName := fmt.Sprintf("custom_report_%s_to_%s.json",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))
	reportPath := filepath.Join(r.outputDir, reportFileName)

	// Save report
	return r.analyzer.GenerateReport(analysis, reportPath)
}

// GenerateAudienceInsightsReport generates a report on audience insights
func (r *ReportGenerator) GenerateAudienceInsightsReport() error {
	// TODO: Implement audience insights report
	return nil
}

// ExportReportCSV exports a performance analysis as CSV
func (r *ReportGenerator) ExportReportCSV(analysis *PerformanceAnalysis, filePath string) error {
	// TODO: Implement CSV export
	return nil
}

// ExportReportHTML generates an HTML report from a performance analysis
func (r *ReportGenerator) ExportReportHTML(analysis *PerformanceAnalysis, filePath string) error {
	// TODO: Implement HTML report generation
	return nil
}
