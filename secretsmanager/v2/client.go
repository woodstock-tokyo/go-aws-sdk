// Package secretsmanager provides instance-scoped, context-aware Secrets Manager
// clients using AWS SDK for Go v2. The legacy parent package remains unchanged.
package secretsmanager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awssecrets "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var (
	ErrClientRequired = errors.New("Secrets Manager client is required")
	ErrSecretID       = errors.New("secret ID is required")
	ErrEmptySecret    = errors.New("secret string is empty or unavailable")
)

// API is the part of the generated AWS client used by Client. It permits
// dependency injection without process-global clients or credentials.
type API interface {
	GetSecretValue(context.Context, *awssecrets.GetSecretValueInput, ...func(*awssecrets.Options)) (*awssecrets.GetSecretValueOutput, error)
}

type Client struct{ api API }

// New uses the standard SDK credential and region chains, including session
// tokens, profiles, web identity and workload IAM roles. It does not require
// static access keys. Options such as config.WithRegion may override defaults.
func New(ctx context.Context, options ...func(*config.LoadOptions) error) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	return NewFromConfig(cfg), nil
}

// NewFromConfig creates an independent client. Multiple clients may use
// different regions, credentials and endpoints in the same process.
func NewFromConfig(cfg aws.Config, options ...func(*awssecrets.Options)) *Client {
	return NewWithAPI(awssecrets.NewFromConfig(cfg, options...))
}

func NewWithAPI(api API) *Client { return &Client{api: api} }

// SecretString returns the raw SecretString, without assuming its JSON schema.
// Binary-only secrets are rejected. No secret or provider error is logged.
// Errors wrap the SDK cause for errors.Is/As; applications should not expose
// provider error details to untrusted callers.
func (c *Client) SecretString(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if c == nil || c.api == nil {
		return "", ErrClientRequired
	}
	if strings.TrimSpace(id) == "" {
		return "", ErrSecretID
	}
	value, err := c.api.GetSecretValue(ctx, &awssecrets.GetSecretValueInput{SecretId: aws.String(id)})
	if err != nil {
		return "", fmt.Errorf("retrieve secret string: %w", err)
	}
	if value == nil || value.SecretString == nil || strings.TrimSpace(*value.SecretString) == "" {
		return "", ErrEmptySecret
	}
	return *value.SecretString, nil
}
