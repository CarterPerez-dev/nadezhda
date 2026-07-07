// ©AngelaMos | 2026
// watch.go

package store

import (
	"fmt"
	"strings"
)

func (s *Store) NewlyFetchedClusters(sinceFetched int64) ([]DigestCluster, error) {
	ids, err := s.clusterIDsFetchedSince(sinceFetched)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	byID, order, err := s.clusterRowsByID(ids)
	if err != nil {
		return nil, err
	}
	if err := s.digestAttachArticles(byID); err != nil {
		return nil, err
	}
	if err := s.digestAttachCVEs(byID); err != nil {
		return nil, err
	}
	out := make([]DigestCluster, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, nil
}

func (s *Store) clusterIDsFetchedSince(sinceFetched int64) ([]int64, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT cm.cluster_id
		FROM cluster_members cm
		JOIN articles a ON a.id = cm.article_id
		WHERE a.fetched_at >= ?
		ORDER BY cm.cluster_id`, sinceFetched)
	if err != nil {
		return nil, fmt.Errorf("newly fetched clusters: %w", err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("newly fetched clusters: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) clusterRowsByID(ids []int64) (map[int64]*DigestCluster, []int64, error) {
	if len(ids) == 0 {
		return map[int64]*DigestCluster{}, nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `SELECT id, cluster_key, size, first_seen, last_seen FROM clusters WHERE id IN (` +
		strings.Join(placeholders, ",") + `) ORDER BY id`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("cluster rows by id: %w", err)
	}
	defer rows.Close()
	return scanClusterRows(rows)
}
