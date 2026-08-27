package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// Regression coverage for Delete()'s TransactionDeleteLockBusy retry: the org-wide Transaction-
// bundle delete lock was busy (another Transaction delete for the same org already in flight),
// nothing was deleted, and retrying the ORIGINAL request is always safe. Reuses the
// fakeRoundTripper/newTestClient/recordedRequest helpers defined in put_retry_test.go.

func TestDelete_TransactionDeleteLockBusy_RetriesOriginalRequest(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"TransactionDeleteLockBusy","errorMessage":"busy"}`, Headers: map[string]string{"Retry-After": "1"}},
		{StatusCode: 200, Body: `{"requestId":"req1","resourceId":"res1"}`},
	}}
	c := newTestClient(rt)

	result, err := c.Delete(context.Background(), "/api/v1/transaction", "tx1")
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
		if call.Method != http.MethodDelete || call.Path != "/api/v1/transaction/tx1" {
			t.Errorf("call %d: expected DELETE /api/v1/transaction/tx1 (retry of the ORIGINAL request), got %s %s", i, call.Method, call.Path)
		}
	}
}

// A 404 is treated as a successful no-op (per delete()'s existing behavior) and must never enter
// the retry loop at all.
func TestDelete_404_ReturnsEmptySuccessNoRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 404},
	}}
	c := newTestClient(rt)

	result, err := c.Delete(context.Background(), "/api/v1/transaction", "already-gone")
	if err != nil {
		t.Fatalf("expected success (404 = already deleted), got error: %v", err)
	}
	if result == nil {
		t.Fatal("expected a non-nil empty ModificationResponse")
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call, got %d: %+v", len(rt.calls), rt.calls)
	}
}

// Only TransactionDeleteLockBusy is retried - any other ErrorCode (e.g. ResourceIsDeleteProtected)
// must return immediately, unretried.
func TestDelete_OtherErrorCode_NoRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"ResourceIsDeleteProtected","errorMessage":"protected"}`},
	}}
	c := newTestClient(rt)

	_, err := c.Delete(context.Background(), "/api/v1/transaction", "protected1")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a structured ApiError, got: %v", err)
	}
	if apiErr.ErrorCode != "ResourceIsDeleteProtected" {
		t.Errorf("expected ResourceIsDeleteProtected, got %s", apiErr.ErrorCode)
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (non-retryable error code), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// A non-JSON error body (e.g. a 5xx from an intermediary) falls back to a plain, unstructured
// error and must not be retried - mirrors TestPut_NonJSONError_ReturnsFallbackWithoutRetry.
func TestDelete_NonJSONError_ReturnsFallbackWithoutRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 502, Body: `<html>Bad Gateway</html>`},
	}}
	c := newTestClient(rt)

	_, err := c.Delete(context.Background(), "/api/v1/transaction", "tx2")
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

// Delete() is deliberately narrower than Put(): no transport-level (*url.Error) retry yet (see the
// "Transaction-Bundle Delete Lock" CLAUDE.md entry for why this is a documented scope cut). A
// transport error must surface immediately, not retry.
func TestDelete_TransportError_NoRetry(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{Err: context.DeadlineExceeded},
	}}
	c := newTestClient(rt)

	_, err := c.Delete(context.Background(), "/api/v1/transaction", "tx3")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (transport errors are not retried by Delete()), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// Proves the short parent deadline actually bounds Delete()'s derived context.WithTimeout(ctx,
// 10*time.Minute) to the shorter of the two - persistent lock contention must give up promptly,
// not hang for the full 10 minutes. Mirrors TestPut_OuterDeadlineExhaustion_GivesUpPromptly.
func TestDelete_OuterDeadlineExhaustion_GivesUpPromptly(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 422, Body: `{"errorCode":"TransactionDeleteLockBusy","errorMessage":"busy"}`, Repeat: true},
	}}
	c := newTestClient(rt)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.Delete(ctx, "/api/v1/transaction", "deadline1")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error (persistent lock contention), got nil")
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected Delete() to give up promptly after the short parent deadline, took %v", elapsed)
	}
	if len(rt.calls) == 0 {
		t.Error("expected at least one call before giving up")
	}
}
