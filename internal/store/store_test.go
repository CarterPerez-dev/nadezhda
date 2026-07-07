// ©AngelaMos | 2026
// store_test.go

package store

import (
	"errors"
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestMigrateAppliesLatest(t *testing.T) {
	s := openTemp(t)
	if s.Version() < 1 {
		t.Fatalf("schema version = %d, want >= 1", s.Version())
	}
	var n int
	if err := s.DB().QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='articles'`,
	).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("articles table not created by migration")
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "idem.db")
	s1, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	v1 := s1.Version()
	_ = s1.Close()

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	if s2.Version() != v1 {
		t.Errorf("reopen version = %d, want %d", s2.Version(), v1)
	}
}

func TestSourceRoundTrip(t *testing.T) {
	s := openTemp(t)
	id, err := s.UpsertSource(SourceInput{
		Name: "krebs", Title: "Krebs", URL: "https://krebsonsecurity.com/feed/",
		Type: "rss", Weight: 1.0, Tags: []string{"news"}, Enabled: true,
	})
	if err != nil {
		t.Fatalf("UpsertSource: %v", err)
	}
	row, err := s.GetSourceByName("krebs")
	if err != nil {
		t.Fatalf("GetSourceByName: %v", err)
	}
	if row.ID != id || row.Weight != 1.0 || !row.Enabled || len(row.Tags) != 1 {
		t.Errorf("round trip mismatch: %+v", row)
	}

	id2, err := s.UpsertSource(SourceInput{
		Name: "krebs", Title: "Krebs Updated", URL: "https://krebsonsecurity.com/feed/",
		Type: "rss", Weight: 0.9, Tags: []string{"news"}, Enabled: true,
	})
	if err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if id2 != id {
		t.Errorf("upsert should keep id %d, got %d", id, id2)
	}
	row2, _ := s.GetSourceByName("krebs")
	if row2.Title != "Krebs Updated" || row2.Weight != 0.9 {
		t.Errorf("upsert did not update fields: %+v", row2)
	}
}

func TestArticleUniqueConstraint(t *testing.T) {
	s := openTemp(t)
	srcID, err := s.UpsertSource(SourceInput{
		Name: "krebs", URL: "https://krebsonsecurity.com/feed/", Type: "rss", Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	a := Article{
		SourceID: srcID, CanonicalURL: "https://krebsonsecurity.com/post-1",
		ContentHash: "hash-1", Title: "Post 1", PublishedAt: 100, FetchedAt: 200,
	}
	if _, err := s.InsertArticle(a); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if _, err := s.InsertArticle(a); !errors.Is(err, ErrDuplicate) {
		t.Errorf("duplicate insert: got %v, want ErrDuplicate", err)
	}

	n, err := s.CountArticles()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("article count = %d, want 1 (dup rejected)", n)
	}
}
