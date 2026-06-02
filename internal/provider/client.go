package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const orgHeader = "X-Dollarbox-Org"

type APIClient struct {
	endpoint   string
	token      string
	org        string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Body       string
}

func NewAPIClient(config ClientConfig) *APIClient {
	return newAPIClient(config, &http.Client{Timeout: 30 * time.Second})
}

func newAPIClient(config ClientConfig, httpClient *http.Client) *APIClient {
	return &APIClient{
		endpoint:   strings.TrimRight(config.Endpoint, "/"),
		token:      config.Token,
		org:        config.Org,
		httpClient: httpClient,
	}
}

func (c *APIClient) do(
	ctx context.Context,
	method string,
	path string,
	requestBody any,
	responseBody any,
) error {
	var body io.Reader
	if requestBody != nil {
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.org != "" {
		req.Header.Set(orgHeader, c.org)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return decodeAPIError(res)
	}
	if res.StatusCode == http.StatusNoContent || responseBody == nil {
		_, _ = io.Copy(io.Discard, res.Body)
		return nil
	}
	if err := json.NewDecoder(res.Body).Decode(responseBody); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

func decodeAPIError(res *http.Response) error {
	data, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return fmt.Errorf("DollarBox API returned HTTP %d and the error body could not be read: %w", res.StatusCode, readErr)
	}
	apiErr := &APIError{StatusCode: res.StatusCode, Body: strings.TrimSpace(string(data))}

	var envelope struct {
		Error struct {
			Code    string          `json:"code"`
			Message string          `json:"message"`
			Detail  json.RawMessage `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil {
		apiErr.Code = envelope.Error.Code
		apiErr.Message = envelope.Error.Message
	}
	return apiErr
}

func (e *APIError) Error() string {
	if e.Code != "" && e.Message != "" {
		return fmt.Sprintf("DollarBox API returned %s (%d): %s", e.Code, e.StatusCode, e.Message)
	}
	if e.Body != "" {
		return fmt.Sprintf("DollarBox API returned HTTP %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("DollarBox API returned HTTP %d", e.StatusCode)
}

func isNotFoundError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func addAPIError(diags *diag.Diagnostics, summary string, err error) {
	if err == nil {
		return
	}
	diags.AddError(summary, err.Error())
}
