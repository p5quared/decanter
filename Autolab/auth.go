package Autolab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// This client implements the device flow for the autolab API
// and is similar to golang.org/x/oauth2 `config`
type AutolabOAuthClient struct {
	host      string
	clientID  string
	clientSec string
	endpoints map[string]string
}

func NewAuthClient(clientID, clientSecret, host string) AutolabOAuthClient {
	endpoints := map[string]string{
		"device_auth": host + "/oauth/device_flow_init",
		"device_code": host + "/oauth/device_flow_authorize",
		"token":       host + "/oauth/token",
		"cb":          host + "/device_flow_auth_cb",
	}

	return AutolabOAuthClient{
		host:      host,
		clientID:  clientID,
		clientSec: clientSecret,
		endpoints: endpoints,
	}
}

// TODO: This seems irrelevant
func (a AutolabOAuthClient) DefaultSetup() (oauth2.Token, error) {
	tok, err := a.DeviceSetup()
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("Failed to setup device: %v", err)
	}
	return tok, nil
}

// DeviceSetup is a convenience method for setting up a new client and getting a token
// TODO: with context
func (a AutolabOAuthClient) DeviceSetup() (oauth2.Token, error) {
	// a.conf.DeviceAuth(ctx) does not work
	// because for some reason, the endpoint is set to GET
	// and not POST as it should be.
	// ??? It says it right in the OAuth2 standard that they reference....
	resp, err := a.DeviceAuth()
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("Failed to start device flow: %v", err)
	}
	fmt.Printf("Visit %v and enter the code %v to authenticate.\n", resp.VerificationURI, resp.UserCode)

	// conf.DeviceAccessToken does not work either???
	// "unsupported_grant_type" error
	code, err := a.DeviceAccessCode(resp)
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("Failed to get code: %v", err)
	}

	fmt.Printf("Your code is: %v\n", code)

	token, err := a.ExchangeCodeForToken(code)
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("Failed to exchange code for token: %v", err)
	}

	return token, nil
}

// Initiate a device flow with the autolab server
// TODO: with context
func (a AutolabOAuthClient) DeviceAuth() (oauth2.DeviceAuthResponse, error) {
	endpoint := a.endpoints["device_auth"]
	params := url.Values{"client_id": {a.clientID}}

	resp, err := http.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return oauth2.DeviceAuthResponse{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return oauth2.DeviceAuthResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var deviceAuthResponse oauth2.DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceAuthResponse); err != nil {
		return oauth2.DeviceAuthResponse{}, fmt.Errorf("error decoding response: %w\n", err)
	}

	return deviceAuthResponse, nil
}

// Poll the autolab server for the access code
// TODO: with context
func (a AutolabOAuthClient) DeviceAccessCode(resp oauth2.DeviceAuthResponse) (string, error) {
	endpoint := a.endpoints["device_code"]
	params := url.Values{
		"client_id":   {a.clientID},
		"device_code": {resp.DeviceCode},
	}

	// TODO: configurable timeout and context
	authorizationTimeout := 2 * time.Minute
	for {
		select {
		case <-time.After(authorizationTimeout):
			return "", fmt.Errorf("timed out")

		case <-time.Tick(6 * time.Second):
			resp, err := http.Get(endpoint + "?" + params.Encode())
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var code struct {
					AccessToken string `json:"code"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&code); err == nil {
					return code.AccessToken, nil
				}
			}
		}
	}
}

// TODO: with context
func (a AutolabOAuthClient) ExchangeCodeForToken(code string) (oauth2.Token, error) {
	endpoint := a.endpoints["token"]

	resp, err := http.PostForm(endpoint, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSec},
		"redirect_uri":  {a.endpoints["cb"]},
	})

	if err != nil {
		return oauth2.Token{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return oauth2.Token{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var token oauth2.Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return oauth2.Token{}, err
	}

	return token, nil
}

// TODO: with context
func (a AutolabOAuthClient) RefreshToken(oldToken oauth2.Token) (*oauth2.Token, error) {
	endpoint := a.endpoints["token"]
	params := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {oldToken.RefreshToken},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSec},
	}

	resp, err := http.PostForm(endpoint, params)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var token *oauth2.Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return token, nil
}
