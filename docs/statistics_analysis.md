# Statistics Analysis System

The statistics analysis system provides comprehensive tools for collecting, analyzing, and exporting Facebook Ads performance data. This system allows you to track key metrics such as impressions, clicks, costs, and conversions, and provides insights into campaign performance trends over time.

## Overview

The system supports:
- Daily data collection from the Facebook Marketing API
- Persistent storage of performance data
- Analysis of performance over time (trends)
- Campaign-specific and aggregate statistics
- Statistical calculations (min, max, average, standard deviation)
- Data export in both JSON and CSV formats
- Command-line interface for all operations

## CLI Commands

### Collecting Statistics

The `collect` subcommand retrieves performance data from Facebook for a specified date range and stores it for later analysis.

```
fbads stats collect [options]
```

Options:
- `--start, -s <date>`: Start date in YYYY-MM-DD format
- `--end, -e <date>`: End date in YYYY-MM-DD format
- `--days, -d <num>`: Number of days to go back from today (default: 30)

Example:
```
# Collect data for the last 30 days
fbads stats collect

# Collect data for a specific date range
fbads stats collect --start 2023-01-01 --end 2023-01-31

# Collect data for the last 7 days
fbads stats collect --days 7
```

### Analyzing Statistics

The `analyze` subcommand performs statistical analysis on the collected data and displays the results.

```
fbads stats analyze [options]
```

Options:
- `--start, -s <date>`: Start date in YYYY-MM-DD format
- `--end, -e <date>`: End date in YYYY-MM-DD format
- `--days, -d <num>`: Number of days to go back from today (default: 30)
- `--campaign, -c <id>`: Filter results to a specific campaign ID
- `--format, -f <format>`: Output format: `json` or `table` (default: `json`)

Example:
```
# Analyze all campaigns for the last 30 days in table format
fbads stats analyze --format table

# Analyze a specific campaign for a custom date range
fbads stats analyze --campaign 123456789 --start 2023-01-01 --end 2023-01-31

# Analyze all campaigns for the last 7 days in JSON format
fbads stats analyze --days 7
```

### Exporting Statistics

The `export` subcommand exports statistics to a CSV file for further analysis in spreadsheet applications.

```
fbads stats export [options]
```

Options:
- `--start, -s <date>`: Start date in YYYY-MM-DD format
- `--end, -e <date>`: End date in YYYY-MM-DD format
- `--days, -d <num>`: Number of days to go back from today (default: 30)
- `--output, -o <file>`: Path to the output CSV file

Example:
```
# Export last 30 days to a CSV file
fbads stats export --output campaign_stats.csv

# Export specific date range to a CSV file
fbads stats export --start 2023-01-01 --end 2023-01-31 --output january_stats.csv
```

## Metrics and Analysis

The statistics system collects and analyzes the following key metrics:

### Primary Metrics
- **Impressions**: Number of times ads were shown
- **Clicks**: Number of clicks on ads
- **Spend**: Amount spent on the campaign
- **Conversions**: Number of actions taken after viewing ads

### Calculated Metrics
- **CTR (Click-Through Rate)**: Percentage of impressions that resulted in clicks
- **CPC (Cost Per Click)**: Average cost per click
- **CPM (Cost Per Mille)**: Cost per 1,000 impressions
- **CPA (Cost Per Acquisition)**: Cost per conversion
- **ROAS (Return On Ad Spend)**: Estimated return on ad spend

### Statistical Analysis
For each metric, the system calculates:
- **Totals**: Sum of values across all campaigns or dates
- **Averages**: Mean value over the time period
- **Minimum and Maximum**: Lowest and highest values
- **Standard Deviation**: Measure of data spread
- **Change over Time**: Percentage change from the first to the last data point

## Data Storage

Performance data is stored in the following location:
- `~/.fbads/stats/daily/`: Contains JSON files with daily performance data per campaign
- Each file is named using the pattern `{campaign_id}_{date}.json`
- Aggregated daily data is stored in `aggregated_{date}.json` files

## Integration with Other Commands

The statistics system is designed to work seamlessly with other fbads commands:

- **Campaign Optimization**: Use statistics to inform optimization decisions
- **Reports**: Generate reports based on collected statistics
- **Dashboard**: Visualize statistics in the web dashboard

## Troubleshooting

### No Data Available
If you see "No statistics found" when running the analyze command:
1. Ensure you've run the `collect` command for the relevant date range
2. Verify your Facebook API token has access to the campaign data
3. Check that the campaign IDs are correct

### API Rate Limiting
Facebook enforces rate limits on API requests. The system implements automatic backoff:
1. If you're collecting data for a large date range, the system may pause between requests
2. For very large datasets, consider collecting data in smaller date ranges

### File Storage Issues
If you encounter errors saving statistics:
1. Ensure the application has write permissions to the `~/.fbads/stats` directory
2. Check for available disk space
3. Try using an absolute path with the `--output` option when exporting