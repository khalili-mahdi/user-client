package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var ErrNotFound = &APIError{Message: "not found", Code: "NOT_FOUND"}

type Client struct {
	internal string
	external string
	net      *http.Client
}

func New(internal, external string) *Client {
	c := &Client{internal: internal, external: external}
	c.net = &http.Client{Timeout: time.Second * 30}
	return c
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.net.Timeout = timeout
	return c
}

func (c *Client) SetNet(net *http.Client) *Client {
	c.net = net
	return c
}

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	TraceID string `json:"trace_id"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.TraceID)
}

func (c *Client) apiCall(ctx context.Context, url string, method string, body io.Reader, response any) (int, error) {

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.net.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return resp.StatusCode, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent &&
		resp.StatusCode != http.StatusAccepted {
		var e APIError
		if err = json.NewDecoder(resp.Body).Decode(&e); err == nil {
			return resp.StatusCode, &e
		}
		return resp.StatusCode, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	if response == nil {
		return resp.StatusCode, nil
	}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return resp.StatusCode, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.StatusCode, nil
}

func (c *Client) httpCall(url, method string, headers map[string]string, req io.Reader, decodeTo any) error {
	r, err := http.NewRequest(method, url, req)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	r.Header.Set("content-type", "application/json")
	res, err := c.net.Do(r)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 300 {
		return StatusErr(res.StatusCode, res.Body)
	}
	if decodeTo == nil {
		return nil
	}
	if err = json.NewDecoder(res.Body).Decode(decodeTo); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}
	return nil
}

func StatusErr(statusCode int, body io.ReadCloser) error {
	var apiErr APIError
	if err := json.NewDecoder(body).Decode(&apiErr); err != nil {
		return fmt.Errorf("failed to decode status err: %v", err)
	}
	apiErr.Code = strconv.Itoa(statusCode)
	return &apiErr
}
