package memory

import (
	"context"
	"sort"

	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

// DefaultRelatedLimit caps how many related tasks the resolver returns when a
// query does not specify its own limit.
const DefaultRelatedLimit = 10

// Query describes the work a new (or in-flight) task will touch. Files and Tests
// are the paths to match against prior tasks; ExcludeTaskID drops a task from the
// results (its own id, when resolving relations for an already-recorded task).
type Query struct {
	Files         []string
	Tests         []string
	ExcludeTaskID string
	Limit         int
}

// RelatedTask is a prior task that shares paths with the query, with the overlap
// broken out so callers can explain why it surfaced.
type RelatedTask struct {
	TaskID      string
	SharedFiles int
	SharedTests int
	Score       int
}

// Resolver finds prior tasks related to a query by shared file/test paths. It is
// stateless; each call operates on the given project's projection store.
type Resolver struct{}

// NewResolver returns a Resolver.
func NewResolver() *Resolver { return &Resolver{} }

// Related returns prior tasks that share files or tests with the query, ranked
// by total shared paths (tests weighted equal to files) with recency as the
// tiebreak, deduplicated per task, and capped to the query limit. A query with
// no paths returns nothing.
func (r *Resolver) Related(ctx context.Context, store *memorydb.Store, q Query) ([]RelatedTask, error) {
	overlaps, err := store.TasksSharingPaths(ctx, dedupe(q.Files), dedupe(q.Tests))
	if err != nil {
		return nil, err
	}
	ranked := make([]memorydb.PathOverlap, 0, len(overlaps))
	for _, o := range overlaps {
		if o.TaskID == q.ExcludeTaskID {
			continue
		}
		if o.SharedFiles == 0 && o.SharedTests == 0 {
			continue
		}
		ranked = append(ranked, o)
	}
	// Strongest overlap first; break ties by most recent, then by id so the
	// order is deterministic.
	sort.Slice(ranked, func(i, j int) bool {
		si, sj := ranked[i].SharedFiles+ranked[i].SharedTests, ranked[j].SharedFiles+ranked[j].SharedTests
		if si != sj {
			return si > sj
		}
		if !ranked[i].OccurredAt.Equal(ranked[j].OccurredAt) {
			return ranked[i].OccurredAt.After(ranked[j].OccurredAt)
		}
		return ranked[i].TaskID < ranked[j].TaskID
	})

	limit := q.Limit
	if limit <= 0 {
		limit = DefaultRelatedLimit
	}
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	out := make([]RelatedTask, 0, len(ranked))
	for _, o := range ranked {
		out = append(out, RelatedTask{
			TaskID:      o.TaskID,
			SharedFiles: o.SharedFiles,
			SharedTests: o.SharedTests,
			Score:       o.SharedFiles + o.SharedTests,
		})
	}
	return out, nil
}

// dedupe returns the distinct non-empty entries of paths, preserving order, so a
// repeated path cannot inflate an overlap count.
func dedupe(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
