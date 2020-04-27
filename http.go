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
	Request(method string, url string, headers map[string]string, options *RequestOptions, dst interface{}) error
}

type DefaultHTTPClient struct {
	client *http.Client
}

func (d *DefaultHTTPClient) Request(method string, url string, headers map[string]string, options *RequestOptions, dst interface{}) (err error) {
	req, err := http.NewRequest(method, url, options.Body)
	if err != nil {
		return
	}

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	res, err := d.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		err = res.Body.Close()
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