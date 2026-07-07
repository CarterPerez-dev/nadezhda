// ©AngelaMos | 2026
// watch_test.go

package store

import "testing"

func TestNewlyFetchedClustersFiltersByFetchTimeNotPublishTime(t *testing.T) {
	s := openTemp(t)
	src, _ := s.UpsertSource(SourceInput{Name: "a", URL: "https://a/f", Type: "rss", Weight: 1, Enabled: true})

	aOld, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/old", ContentHash: "cold", TitleHash: "told", Title: "old", PublishedAt: 100, FetchedAt: 100})
	aRecent, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/recent", ContentHash: "crecent", TitleHash: "trecent", Title: "recently published, freshly fetched", PublishedAt: 940, FetchedAt: 1000})
	aFresh, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/fresh", ContentHash: "cfresh", TitleHash: "tfresh", Title: "fresh", PublishedAt: 1000, FetchedAt: 1000})

	if err := s.ReplaceClusters([]ClusterRow{
		{Key: "old", Members: []int64{aOld}, FirstSeen: 100, LastSeen: 100},
		{Key: "recent", Members: []int64{aRecent}, FirstSeen: 940, LastSeen: 940},
		{Key: "fresh", Members: []int64{aFresh}, FirstSeen: 1000, LastSeen: 1000},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := s.NewlyFetchedClusters(1000)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, c := range got {
		for _, a := range c.Articles {
			seen[a.CanonicalURL] = true
		}
	}
	if len(got) != 2 {
		t.Fatalf("newly fetched clusters = %d, want 2 (recent + fresh)", len(got))
	}
	if !seen["https://a/recent"] {
		t.Error("a story fetched this cycle whose publish time is just before the watermark must be surfaced by the fetch-time query")
	}
	if !seen["https://a/fresh"] {
		t.Error("the fresh story must be surfaced")
	}
	if seen["https://a/old"] {
		t.Error("a story fetched before the watermark must be excluded")
	}

	dig, err := s.DigestClusters(1000)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range dig {
		for _, a := range c.Articles {
			if a.CanonicalURL == "https://a/recent" {
				t.Error("this is the whole point of the separate query: the publish-time DigestClusters filter drops the recent-but-just-fetched story, so watch must use the fetch-time query instead")
			}
		}
	}
}

func TestNewlyFetchedClustersEmpty(t *testing.T) {
	s := openTemp(t)
	got, err := s.NewlyFetchedClusters(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no clusters on an empty store, got %d", len(got))
	}
}

func TestNewlyFetchedClustersAttachesCVEs(t *testing.T) {
	s := openTemp(t)
	src, _ := s.UpsertSource(SourceInput{Name: "a", URL: "https://a/f", Type: "rss", Weight: 1, Enabled: true})
	a1, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/1", ContentHash: "c1", TitleHash: "t1", Title: "kev story", PublishedAt: 5000, FetchedAt: 5000})
	if err := s.UpsertCVEStub("CVE-2026-9"); err != nil {
		t.Fatal(err)
	}
	score := 9.8
	if err := s.UpdateCVEEnrichment(CVE{ID: "CVE-2026-9", CVSSScore: &score, IsKEV: true, EnrichedAt: 1, EnrichStatus: EnrichStatusOK}); err != nil {
		t.Fatal(err)
	}
	if err := s.LinkArticleCVE(a1, "CVE-2026-9"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceClusters([]ClusterRow{{Key: "1", Members: []int64{a1}, FirstSeen: 5000, LastSeen: 5000}}); err != nil {
		t.Fatal(err)
	}

	got, err := s.NewlyFetchedClusters(1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].CVEs) != 1 {
		t.Fatalf("want 1 cluster with 1 CVE, got %d clusters", len(got))
	}
	if !got[0].CVEs[0].IsKEV {
		t.Error("KEV flag should be attached to the newly fetched cluster")
	}
}
