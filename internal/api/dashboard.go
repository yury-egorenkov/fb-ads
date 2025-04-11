package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/user/fb-ads/pkg/utils"
)

// DashboardData represents the data model for the dashboard
type DashboardData struct {
	Title             string                       `json:"title"`
	GeneratedAt       time.Time                    `json:"generated_at"`
	Summary           DashboardSummary             `json:"summary"`
	TopCampaigns      []utils.CampaignPerformance  `json:"top_campaigns"`
	WorstCampaigns    []utils.CampaignPerformance  `json:"worst_campaigns"`
	PerformanceByDay  []DailyPerformance           `json:"performance_by_day"`
	Recommendations   []string                     `json:"recommendations"`
}

// DashboardSummary contains summary metrics for the dashboard
type DashboardSummary struct {
	TotalCampaigns   int     `json:"total_campaigns"`
	ActiveCampaigns  int     `json:"active_campaigns"`
	TotalSpend       float64 `json:"total_spend"`
	TotalImpressions int     `json:"total_impressions"`
	TotalClicks      int     `json:"total_clicks"`
	TotalConversions int     `json:"total_conversions"`
	AverageCTR       float64 `json:"average_ctr"`
	AverageCPM       float64 `json:"average_cpm"`
	AverageCPA       float64 `json:"average_cpa"`
	AverageROAS      float64 `json:"average_roas"`
}

// DailyPerformance represents performance data for a single day
type DailyPerformance struct {
	Date         string  `json:"date"`
	Spend        float64 `json:"spend"`
	Impressions  int     `json:"impressions"`
	Clicks       int     `json:"clicks"`
	Conversions  int     `json:"conversions"`
	CTR          float64 `json:"ctr"`
	CPC          float64 `json:"cpc"`
	CPM          float64 `json:"cpm"`
	CPA          float64 `json:"cpa"`
	ROAS         float64 `json:"roas"`
}

// Dashboard handles the web dashboard for visualizing campaign performance
type Dashboard struct {
	metricsCollector *MetricsCollector
	analyzer         *PerformanceAnalyzer
	port             int
	templateDir      string
	dataDir          string
}

// NewDashboard creates a new dashboard
func NewDashboard(metricsCollector *MetricsCollector, analyzer *PerformanceAnalyzer, port int, templateDir, dataDir string) *Dashboard {
	return &Dashboard{
		metricsCollector: metricsCollector,
		analyzer:         analyzer,
		port:             port,
		templateDir:      templateDir,
		dataDir:          dataDir,
	}
}

// Start starts the dashboard web server
func (d *Dashboard) Start() error {
	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(d.dataDir, 0755); err != nil {
		return fmt.Errorf("error creating data directory: %w", err)
	}

	// Set up HTTP routes
	http.HandleFunc("/", d.handleHome)
	http.HandleFunc("/api/dashboard", d.handleDashboardData)
	http.HandleFunc("/api/campaigns", d.handleCampaigns)
	http.HandleFunc("/api/performance", d.handlePerformance)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(d.templateDir, "static")))))

	// Start the server
	addr := fmt.Sprintf(":%d", d.port)
	fmt.Printf("Dashboard starting on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleHome handles the dashboard home page
func (d *Dashboard) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Parse the template
	tmpl, err := template.ParseFiles(filepath.Join(d.templateDir, "dashboard.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	// Execute the template
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleDashboardData handles API requests for dashboard data
func (d *Dashboard) handleDashboardData(w http.ResponseWriter, r *http.Request) {
	// Get the dashboard data
	data, err := d.generateDashboardData()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating dashboard data: %v", err), http.StatusInternalServerError)
		return
	}

	// Set the content type
	w.Header().Set("Content-Type", "application/json")

	// Encode the data as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleCampaigns handles API requests for campaign data
func (d *Dashboard) handleCampaigns(w http.ResponseWriter, r *http.Request) {
	// Create time range for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	timeRange := TimeRange{
		Since: startDate.Format("2006-01-02"),
		Until: endDate.Format("2006-01-02"),
	}

	// Generate an analysis to get campaign data
	analysis, err := d.analyzer.AnalyzeCampaignPerformance(timeRange)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error analyzing performance: %v", err), http.StatusInternalServerError)
		return
	}

	// Set the content type
	w.Header().Set("Content-Type", "application/json")

	// Encode the data as JSON
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// handlePerformance handles API requests for daily performance data
func (d *Dashboard) handlePerformance(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters
	days := 30
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// Get the performance data
	data, err := d.generateDailyPerformanceData(days)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating performance data: %v", err), http.StatusInternalServerError)
		return
	}

	// Set the content type
	w.Header().Set("Content-Type", "application/json")

	// Encode the data as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// generateDashboardData generates data for the dashboard
func (d *Dashboard) generateDashboardData() (*DashboardData, error) {
	// Create time range for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	timeRange := TimeRange{
		Since: startDate.Format("2006-01-02"),
		Until: endDate.Format("2006-01-02"),
	}

	// Generate an analysis
	analysis, err := d.analyzer.AnalyzeCampaignPerformance(timeRange)
	if err != nil {
		return nil, fmt.Errorf("error analyzing performance: %w", err)
	}

	// Get daily performance data
	dailyPerformance, err := d.generateDailyPerformanceData(30)
	if err != nil {
		return nil, fmt.Errorf("error generating daily performance data: %w", err)
	}

	// Create the dashboard data
	dashboardData := &DashboardData{
		Title:             "Facebook Ads Performance Dashboard",
		GeneratedAt:       time.Now(),
		TopCampaigns:      analysis.TopCampaigns,
		WorstCampaigns:    analysis.WorstCampaigns,
		PerformanceByDay:  dailyPerformance,
		Recommendations:   analysis.Recommendations,
	}

	// Calculate summary metrics
	dashboardData.Summary = DashboardSummary{
		TotalCampaigns:   len(analysis.TopCampaigns) + len(analysis.WorstCampaigns),
		ActiveCampaigns:  0, // To be calculated
		TotalSpend:       analysis.TotalSpend,
		TotalImpressions: analysis.TotalImpressions,
		TotalClicks:      analysis.TotalClicks,
		TotalConversions: analysis.TotalConversions,
		AverageCTR:       analysis.AverageCTR,
		AverageCPM:       0, // To be calculated
		AverageCPA:       analysis.AverageCPA,
		AverageROAS:      analysis.AverageROAS,
	}

	// Save the dashboard data to a file
	dataFile := filepath.Join(d.dataDir, "dashboard_data.json")
	data, err := json.MarshalIndent(dashboardData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling dashboard data: %w", err)
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		return nil, fmt.Errorf("error writing dashboard data: %w", err)
	}

	return dashboardData, nil
}

// generateDailyPerformanceData generates daily performance data for the specified number of days
func (d *Dashboard) generateDailyPerformanceData(days int) ([]DailyPerformance, error) {
	// In a real implementation, this would query the Facebook API for daily performance data
	// For now, we'll generate some sample data
	var result []DailyPerformance

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Check if we have cached data
	cacheFile := filepath.Join(d.dataDir, fmt.Sprintf("daily_performance_%d.json", days))
	if data, err := os.ReadFile(cacheFile); err == nil {
		// Parse the cached data
		if err := json.Unmarshal(data, &result); err == nil {
			return result, nil
		}
	}

	// Generate sample data
	date := startDate
	for date.Before(endDate) || date.Equal(endDate) {
		// Generate random but somewhat realistic metrics
		// In a real implementation, this would come from the Facebook API
		spend := 50.0 + float64(date.Day()%10)*10.0
		impressions := 5000 + date.Day()*200
		clicks := 100 + date.Day()*5
		conversions := 2 + date.Day()%5
		ctr := float64(clicks) / float64(impressions) * 100
		cpc := spend / float64(clicks)
		cpm := spend / float64(impressions) * 1000
		cpa := spend / float64(conversions)
		roas := float64(conversions) * 50.0 / spend

		// Create the daily performance
		performance := DailyPerformance{
			Date:         date.Format("2006-01-02"),
			Spend:        spend,
			Impressions:  impressions,
			Clicks:       clicks,
			Conversions:  conversions,
			CTR:          ctr,
			CPC:          cpc,
			CPM:          cpm,
			CPA:          cpa,
			ROAS:         roas,
		}

		result = append(result, performance)
		date = date.AddDate(0, 0, 1)
	}

	// Cache the data
	data, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		_ = os.WriteFile(cacheFile, data, 0644)
	}

	return result, nil
}

// CreateDashboardFiles creates the necessary files for the dashboard
func (d *Dashboard) CreateDashboardFiles() error {
	// Create the template directory if it doesn't exist
	if err := os.MkdirAll(d.templateDir, 0755); err != nil {
		return fmt.Errorf("error creating template directory: %w", err)
	}

	// Create the static directory
	staticDir := filepath.Join(d.templateDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("error creating static directory: %w", err)
	}

	// Create the CSS directory
	cssDir := filepath.Join(staticDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("error creating CSS directory: %w", err)
	}

	// Create the JS directory
	jsDir := filepath.Join(staticDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return fmt.Errorf("error creating JS directory: %w", err)
	}

	// Create the dashboard HTML template
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Facebook Ads Performance Dashboard</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <header>
        <h1>Facebook Ads Performance Dashboard</h1>
        <p id="updated">Last updated: <span id="last-updated"></span></p>
    </header>
    
    <main>
        <section class="summary-section">
            <h2>Performance Summary</h2>
            <div class="summary-grid">
                <div class="summary-card">
                    <h3>Spend</h3>
                    <p id="total-spend">$0.00</p>
                </div>
                <div class="summary-card">
                    <h3>Impressions</h3>
                    <p id="total-impressions">0</p>
                </div>
                <div class="summary-card">
                    <h3>Clicks</h3>
                    <p id="total-clicks">0</p>
                </div>
                <div class="summary-card">
                    <h3>Conversions</h3>
                    <p id="total-conversions">0</p>
                </div>
                <div class="summary-card">
                    <h3>CTR</h3>
                    <p id="average-ctr">0.00%</p>
                </div>
                <div class="summary-card">
                    <h3>CPA</h3>
                    <p id="average-cpa">$0.00</p>
                </div>
                <div class="summary-card">
                    <h3>ROAS</h3>
                    <p id="average-roas">0.0x</p>
                </div>
                <div class="summary-card">
                    <h3>Active Campaigns</h3>
                    <p id="active-campaigns">0</p>
                </div>
            </div>
        </section>
        
        <section class="chart-section">
            <h2>Performance Trends</h2>
            <div class="chart-container">
                <canvas id="performance-chart"></canvas>
            </div>
        </section>
        
        <div class="dashboard-grid">
            <section class="top-campaigns-section">
                <h2>Top Performing Campaigns</h2>
                <table id="top-campaigns-table">
                    <thead>
                        <tr>
                            <th>Campaign</th>
                            <th>Spend</th>
                            <th>Conv.</th>
                            <th>CPA</th>
                            <th>ROAS</th>
                        </tr>
                    </thead>
                    <tbody id="top-campaigns-body">
                        <!-- Will be populated by JavaScript -->
                    </tbody>
                </table>
            </section>
            
            <section class="recommendations-section">
                <h2>Recommendations</h2>
                <ul id="recommendations-list">
                    <!-- Will be populated by JavaScript -->
                </ul>
            </section>
        </div>
    </main>
    
    <script src="/static/js/dashboard.js"></script>
</body>
</html>`

	if err := os.WriteFile(filepath.Join(d.templateDir, "dashboard.html"), []byte(htmlTemplate), 0644); err != nil {
		return fmt.Errorf("error writing dashboard HTML template: %w", err)
	}

	// Create the CSS file
	cssContent := `/* Reset and base styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f8f9fa;
    padding: 20px;
}

header {
    margin-bottom: 30px;
}

h1 {
    color: #1877f2;
    margin-bottom: 10px;
}

#updated {
    color: #666;
    font-size: 0.9rem;
}

/* Grid layouts */
.summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 20px;
    margin-bottom: 30px;
}

.dashboard-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 30px;
    margin-top: 30px;
}

@media (max-width: 768px) {
    .dashboard-grid {
        grid-template-columns: 1fr;
    }
}

/* Cards */
.summary-card {
    background-color: white;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
    text-align: center;
}

.summary-card h3 {
    font-size: 0.9rem;
    font-weight: 600;
    color: #666;
    margin-bottom: 10px;
}

.summary-card p {
    font-size: 1.5rem;
    font-weight: 700;
    color: #1877f2;
}

/* Sections */
section {
    background-color: white;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
    margin-bottom: 30px;
}

section h2 {
    color: #1877f2;
    margin-bottom: 20px;
    font-size: 1.2rem;
}

/* Chart */
.chart-container {
    height: 400px;
    width: 100%;
}

/* Tables */
table {
    width: 100%;
    border-collapse: collapse;
}

th, td {
    padding: 12px 15px;
    text-align: left;
    border-bottom: 1px solid #e0e0e0;
}

th {
    font-weight: 600;
    color: #666;
    font-size: 0.9rem;
}

tr:hover {
    background-color: #f5f5f5;
}

/* Recommendations */
#recommendations-list {
    padding-left: 20px;
}

#recommendations-list li {
    margin-bottom: 10px;
    line-height: 1.5;
}`

	if err := os.WriteFile(filepath.Join(cssDir, "style.css"), []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("error writing CSS file: %w", err)
	}

	// Create the JavaScript file
	jsContent := `// Fetch dashboard data
async function fetchDashboardData() {
    try {
        const response = await fetch('/api/dashboard');
        if (!response.ok) {
            throw new Error('Failed to fetch dashboard data');
        }
        return await response.json();
    } catch (error) {
        console.error('Error fetching dashboard data:', error);
        return null;
    }
}

// Fetch performance data
async function fetchPerformanceData(days = 30) {
    try {
        const response = await fetch('/api/performance?days=' + days);
        if (!response.ok) {
            throw new Error('Failed to fetch performance data');
        }
        return await response.json();
    } catch (error) {
        console.error('Error fetching performance data:', error);
        return [];
    }
}

// Format currency
function formatCurrency(value) {
    return '$' + parseFloat(value).toFixed(2);
}

// Format number with commas
function formatNumber(value) {
    return value.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

// Format percentage
function formatPercentage(value) {
    return parseFloat(value).toFixed(2) + '%';
}

// Update summary metrics
function updateSummary(data) {
    document.getElementById('total-spend').textContent = formatCurrency(data.summary.total_spend);
    document.getElementById('total-impressions').textContent = formatNumber(data.summary.total_impressions);
    document.getElementById('total-clicks').textContent = formatNumber(data.summary.total_clicks);
    document.getElementById('total-conversions').textContent = formatNumber(data.summary.total_conversions);
    document.getElementById('average-ctr').textContent = formatPercentage(data.summary.average_ctr);
    document.getElementById('average-cpa').textContent = formatCurrency(data.summary.average_cpa);
    document.getElementById('average-roas').textContent = parseFloat(data.summary.average_roas).toFixed(1) + 'x';
    document.getElementById('active-campaigns').textContent = data.summary.active_campaigns;
    
    document.getElementById('last-updated').textContent = new Date(data.generated_at).toLocaleString();
}

// Update top campaigns table
function updateTopCampaigns(campaigns) {
    const tableBody = document.getElementById('top-campaigns-body');
    tableBody.innerHTML = '';
    
    campaigns.forEach(campaign => {
        const row = document.createElement('tr');
        
        const cpa = campaign.spend / campaign.conversions;
        
        row.innerHTML = 
            "<td>" + campaign.name + "</td>" +
            "<td>" + formatCurrency(campaign.spend) + "</td>" +
            "<td>" + campaign.conversions + "</td>" +
            "<td>" + formatCurrency(cpa) + "</td>" +
            "<td>" + parseFloat(campaign.roas).toFixed(1) + "x</td>";
        
        tableBody.appendChild(row);
    });
}

// Update recommendations
function updateRecommendations(recommendations) {
    const list = document.getElementById('recommendations-list');
    list.innerHTML = '';
    
    recommendations.forEach(recommendation => {
        const item = document.createElement('li');
        item.textContent = recommendation;
        list.appendChild(item);
    });
}

// Create performance chart
function createPerformanceChart(data) {
    const ctx = document.getElementById('performance-chart').getContext('2d');
    
    const dates = data.map(item => item.date);
    const spend = data.map(item => item.spend);
    const conversions = data.map(item => item.conversions);
    const cpa = data.map(item => item.cpa);
    
    new Chart(ctx, {
        type: 'line',
        data: {
            labels: dates,
            datasets: [
                {
                    label: 'Spend',
                    data: spend,
                    borderColor: '#1877f2',
                    backgroundColor: 'rgba(24, 119, 242, 0.1)',
                    yAxisID: 'y',
                    fill: true
                },
                {
                    label: 'Conversions',
                    data: conversions,
                    borderColor: '#42b72a',
                    backgroundColor: 'rgba(66, 183, 42, 0.1)',
                    yAxisID: 'y1',
                    fill: true
                },
                {
                    label: 'CPA',
                    data: cpa,
                    borderColor: '#fa3e3e',
                    backgroundColor: 'rgba(250, 62, 62, 0.1)',
                    yAxisID: 'y2',
                    fill: false
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Date'
                    }
                },
                y: {
                    type: 'linear',
                    display: true,
                    position: 'left',
                    title: {
                        display: true,
                        text: 'Spend ($)'
                    }
                },
                y1: {
                    type: 'linear',
                    display: true,
                    position: 'right',
                    title: {
                        display: true,
                        text: 'Conversions'
                    },
                    grid: {
                        drawOnChartArea: false
                    }
                },
                y2: {
                    type: 'linear',
                    display: true,
                    position: 'right',
                    title: {
                        display: true,
                        text: 'CPA ($)'
                    },
                    grid: {
                        drawOnChartArea: false
                    }
                }
            }
        }
    });
}

// Initialize the dashboard
async function initDashboard() {
    const dashboardData = await fetchDashboardData();
    if (!dashboardData) {
        return;
    }
    
    updateSummary(dashboardData);
    updateTopCampaigns(dashboardData.top_campaigns);
    updateRecommendations(dashboardData.recommendations);
    
    const performanceData = await fetchPerformanceData();
    if (performanceData.length > 0) {
        createPerformanceChart(performanceData);
    }
}

// Initialize when the DOM is loaded
document.addEventListener('DOMContentLoaded', initDashboard);`

	if err := os.WriteFile(filepath.Join(jsDir, "dashboard.js"), []byte(jsContent), 0644); err != nil {
		return fmt.Errorf("error writing JavaScript file: %w", err)
	}

	return nil
}