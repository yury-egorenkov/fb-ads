# Facebook Ads Manager CLI

A command line interface tool for managing Facebook Ads campaigns, built with Go.

## Features

- List all Facebook advertising campaigns
- Create campaigns from JSON configuration files
- Analyze audience data for better targeting
- Manage campaigns with minimal test budgets
- Optimize campaigns based on performance

## Installation

### Prerequisites

- Go 1.21 or higher
- Facebook Developer Account
- Facebook Marketing API access

### Building from Source

1. Clone this repository:
   ```
   git clone https://github.com/user/fb-ads.git
   cd fb-ads
   ```

2. Build the application:
   ```
   go build -o fbads ./cmd/fbads
   ```

3. Install the binary (optional):
   ```
   go install ./cmd/fbads
   ```

## Configuration

Before using the application, you need to configure it with your Facebook API credentials:

1. Run the configuration command:
   ```
   fbads config
   ```

2. Enter your Facebook App ID, App Secret, Access Token, and Ad Account ID.

Alternatively, you can manually create a configuration file at `~/.fbads/config.json` using the format in `config.example.json`.

## Usage

```
fbads <command> [arguments]
```

Available commands:

- `list` - List all campaigns
- `create` - Create a new campaign from configuration
- `audience` - Analyze audience data
- `config` - Configure the application
- `help` - Show help information

## Examples

### Listing Campaigns

```
fbads list
```

### Creating a Campaign

```
fbads create -f campaign_config.json
```

### Analyzing Audience Data

```
fbads audience -export audience_data.json
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.