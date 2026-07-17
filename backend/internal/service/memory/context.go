package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

// DefaultContextTokenBudget is the hard ceiling on a context pack when the caller
// does not set one. It sits within the plan's 300-2500 token target.
const DefaultContextTokenBudget = 2000

// maxContextCandidates bounds how many related tasks the pack considers before
// the token budget trims further. Tasks beyond this are the lowest-ranked and
// would be dropped by the budget anyway.
const maxContextCandidates = 500

// Role selects which slice of memory a context pack surfaces. The roles mirror
// the agent stages that consume memory.
type Role string

// The consuming agent stages a context pack can be shaped for.
const (
	RolePlanner  Role = "planner"
	RoleCoder    Role = "coder"
	RoleReviewer Role = "reviewer"
	RoleTester   Role = "tester"
)

// roleSpec documents, per role, which parts of a related task are relevant.
// Planner wants prior implementations and constraints; coder wants the files and
// tests to touch; reviewer wants regression history (bugfixes first), affected
// files, PRs, and protected decisions; tester wants existing tests and prior
// bugs. Symbol/snippet sections arrive with the codebase-memory provider.
type roleSpec struct {
	includeFiles     bool
	includeTests     bool
	includeDecisions bool
	includePRs       bool
	preferTaskType   string
}

var roleSpecs = map[Role]roleSpec{
	RolePlanner:  {includeFiles: true, includeDecisions: true},
	RoleCoder:    {includeFiles: true, includeTests: true, includeDecisions: true},
	RoleReviewer: {includeFiles: true, includeDecisions: true, includePRs: true, preferTaskType: "bugfix"},
	RoleTester:   {includeTests: true, includeDecisions: true, preferTaskType: "bugfix"},
}

// PackTask is one prior task included in a context pack.
type PackTask struct {
	TaskID      string
	Intent      string
	TaskType    string
	Branch      string
	SharedFiles int
	SharedTests int
	Files       []string
	Tests       []string
	Decisions   []string
	PRs         []memoryevents.PRRef
}

// DroppedItem records something omitted from a pack so truncation is never
// silent.
type DroppedItem struct {
	Kind   string // related_task
	Ref    string // task id
	Reason string // token_budget
}

// ContextPack is the compact, role-specific memory handed to an agent stage.
type ContextPack struct {
	Role            Role
	RelatedTasks    []PackTask
	RelevantFiles   []string
	RelevantTests   []string
	Decisions       []string
	Dropped         []DroppedItem
	EstimatedTokens int
}

// ContextResolver assembles token-budgeted, role-specific context packs from the
// projection.
type ContextResolver struct {
	related   *Resolver
	maxTokens int
	log       *slog.Logger
}

// NewContextResolver returns a ContextResolver capping packs at maxTokens (or the
// default when maxTokens <= 0).
func NewContextResolver(maxTokens int, log *slog.Logger) *ContextResolver {
	if maxTokens <= 0 {
		maxTokens = DefaultContextTokenBudget
	}
	if log == nil {
		log = slog.Default()
	}
	return &ContextResolver{related: NewResolver(), maxTokens: maxTokens, log: log}
}

// Pack builds a role-specific context pack for the work described by q. It ranks
// prior tasks by shared paths, shapes each to the role, then greedily includes
// whole tasks until the token budget is reached; tasks that do not fit are
// recorded in Dropped and logged, never silently discarded. EstimatedTokens is
// guaranteed not to exceed the configured budget.
func (cr *ContextResolver) Pack(ctx context.Context, store *memorydb.Store, q Query, role Role) (ContextPack, error) {
	spec, ok := roleSpecs[role]
	if !ok {
		return ContextPack{}, fmt.Errorf("memory: unknown context role %q", role)
	}
	cq := q
	cq.Limit = maxContextCandidates
	related, err := cr.related.Related(ctx, store, cq)
	if err != nil {
		return ContextPack{}, err
	}

	hydrated, err := cr.hydrate(ctx, store, related, spec)
	if err != nil {
		return ContextPack{}, err
	}
	if spec.preferTaskType != "" {
		sort.SliceStable(hydrated, func(i, j int) bool {
			return hydrated[i].TaskType == spec.preferTaskType && hydrated[j].TaskType != spec.preferTaskType
		})
	}

	pack := ContextPack{Role: role}
	files, tests, decisions := newOrderedSet(), newOrderedSet(), newOrderedSet()
	used := 0
	for _, pt := range hydrated {
		cost := packTaskCost(pt)
		if used+cost > cr.maxTokens {
			pack.Dropped = append(pack.Dropped, DroppedItem{Kind: "related_task", Ref: pt.TaskID, Reason: "token_budget"})
			continue
		}
		used += cost
		pack.RelatedTasks = append(pack.RelatedTasks, pt)
		files.addAll(pt.Files)
		tests.addAll(pt.Tests)
		decisions.addAll(pt.Decisions)
	}
	pack.RelevantFiles = files.items
	pack.RelevantTests = tests.items
	pack.Decisions = decisions.items
	pack.EstimatedTokens = used

	if len(pack.Dropped) > 0 {
		cr.log.Debug("memory: context pack trimmed to token budget",
			"role", role, "included", len(pack.RelatedTasks), "dropped", len(pack.Dropped), "budget", cr.maxTokens)
	}
	return pack, nil
}

func (cr *ContextResolver) hydrate(ctx context.Context, store *memorydb.Store, related []RelatedTask, spec roleSpec) ([]PackTask, error) {
	out := make([]PackTask, 0, len(related))
	for _, rt := range related {
		task, ok, err := store.GetTask(ctx, rt.TaskID)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		pt := PackTask{
			TaskID:      rt.TaskID,
			Intent:      task.Intent,
			TaskType:    task.TaskType,
			Branch:      task.Branch,
			SharedFiles: rt.SharedFiles,
			SharedTests: rt.SharedTests,
		}
		if spec.includeFiles {
			if pt.Files, err = store.ListFiles(ctx, rt.TaskID); err != nil {
				return nil, err
			}
		}
		if spec.includeTests {
			if pt.Tests, err = store.ListTests(ctx, rt.TaskID); err != nil {
				return nil, err
			}
		}
		if spec.includeDecisions {
			pt.Decisions = decodeStringArray(task.DecisionsJSON)
		}
		if spec.includePRs {
			pt.PRs = decodePRs(task.PRsJSON)
		}
		out = append(out, pt)
	}
	return out, nil
}

// packTaskCost estimates the token footprint of a PackTask at roughly four
// characters per token.
func packTaskCost(pt PackTask) int {
	n := estTokens(pt.Intent) + estTokens(pt.Branch) + estTokens(pt.TaskType) + 4
	for _, f := range pt.Files {
		n += estTokens(f)
	}
	for _, t := range pt.Tests {
		n += estTokens(t)
	}
	for _, d := range pt.Decisions {
		n += estTokens(d)
	}
	for _, pr := range pt.PRs {
		n += estTokens(pr.URL) + 2
	}
	return n
}

func estTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + 3) / 4
}

func decodeStringArray(s string) []string {
	if s == "" || s == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil
	}
	return out
}

func decodePRs(s string) []memoryevents.PRRef {
	if s == "" || s == "[]" {
		return nil
	}
	var out []memoryevents.PRRef
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil
	}
	return out
}

// orderedSet is a small insertion-ordered string set for building deduplicated
// union lists.
type orderedSet struct {
	seen  map[string]struct{}
	items []string
}

func newOrderedSet() *orderedSet {
	return &orderedSet{seen: map[string]struct{}{}}
}

func (s *orderedSet) addAll(vs []string) {
	for _, v := range vs {
		if v == "" {
			continue
		}
		if _, ok := s.seen[v]; ok {
			continue
		}
		s.seen[v] = struct{}{}
		s.items = append(s.items, v)
	}
}
