package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// FacebookAuth handles authentication with Facebook API
type FacebookAuth struct {
	AppID       string
	AppSecret   string
	AccessToken string
	APIVersion  string
}

// NewFacebookAuth creates a new FacebookAuth instance
func NewFacebookAuth(appID, appSecret, accessToken, apiVersion string) *FacebookAuth {
	return &FacebookAuth{
		AppID:       appID,
		AppSecret:   appSecret,
		AccessToken: accessToken,
		APIVersion:  apiVersion,
	}
}

// ValidateToken checks if the access token is valid
func (fa *FacebookAuth) ValidateToken() (bool, error) {
	if fa.AccessToken == "" {
		return false, errors.New("access token is empty")
	}

	// TODO: Implement actual validation by making a request to Facebook Graph API
	return true, nil
}

// GetAPIBaseURL returns the base URL for the Facebook API
func (fa *FacebookAuth) GetAPIBaseURL() string {
	return fmt.Sprintf("https://graph.facebook.com/%s", fa.APIVersion)
}

// GetAuthenticatedRequest returns an http request with authentication
func (fa *FacebookAuth) GetAuthenticatedRequest(endpoint string, params url.Values) (*http.Request, error) {
	baseURL := fmt.Sprintf("%s/%s", fa.GetAPIBaseURL(), endpoint)
	
	if params == nil {
		params = url.Values{}
	}
	
	params.Set("access_token", fa.AccessToken)
	
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.URL.RawQuery = params.Encode()
	return req, nil
}

// AuthenticateRequest adds authentication to an existing HTTP request
func (fa *FacebookAuth) AuthenticateRequest(req *http.Request) {
	// Add access token to query parameters
	q := req.URL.Query()
	q.Set("access_token", fa.AccessToken)
	req.URL.RawQuery = q.Encode()
}