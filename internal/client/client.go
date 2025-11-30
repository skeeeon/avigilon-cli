package client

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"avigilon-cli/internal/auth"
)

type AvigilonClient struct {
	HTTP   *resty.Client
	Config ClientConfig
}

type ClientConfig struct {
	BaseURL       string
	Username      string
	Password      string
	UserNonce     string // For Auth Token
	UserKey       string // For Auth Token
	IntegrationID string // For Auth Token
}

// LoginPayload matches the JSON body required by POST /login (Page 40)
type LoginPayload struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	ClientName         string `json:"clientName"`
	AuthorizationToken string `json:"authorizationToken"`
}

// LoginResponse captures the session ID returned by the API
type LoginResponse struct {
	Result struct {
		Session string `json:"session"`
	} `json:"result"`
}

func New(cfg ClientConfig) *AvigilonClient {
	r := resty.New()
	r.SetBaseURL(cfg.BaseURL)
	
	// PDF Page 1: "Response Content Type: application/json"
	r.SetHeader("Content-Type", "application/json")
	r.SetHeader("Accept", "application/json")

	// Disable TLS verification for testing (common in on-prem VMS with self-signed certs)
	r.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &AvigilonClient{
		HTTP:   r,
		Config: cfg,
	}
}

// Login authenticates with the VMS, sets the session header internally, 
// and returns the session ID string for persistence.
func (c *AvigilonClient) Login() (string, error) {
	// 1. Generate the cryptographic signature
	authToken := auth.GenerateAuthToken(
		c.Config.UserNonce,
		c.Config.UserKey,
		c.Config.IntegrationID,
	)

	payload := LoginPayload{
		Username:           c.Config.Username,
		Password:           c.Config.Password,
		ClientName:         "Avigilon-Go-CLI",
		AuthorizationToken: authToken,
	}

	// 2. Make Request
	resp, err := c.HTTP.R().
		SetBody(payload).
		SetResult(&LoginResponse{}).
		Post("/login")

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("login failed: %s", resp.String())
	}

	// 3. Extract Session
	loginResult, ok := resp.Result().(*LoginResponse)
	if !ok {
		return "", errors.New("failed to parse login response")
	}

	sessionID := loginResult.Result.Session

	if sessionID == "" {
		return "", errors.New("login successful but no session ID returned")
	}

	// 4. Inject Session into all future requests for this client instance
	// Page 1: "x-avg-session: The session parameter returned by the login request (header)"
	c.HTTP.SetHeader("x-avg-session", sessionID)

	// Return the session ID so it can be saved to config
	return sessionID, nil
}

// GetHealth checks the node status
func (c *AvigilonClient) GetHealth() (string, error) {
	resp, err := c.HTTP.R().Get("/health")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}
