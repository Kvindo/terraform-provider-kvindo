package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
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

// ApiError is the structured shape of a Kvindo API error response (statusCode >= 400). ErrorCode/
// ErrorMessage are parsed best-effort from the JSON body in put(); callers use errors.As instead of
// string-matching err.Error(), which is fragile (a wording change anywhere silently breaks it).
// RetryAfterSeconds is NOT a body field - it's parsed from the real HTTP Retry-After response
// header (RFC 9110 delta-seconds form), set by put() alongside the body parsing.
type ApiError struct {
	StatusCode        int
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
	RetryAfterSeconds *int
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("API error %s (status %d): %s", e.ErrorCode, e.StatusCode, e.ErrorMessage)
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
			Timeout: 60 * time.Second,
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

// do returns the response body, status code, and response headers (needed by put() to read
// Retry-After - no other caller currently needs headers, but returning them uniformly here is
// simpler than a one-off variant just for put()).
func (c *Client) do(ctx context.Context, req *http.Request) ([]byte, int, http.Header, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		// Wrapped as *url.Error (the same type net/http's own Client.Do uses for header-phase
		// transport failures) rather than a plain fmt.Errorf, so a connection dropped mid-response
		// flows through the same retry classification in Put() as a pre-header timeout - both are
		// the same underlying network-flake family, and both are just as safe to retry. Unwrap()
		// still reaches the real io.ReadAll error via errors.Is/errors.As/%w, so nothing is lost.
		return nil, resp.StatusCode, resp.Header, &url.Error{Op: req.Method, URL: req.URL.String(), Err: err}
	}
	tflog.Debug(ctx, "API response", map[string]interface{}{"status": resp.StatusCode, "body": string(data)})
	return data, resp.StatusCode, resp.Header, nil
}

func (c *Client) put(ctx context.Context, path string, body interface{}) (*ModificationResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	data, statusCode, headers, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		var apiErr ApiError
		apiErr.StatusCode = statusCode
		// Set BEFORE the JSON unmarshal below - Unmarshal only touches fields present in the JSON,
		// so pre-setting this from the header (not a body field) survives untouched.
		if ra := headers.Get("Retry-After"); ra != "" {
			if seconds, convErr := strconv.Atoi(ra); convErr == nil {
				apiErr.RetryAfterSeconds = &seconds
			}
		}
		// Best-effort: if the body isn't the expected JSON shape (a 5xx from an intermediary, an
		// HTML error page, etc.), fall back to the plain-text error rather than losing the failure.
		if jsonErr := json.Unmarshal(data, &apiErr); jsonErr == nil && apiErr.ErrorCode != "" {
			return nil, &apiErr
		}
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
			// A transient Get failure doesn't mean the resource is gone or broken - tolerate it
			// and keep polling, same as PollUntilDone already does for its own poll loop. No
			// separate retry cap here either: the 30-minute deadline above (or Put()'s outer
			// 10-minute context, which now preempts it) is already the real ceiling.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			if backoff < 30*time.Second {
				backoff += 2 * time.Second
			}
			continue
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

// Put sends a PUT request to create or update a resource. Three distinct failure modes need
// different retry behavior:
//   - A transport-level error (*url.Error - client timeout, connection reset, DNS blip, or a
//     dropped connection mid-response): the request may have completed successfully server-side
//     even though we never saw the response - retry the ORIGINAL request (safe: Kvindo's
//     create/update is idempotent by client-generated ULID).
//   - SubmitLockBusy: the org's submit lock was busy - nothing was created, retry the ORIGINAL
//     request.
//   - ResourceIsScheduling: the TARGET resource itself exists and is mid-reconcile - poll it until
//     it settles, then retry (handles Ctrl+C interrupted applies cleanly, as before).
//
// Any other error (local/permanent failures like a request body that fails to marshal, or a
// response body that fails to unmarshal) is returned immediately, unretried.
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*ModificationResponse, error) {
	// A real ceiling, not just a between-attempts check: derive a deadline-bound context once and
	// use it (not the original ctx) for every inner call below. Since put()/WaitUntilNotReconciling
	// already thread ctx through to their own HTTP calls, this makes the 10-minute cap preemptive -
	// a long-running WaitUntilNotReconciling call gets cancelled at the 10-minute mark instead of
	// running to its own independent 30-minute budget.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	lockBusyAttempt := 0
	networkErrAttempt := 0

	for {
		result, err := c.put(ctx, path, body)
		if err == nil {
			return result, nil
		}
		if ctx.Err() != nil {
			tflog.Warn(ctx, "giving up on Put retry loop", map[string]interface{}{"path": path, "lastError": err.Error()})
			return nil, fmt.Errorf("giving up after retrying for 10m, last error: %w", err)
		}

		var urlErr *url.Error
		if errors.As(err, &urlErr) && !errors.Is(err, context.Canceled) {
			// Transport-level failure (client timeout, connection reset, DNS blip, or a dropped
			// connection mid-response per do()'s wrapping). The request may have completed
			// successfully server-side even though we never saw the response - this is exactly
			// what happened live: a folder PUT succeeded server-side in ~30s+ while the client's
			// HTTP timeout fired first, orphaning the resource because Terraform never recorded
			// its id. Safe to retry the ORIGINAL request because Kvindo's create/upsert is
			// idempotent by client-generated ULID (metadata.id already in the body) - a
			// byte-identical re-PUT against an existing id is a documented no-op that returns the
			// existing resource.
			//
			// The explicit context.Canceled exclusion is defensive documentation, not a
			// load-bearing guard: this branch is only reached when ctx.Err() == nil (checked
			// above), so a cancelled or outer-deadline-expired ctx always hits the "giving up"
			// path first. That same ordering invariant is also why the context.WithTimeout(ctx,
			// wait) call below can never spin with a zero-length wait - it's only ever reached
			// when ctx.Err() == nil, so wait is applied against a live, non-expired context.
			networkErrAttempt++
			multiplier := networkErrAttempt
			if multiplier > 5 {
				multiplier = 5
			}
			wait := time.Duration(rand.Int63n(int64(2 * time.Second * time.Duration(multiplier))))
			tflog.Debug(ctx, "transport error, retrying original request", map[string]interface{}{"path": path, "attempt": networkErrAttempt, "waitMs": wait.Milliseconds(), "error": err.Error()})
			waitCtx, waitCancel := context.WithTimeout(ctx, wait)
			<-waitCtx.Done()
			waitCancel()
			continue
		}

		var apiErr *ApiError
		if !errors.As(err, &apiErr) {
			return nil, err // local/permanent error (marshal/unmarshal failure) - not retryable
		}

		switch apiErr.ErrorCode {
		case "SubmitLockBusy":
			// Nothing was created - retry the ORIGINAL request, not a poll on some id. Base wait
			// comes from the server's hint (tunable server-side without a client release, and
			// clamped defensively - a buggy or malicious server sending an absurd value can't stall
			// the client for hours); grows per attempt (capped) so a sustained busy org backs off
			// instead of hammering at a fixed interval, with full jitter so many waiters on the same
			// busy lock spread out instead of clustering in a narrow window.
			base := 3 * time.Second
			if apiErr.RetryAfterSeconds != nil {
				hint := time.Duration(*apiErr.RetryAfterSeconds) * time.Second
				if hint >= 1*time.Second && hint <= 60*time.Second {
					base = hint
				}
			}
			lockBusyAttempt++
			multiplier := lockBusyAttempt
			if multiplier > 5 {
				multiplier = 5 // growth plateaus at 5x base after attempt 5; attempt counter itself is unbounded (only the wait is capped)
			}
			grown := base * time.Duration(multiplier)
			wait := time.Duration(rand.Int63n(int64(grown))) // full jitter: [0, grown)
			tflog.Debug(ctx, "SubmitLockBusy, retrying original request", map[string]interface{}{"path": path, "attempt": lockBusyAttempt, "waitMs": wait.Milliseconds()})
			// context.WithTimeout instead of a bare time.After: if ctx is cancelled first, the timer
			// is cleaned up via waitCancel() rather than leaking until wait naturally elapses.
			waitCtx, waitCancel := context.WithTimeout(ctx, wait)
			<-waitCtx.Done()
			waitCancel()
			continue

		case "ResourceIsScheduling":
			// The resource genuinely exists and is mid-transition - poll it, as before.
			id := extractId(body)
			if id == "" {
				tflog.Warn(ctx, "ResourceIsScheduling but could not extract id from submitted body - returning original error instead of polling", map[string]interface{}{"path": path})
				return nil, err
			}
			tflog.Debug(ctx, "ResourceIsScheduling, polling for settle", map[string]interface{}{"path": path, "id": id})
			if waitErr := c.WaitUntilNotReconciling(ctx, path, id); waitErr != nil {
				return nil, fmt.Errorf("waiting for resource %s to settle before retry: %w", id, waitErr)
			}
			continue

		default:
			return nil, err
		}
	}
}

// extractId pulls metadata.id out of the submitted request body - the id the client itself chose
// for this create/upsert. Behavior unchanged from the logic this was extracted from: silently
// returns "" for an unrecognized body shape, matching Put()'s existing pre-fix handling of that case.
func extractId(body interface{}) string {
	if m, ok := body.(map[string]interface{}); ok {
		if meta, ok := m["metadata"].(map[string]interface{}); ok {
			if s, ok := meta["id"].(string); ok {
				return s
			}
		}
	}
	return ""
}

// Get fetches a resource by ID.
func (c *Client) Get(ctx context.Context, path string, id string) (map[string]interface{}, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	data, statusCode, _, err := c.do(ctx, req)
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

	data, statusCode, _, err := c.do(ctx, req)
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
// The deadline is deliberately generous (not a tight per-resource estimate): dev's shared-environment
// reconciler contention (org-wide in-progress gates serializing all in-flight changes of a type -
// see e.g. OpenVpn/LoadBalancer/SecurityGroup reconcilers) can legitimately leave a real, healthy
// change request "Waiting for all changes to finish first" for a long time with nothing actually
// wrong. Since this only bounds a poll loop that returns immediately on success, a long ceiling
// costs nothing on the fast path - it only matters for a genuinely hung/broken backend, which is
// better diagnosed by querying the resource's real state directly than by a client timing out early.
func (c *Client) PollUntilDone(ctx context.Context, path string, requestId string) error {
	if requestId == "" {
		return nil
	}

	deadline := time.Now().Add(6 * time.Hour)
	backoff := 2 * time.Second

	for time.Now().Before(deadline) {
		pollPath := path + "/request/" + requestId
		req, err := c.newRequest(ctx, http.MethodGet, pollPath, nil)
		if err != nil {
			return err
		}

		// A transient network error (VPN blip, DNS hiccup, connection reset) on a single poll
		// attempt does NOT mean the operation failed - the backend is still working regardless.
		// Treat it the same as "not done yet" and retry with the loop's own backoff, same as the
		// C# test harness's KvindoCloudClient already does for its own read/poll calls. Bounded
		// by the outer deadline above, so a genuinely unreachable backend still eventually errors.
		data, statusCode, _, err := c.do(ctx, req)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			if backoff < 30*time.Second {
				backoff += 2 * time.Second
			}
			continue
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

	data, statusCode, _, err := c.do(ctx, req)
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
