# Creating Facebook Ad Campaigns

Facebook's Marketing API has evolved over time, and campaign configuration requirements have changed. This document outlines the current requirements for creating campaigns with the `fbads` tool.

## Current API Version

The tool uses Facebook Marketing API v22.0 by default. If you encounter API version errors, update your configuration file.

## Campaign Objectives

Facebook has updated their campaign objective values to be more outcome-focused. Here are the current valid objective values:

1. OUTCOME_LEADS - For lead generation campaigns
2. OUTCOME_SALES - For conversion and sales campaigns
3. OUTCOME_ENGAGEMENT - For engagement campaigns (formerly PAGE_LIKES, POST_ENGAGEMENT, etc.)
4. OUTCOME_AWARENESS - For brand awareness campaigns
5. OUTCOME_TRAFFIC - For traffic campaigns (formerly LINK_CLICKS)
6. OUTCOME_APP_PROMOTION - For app install and engagement campaigns

## Required Page ID

All ad creatives require a Facebook Page ID. This is the Page that will be shown as the advertiser for your ads. You must add a `page_id` field to your creative configuration:

```json
"creative": {
  "title": "My Website",
  "body": "Check out our website for more information",
  "image_url": "https://example.com/image.jpg",
  "link_url": "https://example.com",
  "call_to_action": "LEARN_MORE",
  "page_id": "YOUR_FACEBOOK_PAGE_ID"  // This is required
}
```

You can find your Page ID by going to your Facebook Page and looking at the URL, or through the Facebook Business Manager.

## Creating Campaigns

To create a campaign, run:
```
./fbads create campaign_template.json
```

You'll be shown a summary of the campaign configuration and prompted to confirm before creation.

## API Evolution

As Facebook's API continues to evolve, you may need to update other fields or parameters in the future. This flexibility is one of the advantages of using a tool like this that can be updated to accommodate API changes while maintaining a consistent interface for your campaign management workflow.
