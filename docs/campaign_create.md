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

1. Changed the bid_strategy to "COST_CAP" - This tells Facebook to use cost cap bidding where you set a maximum cost you're willing to pay
 2. Set a bid_amount of 5.00 - This means you're willing to pay up to $5.00 for 1000 impressions

 The creator.go code properly handles this by:
 - Converting the dollar amount to cents (multiplying by 100) as required by Facebook's API
 - Setting the bid_amount parameter in the API request

 Additional notes about Facebook bidding strategies:

 1. COST_CAP: Sets a maximum average cost per optimization event (in this case, per 1000 impressions since your billing_event is "IMPRESSIONS")
 2. BID_CAP: Sets a maximum bid for each auction (more strict than COST_CAP)
 3. LOWEST_COST_WITH_BID_CAP: Similar to BID_CAP but allows Facebook to bid lower when possible
 4. LOWEST_COST_WITHOUT_CAP: Lets Facebook automatically bid to get the most results at the lowest cost (no maximum bid)

 With these changes, your campaign will now use cost cap bidding with a maximum of $5.00 per 1000 impressions. You can adjust the bid_amount value
 based on your target CPM.

--

Here's how to fix this issue:

1. Go to your Facebook Developers Dashboard (https://developers.facebook.com/apps/)
2. Select the app you're using for these ads
3. Navigate to App Review > Requests
4. Submit your app for review to move it from development to public mode
5. Alternatively, you can use an existing Facebook page that's already associated with a live app

You can also try these workarounds:

1. Use a different Facebook marketing API access token that's associated with a public app
2. Create the ad directly in Facebook Ads Manager and export the settings for future reference

The key issue is that Facebook restricts ad creation from apps in development mode as a security measure. Your app needs to complete the review
process and be made public before it can create ads programmatically.
