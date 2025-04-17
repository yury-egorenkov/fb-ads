# Facebook Ads Campaign Manager

## Project Objective
Create a comprehensive Go CLI application for Facebook Ads management that enables:
1. Listing all Facebook advertising campaigns via API
2. Creating new campaigns from JSON configuration files
3. Managing campaigns for narrowly targeted audiences
4. Optimizing campaign performance based on results
5. Extracting and analyzing audience data to identify optimal targeting segments

## Technical Requirements
- Implement in Golang
- Use the official Facebook Marketing API
- Authenticate with proper credentials
- Store all settings in JSON and/or Markdown files
- Display and manage campaign details (ID, name, status, budget)
- Handle pagination for accounts with many campaigns
- Implement proper error handling and logging
- Support creation of campaigns with minimal test budgets
- Enable dynamic CPM adjustments based on conversion performance
- Allow campaign deactivation based on performance metrics
- Export audience data with performance statistics

## Audience Analysis
- Export all available audience segments and their statistics
- Analyze performance metrics to identify high-potential audiences
- Create data visualization for audience performance comparison
- Implement filtering capabilities to find suitable audience segments
- Store audience insights in structured format for future campaign creation

## Campaign Strategy
- Create multiple campaigns for narrow audience segments consisting of single categories
- Launch with minimal test budgets
- Adjust maximum CPM pricing to achieve acceptable conversion costs
- Create campaigns for different creative combinations (text + image) targeting various audiences
- Disable campaigns with poor performance metrics
- Continuously refine audience targeting based on performance data

## Implementation Details
- Store all Facebook Ads settings in JSON/Markdown files
- Implement complete management via CLI tool
- Enable creation of campaign structures from configuration files
- Support various ad creative combinations
- Build modular architecture to separate audience analysis, campaign creation, and performance tracking

## Getting Started
1. Set up a Facebook Developer account
2. Create a Facebook App and configure Marketing API access
3. Generate access tokens with appropriate permissions
4. Install required Go dependencies
5. Implement the API client following best practices
6. Set up audience data extraction and analysis pipeline

## Core Features

### Performance Analytics and Statistics
- Comprehensive statistics collection system for impressions, clicks, and costs
- Daily performance data gathering and persistent storage 
- Trend analysis with statistical calculations (min, max, avg, std dev)
- Campaign-specific and aggregate performance metrics
- CSV export for further analysis in spreadsheet applications
- Visualization of campaign performance over time
- Automated data collection and reporting

### Campaign Management
- Creating campaigns from JSON configuration files
- Updating existing campaigns with new parameters
- Duplicating campaigns with customized settings
- Exporting campaigns to JSON and YAML formats
- Pausing, resuming, and archiving campaigns
- Bulk management of multiple campaigns

### Audience Targeting and Analysis
- Searching and filtering audience segments
- Analyzing audience performance metrics
- Exporting audience data for further analysis
- Creating optimized targeting combinations
- Testing multiple audience segments with minimal budgets

### Campaign Optimization
- YAML-based configuration for optimization testing
- Automatic test campaign generation for creative/audience combinations
- Budget optimization based on performance metrics
- CPM adjustment based on conversion data
- Automated underperforming campaign deactivation

## Deliverables
- Go source code for the CLI application
- Comprehensive documentation for setup and usage
- Sample configuration files for campaigns and optimization
- Performance tracking and statistical analysis system
- Audience analysis toolkit with data visualization
- Command reference for all CLI operations
- Reporting and data export capabilities