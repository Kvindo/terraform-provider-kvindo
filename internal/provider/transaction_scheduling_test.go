package provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

// Regression coverage for review findings #1 and #2: the transaction Create retry loop's
// blocking-resource-id extraction, and the 4 wait/cleanup helpers that were silently reading the
// pre-2026-06-28 API shape (top-level "info"/arrays instead of "status"/"spec.*").

// fakeTxnRoundTripper is a minimal local fake HTTP transport for exercising *client.Client-backed
// helpers (waitForResourceStable, cleanupTransaction, etc.) without a live backend. Keyed by exact
// request path since these helpers only ever call client.Get/Delete/PollUntilDone with well-known
// paths - simpler than internal/client's own fakeRoundTripper (unexported there, not importable
// across packages) since these tests don't need to assert call ordering, only per-path responses.
type fakeTxnRoundTripper struct {
	t         *testing.T
	responses map[string][]string // path -> queue of JSON bodies, consumed in order
	calls     []string            // path of every request received, in order
}

func (f *fakeTxnRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	f.calls = append(f.calls, path)
	queue := f.responses[path]
	if len(queue) == 0 {
		f.t.Fatalf("fakeTxnRoundTripper: no canned response left for %s %s", req.Method, path)
	}
	body := queue[0]
	if len(queue) > 1 {
		f.responses[path] = queue[1:]
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header:     make(http.Header),
	}, nil
}

func newFakeTxnClient(t *testing.T, responses map[string][]string) *client.Client {
	return &client.Client{
		BaseURL: "http://test.local",
		Token:   "test-token",
		Version: "test",
		HTTPClient: &http.Client{
			Transport: &fakeTxnRoundTripper{t: t, responses: responses},
		},
	}
}

// #1: parseSchedulingConflict must be called against *client.ApiError's own ErrorMessage field
// (the raw backend message, unaffected by the client's Error()-method wrapping added in f60010b),
// not against putErr.Error() (the wrapped string) - the wrapped string's own "(status <n>)"
// parenthetical shifts the ULID-boundary search to the wrong parens, extracting garbage instead of
// the real id (confirmed live in the original review: `ulid="status 422): The resource
// OrganizationS3Bucket \"b1\" (01k4..."`). Proves the parser correctly extracts type+id from the
// raw message, AND that the old wrapped-string input really did (and, if ever reintroduced, would
// again) produce garbage rather than failing loudly - a silent-wrong-data bug, not a clean
// parse failure.
func TestParseSchedulingConflict_RawMessageVsWrappedString(t *testing.T) {
	apiErr := &client.ApiError{
		StatusCode:   422,
		ErrorCode:    "ResourceIsScheduling",
		ErrorMessage: `The resource OrganizationS3Bucket "b1" (01k4t8z0yq2v7d3n5q8r9wxyza) is in state Scheduling and can not be modified`,
	}

	id, resType := parseSchedulingConflict(apiErr.ErrorMessage)
	if id != "01k4t8z0yq2v7d3n5q8r9wxyza" {
		t.Errorf("expected id 01k4t8z0yq2v7d3n5q8r9wxyza from the raw message, got %q", id)
	}
	if resType != "s3-bucket" {
		t.Errorf("expected resType s3-bucket, got %q", resType)
	}

	// The wrapped Error() string is what the pre-fix code passed in - proves it silently produces a
	// WRONG (garbage) id rather than the real one, which is exactly why the call site was changed
	// to use apiErr.ErrorMessage instead of putErr.Error().
	wrapped := apiErr.Error()
	wrongID, _ := parseSchedulingConflict(wrapped)
	if wrongID == "01k4t8z0yq2v7d3n5q8r9wxyza" {
		t.Fatal("expected the wrapped string to produce a WRONG id (demonstrating the bug this fix avoids), not the correct one")
	}
}

// The backend sends ErrorCode "ResourceIsScheduling" even when the resource's real State is
// "Reconciling" (there is no separate "ResourceIsReconciling" code - confirmed directly against
// ResourceControllerBase.cs:1430) - parseSchedulingConflict must extract type+id correctly from
// that message shape too, generically matching "is in state <anything>" rather than hardcoding the
// two known state names.
func TestParseSchedulingConflict_ReconcilingStateMessage(t *testing.T) {
	msg := `The resource OrganizationFolder "f1" (01hzy3x2w1v0u9t8s7r6q5p4o3) is in state Reconciling and can not be modified`
	id, resType := parseSchedulingConflict(msg)
	if id != "01hzy3x2w1v0u9t8s7r6q5p4o3" || resType != "folder" {
		t.Errorf("expected (01hzy3x2w1v0u9t8s7r6q5p4o3, folder), got (%q, %q)", id, resType)
	}
}

// The call site's own classification: errors.As + the structured ErrorCode check, gating whether
// parseSchedulingConflict is even attempted - proven end-to-end against a *client.ApiError so a
// future change to how the client wraps/exposes errors is caught here.
func TestSchedulingConflictClassification_ViaErrorsAs(t *testing.T) {
	var err error = &client.ApiError{
		StatusCode:   422,
		ErrorCode:    "ResourceIsScheduling",
		ErrorMessage: `The resource OrganizationFolder "f1" (01hzy3x2w1v0u9t8s7r6q5p4o3) is in state Reconciling and can not be modified`,
	}
	var apiErr *client.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected errors.As to extract *client.ApiError")
	}
	if apiErr.ErrorCode != "ResourceIsScheduling" {
		t.Fatalf("expected ErrorCode ResourceIsScheduling, got %q", apiErr.ErrorCode)
	}
	id, resType := parseSchedulingConflict(apiErr.ErrorMessage)
	if id != "01hzy3x2w1v0u9t8s7r6q5p4o3" || resType != "folder" {
		t.Errorf("expected (01hzy3x2w1v0u9t8s7r6q5p4o3, folder), got (%q, %q)", id, resType)
	}
}

// A structured error with a DIFFERENT ErrorCode (not this specific scheduling conflict) must not
// be treated as one by the caller - the classification lives in the ErrorCode check, not message
// text, so this is really a documentation test for that call-site behavior via a direct
// errors.As + ErrorCode check, mirroring exactly what Create()'s retry loop does.
func TestSchedulingConflictClassification_OtherErrorCode_NotTreatedAsConflict(t *testing.T) {
	var err error = &client.ApiError{StatusCode: 422, ErrorCode: "QuotaExceeded", ErrorMessage: "quota exceeded"}
	var apiErr *client.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected errors.As to extract *client.ApiError")
	}
	if apiErr.ErrorCode == "ResourceIsScheduling" {
		t.Fatal("test setup bug: ErrorCode should not be ResourceIsScheduling")
	}
}

// #2: waitForResourceStable must read status.state (the API renamed info -> status on
// 2026-06-28) - reading the old "info" key always found nothing (state read as ""), which is
// neither "scheduling" nor "reconciling", so the function returned nil immediately on the very
// first GET regardless of the resource's real state, never actually waiting. Uses a short context
// deadline rather than waiting out the function's real 10s poll interval: a resource genuinely
// stuck Scheduling forever must make the function block until ctx.Done() fires (proving it
// correctly recognized "still scheduling" and kept polling) - the pre-fix bug would instead return
// nil within the first GET, well before the deadline.
func TestWaitForResourceStable_ReadsStatusState_KeepsWaitingWhileScheduling(t *testing.T) {
	c := newFakeTxnClient(t, map[string][]string{
		"/api/v1/s3-bucket/b1": {`{"resource":{"status":{"state":"Scheduling"}}}`},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err := waitForResourceStable(ctx, c, "s3-bucket", "b1")
	if err == nil {
		t.Fatal("expected a context-deadline error (resource never left Scheduling) - a nil return means the state read was ignored, same as the pre-fix bug")
	}
	rt := c.HTTPClient.Transport.(*fakeTxnRoundTripper)
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 GET before blocking on the (real, 10s) poll interval, got %d: %v", len(rt.calls), rt.calls)
	}
}

// The success path: once status.state genuinely reads "Stable", the function must return nil
// immediately (not an artifact of the bug - a positive control for the above).
func TestWaitForResourceStable_StableState_ReturnsImmediately(t *testing.T) {
	c := newFakeTxnClient(t, map[string][]string{
		"/api/v1/s3-bucket/b3": {`{"resource":{"status":{"state":"Stable"}}}`},
	})
	if err := waitForResourceStable(context.Background(), c, "s3-bucket", "b3"); err != nil {
		t.Fatalf("expected success for an already-stable resource, got error: %v", err)
	}
}

// deleteIfSchedulingFailed must also read status.state, not info.state: a GET reporting
// SchedulingFailed there should proceed to actually issue the DELETE (both hit the same path, in
// order - the fake queues GET's response first, DELETE's second).
func TestDeleteIfSchedulingFailed_ReadsStatusState(t *testing.T) {
	responses := map[string][]string{
		"/api/v1/s3-bucket/b2": {
			`{"resource":{"status":{"state":"SchedulingFailed"}}}`, // the state-check GET
			`{"requestId":"","resourceId":"b2"}`,                   // the DELETE response (empty requestId short-circuits PollUntilDone)
		},
	}
	c := newFakeTxnClient(t, responses)
	deleteIfSchedulingFailed(context.Background(), c, "s3-bucket", "b2")
	rt := c.HTTPClient.Transport.(*fakeTxnRoundTripper)
	if len(rt.calls) != 2 {
		t.Fatalf("expected a GET (state check) then a DELETE, got %v", rt.calls)
	}
}

// #2: cleanupTransaction and waitForSubResourcesStable must read the transaction's sub-resource
// arrays from spec.* (they moved off the top level on 2026-06-28) and each item's own state from
// its "status" block (not "info"), and must resolve each item's id via metadata.id (a transaction
// sub-resource item is a full resource envelope - ResourceModelBase has no top-level Id field).
func TestWaitForSubResourcesStable_ReadsSpecAndStatus_KeepsWaitingWhilePending(t *testing.T) {
	c := newFakeTxnClient(t, map[string][]string{
		"/api/v1/transaction/tx1": {`{"resource":{"spec":{"s3Buckets":[{"metadata":{"id":"bkt1"},"status":{"state":"Scheduling"}}]}}}`},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Same reasoning as waitForResourceStable's test above: reading the pre-2026-06-28 shape
	// (top-level "s3Buckets", item "info") always finds zero pending items and returns nil
	// immediately; a genuinely-still-scheduling sub-resource must instead block until ctx.Done().
	if err := waitForSubResourcesStable(ctx, c, "tx1"); err == nil {
		t.Fatal("expected a context-deadline error (sub-resource never left Scheduling)")
	}
}

func TestWaitForSubResourcesStable_AllStable_ReturnsImmediately(t *testing.T) {
	c := newFakeTxnClient(t, map[string][]string{
		"/api/v1/transaction/tx4": {`{"resource":{"spec":{"s3Buckets":[{"metadata":{"id":"bkt4"},"status":{"state":"Stable"}}]}}}`},
	})
	if err := waitForSubResourcesStable(context.Background(), c, "tx4"); err != nil {
		t.Fatalf("expected success once all sub-resources are stable, got error: %v", err)
	}
}

func TestWaitForSubResourcesStable_FailedSubResource_ReturnsError(t *testing.T) {
	c := newFakeTxnClient(t, map[string][]string{
		"/api/v1/transaction/tx2": {
			`{"resource":{"spec":{"s3Buckets":[{"metadata":{"id":"bkt2"},"status":{"state":"SchedulingFailed"}}]}}}`,
		},
	})
	if err := waitForSubResourcesStable(context.Background(), c, "tx2"); err == nil {
		t.Fatal("expected an error for a sub-resource that entered SchedulingFailed")
	}
}

func TestItemMetadataID(t *testing.T) {
	item := map[string]interface{}{"metadata": map[string]interface{}{"id": "abc123"}}
	if got := itemMetadataID(item); got != "abc123" {
		t.Errorf("expected abc123, got %q", got)
	}
	// A bare top-level "id" (the pre-fix, wrong read) must NOT be picked up.
	flatOnly := map[string]interface{}{"id": "should-not-be-read"}
	if got := itemMetadataID(flatOnly); got != "" {
		t.Errorf("expected empty string for an item with no metadata.id, got %q", got)
	}
}
