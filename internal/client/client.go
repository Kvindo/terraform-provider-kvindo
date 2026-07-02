package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ModificationResponse is returned from PUT and DELETE operations.
type ModificationResponse struct {
	RequestId    string `json:"requestId"`
	ResourceId   string `json:"resourceId"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// RequestStatusResponse is returned from polling the async status endpoint.
// NOTE: the API previously misspelled this field as "succeded"; it was corrected
// to "succeeded" API-wide. This json tag must match the current API field exactly.
type RequestStatusResponse struct {
	Succeeded            bool   `json:"succeeded"`
	ScheduledResourceId string `json:"scheduledResourceId"`
	ErrorCode           string `json:"errorCode"`
	ErrorMessage        string `json:"errorMessage"`
}

// Client is an HTTP client for the Kvindo Cloud API.
type Client struct {
	BaseURL    string
	Token      string
	Version    string
	HTTPClient *http.Client
}

// New creates a new Kvindo API client.
func New(baseURL, token, version string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		Version: version,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		tflog.Debug(ctx, "API request", map[string]interface{}{"method": method, "url": url, "body": string(data)})
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-kvindo/"+c.Version)
	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) ([]byte, int, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}
	tflog.Debug(ctx, "API response", map[string]interface{}{"status": resp.StatusCode, "body": string(data)})
	return data, resp.StatusCode, nil
}

func (c *Client) put(ctx context.Context, path string, body interface{}) (*ModificationResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	data, statusCode, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("PUT %s returned status %d: %s", path, statusCode, string(data))
	}

	var result ModificationResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling PUT response: %w (body: %s)", err, string(data))
	}

	if result.ErrorCode != "" {
		return nil, fmt.Errorf("API error %s: %s", result.ErrorCode, result.ErrorMessage)
	}

	return &result, nil
}

// WaitUntilNotReconciling polls Get until the resource exits the Reconciling state.
func (c *Client) WaitUntilNotReconciling(ctx context.Context, path, id string) error {
	deadline := time.Now().Add(30 * time.Minute)
	backoff := 2 * time.Second

	for time.Now().Before(deadline) {
		data, err := c.Get(ctx, path, id)
		if err != nil {
			return err
		}
		if data != nil {
			state := ""
			if s, ok := data["state"].(string); ok {
				state = s
			} else if status, ok := data["status"].(map[string]interface{}); ok {
				if s, ok := status["state"].(string); ok {
					state = s
				}
			} else if info, ok := data["info"].(map[string]interface{}); ok {
				// ponytail: keep old "info" key as fallback during API transition
				if s, ok := info["state"].(string); ok {
					state = s
				}
			}
			if !strings.HasPrefix(state, "Reconcil") {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < 30*time.Second {
			backoff += 2 * time.Second
		}
	}

	return fmt.Errorf("timed out waiting for %s/%s to exit Reconciling state", path, id)
}

// Put sends a PUT request to create or update a resource.
// If the resource is currently reconciling (ResourceIsScheduling), it waits for it to
// settle and retries automatically — handles Ctrl+C interrupted applies cleanly.
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*ModificationResponse, error) {
	for {
		result, err := c.put(ctx, path, body)
		if err == nil {
			return result, nil
		}
		if !strings.Contains(err.Error(), "ResourceIsScheduling") {
			return nil, err
		}

		id := ""
		if m, ok := body.(map[string]interface{}); ok {
			if meta, ok := m["metadata"].(map[string]interface{}); ok {
				if s, ok := meta["id"].(string); ok {
					id = s
				}
			}
		}
		if id == "" {
			return nil, err
		}

		if waitErr := c.WaitUntilNotReconciling(ctx, path, id); waitErr != nil {
			return nil, fmt.Errorf("waiting for resource to settle before retry: %w", waitErr)
		}
	}
}

// Get fetches a resource by ID.
func (c *Client) Get(ctx context.Context, path string, id string) (map[string]interface{}, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	data, statusCode, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	// This API signals "does not exist" with 422 + errorCode "NotFound" (not 404). Treat that as
	// not-found so Read removes the resource from state instead of erroring — otherwise an
	// out-of-band-deleted resource permanently blocks refresh/plan.
	if statusCode == 422 {
		var env map[string]interface{}
		if json.Unmarshal(data, &env) == nil {
			if ec, _ := env["errorCode"].(string); ec == "NotFound" {
				return nil, nil
			}
		}
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("GET %s/%s returned status %d: %s", path, id, statusCode, string(data))
	}

	var envelope map[string]interface{}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling GET response: %w (body: %s)", err, string(data))
	}

	// All GET responses wrap the resource in a "resource" key.
	if resource, ok := envelope["resource"].(map[string]interface{}); ok {
		return resource, nil
	}

	return envelope, nil
}

// Delete sends a DELETE request for a resource.
func (c *Client) Delete(ctx context.Context, path string, id string) (*ModificationResponse, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, path+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	data, statusCode, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return &ModificationResponse{}, nil
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("DELETE %s/%s returned status %d: %s", path, id, statusCode, string(data))
	}

	var result ModificationResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling DELETE response: %w (body: %s)", err, string(data))
	}

	if result.ErrorCode != "" {
		return nil, fmt.Errorf("API error %s: %s", result.ErrorCode, result.ErrorMessage)
	}

	return &result, nil
}

// PollUntilDone polls the async request status endpoint until the operation succeeds or times out.
func (c *Client) PollUntilDone(ctx context.Context, path string, requestId string) error {
	if requestId == "" {
		return nil
	}

	deadline := time.Now().Add(30 * time.Minute)
	backoff := 2 * time.Second

	for time.Now().Before(deadline) {
		pollPath := path + "/request/" + requestId
		req, err := c.newRequest(ctx, http.MethodGet, pollPath, nil)
		if err != nil {
			return err
		}

		data, statusCode, err := c.do(ctx, req)
		if err != nil {
			return err
		}

		if statusCode >= 400 {
			return fmt.Errorf("polling %s returned status %d: %s", pollPath, statusCode, string(data))
		}

		var status RequestStatusResponse
		if err := json.Unmarshal(data, &status); err != nil {
			return fmt.Errorf("unmarshaling poll response: %w (body: %s)", err, string(data))
		}

		if status.ErrorCode != "" {
			return fmt.Errorf("async operation error %s: %s", status.ErrorCode, status.ErrorMessage)
		}

		if status.Succeeded {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < 30*time.Second {
			backoff += 2 * time.Second
		}
	}

	return fmt.Errorf("timed out waiting for operation on %s (requestId: %s)", path, requestId)
}

// GetByLabels fetches resources filtered by labels.
func (c *Client) GetByLabels(ctx context.Context, path string, labels map[string]string) ([]map[string]interface{}, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path+"/get-by-labels", nil)
	if err != nil {
		return nil, err
	}

	if len(labels) > 0 {
		q := req.URL.Query()
		for k, v := range labels {
			q.Set("label."+k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	data, statusCode, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("GET %s/get-by-labels returned status %d: %s", path, statusCode, string(data))
	}

	// get-by-labels returns the same {"resources": [...], "pagination": {...}} envelope as every
	// other list endpoint, not a bare JSON array — this was never caught because a separate
	// datasource-side bug (metadata null-conversion) always crashed before any list response ever
	// reached this unmarshal.
	var envelope struct {
		Resources []map[string]interface{} `json:"resources"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling list response: %w (body: %s)", err, string(data))
	}

	return envelope.Resources, nil
}

// GetByName fetches a single resource by its metadata.name. It lists all resources of the type and
// filters by name client-side. Errors if zero or more than one resource matches (names are not
// guaranteed unique, so the caller should fall back to id in that case).
func (c *Client) GetByName(ctx context.Context, path string, name string) (map[string]interface{}, error) {
	items, err := c.GetByLabels(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var matches []map[string]interface{}
	for _, it := range items {
		res := it
		if r, ok := it["resource"].(map[string]interface{}); ok {
			res = r
		}
		meta, _ := res["metadata"].(map[string]interface{})
		if meta == nil {
			continue
		}
		if n, _ := meta["name"].(string); n == name {
			matches = append(matches, res)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no resource found at %s with name %q", path, name)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple resources at %s named %q — use id instead", path, name)
	}
	return matches[0], nil
}
