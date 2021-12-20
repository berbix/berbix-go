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
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sdkVersion     = "1.0.0"
	clockDrift     = 300
	v0Transactions = "/v0/transactions"
)

type Client interface {
	CreateTransaction(options *CreateTransactionOptions) (*Tokens, error)
	CreateHostedTransaction(options *CreateHostedTransactionOptions) (*CreateHostedTransactionResponse, error)
	CreateAPIOnlyTransaction(options *CreateAPIOnlyTransactionOptions) (*CreateAPIOnlyTransactionResponse, error)
	RefreshTokens(tokens *Tokens) (*Tokens, error)
	FetchTransaction(tokens *Tokens) (*TransactionMetadata, error)
	DeleteTransaction(tokens *Tokens) error
	UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error)
	OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error
	ValidateSignature(secret, body, header string) error
	UploadImage(image []byte, subject ImageSubject, format ImageFormat, tokens *Tokens) (*ImageUploadResponse, error)
}

type defaultClient struct {
	secret string
	host   string
	client HTTPClient
}

type ClientOptions struct {
	Host       string
	HTTPClient HTTPClient
}

func NewClient(secret string, options *ClientOptions) Client {
	client := options.HTTPClient
	if client == nil {
		client = &DefaultHTTPClient{client: http.DefaultClient}
	}
	host := options.Host
	if host == "" {
		host = "https://api.berbix.com"
	}
	return &defaultClient{
		secret: secret,
		host:   host,
		client: client,
	}
}

func (c *defaultClient) CreateTransaction(options *CreateTransactionOptions) (*Tokens, error) {
	if options == nil {
		return nil, errors.New("options cannot be nil")
	}
	return c.fetchTokens(v0Transactions, options)
}

func (c *defaultClient) CreateHostedTransaction(options *CreateHostedTransactionOptions) (*CreateHostedTransactionResponse, error) {
	if options == nil {
		return nil, errors.New("options cannot be nil")
	}
	response := &hostedTransactionResponse{}
	if err := c.postBasicAuth(v0Transactions, options, response); err != nil {
		return nil, err
	}

	tokens := fromTokenResponse(&response.tokenResponse)
	return &CreateHostedTransactionResponse{
		Tokens:    *tokens,
		HostedURL: response.HostedURL,
	}, nil
}

func (c *defaultClient) CreateAPIOnlyTransaction(options *CreateAPIOnlyTransactionOptions) (*CreateAPIOnlyTransactionResponse, error) {
	if options == nil {
		return nil, errors.New("options cannot be nil")
	}
	response := &tokenResponse{}
	if err := c.postBasicAuth(v0Transactions, options, response); err != nil {
		return nil, err
	}

	tokens := fromTokenResponse(response)
	return &CreateAPIOnlyTransactionResponse{
		Tokens: *tokens,
	}, nil
}

func (c *defaultClient) RefreshTokens(tokens *Tokens) (*Tokens, error) {
	if tokens == nil {
		return nil, errors.New("tokens cannot be nil")
	}
	payload := &struct {
		RefreshToken string `json:"refresh_token"`
		GrantType    string `json:"grant_type"`
	}{
		RefreshToken: tokens.RefreshToken,
		GrantType:    "refresh_token",
	}
	return c.fetchTokens("/v0/tokens", payload)
}

func (c *defaultClient) FetchTransaction(tokens *Tokens) (*TransactionMetadata, error) {
	if tokens == nil {
		return nil, errors.New("tokens cannot be nil")
	}
	metadata := &TransactionMetadata{}
	return metadata, c.tokenAuthRequest(http.MethodGet, tokens, v0Transactions, nil, metadata)
}

func (c *defaultClient) DeleteTransaction(tokens *Tokens) error {
	if tokens == nil {
		return errors.New("tokens cannot be nil")
	}
	return c.tokenAuthRequest(http.MethodDelete, tokens, v0Transactions, nil, nil)
}

func (c *defaultClient) UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error) {
	if tokens == nil {
		return nil, errors.New("tokens cannot be nil")
	}

	if options == nil {
		return nil, errors.New("options cannot be nil")
	}

	metadata := &TransactionMetadata{}
	return metadata, c.tokenAuthRequest(http.MethodPatch, tokens, v0Transactions, options, metadata)
}

func (c *defaultClient) OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error {
	return c.tokenAuthRequest(http.MethodPatch, tokens, "/v0/transactions/override", options, nil)
}

func (c *defaultClient) UploadImage(image []byte, subject ImageSubject, format ImageFormat, tokens *Tokens) (*ImageUploadResponse, error) {
	encoded := base64.StdEncoding.EncodeToString(image)
	log.Printf("encoded image is %d bytes\n", len(encoded))
	req := &ImageUploadRequest{
		Image: ImageData{
			ImageSubject: subject,
			Format:       format,
			Data:         encoded,
		},
	}

	resp := &ImageUploadResponse{}
	// TODO this should cover error cases
	return resp, c.tokenAuthRequest(http.MethodPost, tokens, "/v0/images/upload", req, resp)
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
	if timestamp < time.Now().Unix()-clockDrift {
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
		"Content-Type":  "application/json",
		"User-Agent":    fmt.Sprintf("BerbixGo/%s", sdkVersion),
		"Authorization": fmt.Sprintf("Bearer %s", tokens.AccessToken),
	}
	return c.client.Request(method, c.makeURL(path), headers, &RequestOptions{Body: body}, dst)
}

func (c *defaultClient) fetchTokens(path string, payload interface{}) (*Tokens, error) {
	response := &tokenResponse{}
	if err := c.postBasicAuth(path, payload, response); err != nil {
		return nil, err
	}

	return fromTokenResponse(response), nil
}

func (c *defaultClient) postBasicAuth(path string, payload interface{}, dst interface{}) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"User-Agent":    fmt.Sprintf("BerbixGo/%s", sdkVersion),
		"Authorization": c.basicAuth(),
	}
	return c.client.Request(http.MethodPost, c.makeURL(path), headers, &RequestOptions{Body: body}, dst)
}

func (c *defaultClient) makeURL(path string) string {
	return fmt.Sprintf("%s%s", c.host, path)
}

func (c *defaultClient) basicAuth() string {
	auth := fmt.Sprintf("%s:", c.secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return fmt.Sprintf("Basic %s", encoded)
}
