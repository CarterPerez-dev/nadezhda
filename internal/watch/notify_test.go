// ©AngelaMos | 2026
// notify_test.go

package watch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sampleReport() Report {
	return Report{
		Notable: []NotableItem{
			{Title: "Critical bug in Foo", URL: "https://x/1", Score: 0.91, MaxCVSS: 9.8, IsKEV: true, CVEs: []string{"CVE-2026-1"}, Sources: 3},
			{Title: "Breach at Bar", URL: "https://x/2", Score: 0.62, Sources: 2},
		},
	}
}

func TestWebhookNotifierPostsPayload(t *testing.T) {
	got := make(chan webhookPayload, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %q, want application/json", ct)
		}
		var p webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("decode: %v", err)
		}
		got <- p
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := WebhookNotifier{URL: srv.URL, Client: srv.Client()}
	if err := n.Notify(context.Background(), sampleReport()); err != nil {
		t.Fatal(err)
	}

	p := <-got
	if len(p.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(p.Items))
	}
	if p.Items[0].Title != "Critical bug in Foo" || !p.Items[0].IsKEV {
		t.Errorf("first item wrong: %+v", p.Items[0])
	}
	if p.Text == "" || p.Content == "" {
		t.Error("text and content must both be set for Slack/Discord compatibility")
	}
	if !strings.Contains(p.Text, "[KEV]") {
		t.Errorf("summary should flag KEV items, got %q", p.Text)
	}
}

func TestWebhookNotifierNoopWhenNothingNotable(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := WebhookNotifier{URL: srv.URL, Client: srv.Client()}
	if err := n.Notify(context.Background(), Report{NewArticles: 5}); err != nil {
		t.Fatal(err)
	}
	if hits != 0 {
		t.Errorf("webhook was called %d times, want 0 when nothing is notable", hits)
	}
}

func TestWebhookNotifierErrorsOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := WebhookNotifier{URL: srv.URL, Client: srv.Client()}
	if err := n.Notify(context.Background(), sampleReport()); err == nil {
		t.Error("expected an error when the webhook returns 500")
	}
}
