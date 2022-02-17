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
	UploadImages(tokens *Tokens, options *UploadImagesOptions) (*ImageUploadResult, error)
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
	return metadata, c.tokenAuthRequestExpecting2XX(http.MethodGet, tokens, v0Transactions, nil, metadata)
}

func (c *defaultClient) DeleteTransaction(tokens *Tokens) error {
	if tokens == nil {
		return errors.New("tokens cannot be nil")
	}
	return c.tokenAuthRequestExpecting2XX(http.MethodDelete, tokens, v0Transactions, nil, nil)
}

func (c *defaultClient) UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error) {
	if tokens == nil {
		return nil, errors.New("tokens cannot be nil")
	}

	if options == nil {
		return nil, errors.New("options cannot be nil")
	}

	metadata := &TransactionMetadata{}
	return metadata, c.tokenAuthRequestExpecting2XX(http.MethodPatch, tokens, v0Transactions, options, metadata)
}

func (c *defaultClient) OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error {
	return c.tokenAuthRequestExpecting2XX(http.MethodPatch, tokens, "/v0/transactions/override", options, nil)
}

type RawImage struct {
	// Bytes representing the image. This should represent the image in a supported
	// format, such as JPEG or PNG, without extra encoding, such as base 64, encoding applied.
	Image   []byte
	Subject ImageSubject
	Format  ImageFormat
}

type UploadImagesOptions struct {
	Images []RawImage
}

// UploadImages uploads image(s) to Berbix and provides a response that indicates the next upload, if any,
// that is expected, along with flags indicating any issues with the image(s) that could immediately be detected.
// The options argument is required.
// Returns a InvalidStateErr if the upload was invalid for the current state of the transaction.
// Returns a TransactionDoesNotExistErr if the transaction for which the image is being uploaded no longer exists.
// Returns a PayloadTooLargeErr if the uploaded payload or underlying image is too large.
func (c *defaultClient) UploadImages(tokens *Tokens, options *UploadImagesOptions) (*ImageUploadResult, error) {
	if options == nil {
		return nil, errors.New("must specify non-nil UploadImageOptions")
	}
	if options.Images == nil {
		return nil, errors.New("must specify non-nil Images")
	}
	imageDatas := make([]ImageData, len(options.Images))
	for i, rawImage := range options.Images {
		encoded := base64.StdEncoding.EncodeToString(rawImage.Image)
		imageDatas[i] = ImageData{
			ImageSubject: rawImage.Subject,
			Format:       rawImage.Format,
			Data:         encoded,
		}
	}

	req := &ImageUploadRequest{
		Images: imageDatas,
	}

	httpResp, err := c.tokenAuthRequest(http.MethodPost, tokens, "/v0/images/upload", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	bodyDec := json.NewDecoder(httpResp.Body)
	switch httpResp.StatusCode {
	case http.StatusOK, http.StatusUnprocessableEntity:
		imageResp := ImageUploadResponse{}
		if err := bodyDec.Decode(&imageResp); err != nil {
			return nil, fmt.Errorf("error unmarshalling response: %v", err)
		}
		return &ImageUploadResult{
			IsAcceptableIDType:  httpResp.StatusCode != http.StatusUnprocessableEntity,
			ImageUploadResponse: imageResp,
		}, nil
	case http.StatusConflict:
		invalidStateRes := InvalidUploadForStateResponse{}
		if err := bodyDec.Decode(&invalidStateRes); err != nil {
			return nil, fmt.Errorf("got malformed response body for status code %d", httpResp.StatusCode)
		}

		return nil, InvalidStateErr{
			InvalidUploadForStateResponse: invalidStateRes,
		}
	case http.StatusGone:
		const defaultMsg = "The transaction for this upload does not exist. It may have been deleted."
		msgErr := makeErrorMessage(bodyDec, defaultMsg)
		return nil, TransactionDoesNotExistErr{errorMessage: msgErr}
	case http.StatusRequestEntityTooLarge:
		const defaultMsg = "Request body or image too large."
		msgErr := makeErrorMessage(bodyDec, defaultMsg)
		return nil, PayloadTooLargeErr{errorMessage: msgErr}
	default:
		return nil, GenericErr{
			StatusCode: httpResp.StatusCode,
			Message:    fmt.Sprintf("got unexpected response code %d", httpResp.StatusCode),
		}
	}
}

func makeErrorMessage(bodyDec *json.Decoder, defaultMsg string) errorMessage {
	var msg string
	genErr := GenericErrorResponse{}
	if err := bodyDec.Decode(&genErr); err == nil {
		msg = genErr.Message
	} else {
		msg = defaultMsg
	}
	return errorMessage{Message: msg}
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

func (c *defaultClient) tokenAuthRequestExpecting2XX(method string, tokens *Tokens, path string, payload interface{}, dst interface{}) error {
	body, headers, err := c.prepTokenAuthRequest(tokens, payload)
	if err != nil {
		return err
	}
	return requestExpecting2XX(c.client, method, c.makeURL(path), headers, &RequestOptions{Body: body}, dst)
}

func (c *defaultClient) tokenAuthRequest(method string, tokens *Tokens, path string, payload interface{}) (*HTTPResponse, error) {
	body, headers, err := c.prepTokenAuthRequest(tokens, payload)
	if err != nil {
		return nil, err
	}

	return c.client.Request(method, c.makeURL(path), headers, &RequestOptions{Body: body})
}

func (c *defaultClient) prepTokenAuthRequest(tokens *Tokens, payload interface{},
) (reqBody io.Reader, reqHeaders map[string]string, err error) {
	if tokens.NeedsRefresh() {
		tokens, err = c.RefreshTokens(tokens)
		if err != nil {
			return nil, nil, err
		}
	}
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, err
		}
		body = bytes.NewReader(data)
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"User-Agent":    fmt.Sprintf("BerbixGo/%s", sdkVersion),
		"Authorization": fmt.Sprintf("Bearer %s", tokens.AccessToken),
	}
	return body, headers, nil
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
	return requestExpecting2XX(c.client, http.MethodPost, c.makeURL(path), headers, &RequestOptions{Body: body}, dst)
}

func (c *defaultClient) makeURL(path string) string {
	return fmt.Sprintf("%s%s", c.host, path)
}

func (c *defaultClient) basicAuth() string {
	auth := fmt.Sprintf("%s:", c.secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return fmt.Sprintf("Basic %s", encoded)
}
