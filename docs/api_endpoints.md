# Facebook Ads Manager API Endpoints

This document describes the Facebook Marketing API endpoints used by the Facebook Ads Manager CLI tool.

## Authentication

All API requests require authentication with an access token. The Facebook Ads Manager uses OAuth 2.0 for authentication.

Base URL: `https://graph.facebook.com/{api-version}/`

## Campaigns

### List Campaigns

Retrieves a list of all campaigns for an ad account.

- **Endpoint**: `act_{ad_account_id}/campaigns`
- **Method**: GET
- **Parameters**:
  - `fields`: Comma-separated list of fields to retrieve
  - `limit`: Maximum number of campaigns to return (default: 25)
  - `after`: Pagination cursor
  - `time_range`: JSON object with `since` and `until` date filters

**Example Response**:
```json
{
  "data": [
    {
      "id": "123456789",
      "name": "Test Campaign",
      "status": "ACTIVE",
      "objective": "CONVERSIONS",
      "spend_cap": 1000,
      "daily_budget": 50,
      "bid_strategy": "LOWEST_COST_WITHOUT_CAP",
      "buying_type": "AUCTION",
      "created_time": "2023-04-10T12:00:00+0000",
      "updated_time": "2023-04-10T12:00:00+0000"
    }
  ],
  "paging": {
    "cursors": {
      "before": "cursor_before",
      "after": "cursor_after"
    },
    "next": "https://graph.facebook.com/v18.0/act_123456789/campaigns?limit=25&after=cursor_after"
  }
}
```

### Create Campaign

Creates a new campaign for an ad account.

- **Endpoint**: `act_{ad_account_id}/campaigns`
- **Method**: POST
- **Parameters**:
  - `name`: Campaign name
  - `status`: Campaign status (ACTIVE, PAUSED)
  - `objective`: Campaign objective
  - `buying_type`: Buying type (AUCTION, RESERVED)
  - `special_ad_categories`: JSON array of special ad categories
  - `bid_strategy`: Bid strategy
  - `daily_budget`: Daily budget in cents
  - `lifetime_budget`: Lifetime budget in cents (optional)

**Example Response**:
```json
{
  "id": "123456789"
}
```

## Ad Sets

### List Ad Sets

Retrieves a list of all ad sets for a campaign.

- **Endpoint**: `act_{ad_account_id}/adsets`
- **Method**: GET
- **Parameters**:
  - `campaign_id`: Filter by campaign ID
  - `fields`: Comma-separated list of fields to retrieve
  - `limit`: Maximum number of ad sets to return (default: 25)
  - `after`: Pagination cursor

**Example Response**:
```json
{
  "data": [
    {
      "id": "123456789",
      "name": "Test Ad Set",
      "campaign_id": "987654321",
      "status": "ACTIVE",
      "targeting": { /* targeting object */ },
      "optimization_goal": "OFFSITE_CONVERSIONS",
      "billing_event": "IMPRESSIONS",
      "bid_amount": 500,
      "start_time": "2023-04-10T12:00:00+0000",
      "end_time": "2023-05-10T12:00:00+0000"
    }
  ],
  "paging": {
    "cursors": {
      "before": "cursor_before",
      "after": "cursor_after"
    }
  }
}
```

### Create Ad Set

Creates a new ad set for a campaign.

- **Endpoint**: `act_{ad_account_id}/adsets`
- **Method**: POST
- **Parameters**:
  - `name`: Ad set name
  - `campaign_id`: Parent campaign ID
  - `status`: Ad set status (ACTIVE, PAUSED)
  - `targeting`: JSON object with targeting specifications
  - `optimization_goal`: Optimization goal
  - `billing_event`: Billing event
  - `bid_amount`: Bid amount in cents
  - `start_time`: Start time in ISO 8601 format
  - `end_time`: End time in ISO 8601 format (optional)

**Example Response**:
```json
{
  "id": "123456789"
}
```

## Ads

### List Ads

Retrieves a list of all ads for an ad set.

- **Endpoint**: `act_{ad_account_id}/ads`
- **Method**: GET
- **Parameters**:
  - `adset_id`: Filter by ad set ID
  - `fields`: Comma-separated list of fields to retrieve
  - `limit`: Maximum number of ads to return (default: 25)
  - `after`: Pagination cursor

**Example Response**:
```json
{
  "data": [
    {
      "id": "123456789",
      "name": "Test Ad",
      "adset_id": "987654321",
      "status": "ACTIVE",
      "creative": { /* creative object */ }
    }
  ],
  "paging": {
    "cursors": {
      "before": "cursor_before",
      "after": "cursor_after"
    }
  }
}
```

### Create Ad

Creates a new ad for an ad set.

- **Endpoint**: `act_{ad_account_id}/ads`
- **Method**: POST
- **Parameters**:
  - `name`: Ad name
  - `adset_id`: Parent ad set ID
  - `status`: Ad status (ACTIVE, PAUSED)
  - `creative`: Ad creative ID or creative specifications

**Example Response**:
```json
{
  "id": "123456789"
}
```

## Insights

### Get Campaign Insights

Retrieves performance metrics for campaigns.

- **Endpoint**: `act_{ad_account_id}/insights`
- **Method**: GET
- **Parameters**:
  - `level`: Level of results (campaign, adset, ad)
  - `fields`: Comma-separated list of fields to retrieve
  - `time_range`: JSON object with `since` and `until` date filters
  - `filtering`: JSON array of filters
  - `breakdowns`: Breakdown dimensions

**Example Response**:
```json
{
  "data": [
    {
      "campaign_id": "123456789",
      "campaign_name": "Test Campaign",
      "spend": "100.50",
      "impressions": "10000",
      "clicks": "500",
      "ctr": "0.05",
      "cpm": "10.05",
      "cpc": "0.201",
      "actions": [
        {
          "action_type": "offsite_conversion",
          "value": "10"
        }
      ]
    }
  ],
  "paging": {
    "cursors": {
      "before": "cursor_before",
      "after": "cursor_after"
    }
  }
}
```

## Targeting

### Search Targeting Options

Searches for targeting options like interests, behaviors, and demographics.

- **Endpoint**: `search`
- **Method**: GET
- **Parameters**:
  - `type`: Type of targeting option (adinterest, adeducationschool, adlocale, etc.)
  - `q`: Search query
  - `limit`: Maximum number of results to return

**Example Response**:
```json
{
  "data": [
    {
      "id": "6003107902433",
      "name": "Online shopping",
      "description": "People interested in Online shopping",
      "type": "interest",
      "audience_size": 910000000
    }
  ],
  "paging": {
    "cursors": {
      "before": "cursor_before",
      "after": "cursor_after"
    }
  }
}
```

## Error Handling

All API errors return a standard error object:

```json
{
  "error": {
    "message": "Error message",
    "type": "ErrorType",
    "code": 100,
    "error_subcode": 1000,
    "fbtrace_id": "trace_id"
  }
}
```

## Rate Limiting

The Facebook Marketing API is subject to rate limits. When a rate limit is exceeded, the API will return a 429 status code with information about when the limit will reset.

For detailed information on rate limits, see the [Facebook Marketing API documentation](https://developers.facebook.com/docs/marketing-api/overview/rate-limiting).

## Reference

For complete API documentation, refer to the [Facebook Marketing API Reference](https://developers.facebook.com/docs/marketing-api/reference).