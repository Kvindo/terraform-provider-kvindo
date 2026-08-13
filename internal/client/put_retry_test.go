package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// Regression coverage for Put()'s two-error-code branching: SubmitLockBusy (org submit lock busy,
// nothing created - retry the ORIGINAL request) vs ResourceIsScheduling (the target resource itself
// is mid-reconcile - poll it). These used to be indistinguishable (both matched the same
// "ResourceIsScheduling" substring check), which made a lock-busy rejection get treated as "poll a
// resource id that was never created", timing out after 30 minutes with a misleading error.

type recordedRequest struct {
	Method string
	Path   string
}

type fakeResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	// Err, if set, makes RoundTrip return (nil, Err) instead of a real response - simulates a
	// transport-level failure (client timeout, connection reset, DNS blip). http.Client.Do wraps
	// whatever RoundTrip returns here in *url.Error automatically, same as a real Transport would.
	Err error
	// BodyErrors, if true, gives the response a Body that fails on Read() - simulates a connection
	// dropped mid-response (headers already received, body read fails).
	BodyErrors bool
	// Repeat, if true, does not advance callIndex past this entry - it's returned for every
	// subsequent call too. Used to simulate a persistent failure without needing to pre-size the
	// responses slice to an exact retry count.
	Repeat bool
}

// errorReader simulates a connection dropped mid-response: Read always fails.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (e *errorReader) Close() error               { return nil }

// fakeRoundTripper returns canned responses in order and records every request it receives, so
// tests can assert on the exact HTTP call sequence Put() produces - not just its final return value.
type fakeRoundTripper struct {
	t         *testing.T
	responses []fakeResponse
	calls     []recordedRequest
	callIndex int
}

func (f *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls = append(f.calls, recordedRequest{Method: req.Method, Path: req.URL.Path})

	// A real Transport checks the request's context and fails fast on an already-cancelled/expired
	// one rather than proceeding - mirror that here so tests can exercise Put()'s handling of a
	// pre-cancelled context without needing a real network stack.
	if err := req.Context().Err(); err != nil {
		return nil, err
	}

	if f.callIndex >= len(f.responses) {
		f.t.Fatalf("fakeRoundTripper: no more canned responses, but got %s %s (call #%d)", req.Method, req.URL.Path, len(f.calls))
	}
	resp := f.responses[f.callIndex]
	if !resp.Repeat {
		f.callIndex++
	}

	if resp.Err != nil {
		return nil, resp.Err
	}

	header := make(http.Header)
	for k, v := range resp.Headers {
		header.Set(k, v)
	}

	var body io.ReadCloser = io.NopCloser(bytes.NewReader([]byte(resp.Body)))
	if resp.BodyErrors {
		body = &errorReader{}
	}

	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Header:     header,
	}, nil
}

func newTestClient(rt *fakeRoundTripper) *Client {
	return &Client{
		BaseURL: "http://test.local",
		Token:   "test-token",
		Version: "test",
		HTTPClient: &http.Client{
			Transport: rt,
			Timeout:   5 * time.Second,
		},
	}
}

func testBody(id string) map[string]interface{} {
	return map[string]interface{}{"metadata": map[string]interface{}{"id": id, "name": "test"}}
}

func TestPut_SubmitLockBusy_RetriesOriginalRequest(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"SubmitLockBusy","errorMessage":"busy"}`, Headers: map[string]string{"Retry-After": "1"}},
		{StatusCode: 200, Body: `{"requestId":"req1","resourceId":"res1"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Put(context.Background(), "/api/v1/folder", testBody("abc123"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ResourceId != "res1" {
		t.Errorf("expected resourceId res1, got %s", result.ResourceId)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d: %+v", len(rt.calls), rt.calls)
	}
	for i, call := range rt.calls {
		if call.Method != http.MethodPut || call.Path != "/api/v1/folder" {
			t.Errorf("call %d: expected PUT /api/v1/folder (retry of the ORIGINAL request), got %s %s", i, call.Method, call.Path)
		}
	}
}

// Direct coverage that RetryAfterSeconds is parsed from the real Retry-After HTTP header, not a
// JSON body field - the other tests only exercise this indirectly through timing behavior.
func TestPut_ParsesRetryAfterFromHeader_NotBody(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"SubmitLockBusy","errorMessage":"busy"}`, Headers: map[string]string{"Retry-After": "7"}},
	}}
	c := newTestClient(rt)

	_, err := c.put(context.Background(), "/api/v1/folder", testBody("abc123"))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a structured ApiError, got: %v", err)
	}
	if apiErr.RetryAfterSeconds == nil {
		t.Fatal("expected RetryAfterSeconds to be parsed from the Retry-After header, got nil")
	}
	if *apiErr.RetryAfterSeconds != 7 {
		t.Errorf("expected RetryAfterSeconds=7, got %d", *apiErr.RetryAfterSeconds)
	}
}

// A response with no Retry-After header at all (e.g. a different error code, or an older server)
// must not crash header parsing and must leave the field nil.
func TestPut_NoRetryAfterHeader_LeavesFieldNil(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"ResourceIsScheduling","errorMessage":"scheduling"}`},
	}}
	c := newTestClient(rt)

	_, err := c.put(context.Background(), "/api/v1/folder", testBody("abc123"))
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a structured ApiError, got: %v", err)
	}
	if apiErr.RetryAfterSeconds != nil {
		t.Errorf("expected RetryAfterSeconds to be nil with no header present, got %d", *apiErr.RetryAfterSeconds)
	}
}

func TestPut_ResourceIsScheduling_PollsResourceId(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"ResourceIsScheduling","errorMessage":"still scheduling"}`},
		{StatusCode: 200, Body: `{"resource":{"state":"stable"}}`},
		{StatusCode: 200, Body: `{"requestId":"req2","resourceId":"res2"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Put(context.Background(), "/api/v1/folder", testBody("xyz789"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ResourceId != "res2" {
		t.Errorf("expected resourceId res2, got %s", result.ResourceId)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("expected 3 calls (PUT, GET, PUT), got %d: %+v", len(rt.calls), rt.calls)
	}
	if rt.calls[0].Method != http.MethodPut {
		t.Errorf("call 0: expected PUT, got %s", rt.calls[0].Method)
	}
	if rt.calls[1].Method != http.MethodGet || rt.calls[1].Path != "/api/v1/folder/xyz789" {
		t.Errorf("call 1: expected GET /api/v1/folder/xyz789 (poll on the submitted id), got %s %s", rt.calls[1].Method, rt.calls[1].Path)
	}
	if rt.calls[2].Method != http.MethodPut {
		t.Errorf("call 2: expected PUT (retry after settling), got %s", rt.calls[2].Method)
	}
}

func TestPut_NonJSONError_ReturnsFallbackWithoutRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 502, Body: `<html>Bad Gateway</html>`},
	}}
	c := newTestClient(rt)

	_, err := c.Put(context.Background(), "/api/v1/folder", testBody("abc"))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		t.Errorf("expected a plain fallback error (unparseable body), got a structured ApiError: %v", apiErr)
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (non-retryable error, no retry loop entered), got %d", len(rt.calls))
	}
}

// The scenario that justifies Put()'s single shared outer deadline: a request can hit BOTH retry
// reasons in sequence, not just one in isolation. Asserts the exact call sequence, not just the
// final result, so a wrong branch taken in the middle can't accidentally still produce success.
func TestPut_AlternatingErrorCodes_ComposesCorrectly(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"SubmitLockBusy","errorMessage":"busy"}`, Headers: map[string]string{"Retry-After": "1"}},
		{StatusCode: 422, Body: `{"errorCode":"ResourceIsScheduling","errorMessage":"scheduling"}`},
		{StatusCode: 200, Body: `{"resource":{"state":"stable"}}`},
		{StatusCode: 200, Body: `{"requestId":"req3","resourceId":"res3"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Put(context.Background(), "/api/v1/folder", testBody("alt123"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ResourceId != "res3" {
		t.Errorf("expected resourceId res3, got %s", result.ResourceId)
	}

	wantSequence := []recordedRequest{
		{Method: http.MethodPut, Path: "/api/v1/folder"},
		{Method: http.MethodPut, Path: "/api/v1/folder"},
		{Method: http.MethodGet, Path: "/api/v1/folder/alt123"},
		{Method: http.MethodPut, Path: "/api/v1/folder"},
	}
	if len(rt.calls) != len(wantSequence) {
		t.Fatalf("expected %d calls, got %d: %+v", len(wantSequence), len(rt.calls), rt.calls)
	}
	for i, want := range wantSequence {
		if rt.calls[i] != want {
			t.Errorf("call %d: expected %+v, got %+v", i, want, rt.calls[i])
		}
	}
}

// Regression coverage for Put()'s transport-error retry branch (*url.Error - client timeout,
// connection reset, DNS blip, or a dropped connection mid-response). Live-observed: a folder PUT
// succeeded server-side in ~30s+ while the client's HTTP timeout fired first, orphaning the resource
// because Terraform never recorded its id. Safe to retry the original request by the same
// idempotency argument SubmitLockBusy's retry already relies on.

func TestPut_TransportError_RetriesOriginalRequest(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{Err: context.DeadlineExceeded},
		{StatusCode: 200, Body: `{"requestId":"req4","resourceId":"res4"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Put(context.Background(), "/api/v1/folder", testBody("neterr1"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ResourceId != "res4" {
		t.Errorf("expected resourceId res4, got %s", result.ResourceId)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d: %+v", len(rt.calls), rt.calls)
	}
	for i, call := range rt.calls {
		if call.Method != http.MethodPut || call.Path != "/api/v1/folder" {
			t.Errorf("call %d: expected PUT /api/v1/folder (retry of the ORIGINAL request), got %s %s", i, call.Method, call.Path)
		}
	}
}

// Local/permanent errors (a request body that can't be marshaled) must never enter the retry loop -
// they happen in newRequest(), before any HTTP call is attempted, and will never succeed on retry.
func TestPut_MarshalError_NoRetryZeroCalls(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: `{}`},
	}}
	c := newTestClient(rt)

	body := map[string]interface{}{"metadata": map[string]interface{}{"id": "x", "bad": make(chan int)}}
	_, err := c.Put(context.Background(), "/api/v1/folder", body)
	if err == nil {
		t.Fatal("expected a marshal error, got nil")
	}
	if len(rt.calls) != 0 {
		t.Errorf("expected zero HTTP calls (marshal failure happens before any request is sent), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// A 2xx response with a malformed body is a different code path in put() than the existing
// TestPut_NonJSONError_ReturnsFallbackWithoutRetry (which covers a >=400 malformed body) - this is
// not a *url.Error or *ApiError, so it must not be retried either.
func TestPut_Malformed2xxJSON_NoRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: `not json`},
	}}
	c := newTestClient(rt)

	_, err := c.Put(context.Background(), "/api/v1/folder", testBody("badjson"))
	if err == nil {
		t.Fatal("expected an unmarshal error, got nil")
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		t.Errorf("expected a plain unmarshal error, not a *url.Error: %v", err)
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (a malformed 2xx body is not retried), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// Locks in the do() fix wrapping a body-read failure (connection dropped after headers, mid-body) as
// *url.Error too - the same underlying network-flake family as a pre-header timeout, and just as
// safe to retry, but previously fell through as a plain fmt.Errorf with no retry at all.
func TestPut_BodyReadError_Retries(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, BodyErrors: true},
		{StatusCode: 200, Body: `{"requestId":"req5","resourceId":"res5"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Put(context.Background(), "/api/v1/folder", testBody("bodyerr1"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ResourceId != "res5" {
		t.Errorf("expected resourceId res5, got %s", result.ResourceId)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("expected 2 calls (body-read failure retried like a transport error), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// Proves the short parent deadline actually bounds Put()'s derived context.WithTimeout(ctx,
// 10*time.Minute) to the shorter of the two, rather than being silently ignored - a persistent
// transport failure must give up promptly, not hang for the full 10 minutes.
func TestPut_OuterDeadlineExhaustion_GivesUpPromptly(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{Err: context.DeadlineExceeded, Repeat: true},
	}}
	c := newTestClient(rt)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.Put(ctx, "/api/v1/folder", testBody("deadline1"))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error (persistent transport failure), got nil")
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected Put() to give up promptly after the short parent deadline, took %v", elapsed)
	}
	if len(rt.calls) == 0 {
		t.Error("expected at least one call before giving up")
	}
}

// A pre-cancelled context must not be retried. c.put(ctx, ...) is called unconditionally as the
// first statement in Put()'s loop body, before ctx.Err() is ever checked - so exactly one call
// happens, and the ctx.Err() != nil check right after it (not the context.Canceled exclusion in the
// transport-error branch, which is never even reached here) is what stops it from retrying.
func TestPut_CancelledContext_MakesOneCallNoRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: `{"requestId":"req6","resourceId":"res6"}`},
	}}
	c := newTestClient(rt)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before Put() is ever called

	_, err := c.Put(ctx, "/api/v1/folder", testBody("cancelled1"))
	if err == nil {
		t.Fatal("expected an error for a pre-cancelled context, got nil")
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (c.put is called before ctx.Err() is checked) with no retry, got %d: %+v", len(rt.calls), rt.calls)
	}
}

// Extends TestPut_AlternatingErrorCodes_ComposesCorrectly to all three retryable failure modes in
// one sequence (transport error, SubmitLockBusy, transport error, ResourceIsScheduling, success),
// asserting the exact call sequence so a wrong branch taken partway through can't accidentally still
// produce the right final result.
func TestPut_ThreeErrorModes_ComposesCorrectly(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{Err: context.DeadlineExceeded},
		{StatusCode: 422, Body: `{"errorCode":"SubmitLockBusy","errorMessage":"busy"}`, Headers: map[string]string{"Retry-After": "1"}},
		{Err: context.DeadlineExceeded},
		{StatusCode: 422, Body: `{"errorCode":"ResourceIsScheduling","errorMessage":"scheduling"}`},
		{StatusCode: 200, Body: `{"resource":{"state":"stable"}}`},
		{StatusCode: 200, Body: `{"requestId":"req7","resourceId":"res7"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Put(context.Background(), "/api/v1/folder", testBody("three1"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ResourceId != "res7" {
		t.Errorf("expected resourceId res7, got %s", result.ResourceId)
	}

	wantSequence := []recordedRequest{
		{Method: http.MethodPut, Path: "/api/v1/folder"},
		{Method: http.MethodPut, Path: "/api/v1/folder"},
		{Method: http.MethodPut, Path: "/api/v1/folder"},
		{Method: http.MethodPut, Path: "/api/v1/folder"},
		{Method: http.MethodGet, Path: "/api/v1/folder/three1"},
		{Method: http.MethodPut, Path: "/api/v1/folder"},
	}
	if len(rt.calls) != len(wantSequence) {
		t.Fatalf("expected %d calls, got %d: %+v", len(wantSequence), len(rt.calls), rt.calls)
	}
	for i, want := range wantSequence {
		if rt.calls[i] != want {
			t.Errorf("call %d: expected %+v, got %+v", i, want, rt.calls[i])
		}
	}
}
