package berbix

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sdkVersion = "0.0.1"
	clockDrift = 300
)

type Client interface {
	CreateTransaction(options *CreateTransactionOptions) (*Tokens, error)
	RefreshTokens(tokens *Tokens) (*Tokens, error)
	FetchTransaction(tokens *Tokens) (*TransactionMetadata, error)
	DeleteTransaction(tokens *Tokens) error
	UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error)
	OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error
	ValidateSignature(secret, body, header string) error
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
	client := options.HTTPClient
	if client == nil {
		client = &DefaultHTTPClient{client: http.DefaultClient}
	}
	return &defaultClient{
		secret: secret,
		host:   options.Host,
		client: client,
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

func computeHMACSHA256(secret, message string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *defaultClient) ValidateSignature(secret, body, header string) error {
	parts := strings.Split(header, ",")
	if len(parts) != 3 {
		return errors.New("incorrect number of parts in header for validation")
	}
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	signature := parts[2]
	if timestamp < time.Now().Unix() - clockDrift {
		return errors.New("hook is outside of drift range, signature invalid")
	}
	toSign := fmt.Sprintf("%d,%s,%s", timestamp, secret, body)
	if signature != computeHMACSHA256(secret, toSign) {
		return errors.New("signature does not match")
	}
	return nil
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
		"User-Agent": fmt.Sprintf("BerbixGo/%s", sdkVersion),
		"Authorization": fmt.Sprintf("Bearer %s", tokens.AccessToken),
	}
	return c.client.Request(method, c.makeURL(path), headers, &RequestOptions{Body: body}, dst)
}

func (c *defaultClient) fetchTokens(path string, payload interface{}) (*Tokens, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent": fmt.Sprintf("BerbixGo/%s", sdkVersion),
		"Authorization": c.basicAuth(),
	}
	response := &tokenResponse{}
	if err := c.client.Request(http.MethodPost, c.makeURL(path), headers, &RequestOptions{Body: body}, response); err != nil {
		return nil, err
	}
	return fromTokenResponse(response), nil
}

func (c *defaultClient) makeURL(path string) string {
	return fmt.Sprintf("%s%s", c.host, path)
}

func (c *defaultClient) basicAuth() string {
	auth := fmt.Sprintf("%s:", c.secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return fmt.Sprintf("Basic %s", encoded)
}
