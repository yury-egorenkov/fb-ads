# Campaign Optimization System

## Overview

The Campaign Optimization System is designed to help advertisers systematically test different combinations of ad creatives, audiences, and placements to find the most cost-effective configurations. 

The system works by creating multiple small test campaigns, each testing one specific variable (either an audience segment or a placement type), collecting performance data, and then optimizing budgets based on results. This approach prevents Facebook's algorithm from favoring the combinations that are most profitable for Facebook but potentially less effective for advertisers.

## Problem Statement

When all targeting parameters are set in a single ad campaign, the ad network tends to find the least cost-effective parameter combination and consume the entire advertising budget on it. These parameters maximize the ad network's revenue (high frequency and high price) while being least beneficial for the advertiser.

## Key Components

1. **YAML Configuration**: Define campaign parameters, creatives, and targeting options
2. **Combination Generator**: Create all possible combinations of creatives with audiences and placements
3. **Budget Allocation**: Distribute test budget efficiently across test campaigns
4. **Batch Processing**: Create campaigns in batches with rate limiting to respect API constraints
5. **Performance Analysis**: Collect and analyze statistics from test campaigns
6. **Optimization Decision**: Adjust CPM bids or terminate underperforming campaigns
7. **Campaign Management**: Update bids and budgets based on performance metrics

## Getting Started

### Exporting Existing Campaigns to YAML

If you have existing Facebook campaigns that you want to use as a starting point for optimization, you can export them to the optimization YAML format:

```bash
fbads exportyaml <campaign_id> [output_file] [options]
```

Options:
- `--budget <amount>`: Set the total budget for optimization testing (default: 1000.00)
- `--test-percent <pct>`: Set the percentage of the budget to use for testing (default: 20)
- `--max-cpm <amount>`: Set the maximum CPM bid amount (default: 15.00)

Example:
```bash
fbads exportyaml 23843791234897 my_campaign.yaml --budget 2000 --test-percent 15 --max-cpm 12.50
```

This will:
1. Fetch the campaign details from Facebook
2. Extract creatives, audiences, and placements
3. Format them into the optimization YAML structure
4. Save to the specified output file (or `<campaign_id>.yaml` if not specified)

### Creating a YAML Configuration File

Alternatively, you can create a YAML configuration file from scratch. The file should have the following structure:

```yaml
campaign:
  name: "Campaign Name"
  total_budget: 1000.00
  test_budget_percentage: 20
  max_cpm: 15.00

creatives:
  - id: "creative1"
    title: "Creative Title 1"
    description: "Creative Description 1"
    image_url: "https://example.com/image1.jpg"
    link_url: "https://example.com/page1"
    call_to_action: "SHOP_NOW"
    page_id: "123456789"
  - id: "creative2"
    title: "Creative Title 2"
    description: "Creative Description 2"
    image_url: "https://example.com/image2.jpg"
    link_url: "https://example.com/page2"
    call_to_action: "LEARN_MORE"
    page_id: "123456789"

targeting_options:
  audiences:
    - id: "audience1"
      name: "Young Adults"
      parameters:
        age_min: 18
        age_max: 24
        genders: [1]
    - id: "audience2"
      name: "Middle-aged Women"
      parameters:
        age_min: 35
        age_max: 44
        genders: [2]
    - id: "audience3"
      name: "Interest Group"
      parameters:
        interests: [
          {"id": "6003139266461", "name": "Shopping"}
        ]

  placements:
    - id: "placement1"
      name: "Facebook Feed"
      position: "feed"
    - id: "placement2"
      name: "Instagram Stories"
      position: "story"
```

### Configuration Fields

#### Campaign Section
- `name`: Name of the campaign
- `total_budget`: Total budget for the entire optimization process
- `test_budget_percentage`: Percentage of the total budget to allocate for testing
- `max_cpm`: Maximum cost per thousand impressions to bid

#### Creatives Section
Each creative should include:
- `id`: Unique identifier for the creative
- `title`: Title text of the ad
- `description`: Body text of the ad
- `image_url`: URL to the image to use
- `link_url`: Destination URL when the ad is clicked
- `call_to_action`: Type of call-to-action button
- `page_id`: Facebook Page ID that will be shown as the advertiser

#### Targeting Options
The targeting_options section contains two subsections:

##### Audiences
Each audience should include:
- `id`: Unique identifier for the audience
- `name`: Descriptive name for the audience
- `parameters`: Facebook targeting parameters as key-value pairs

Common audience parameters:
- `age_min`, `age_max`: Age range
- `genders`: List of gender IDs (1 for male, 2 for female)
- `interests`: List of interest objects with `id` and `name`
- `behaviors`: List of behavior objects
- `geo_locations`: Geographic targeting
- `family_statuses`: Family status targeting

##### Placements
Each placement should include:
- `id`: Unique identifier for the placement
- `name`: Descriptive name for the placement
- `position`: Type of placement (e.g., `feed`, `story`, `right_hand_column`)

## Using the Optimization System

### Validating a Configuration

Before creating test campaigns, validate the YAML configuration:

```bash
fbads optimize validate <yaml_file>
```

This will check that the configuration is valid and display a summary of the test campaigns that would be created.

Example:
```bash
fbads optimize validate my_campaign.yaml
```

### Creating Test Campaigns

To create the test campaigns from your configuration:

```bash
fbads optimize create <yaml_file> [options]
```

Options:
- `--template <file>`: Use a JSON campaign file as template
- `--limit <num>`: Limit the number of test combinations to create
- `--batch-size <num>`: Number of campaigns to create in each batch (default: 3)
- `--priority <type>`: Priority for combinations: audience or placement (default: audience)
- `--dry-run`: Preview campaigns without creating them

Example:
```bash
fbads optimize create my_campaign.yaml --limit 10 --batch-size 5 --priority audience
```

Using a template campaign:
```bash
fbads optimize create my_campaign.yaml --template examples/works.json
```

This will:
1. Parse the YAML configuration
2. If a template is provided, use it as the base campaign structure for all test campaigns
3. Generate all possible combinations of creatives with audiences and placements
4. Calculate the budget for each test campaign
5. Create the campaigns in batches with rate limiting
6. Show progress and results

Use the `--dry-run` flag to preview the campaigns that would be created without actually creating them.

#### Using Campaign Templates

The template option allows you to use an existing campaign JSON file as a base for all test campaigns. This is useful when you want to:

1. Maintain specific campaign settings (objective, buying type, bid strategy)
2. Keep advanced targeting options that aren't specified in the YAML file
3. Use a campaign structure that has already been successful
4. Ensure consistency across all test campaigns

When using a template, the system will:
- Preserve the campaign structure and settings from the template
- Override the campaign name, status, and budget based on the test combination
- Apply the targeting variations from the YAML file to the template's ad sets
- Apply the creative variations from the YAML file to the template's ads

### Updating Campaign CPM Based on Performance

After your test campaigns have run for a sufficient period (typically 1-2 days), you can update their CPM bids based on performance:

```bash
fbads optimize update <campaign_id1,campaign_id2,...> [options]
```

Options:
- `--max-cpm <value>`: Maximum CPM price allowed (default: 15.0)

Example:
```bash
fbads optimize update 123456789,987654321 --max-cpm 12.5
```

This will:
1. Analyze the performance of the specified campaigns
2. Calculate the optimal CPM for each campaign
3. Update the campaigns with new CPM bids, respecting the maximum limit
4. Terminate underperforming campaigns

## How It Works

### Test Campaign Generation

1. Each creative is combined with each audience and placement
2. For a configuration with C creatives, A audiences, and P placements, there will be C × (A + P) possible combinations
3. The test budget is divided equally among all test combinations
4. Campaigns are created in batches to avoid API rate limits

### Budget Allocation

1. The test budget is calculated as: `total_budget × test_budget_percentage / 100`
2. The budget per test campaign is: `test_budget / number_of_combinations`
3. The system estimates the expected impressions for each campaign based on the budget and maximum CPM

### Performance Analysis

1. After campaigns have run for 24-48 hours, performance data is collected
2. Campaigns with fewer impressions than the worst performing active campaign are considered for termination
3. CPM bids are adjusted based on campaign performance, with a maximum cap of the mean CPM of all active campaigns plus one standard deviation
4. The maximum CPM specified in the configuration is always respected

### API Rate Limiting

The system implements exponential backoff with jitter for API rate limiting:
1. Starts with a base delay between requests
2. If rate limits are hit, exponentially increases the delay
3. Adds random jitter to prevent synchronized retries
4. Respects a maximum number of retries

## Best Practices

1. **Start with a small test budget**: Begin with a modest test budget (10-20% of your total budget) to evaluate performance before scaling up.

2. **Use diverse creatives**: Include a variety of creative styles, messages, and images to find what resonates best with your audience.

3. **Test specific audience segments**: Create targeted audience segments based on demographics, interests, or behaviors rather than broad audiences.

4. **Include multiple placement types**: Test different placements (Feed, Stories, Right Column) to find where your ads perform best.

5. **Allow sufficient test time**: Give test campaigns at least 24-48 hours to gather meaningful data before making optimization decisions.

6. **Set reasonable CPM limits**: Set your max_cpm value based on industry benchmarks for your vertical.

7. **Use batch size wisely**: Smaller batch sizes are safer but slower; larger batches are faster but may hit API limits.

8. **Prioritize by importance**: Use the `--priority` flag to test your most important variables first (e.g., `audience` if testing audience segments is your primary goal).

9. **Review before committing**: Always use the `--dry-run` flag first to preview campaigns before creating them.

10. **Use campaign templates**: Leverage the `--template` option with a proven campaign structure to maintain consistency and settings across test campaigns.

11. **Export existing successful campaigns**: Use the `exportyaml` command to start from campaigns that have already shown some success.