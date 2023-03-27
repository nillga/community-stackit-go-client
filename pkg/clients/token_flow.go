package clients

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/SchwarzIT/community-stackit-go-client/pkg/env"
	"golang.org/x/oauth2"
)

const (
	// Service Account Token Flow
	// Auth flow env variables
	ServiceAccountEmail = "STACKIT_SERVICE_ACCOUNT_EMAIL"
	ServiceAccountToken = "STACKIT_SERVICE_ACCOUNT_TOKEN"
)

// TokenFlow handles auth with SA static token
type TokenFlow struct {
	client *http.Client
	config *TokenFlowConfig
}

// TokenFlowConfig is the flow config
type TokenFlowConfig struct {
	ServiceAccountEmail string
	ServiceAccountToken string
	Environment         env.Environment
	ClientRetry         *RetryConfig
}

// GetEnvironment returns the defined API environment
func (c *TokenFlow) GetEnvironment() env.Environment {
	return c.GetConfig().Environment
}

// GetServiceAccountEmail returns the service account email
func (c *TokenFlow) GetServiceAccountEmail() string {
	return c.GetConfig().ServiceAccountEmail
}

// GetConfig returns the flow configuration
func (c *TokenFlow) GetConfig() TokenFlowConfig {
	if c.config == nil {
		return TokenFlowConfig{}
	}
	return *c.config
}

func (c *TokenFlow) Init(ctx context.Context, cfg ...TokenFlowConfig) error {
	c.processConfig(cfg...)
	c.configureHTTPClient(ctx)
	if c.config.ClientRetry == nil {
		c.config.ClientRetry = NewRetryConfig()
	}
	return c.validate()
}

// processConfig processes the given configuration
func (c *TokenFlow) processConfig(cfg ...TokenFlowConfig) {
	c.config = c.getConfigFromEnvironment()
	for _, m := range cfg {
		c.config = c.mergeConfigs(&m, c.config)
	}
}

// getConfigFromEnvironment returns a TokenFlowConfig populated with environment variables.
func (c *TokenFlow) getConfigFromEnvironment() *TokenFlowConfig {
	return &TokenFlowConfig{
		ServiceAccountEmail: os.Getenv(ServiceAccountEmail),
		ServiceAccountToken: os.Getenv(ServiceAccountToken),
		Environment:         env.Parse(os.Getenv(Environment)),
	}
}

// mergeConfigs returns a new TokenFlowConfig that combines the values of cfg and currentCfg.
func (c *TokenFlow) mergeConfigs(cfg, currentCfg *TokenFlowConfig) *TokenFlowConfig {
	merged := *currentCfg

	if cfg.ServiceAccountEmail != "" {
		merged.ServiceAccountEmail = cfg.ServiceAccountEmail
	}
	if cfg.ServiceAccountToken != "" {
		merged.ServiceAccountToken = cfg.ServiceAccountToken
	}
	if cfg.Environment != "" {
		merged.Environment = cfg.Environment
	}

	return &merged
}

// configureHTTPClient configures the HTTP client
func (c *TokenFlow) configureHTTPClient(ctx context.Context) {
	sts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.config.ServiceAccountToken},
	)
	o2nc := oauth2.NewClient(ctx, sts)
	o2nc.Timeout = DefaultClientTimeout
	c.client = o2nc
}

// validate the client is configured well
func (c *TokenFlow) validate() error {
	if c.config.ServiceAccountToken == "" {
		return errors.New("Service Account Access Token cannot be empty")
	}
	if c.config.ServiceAccountEmail == "" {
		return errors.New("Service Account Email cannot be empty")
	}
	return nil
}

// Do performs the request
func (c *TokenFlow) Do(req *http.Request) (*http.Response, error) {
	if c.client == nil {
		return nil, errors.New("please run Init()")
	}
	return do(c.client, req, c.config.ClientRetry)
}
