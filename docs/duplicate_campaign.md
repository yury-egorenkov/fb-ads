# Duplicating Facebook Ad Campaigns

This guide explains how to use the campaign duplication feature to clone existing Facebook ad campaigns with all their internal components (ad sets and ads).

## Overview

The campaign duplication feature allows you to:

1. Create an exact copy of an existing campaign with all its ad sets and ads
2. Customize aspects of the duplicate campaign (name, status, dates, budget)
3. Preview the duplicated campaign before creation
4. Automatically rename components to indicate they are copies

This is particularly useful for:
- Creating A/B test variations of successful campaigns
- Quickly deploying seasonal campaigns based on previous templates
- Creating new campaigns while preserving complex targeting settings
- Testing different budget levels with identical creative and targeting

## Command Usage

```bash
fbads duplicate <campaign_id> [options]
```

### Required Parameters

- `campaign_id`: The ID of the Facebook ad campaign to duplicate

### Optional Parameters

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `--name=NAME` | Custom name for the duplicated campaign | "Copy of [original name]" |
| `--status=STATUS` | Status for the duplicated campaign | PAUSED |
| `--start=YYYY-MM-DD` | New start date for the duplicated campaign | Same as original |
| `--end=YYYY-MM-DD` | New end date for the duplicated campaign | Same as original |
| `--budget-factor=X` | Multiply budget by factor X | 1.0 (same budget) |
| `--dry-run`, `-d` | Preview without creating the duplicate | - |

## Examples

### Basic Duplication

Create an exact copy of a campaign with default settings (paused status, same budget):

```bash
fbads duplicate 123456789
```

### Custom Name and Increased Budget

Duplicate a campaign with a custom name and 50% more budget:

```bash
fbads duplicate 123456789 --name="Q2 Promotion 2025" --budget-factor=1.5
```

### New Date Range

Duplicate a campaign and set a new time period:

```bash
fbads duplicate 123456789 --start=2025-05-01 --end=2025-06-30
```

### Active Status with Lower Budget

Duplicate a campaign, set it active, but with a reduced budget for testing:

```bash
fbads duplicate 123456789 --status=ACTIVE --budget-factor=0.25
```

### Preview Without Creating

Preview what would be created without actually creating the campaign:

```bash
fbads duplicate 123456789 --budget-factor=2.0 --status=ACTIVE --dry-run
```

## How It Works

When you duplicate a campaign:

1. The system fetches all details of the original campaign, including ad sets and ads
2. All components are renamed to indicate they are copies (prefix "Copy of")
3. The specified customizations (name, status, dates, budget) are applied
4. IDs are not carried over, ensuring new objects are created
5. Ads and ad sets inherit the status of the parent campaign
6. A summary of the duplication is displayed for review
7. If not in dry-run mode, you're asked for confirmation before creation
8. The campaign with all its components is created via the Facebook Marketing API

## Notes

- The original campaign remains unchanged
- Ad creative elements (images, videos, text) are reused, not duplicated
- All targeting settings from the original campaign are preserved
- Duplicated campaigns start with no performance history
- By default, campaigns are created in PAUSED status for safety
- Date parameters must use the format YYYY-MM-DD (e.g., 2025-12-31)
- Budget factor can be any positive decimal (e.g., 0.5 = half budget, 2.0 = double)
- Due to Facebook API changes, the `image_url` field is no longer supported in ad creatives.
  Images will need to be uploaded separately or referenced by ID when modifying the duplicated ads.
- The `link_url` field is required for all ad creatives. If the original campaign doesn't have a valid link URL,
  the duplication tool will automatically set a default URL to prevent API errors.

## Troubleshooting

### Common Issues

- **API Permission Errors**: Ensure your app has the necessary permissions to create campaigns
- **Invalid Campaign ID**: Verify the campaign ID exists and you have access to it
- **Budget Issues**: Check that your ad account has sufficient funds
- **Date Format Errors**: Ensure dates are in the format YYYY-MM-DD
- **Status Errors**: Status must be one of: ACTIVE, PAUSED, ARCHIVED
- **Missing Link URL**: All ad creatives require a link URL; if you see "The link field is required" error, ensure
  your original campaign has valid link URLs for all ads
- **Unsupported Image URL**: Facebook no longer supports specifying images via URL directly; instead, 
  upload images via the Facebook API first and use the resulting image IDs

### Getting Help

If you encounter issues with campaign duplication, you can:

1. Run with the `--dry-run` flag to validate the configuration
2. Check the Facebook Marketing API documentation for specific error codes
3. View your App permissions in the Facebook Developers portal
4. Contact support if persistent issues occur

## Best Practices

- Always start with a `--dry-run` when duplicating important campaigns
- Use PAUSED status (the default) initially to review before activating
- Consider setting a lower budget initially and scaling up after performance review
- Update the date range when duplicating seasonal campaigns
- Give duplicates meaningful names to distinguish them from originals
- Review all ads after duplication to ensure they remain relevant