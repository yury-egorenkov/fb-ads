# Facebook Ads Manager CLI - Testing Plan

This document outlines the tests needed to ensure the Facebook Ads Manager CLI functions correctly and reliably.

## Unit Tests

### Authentication (pkg/auth)
- Test token validation
- Test authentication request generation
- Test token refresh mechanism
- Test error handling for invalid credentials

### API Client (internal/api)
- Test campaign listing with mock responses
- Test error handling for API failures
- Test pagination for large result sets
- Test metrics collection and parsing
- Test reporting functionality

### Audience Analyzer (internal/audience)
- Test interest and behavior search functionality
- Test audience size estimation with cache hits/misses
- Test concurrent processing of segments
- Test segment filtering by various criteria
- Test fallback handling when API errors occur

### Campaign Creation (internal/campaign)
- Test campaign configuration validation
- Test conversion from config to API parameters
- Test ad set creation within campaigns
- Test creative generation and validation

### Configuration (internal/config)
- Test loading from file
- Test environment variable overrides
- Test defaults when config is missing
- Test validation of required fields

## Integration Tests

### API Integration
- Test actual API calls with sandbox account
- Test rate limiting behavior
- Test pagination with real data
- Test error handling with real API responses

### End-to-End Command Tests
- Test `list` command with various filters
- Test `create` command with sample configuration
- Test `update` command for campaign modifications
- Test `duplicate` command 
- Test `audience search` command with various queries
- Test `audience stats` command with real campaigns
- Test `report` command generation

## Mock Testing

### Mock HTTP Server
- Create mock Facebook API server responses
- Test all commands against mock server
- Simulate various error conditions
- Test retry logic and error handling

### CLI Input/Output Testing
- Test command-line argument parsing
- Test output formatting (table, JSON, CSV)
- Test interactive prompts
- Test error messages and help display

## Performance Tests

### Large Data Sets
- Test with large number of campaigns (100+)
- Test audience search with many results
- Test report generation with extensive historical data

### Concurrency
- Test concurrent API requests
- Test audience size estimation with many segments
- Test throttling and backoff strategies

## Error Handling Tests

### API Error Scenarios
- Test handling of authentication failures
- Test handling of rate limits
- Test handling of invalid requests
- Test network connectivity issues

### Input Error Scenarios
- Test with invalid command syntax
- Test with missing required parameters
- Test with invalid configuration files
- Test with malformed JSON input

## Security Tests

### Credential Handling
- Test secure storage of tokens
- Test masking of sensitive information in logs
- Test credential rotation

### API Permissions
- Test behavior with limited permission tokens
- Test error messages for permission issues

## Implementation Plan

### Phase 1: Core Unit Tests
1. Set up testing framework with mocks
2. Implement tests for auth and config
3. Add API client tests with mock responses
4. Test campaign creation and validation

### Phase 2: Integration Tests
1. Create sandbox Facebook Ad Account
2. Implement integration tests with real API
3. Add end-to-end command tests
4. Create mock server for offline testing

### Phase 3: Advanced Tests
1. Add performance tests
2. Implement error handling tests
3. Add security and permission tests
4. Create CI/CD pipeline for automated testing

## Test Utilities Needed

1. **Mock HTTP Server**: To simulate Facebook API responses
2. **Test Campaign Configs**: Sample configurations for testing
3. **Test Data Generator**: To create large test datasets
4. **API Response Fixtures**: Saved API responses for offline testing
5. **CI Integration**: GitHub Actions or similar for automated testing