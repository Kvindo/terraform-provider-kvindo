package client

import (
	"context"
	"strings"
	"testing"
)

// Regression coverage for review finding #15: PollUntilDone used to return on any statusCode >=
// 400 BEFORE ever parsing the body, so (a) a transient 5xx from an intermediary aborted the whole
// operation client-side while the backend kept working, and (b) a genuine business failure (e.g.
// UnableToReconcile) served as a 4xx WITH a real RequestStatusResponse-shaped body was shown as
// raw JSON instead of the intended clean "async operation error <code>: <message>". Reuses the
// fakeRoundTripper/newTestClient helpers from put_retry_test.go.

func TestPollUntilDone_Success(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: `{"succeeded":true}`},
	}}
	c := newTestClient(rt)

	if err := c.PollUntilDone(context.Background(), "/api/v1/vm", "req1"); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

// A genuine business failure (e.g. UnableToReconcile) is legitimately served as a 4xx WITH a real
// RequestStatusResponse envelope - this must produce the clean async-error message, not the raw
// status/body fallback, regardless of the 4xx status.
func TestPollUntilDone_ParseableErrorCodeAt4xx_ReturnsCleanAsyncError(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"succeeded":false,"errorCode":"UnableToReconcile","errorMessage":"could not attach volume"}`},
	}}
	c := newTestClient(rt)

	err := c.PollUntilDone(context.Background(), "/api/v1/vm", "req2")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	want := "async operation error UnableToReconcile: could not attach volume"
	if err.Error() != want {
		t.Errorf("expected clean async-error message %q, got %q", want, err.Error())
	}
}

// A parseable body at 5xx (the backend really did answer, just with an unusual empty-ErrorCode
// failure shape) must NOT be retried - only an unparseable 5xx (a likely intermediary) is.
func TestPollUntilDone_ParseableBodyNoErrorCodeAt5xx_TerminalNotRetried(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 503, Body: `{"succeeded":false}`},
	}}
	c := newTestClient(rt)

	err := c.PollUntilDone(context.Background(), "/api/v1/vm", "req3")
	if err == nil {
		t.Fatal("expected a terminal error, got nil")
	}
	if !strings.Contains(err.Error(), "status 503") {
		t.Errorf("expected the error to mention the status code, got %q", err.Error())
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (parseable body is not retried), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// An UNPARSEABLE body at 5xx (e.g. an HTML error page from an ingress) IS treated as transient and
// retried, matching the existing transport-error retry behavior just above this branch.
func TestPollUntilDone_UnparseableBodyAt5xx_Retries(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 502, Body: `<html>Bad Gateway</html>`},
		{StatusCode: 200, Body: `{"succeeded":true}`},
	}}
	c := newTestClient(rt)

	if err := c.PollUntilDone(context.Background(), "/api/v1/vm", "req4"); err != nil {
		t.Fatalf("expected success after retrying past the transient 502, got error: %v", err)
	}
	if len(rt.calls) != 2 {
		t.Errorf("expected 2 calls (one transient 502, one success), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// An unparseable body at 4xx is NOT retried - only 5xx is treated as a likely intermediary.
func TestPollUntilDone_UnparseableBodyAt4xx_NotRetried(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 404, Body: `not found`},
	}}
	c := newTestClient(rt)

	err := c.PollUntilDone(context.Background(), "/api/v1/vm", "req5")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call, got %d: %+v", len(rt.calls), rt.calls)
	}
}
