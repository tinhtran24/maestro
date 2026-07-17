package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// errWriter accumulates the first write error so a run of prints can be checked
// once instead of at every call site.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, a ...any) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintf(e.w, format, a...)
}

// memoryTaskDTO mirrors controllers.MemoryTaskDTO for GET
// /api/v1/projects/{id}/memory/tasks.
type memoryTaskDTO struct {
	ID           string   `json:"id"`
	SessionID    string   `json:"sessionId"`
	Kind         string   `json:"kind"`
	Intent       string   `json:"intent"`
	TaskType     string   `json:"taskType"`
	Branch       string   `json:"branch"`
	OccurredAt   string   `json:"occurredAt"`
	ChangedFiles []string `json:"changedFiles"`
	ChangedTests []string `json:"changedTests"`
}

type memoryTaskListResult struct {
	Tasks []memoryTaskDTO `json:"tasks"`
}

// memoryPackTaskDTO mirrors controllers.MemoryPackTaskDTO.
type memoryPackTaskDTO struct {
	TaskID      string `json:"taskId"`
	Intent      string `json:"intent"`
	TaskType    string `json:"taskType"`
	SharedFiles int    `json:"sharedFiles"`
	SharedTests int    `json:"sharedTests"`
}

type memoryDroppedDTO struct {
	Kind   string `json:"kind"`
	Ref    string `json:"ref"`
	Reason string `json:"reason"`
}

// memoryContextResult mirrors controllers.MemoryContextResponse.
type memoryContextResult struct {
	Role            string              `json:"role"`
	RelatedTasks    []memoryPackTaskDTO `json:"relatedTasks"`
	RelevantFiles   []string            `json:"relevantFiles"`
	RelevantTests   []string            `json:"relevantTests"`
	Decisions       []string            `json:"decisions"`
	Dropped         []memoryDroppedDTO  `json:"dropped"`
	EstimatedTokens int                 `json:"estimatedTokens"`
}

// memoryRebuildResult mirrors controllers.RebuildMemoryResponse.
type memoryRebuildResult struct {
	ProjectID string `json:"projectId"`
	Tasks     int    `json:"tasks"`
}

func newMemoryCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Inspect a project's task memory",
	}
	cmd.AddCommand(newMemoryListCommand(ctx))
	cmd.AddCommand(newMemoryContextCommand(ctx))
	cmd.AddCommand(newMemoryRebuildCommand(ctx))
	return cmd
}

func newMemoryListCommand(ctx *commandContext) *cobra.Command {
	var opts struct {
		json  bool
		limit int
	}
	cmd := &cobra.Command{
		Use:     "ls <project-id>",
		Aliases: []string{"list"},
		Short:   "List a project's completed-task memory, most recent first",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := fmt.Sprintf("projects/%s/memory/tasks", url.PathEscape(args[0]))
			if opts.limit > 0 {
				path += "?" + url.Values{"limit": {fmt.Sprint(opts.limit)}}.Encode()
			}
			var res memoryTaskListResult
			if err := ctx.getJSON(cmd.Context(), path, &res); err != nil {
				return err
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), res)
			}
			return writeMemoryTaskList(cmd, res.Tasks)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output tasks as JSON")
	cmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum tasks to return (default 100, max 500)")
	return cmd
}

func newMemoryContextCommand(ctx *commandContext) *cobra.Command {
	var opts struct {
		json      bool
		role      string
		intent    string
		files     []string
		tests     []string
		maxTokens int
	}
	cmd := &cobra.Command{
		Use:   "context <project-id>",
		Short: "Build a role-specific, token-budgeted context pack for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.role == "" {
				return usageError{fmt.Errorf("--role is required (planner|coder|reviewer|tester)")}
			}
			q := url.Values{}
			q.Set("role", opts.role)
			if opts.intent != "" {
				q.Set("intent", opts.intent)
			}
			if len(opts.files) > 0 {
				q.Set("files", strings.Join(opts.files, ","))
			}
			if len(opts.tests) > 0 {
				q.Set("tests", strings.Join(opts.tests, ","))
			}
			if opts.maxTokens > 0 {
				q.Set("maxTokens", fmt.Sprint(opts.maxTokens))
			}
			path := fmt.Sprintf("projects/%s/memory/context?%s", url.PathEscape(args[0]), q.Encode())
			var res memoryContextResult
			if err := ctx.getJSON(cmd.Context(), path, &res); err != nil {
				return err
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), res)
			}
			return writeMemoryContext(cmd, res)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output the pack as JSON")
	cmd.Flags().StringVar(&opts.role, "role", "", "Consuming stage: planner|coder|reviewer|tester (required)")
	cmd.Flags().StringVar(&opts.intent, "intent", "", "Free-text description of the work")
	cmd.Flags().StringSliceVar(&opts.files, "file", nil, "Changed file path to match (repeatable)")
	cmd.Flags().StringSliceVar(&opts.tests, "test", nil, "Changed test path to match (repeatable)")
	cmd.Flags().IntVar(&opts.maxTokens, "max-tokens", 0, "Override the pack's token budget")
	return cmd
}

func newMemoryRebuildCommand(ctx *commandContext) *cobra.Command {
	var opts struct{ json bool }
	cmd := &cobra.Command{
		Use:   "rebuild <project-id>",
		Short: "Rebuild a project's memory projection from its event log",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := fmt.Sprintf("projects/%s/memory/rebuild", url.PathEscape(args[0]))
			var res memoryRebuildResult
			if err := ctx.postJSON(cmd.Context(), path, nil, &res); err != nil {
				return err
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), res)
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Rebuilt memory for %s: %d task(s).\n", res.ProjectID, res.Tasks)
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output the result as JSON")
	return cmd
}

func writeMemoryTaskList(cmd *cobra.Command, tasks []memoryTaskDTO) error {
	out := cmd.OutOrStdout()
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(out, "No task memory recorded yet.")
		return err
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	ew := &errWriter{w: tw}
	ew.printf("TASK\tTYPE\tKIND\tFILES\tTESTS\tINTENT\n")
	for _, t := range tasks {
		ew.printf("%s\t%s\t%s\t%d\t%d\t%s\n",
			t.ID, dashIfEmpty(t.TaskType), t.Kind, len(t.ChangedFiles), len(t.ChangedTests), truncateText(t.Intent, 60))
	}
	if ew.err != nil {
		return ew.err
	}
	return tw.Flush()
}

func writeMemoryContext(cmd *cobra.Command, res memoryContextResult) error {
	ew := &errWriter{w: cmd.OutOrStdout()}
	ew.printf("Role: %s  (~%d tokens", res.Role, res.EstimatedTokens)
	if len(res.Dropped) > 0 {
		ew.printf(", %d related task(s) dropped for budget", len(res.Dropped))
	}
	ew.printf(")\n")
	if len(res.RelatedTasks) == 0 {
		ew.printf("No related prior tasks.\n")
	} else {
		var buf bytes.Buffer
		tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		tew := &errWriter{w: tw}
		tew.printf("RELATED\tTYPE\tSHARED\tINTENT\n")
		for _, t := range res.RelatedTasks {
			tew.printf("%s\t%s\t%d\t%s\n", t.TaskID, dashIfEmpty(t.TaskType), t.SharedFiles+t.SharedTests, truncateText(t.Intent, 60))
		}
		if tew.err != nil {
			return tew.err
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		ew.printf("%s", buf.String())
	}
	if len(res.Decisions) > 0 {
		ew.printf("\nProtected decisions:\n")
		for _, d := range res.Decisions {
			ew.printf("  - %s\n", d)
		}
	}
	return ew.err
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func truncateText(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
