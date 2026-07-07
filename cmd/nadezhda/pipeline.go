// ©AngelaMos | 2026
// pipeline.go

package main

import (
	"context"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/cluster"
	"github.com/CarterPerez-dev/nadezhda/internal/config"
	"github.com/CarterPerez-dev/nadezhda/internal/fetch"
	"github.com/CarterPerez-dev/nadezhda/internal/ingest"
	"github.com/CarterPerez-dev/nadezhda/internal/source"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

func ingestAndCluster(ctx context.Context, fc *fetch.Client, st *store.Store, cfg config.Config, targets []source.Source, start time.Time) (ingest.Summary, cluster.Stats, error) {
	summary, err := ingest.Run(ctx, fc, st, cfg, targets, start)
	if err != nil {
		return ingest.Summary{}, cluster.Stats{}, err
	}
	sinceUnix := start.Unix() - int64(cfg.Cluster.LookbackHours)*secondsPerHour
	windowSeconds := int64(cfg.Cluster.WindowHours) * secondsPerHour
	stats, err := cluster.Rebuild(st, cfg.Cluster.TitleJaccard, windowSeconds, sinceUnix)
	if err != nil {
		return summary, cluster.Stats{}, err
	}
	return summary, stats, nil
}
