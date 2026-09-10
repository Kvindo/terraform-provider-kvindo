package client

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// Regression coverage for review finding #4: GetByLabels never sent maxPageSize or followed
// pagination.enumeratorId, so it silently returned only the first (default 10-item) page - every
// datasource's `name = ...` lookup (GetByName below) failed for any org with more resources of a
// type than that. Reuses the fakeRoundTripper/newTestClient helpers from put_retry_test.go.

func makeResourcesPage(n int, startIndex int) string {
	s := `{"resources":[`
	for i := 0; i < n; i++ {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf(`{"metadata":{"id":"id%d","name":"name%d"}}`, startIndex+i, startIndex+i)
	}
	s += `]`
	return s
}

func TestGetByLabels_FollowsPaginationAcrossPages(t *testing.T) {
	page1 := makeResourcesPage(maxGetByLabelsPageSize, 0) + `,"pagination":{"enumeratorId":"cursor1"}}`
	page2 := makeResourcesPage(5, maxGetByLabelsPageSize) + `}` // short page, no cursor - last page
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: page1},
		{StatusCode: 200, Body: page2},
	}}
	c := newTestClient(rt)

	items, err := c.GetByLabels(context.Background(), "/api/v1/vm", nil)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if want := maxGetByLabelsPageSize + 5; len(items) != want {
		t.Errorf("expected %d resources across both pages, got %d", want, len(items))
	}
	if len(rt.calls) != 2 {
		t.Fatalf("expected 2 GET calls (one per page), got %d: %+v", len(rt.calls), rt.calls)
	}
	// The second request must carry the first response's enumeratorId back as the cursor.
	if !strings.Contains(rt.calls[1].RawQuery, "enumeratorId=cursor1") {
		t.Errorf("expected second request to carry enumeratorId=cursor1, got query %q", rt.calls[1].RawQuery)
	}
	for i, call := range rt.calls {
		want := fmt.Sprintf("maxPageSize=%d", maxGetByLabelsPageSize)
		if !strings.Contains(call.RawQuery, want) {
			t.Errorf("call %d: expected query to contain %q, got %q", i, want, call.RawQuery)
		}
	}
}

func TestGetByLabels_ShortFirstPage_StopsAfterOneCall(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: makeResourcesPage(3, 0) + `}`},
	}}
	c := newTestClient(rt)

	items, err := c.GetByLabels(context.Background(), "/api/v1/vm", nil)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 resources, got %d", len(items))
	}
	if len(rt.calls) != 1 {
		t.Errorf("expected exactly 1 call (short page signals no more results), got %d: %+v", len(rt.calls), rt.calls)
	}
}

// The label query key must match the backend's binder shape (labels[<k>]=<v>), not the previous
// label.<k>=<v> - confirmed against KvindoCloudClient.cs/kc_api.py's own encoding.
func TestGetByLabels_EncodesLabelsWithBracketSyntax(t *testing.T) {
	rt := &fakeRoundTripper{t: t, responses: []fakeResponse{
		{StatusCode: 200, Body: `{"resources":[]}`},
	}}
	c := newTestClient(rt)

	_, err := c.GetByLabels(context.Background(), "/api/v1/vm", map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	q := rt.calls[0].RawQuery
	// url.Values.Encode() percent-encodes brackets.
	if !strings.Contains(q, "labels%5Benv%5D=prod") {
		t.Errorf("expected query to encode the label as labels[env]=prod (URL-encoded), got %q", q)
	}
	if strings.Contains(q, "label.env=prod") {
		t.Errorf("expected the old label.env=prod encoding to be gone, got %q", q)
	}
}
