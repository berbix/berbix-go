package berbix

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const SDKVersion = "0.0.1"

type Client interface {
	CreateTransaction(options *CreateTransactionOptions) (*Tokens, error)
	RefreshTokens(tokens *Tokens) (*Tokens, error)
	FetchTransaction(tokens *Tokens) (*TransactionMetadata, error)
	DeleteTransaction(tokens *Tokens) error
	UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error)
	OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error
}

type defaultClient struct {
	secret string
	host string
	client HTTPClient
}

type ClientOptions struct {
	Host string
	HTTPClient HTTPClient
}

func NewClient(secret string, options *ClientOptions) Client {
	return &defaultClient{
		secret: secret,
		host:   options.Host,
		client: options.HTTPClient,
	}
}

func (c *defaultClient) CreateTransaction(options *CreateTransactionOptions) (*Tokens, error) {
	return c.fetchTokens("/v0/transactions", options)
}

func (c *defaultClient) RefreshTokens(tokens *Tokens) (*Tokens, error) {
	payload := &struct {
		RefreshToken string `json:"refresh_token"`
		GrantType string `json:"grant_type"`
	}{
		RefreshToken: tokens.RefreshToken,
		GrantType: "refresh_token",
	}
	return c.fetchTokens("/v0/tokens", payload)
}

func (c *defaultClient) FetchTransaction(tokens *Tokens) (*TransactionMetadata, error) {
	metadata := &TransactionMetadata{}
	return metadata, c.tokenAuthRequest(http.MethodGet, tokens, "/v0/transactions", nil, metadata)
}

func (c *defaultClient) DeleteTransaction(tokens *Tokens) error {
	return c.tokenAuthRequest(http.MethodDelete, tokens, "/v0/transactions", nil, nil)
}

func (c *defaultClient) UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error) {
	metadata := &TransactionMetadata{}
	return metadata, c.tokenAuthRequest(http.MethodPatch, tokens, "/v0/transactions", options, metadata)
}

func (c *defaultClient) OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error {
	return c.tokenAuthRequest(http.MethodPatch, tokens, "/v0/transactions/override", options, nil)
}

func (c *defaultClient) tokenAuthRequest(method string, tokens *Tokens, path string, payload interface{}, dst interface{}) error {
	var err error
	if tokens.NeedsRefresh() {
		tokens, err = c.RefreshTokens(tokens)
		if err != nil {
			return err
		}
	}
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent": fmt.Sprintf("BerbixGo/%s", SDKVersion),
		"Authorization": fmt.Sprintf("Bearer %s", tokens.AccessToken),
	}
	return c.client.Request(method, c.makeURL(path), headers, &RequestOptions{Body: body}, dst)
}

func (c *defaultClient) fetchTokens(path string, payload interface{}) (*Tokens, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent": fmt.Sprintf("BerbixGo/%s", SDKVersion),
		"Authorization": c.basicAuth(),
	}
	response := &tokenResponse{}
	if err := c.client.Request(http.MethodPost, c.makeURL(path), headers, &RequestOptions{}, response); err != nil {
		return nil, err
	}
	return fromTokenResponse(response), nil
}

func (c *defaultClient) makeURL(path string) string {
	http.Request{}.BasicAuth()
	return fmt.Sprintf("%s%s", c.host, path)
}

func (c *defaultClient) basicAuth() string {
	auth := fmt.Sprintf("%s:", c.secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return fmt.Sprintf("Basic %s", encoded)
}