package api

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/user/fb-ads/pkg/utils"
)

// StorageType defines how campaign metrics are stored
type StorageType string

const (
	// StorageTypeFile stores metrics in JSON files
	StorageTypeFile StorageType = "file"
	// StorageTypeMemory stores metrics in memory only
	StorageTypeMemory StorageType = "memory"
	// Default directory for storing statistics
	DefaultStatsDir = "stats"
)

// StatisticsManager handles the storage, analysis, and retrieval of campaign performance statistics
type StatisticsManager struct {
	metricsCollector *MetricsCollector
	storageType      StorageType
	storageDir       string
	memoryStore      map[string][]utils.CampaignPerformance
	mu               sync.RWMutex
}

// StatisticsTrend represents a trend in a specific metric over time
type StatisticsTrend struct {
	Metric     string      `json:"metric"`
	Values     []float64   `json:"values"`
	Timestamps []time.Time `json:"timestamps"`
	AvgValue   float64     `json:"avg_value"`
	MinValue   float64     `json:"min_value"`
	MaxValue   float64     `json:"max_value"`
	StdDev     float64     `json:"std_dev"`
	Change     float64     `json:"change"` // Percentage change from first to last value
}

// AggregateStatistics represents aggregated statistics across multiple campaigns
type AggregateStatistics struct {
	StartDate       time.Time                  `json:"start_date"`
	EndDate         time.Time                  `json:"end_date"`
	TotalSpend      float64                    `json:"total_spend"`
	TotalImpressions int                       `json:"total_impressions"`
	TotalClicks     int                        `json:"total_clicks"`
	TotalConversions int                       `json:"total_conversions"`
	AvgCTR          float64                    `json:"avg_ctr"`
	AvgCPM          float64                    `json:"avg_cpm"`
	AvgCPC          float64                    `json:"avg_cpc"`
	AvgCPA          float64                    `json:"avg_cpa"`
	TrendImpressions *StatisticsTrend          `json:"trend_impressions,omitempty"`
	TrendClicks      *StatisticsTrend          `json:"trend_clicks,omitempty"`
	TrendCTR         *StatisticsTrend          `json:"trend_ctr,omitempty"`
	TrendCPM         *StatisticsTrend          `json:"trend_cpm,omitempty"`
	TrendSpend       *StatisticsTrend          `json:"trend_spend,omitempty"`
	TrendConversions *StatisticsTrend          `json:"trend_conversions,omitempty"`
	CampaignStats    map[string]CampaignStats  `json:"campaign_stats,omitempty"`
}

// CampaignStats represents statistics for a single campaign
type CampaignStats struct {
	CampaignID      string    `json:"campaign_id"`
	Name            string    `json:"name"`
	FirstDataPoint  time.Time `json:"first_data_point"`
	LastDataPoint   time.Time `json:"last_data_point"`
	NumDataPoints   int       `json:"num_data_points"`
	TotalSpend      float64   `json:"total_spend"`
	TotalImpressions int      `json:"total_impressions"`
	TotalClicks     int       `json:"total_clicks"`
	TotalConversions int      `json:"total_conversions"`
	AvgCTR          float64   `json:"avg_ctr"`
	AvgCPM          float64   `json:"avg_cpm"`
	AvgCPC          float64   `json:"avg_cpc"`
	AvgCPA          float64   `json:"avg_cpa"`
	MinCPM          float64   `json:"min_cpm"`
	MaxCPM          float64   `json:"max_cpm"`
	ROI             float64   `json:"roi"`
}

// NewStatisticsManager creates a new statistics manager
func NewStatisticsManager(metricsCollector *MetricsCollector, storageType StorageType, storageDir string) *StatisticsManager {
	if storageDir == "" {
		storageDir = DefaultStatsDir
	}

	return &StatisticsManager{
		metricsCollector: metricsCollector,
		storageType:      storageType,
		storageDir:       storageDir,
		memoryStore:      make(map[string][]utils.CampaignPerformance),
		mu:               sync.RWMutex{},
	}
}

// CollectAndStoreStatistics collects statistics for the given time range and stores them
func (s *StatisticsManager) CollectAndStoreStatistics(timeRange TimeRange) error {
	// Collect metrics
	performances, err := s.metricsCollector.CollectCampaignMetrics(InsightsRequest{
		Level:     "campaign",
		TimeRange: timeRange,
	})
	if err != nil {
		return fmt.Errorf("error collecting metrics: %w", err)
	}

	// Store metrics
	return s.StoreStatistics(performances)
}

// StoreStatistics stores collected campaign performance data
func (s *StatisticsManager) StoreStatistics(performances []utils.CampaignPerformance) error {
	if len(performances) == 0 {
		return nil // No data to store
	}

	switch s.storageType {
	case StorageTypeFile:
		// Create date-based filename for today's statistics
		today := time.Now().Format("2006-01-02")
		dirPath := filepath.Join(s.storageDir, "daily")
		
		// Ensure directory exists
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("error creating statistics directory: %w", err)
		}
		
		// Create a file for each campaign to allow easier retrieval by campaign ID
		for _, perf := range performances {
			// Use campaign ID in filename for easy lookup
			filename := fmt.Sprintf("%s_%s.json", perf.CampaignID, today)
			filePath := filepath.Join(dirPath, filename)
			
			// Write performance data to file
			data, err := json.MarshalIndent(perf, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshaling performance data: %w", err)
			}
			
			if err := os.WriteFile(filePath, data, 0644); err != nil {
				return fmt.Errorf("error writing performance data to file: %w", err)
			}
		}
		
		// Also store aggregated data for the day
		aggregatedFilename := fmt.Sprintf("aggregated_%s.json", today)
		aggregatedFilePath := filepath.Join(dirPath, aggregatedFilename)
		
		// Marshal to JSON
		aggregatedData, err := json.MarshalIndent(performances, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling aggregated performance data: %w", err)
		}
		
		// Write to file
		if err := os.WriteFile(aggregatedFilePath, aggregatedData, 0644); err != nil {
			return fmt.Errorf("error writing aggregated performance data to file: %w", err)
		}
		
	case StorageTypeMemory:
		// Store in memory by campaign ID
		s.mu.Lock()
		defer s.mu.Unlock()
		
		for _, perf := range performances {
			s.memoryStore[perf.CampaignID] = append(s.memoryStore[perf.CampaignID], perf)
		}
	}
	
	return nil
}

// GetCampaignStatistics retrieves statistics for a specific campaign for the given time range
func (s *StatisticsManager) GetCampaignStatistics(campaignID string, startDate, endDate time.Time) ([]utils.CampaignPerformance, error) {
	var performances []utils.CampaignPerformance
	
	switch s.storageType {
	case StorageTypeFile:
		// Get list of dates to check within the range
		var dates []string
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d.Format("2006-01-02"))
		}
		
		// For each date, check if there's a file for the campaign
		for _, date := range dates {
			filename := fmt.Sprintf("%s_%s.json", campaignID, date)
			filePath := filepath.Join(s.storageDir, "daily", filename)
			
			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				continue // Skip if file doesn't exist
			}
			
			// Read file content
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading performance data: %w", err)
			}
			
			// Unmarshal into a campaign performance object
			var perf utils.CampaignPerformance
			if err := json.Unmarshal(data, &perf); err != nil {
				return nil, fmt.Errorf("error unmarshaling performance data: %w", err)
			}
			
			performances = append(performances, perf)
		}
		
	case StorageTypeMemory:
		s.mu.RLock()
		defer s.mu.RUnlock()
		
		// Get stored performances for the campaign
		campaignPerfs, ok := s.memoryStore[campaignID]
		if !ok {
			return nil, nil // No data found for this campaign
		}
		
		// Filter by date range
		for _, perf := range campaignPerfs {
			if !perf.LastUpdated.Before(startDate) && !perf.LastUpdated.After(endDate) {
				performances = append(performances, perf)
			}
		}
	}
	
	return performances, nil
}

// GetAllCampaignStatistics retrieves statistics for all campaigns for the given time range
func (s *StatisticsManager) GetAllCampaignStatistics(startDate, endDate time.Time) (map[string][]utils.CampaignPerformance, error) {
	result := make(map[string][]utils.CampaignPerformance)
	
	switch s.storageType {
	case StorageTypeFile:
		// Get the daily directory listing
		dirPath := filepath.Join(s.storageDir, "daily")
		files, err := os.ReadDir(dirPath)
		if err != nil {
			if os.IsNotExist(err) {
				return result, nil // No data yet
			}
			return nil, fmt.Errorf("error reading statistics directory: %w", err)
		}
		
		// Process each file within the date range
		for _, file := range files {
			// Skip aggregated files
			if file.IsDir() || len(file.Name()) < 10 {
				continue
			}
			
			// Extract date from filename
			var fileDate time.Time
			var campaignID string
			
			// Parse date and campaign ID (format: campaignID_YYYY-MM-DD.json)
			parts := filepath.Base(file.Name())
			if len(parts) > 11 {
				// Extract date part (last 10 chars + .json)
				datePart := parts[len(parts)-15:len(parts)-5]
				fileDate, err = time.Parse("2006-01-02", datePart)
				if err != nil {
					continue // Skip files with invalid date format
				}
				
				// Extract campaign ID
				campaignID = parts[:len(parts)-16]
			}
			
			// Skip if outside date range
			if fileDate.Before(startDate) || fileDate.After(endDate) {
				continue
			}
			
			// Read file
			filePath := filepath.Join(dirPath, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading performance data: %w", err)
			}
			
			// Unmarshal into a campaign performance object
			var perf utils.CampaignPerformance
			if err := json.Unmarshal(data, &perf); err != nil {
				return nil, fmt.Errorf("error unmarshaling performance data: %w", err)
			}
			
			// Add to result
			result[campaignID] = append(result[campaignID], perf)
		}
		
	case StorageTypeMemory:
		s.mu.RLock()
		defer s.mu.RUnlock()
		
		// Copy from memory store, filtering by date range
		for campaignID, perfs := range s.memoryStore {
			var filteredPerfs []utils.CampaignPerformance
			
			for _, perf := range perfs {
				if !perf.LastUpdated.Before(startDate) && !perf.LastUpdated.After(endDate) {
					filteredPerfs = append(filteredPerfs, perf)
				}
			}
			
			if len(filteredPerfs) > 0 {
				result[campaignID] = filteredPerfs
			}
		}
	}
	
	return result, nil
}

// AnalyzeStatistics performs statistical analysis on campaign performance data
func (s *StatisticsManager) AnalyzeStatistics(startDate, endDate time.Time) (*AggregateStatistics, error) {
	// Get all campaign statistics for the date range
	allStats, err := s.GetAllCampaignStatistics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error retrieving campaign statistics: %w", err)
	}
	
	// Initialize aggregate statistics
	stats := &AggregateStatistics{
		StartDate:       startDate,
		EndDate:         endDate,
		CampaignStats:   make(map[string]CampaignStats),
	}
	
	// Variables for trend analysis
	allImpressions := make(map[time.Time]int)
	allClicks := make(map[time.Time]int)
	allSpend := make(map[time.Time]float64)
	allCTR := make(map[time.Time]float64)
	allCPM := make(map[time.Time]float64)
	allConversions := make(map[time.Time]int)
	
	// Process each campaign's statistics
	for campaignID, performances := range allStats {
		// Initialize campaign statistics
		campaignStats := CampaignStats{
			CampaignID: campaignID,
			MinCPM:     math.MaxFloat64,
		}
		
		if len(performances) == 0 {
			continue
		}
		
		// Set campaign name from the first performance record
		campaignStats.Name = performances[0].Name
		
		// Track the earliest and latest data points
		campaignStats.FirstDataPoint = performances[0].LastUpdated
		campaignStats.LastDataPoint = performances[0].LastUpdated
		
		// Accumulate statistics across all performance records
		for _, perf := range performances {
			// Update first/last data points
			if perf.LastUpdated.Before(campaignStats.FirstDataPoint) {
				campaignStats.FirstDataPoint = perf.LastUpdated
			}
			if perf.LastUpdated.After(campaignStats.LastDataPoint) {
				campaignStats.LastDataPoint = perf.LastUpdated
			}
			
			// Accumulate metrics
			campaignStats.TotalSpend += perf.Spend
			campaignStats.TotalImpressions += perf.Impressions
			campaignStats.TotalClicks += perf.Clicks
			campaignStats.TotalConversions += perf.Conversions
			campaignStats.NumDataPoints++
			
			// Track min/max CPM
			if perf.CPM < campaignStats.MinCPM {
				campaignStats.MinCPM = perf.CPM
			}
			if perf.CPM > campaignStats.MaxCPM {
				campaignStats.MaxCPM = perf.CPM
			}
			
			// Aggregate for global trends
			day := time.Date(perf.LastUpdated.Year(), perf.LastUpdated.Month(), perf.LastUpdated.Day(), 0, 0, 0, 0, time.Local)
			allImpressions[day] += perf.Impressions
			allClicks[day] += perf.Clicks
			allSpend[day] += perf.Spend
			allConversions[day] += perf.Conversions
			
			// We'll calculate the daily averages later
			if _, ok := allCTR[day]; !ok {
				allCTR[day] = 0
				allCPM[day] = 0
			}
		}
		
		// Calculate averages
		if campaignStats.TotalClicks > 0 {
			campaignStats.AvgCPC = campaignStats.TotalSpend / float64(campaignStats.TotalClicks)
		}
		
		if campaignStats.TotalImpressions > 0 {
			campaignStats.AvgCTR = float64(campaignStats.TotalClicks) / float64(campaignStats.TotalImpressions) * 100
			campaignStats.AvgCPM = campaignStats.TotalSpend / float64(campaignStats.TotalImpressions) * 1000
		}
		
		if campaignStats.TotalConversions > 0 {
			campaignStats.AvgCPA = campaignStats.TotalSpend / float64(campaignStats.TotalConversions)
			// Calculate ROI - assuming $50 average order value per conversion
			avgOrderValue := 50.0
			campaignStats.ROI = (float64(campaignStats.TotalConversions) * avgOrderValue - campaignStats.TotalSpend) / campaignStats.TotalSpend * 100
		}
		
		// Add to total statistics
		stats.TotalSpend += campaignStats.TotalSpend
		stats.TotalImpressions += campaignStats.TotalImpressions
		stats.TotalClicks += campaignStats.TotalClicks
		stats.TotalConversions += campaignStats.TotalConversions
		
		// Add to campaign-specific stats
		stats.CampaignStats[campaignID] = campaignStats
	}
	
	// Calculate global averages
	if stats.TotalClicks > 0 {
		stats.AvgCPC = stats.TotalSpend / float64(stats.TotalClicks)
	}
	
	if stats.TotalImpressions > 0 {
		stats.AvgCTR = float64(stats.TotalClicks) / float64(stats.TotalImpressions) * 100
		stats.AvgCPM = stats.TotalSpend / float64(stats.TotalImpressions) * 1000
	}
	
	if stats.TotalConversions > 0 {
		stats.AvgCPA = stats.TotalSpend / float64(stats.TotalConversions)
	}
	
	// Calculate daily averages for CTR and CPM
	for day, impressions := range allImpressions {
		if impressions > 0 {
			clicks := allClicks[day]
			spend := allSpend[day]
			
			allCTR[day] = float64(clicks) / float64(impressions) * 100
			allCPM[day] = spend / float64(impressions) * 1000
		}
	}
	
	// Generate trend data
	dates := make([]time.Time, 0, len(allImpressions))
	for date := range allImpressions {
		dates = append(dates, date)
	}
	
	// Sort dates chronologically
	sortDates(dates)
	
	// Create trend data structures
	if len(dates) > 0 {
		stats.TrendImpressions = s.createTrend("impressions", dates, func(date time.Time) float64 {
			return float64(allImpressions[date])
		})
		
		stats.TrendClicks = s.createTrend("clicks", dates, func(date time.Time) float64 {
			return float64(allClicks[date])
		})
		
		stats.TrendCTR = s.createTrend("ctr", dates, func(date time.Time) float64 {
			return allCTR[date]
		})
		
		stats.TrendCPM = s.createTrend("cpm", dates, func(date time.Time) float64 {
			return allCPM[date]
		})
		
		stats.TrendSpend = s.createTrend("spend", dates, func(date time.Time) float64 {
			return allSpend[date]
		})
		
		stats.TrendConversions = s.createTrend("conversions", dates, func(date time.Time) float64 {
			return float64(allConversions[date])
		})
	}
	
	return stats, nil
}

// createTrend creates a trend analysis for a specific metric
func (s *StatisticsManager) createTrend(metricName string, dates []time.Time, valueFunc func(time.Time) float64) *StatisticsTrend {
	if len(dates) == 0 {
		return nil
	}
	
	trend := &StatisticsTrend{
		Metric:     metricName,
		Timestamps: dates,
		Values:     make([]float64, len(dates)),
		MinValue:   math.MaxFloat64,
		MaxValue:   -math.MaxFloat64,
	}
	
	// Populate values
	sum := 0.0
	for i, date := range dates {
		value := valueFunc(date)
		trend.Values[i] = value
		sum += value
		
		if value < trend.MinValue {
			trend.MinValue = value
		}
		if value > trend.MaxValue {
			trend.MaxValue = value
		}
	}
	
	// Calculate average
	trend.AvgValue = sum / float64(len(dates))
	
	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, value := range trend.Values {
		diff := value - trend.AvgValue
		sumSquaredDiff += diff * diff
	}
	trend.StdDev = math.Sqrt(sumSquaredDiff / float64(len(dates)))
	
	// Calculate change percentage (if at least 2 data points)
	if len(trend.Values) >= 2 {
		firstValue := trend.Values[0]
		lastValue := trend.Values[len(trend.Values)-1]
		
		if firstValue != 0 {
			trend.Change = (lastValue - firstValue) / firstValue * 100
		}
	}
	
	return trend
}

// sortDates sorts dates in ascending order
func sortDates(dates []time.Time) {
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[j].Before(dates[i]) {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}
}

// ExportStatisticsCSV exports campaign statistics to a CSV file
func (s *StatisticsManager) ExportStatisticsCSV(stats *AggregateStatistics, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}
	
	// Create CSV file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %w", err)
	}
	defer file.Close()
	
	// Write header
	header := "Campaign ID,Campaign Name,Impressions,Clicks,CTR (%),Spend ($),CPM ($),CPC ($),Conversions,CPA ($),ROI (%)\n"
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("error writing CSV header: %w", err)
	}
	
	// Write campaign data
	for _, campaign := range stats.CampaignStats {
		line := fmt.Sprintf(
			"%s,%s,%d,%d,%.2f,%.2f,%.2f,%.2f,%d,%.2f,%.2f\n",
			campaign.CampaignID,
			escapeCsvField(campaign.Name),
			campaign.TotalImpressions,
			campaign.TotalClicks,
			campaign.AvgCTR,
			campaign.TotalSpend,
			campaign.AvgCPM,
			campaign.AvgCPC,
			campaign.TotalConversions,
			campaign.AvgCPA,
			campaign.ROI,
		)
		
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("error writing CSV line: %w", err)
		}
	}
	
	// Write totals
	totalsLine := fmt.Sprintf(
		"TOTAL,All Campaigns,%d,%d,%.2f,%.2f,%.2f,%.2f,%d,%.2f,\n",
		stats.TotalImpressions,
		stats.TotalClicks,
		stats.AvgCTR,
		stats.TotalSpend,
		stats.AvgCPM,
		stats.AvgCPC,
		stats.TotalConversions,
		stats.AvgCPA,
	)
	
	if _, err := file.WriteString("\n" + totalsLine); err != nil {
		return fmt.Errorf("error writing CSV totals: %w", err)
	}
	
	return nil
}

// Escape CSV field to handle commas and quotes
func escapeCsvField(field string) string {
	needsQuotes := false
	for i := 0; i < len(field); i++ {
		if field[i] == '"' || field[i] == ',' || field[i] == '\n' || field[i] == '\r' {
			needsQuotes = true
			break
		}
	}
	
	if !needsQuotes {
		return field
	}
	
	// Replace double quotes with two double quotes and wrap in quotes
	result := `"`
	for i := 0; i < len(field); i++ {
		if field[i] == '"' {
			result += "\"\""
		} else {
			result += string(field[i])
		}
	}
	result += `"`
	
	return result
}