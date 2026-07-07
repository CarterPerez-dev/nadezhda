// ©AngelaMos | 2026
// notify.go

package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	notifyMaxStatus      = 300
	defaultNotifyTimeout = 15 * time.Second
	headerContentType    = "Content-Type"
	mimeJSON             = "application/json"
)

type Notifier interface {
	Notify(ctx context.Context, r Report) error
}

type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

type webhookPayload struct {
	Text    string        `json:"text"`
	Content string        `json:"content"`
	Items   []webhookItem `json:"items"`
}

type webhookItem struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Score   float64  `json:"score"`
	MaxCVSS float64  `json:"max_cvss"`
	IsKEV   bool     `json:"is_kev"`
	CVEs    []string `json:"cves,omitempty"`
	Sources int      `json:"sources"`
}

func (w WebhookNotifier) Notify(ctx context.Context, r Report) error {
	if len(r.Notable) == 0 {
		return nil
	}
	body, err := json.Marshal(buildPayload(r))
	if err != nil {
		return fmt.Errorf("notify: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: new request: %w", err)
	}
	req.Header.Set(headerContentType, mimeJSON)
	resp, err := w.client().Do(req)
	if err != nil {
		return fmt.Errorf("notify: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= notifyMaxStatus {
		return fmt.Errorf("notify: webhook returned %s", resp.Status)
	}
	return nil
}

func (w WebhookNotifier) client() *http.Client {
	if w.Client != nil {
		return w.Client
	}
	return &http.Client{Timeout: defaultNotifyTimeout}
}

func buildPayload(r Report) webhookPayload {
	summary := summarize(r)
	items := make([]webhookItem, len(r.Notable))
	for i, n := range r.Notable {
		items[i] = webhookItem{
			Title:   n.Title,
			URL:     n.URL,
			Score:   n.Score,
			MaxCVSS: n.MaxCVSS,
			IsKEV:   n.IsKEV,
			CVEs:    n.CVEs,
			Sources: n.Sources,
		}
	}
	return webhookPayload{Text: summary, Content: summary, Items: items}
}

func summarize(r Report) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "nadezhda: %d notable %s\n", len(r.Notable), storyWord(len(r.Notable)))
	for _, n := range r.Notable {
		tag := ""
		if n.IsKEV {
			tag = " [KEV]"
		}
		fmt.Fprintf(&sb, "- %s%s (%s)\n", n.Title, tag, n.URL)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func storyWord(n int) string {
	if n == 1 {
		return "story"
	}
	return "stories"
}
