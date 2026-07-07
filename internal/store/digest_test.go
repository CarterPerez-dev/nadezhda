// ©AngelaMos | 2026
// digest_test.go

package store

import "testing"

func TestDigestClustersAggregates(t *testing.T) {
	s := openTemp(t)

	src1, _ := s.UpsertSource(SourceInput{Name: "a", URL: "https://a/f", Type: "rss", Weight: 1.0, Enabled: true})
	src2, _ := s.UpsertSource(SourceInput{Name: "b", URL: "https://b/f", Type: "rss", Weight: 0.8, Enabled: true})

	a1, _ := s.InsertArticle(Article{SourceID: src1, CanonicalURL: "https://a/1", ContentHash: "c1", TitleHash: "t1", Title: "Story one", PublishedAt: 1000})
	a2, _ := s.InsertArticle(Article{SourceID: src2, CanonicalURL: "https://b/1", ContentHash: "c2", TitleHash: "t2", Title: "Story two", PublishedAt: 2000})

	if err := s.UpsertCVEStub("CVE-2026-1"); err != nil {
		t.Fatal(err)
	}
	score := 9.8
	if err := s.UpdateCVEEnrichment(CVE{ID: "CVE-2026-1", CVSSScore: &score, IsKEV: true, EnrichedAt: 1, EnrichStatus: EnrichStatusOK}); err != nil {
		t.Fatal(err)
	}
	if err := s.LinkArticleCVE(a1, "CVE-2026-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.LinkArticleCVE(a2, "CVE-2026-1"); err != nil {
		t.Fatal(err)
	}

	if err := s.ReplaceClusters([]ClusterRow{
		{Key: "1", Members: []int64{a1, a2}, FirstSeen: 1000, LastSeen: 2000},
	}); err != nil {
		t.Fatal(err)
	}

	clusters, err := s.DigestClusters(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 1 {
		t.Fatalf("clusters = %d, want 1", len(clusters))
	}
	c := clusters[0]
	if len(c.Articles) != 2 {
		t.Errorf("articles = %d, want 2", len(c.Articles))
	}
	if len(c.CVEs) != 1 {
		t.Fatalf("cves = %d, want 1 (deduped across both articles)", len(c.CVEs))
	}
	if !c.CVEs[0].IsKEV || c.CVEs[0].CVSSScore == nil || *c.CVEs[0].CVSSScore != 9.8 {
		t.Errorf("cve signals not aggregated: %+v", c.CVEs[0])
	}
}

func TestDigestClustersAttachToCorrectCluster(t *testing.T) {
	s := openTemp(t)
	src, _ := s.UpsertSource(SourceInput{Name: "a", URL: "https://a/f", Type: "rss", Weight: 1, Enabled: true})

	a1, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/1", ContentHash: "c1", TitleHash: "t1", Title: "one"})
	a2, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/2", ContentHash: "c2", TitleHash: "t2", Title: "two"})

	for _, id := range []string{"CVE-2026-1", "CVE-2026-2"} {
		if err := s.UpsertCVEStub(id); err != nil {
			t.Fatal(err)
		}
	}
	_ = s.LinkArticleCVE(a1, "CVE-2026-1")
	_ = s.LinkArticleCVE(a2, "CVE-2026-2")

	if err := s.ReplaceClusters([]ClusterRow{
		{Key: "1", Members: []int64{a1}, FirstSeen: 1000, LastSeen: 1000},
		{Key: "2", Members: []int64{a2}, FirstSeen: 2000, LastSeen: 2000},
	}); err != nil {
		t.Fatal(err)
	}

	clusters, err := s.DigestClusters(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 2 {
		t.Fatalf("clusters = %d, want 2", len(clusters))
	}
	byArticle := map[int64]DigestCluster{}
	for _, c := range clusters {
		if len(c.Articles) != 1 || len(c.CVEs) != 1 {
			t.Fatalf("cluster %d: articles=%d cves=%d, want 1/1", c.ClusterID, len(c.Articles), len(c.CVEs))
		}
		byArticle[c.Articles[0].ID] = c
	}
	if byArticle[a1].CVEs[0].ID != "CVE-2026-1" || byArticle[a2].CVEs[0].ID != "CVE-2026-2" {
		t.Error("CVEs attached to the wrong cluster")
	}
}

func TestDigestClustersSinceFilter(t *testing.T) {
	s := openTemp(t)
	src, _ := s.UpsertSource(SourceInput{Name: "a", URL: "https://a/f", Type: "rss", Weight: 1, Enabled: true})
	a1, _ := s.InsertArticle(Article{SourceID: src, CanonicalURL: "https://a/1", ContentHash: "c1", TitleHash: "t1", Title: "old"})
	if err := s.ReplaceClusters([]ClusterRow{{Key: "1", Members: []int64{a1}, FirstSeen: 500, LastSeen: 500}}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.DigestClusters(1000); len(got) != 0 {
		t.Errorf("cluster with last_seen 500 should be excluded by since=1000, got %d", len(got))
	}
	if got, _ := s.DigestClusters(0); len(got) != 1 {
		t.Errorf("since=0 should include it, got %d", len(got))
	}
}
