package berbix

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type RequestOptions struct {
	Body io.Reader
}

type HTTPClient interface {
	Request(method string, url string, headers map[string]string, options *RequestOptions) (*HTTPResponse, error)
}

type DefaultHTTPClient struct {
	client *http.Client
}

type HTTPResponse struct {
	StatusCode int
	Body       io.ReadCloser
	Headers    map[string][]string
}

func (d *DefaultHTTPClient) Request(method string, url string, headers map[string]string, options *RequestOptions) (*HTTPResponse, error) {
	req, err := http.NewRequest(method, url, options.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing HTTP request: %v", err)
	}

	return &HTTPResponse{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Headers:    res.Header,
	}, nil
}

func requestExpecting2XX(c HTTPClient, method string, url string, headers map[string]string, options *RequestOptions, dst interface{}) (err error) {
	res, err := c.Request(method, url, headers, options)
	if err != nil {
		return err
	}
	defer func() {
		bodyCloseErr := res.Body.Close()
		if err == nil && bodyCloseErr != nil {
			err = bodyCloseErr
		}
	}()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("non-2XX response from Berbix backend %d", res.StatusCode)
	}

	if res.StatusCode != http.StatusNoContent {
		if dst != nil {
			if err = json.NewDecoder(res.Body).Decode(dst); err != nil {
				return
			}
		} else {
			return errors.New("received a non-204 response but a nil destination")
		}
	}

	return
}
