# Updating Facebook Ad Campaigns

This document explains how to update existing Facebook ad campaigns using the campaign update functionality.

## Overview

The campaign update functionality allows you to modify parameters of an existing Facebook ad campaign, such as:

- Campaign status (ACTIVE, PAUSED, ARCHIVED)
- Campaign name
- Budget (daily or lifetime)
- Bid strategy
- And other campaign-level settings

## Command-Line Usage

To update a campaign via the command line, use the `update_campaign` command with the appropriate flags:

```
go run cmd/update_campaign.go -id=CAMPAIGN_ID [options]
```

### Required Parameters

- `-id`: The ID of the campaign to update (required)

### Optional Parameters

At least one of the following update parameters must be provided:

- `-status`: New campaign status (ACTIVE, PAUSED, ARCHIVED)
- `-name`: New campaign name
- `-daily_budget`: New daily budget in your currency (e.g., 50.00)
- `-lifetime_budget`: New lifetime budget in your currency (e.g., 1000.00)
- `-bid_strategy`: New bid strategy (e.g., LOWEST_COST_WITHOUT_CAP)
- `-file`: Path to a JSON file containing campaign update details

### Examples

```bash
# Update a campaign's status to PAUSED
go run cmd/update_campaign.go -id=23847239847 -status=PAUSED

# Update a campaign's name and daily budget
go run cmd/update_campaign.go -id=23847239847 -name="Summer Sale 2025" -daily_budget=75.50

# Update a campaign using a JSON configuration file
go run cmd/update_campaign.go -id=23847239847 -file=examples/update_campaign.json
```

## JSON Update File Format

You can use a JSON file to specify multiple update parameters. Create a JSON file with the following structure:

```json
{
  "name": "Updated Campaign Name",
  "status": "ACTIVE",
  "daily_budget": 65.50,
  "lifetime_budget": 1000.00,
  "bid_strategy": "LOWEST_COST_WITHOUT_CAP"
}
```

Only include the fields you want to update. Fields not included will remain unchanged.

## Notes

1. You can only update one campaign at a time.
2. Command-line parameters take precedence over parameters in the JSON file.
3. You cannot update both daily_budget and lifetime_budget at the same time.
4. Dollar/currency amounts should be specified as decimal values (e.g., 50.00).
5. The Facebook API requires budget values in cents, but our tool handles the conversion for you.

## Error Handling

Common errors you might encounter:

- Invalid campaign ID: Check that the campaign ID exists and is owned by your account.
- Invalid status value: Status must be one of ACTIVE, PAUSED, or ARCHIVED.
- Budget errors: Cannot set both daily and lifetime budgets; values must be positive.
- Authentication errors: Check your API credentials in the config file.

## API Integration

If you want to integrate campaign updates into your own code, use the `UpdateCampaign` method from the API client:

```go
import (
    "net/url"
    "github.com/user/fb-ads/internal/api"
)

// Create API client
client := api.NewClient(fbAuth, accountID)

// Create update parameters
params := url.Values{}
params.Set("status", "PAUSED")
params.Set("name", "Updated Campaign Name")

// Update the campaign
err := client.UpdateCampaign(campaignID, params)
```