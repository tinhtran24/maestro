package cli

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type orchestratorListOptions struct {
	json bool
}

type orchestratorListOutput struct {
	Data []sessionListEntry `json:"data"`
}

func newOrchestratorCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orchestrator",
		Short: "Manage orchestrator sessions",
	}
	cmd.AddCommand(newOrchestratorListCommand(ctx))
	cmd.AddCommand(newOrchestratorFinalizeCommand(ctx))
	return cmd
}

func newOrchestratorFinalizeCommand(ctx *commandContext) *cobra.Command {
	var state string
	cmd := &cobra.Command{
		Use:   "finalize <worker-id>",
		Short: "Record one completed task-finalization step",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestratorID := strings.TrimSpace(os.Getenv("THANOS_SESSION_ID"))
			if orchestratorID == "" {
				return usageError{errors.New("orchestrator finalize must run inside a managed orchestrator session")}
			}
			var out sessionFinalizationResponse
			endpoint := "orchestrators/" + url.PathEscape(orchestratorID) + "/finalizations/" + url.PathEscape(args[0])
			if err := ctx.postJSON(cmd.Context(), endpoint, advanceFinalizationRequest{State: state}, &out); err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), out)
		},
	}
	cmd.Flags().StringVar(&state, "state", "", "Completed workflow state")
	_ = cmd.MarkFlagRequired("state")
	return cmd
}

func newOrchestratorListCommand(ctx *commandContext) *cobra.Command {
	var opts orchestratorListOptions
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List orchestrator sessions",
		Args:    noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.listOrchestrators(cmd.Context(), cmd, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")
	return cmd
}

func (c *commandContext) listOrchestrators(ctx context.Context, cmd *cobra.Command, opts orchestratorListOptions) error {
	var res sessionListResponse
	if err := c.getJSON(ctx, "orchestrators", &res); err != nil {
		return err
	}
	orchestrators := filterAndSortOrchestrators(res.Sessions)
	if opts.json {
		return writeJSON(cmd.OutOrStdout(), orchestratorListOutput{Data: sessionListEntries(orchestrators)})
	}
	return writeOrchestratorList(cmd, orchestrators)
}

func filterAndSortOrchestrators(sessions []sessionDTO) []sessionDTO {
	out := make([]sessionDTO, 0, len(sessions))
	for _, sess := range sessions {
		if sess.Kind != "orchestrator" {
			continue
		}
		out = append(out, sess)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ProjectID != out[j].ProjectID {
			return out[i].ProjectID < out[j].ProjectID
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func writeOrchestratorList(cmd *cobra.Command, sessions []sessionDTO) error {
	out := cmd.OutOrStdout()
	if len(sessions) == 0 {
		_, err := fmt.Fprintln(out, "(no orchestrators)")
		return err
	}
	currentProject := ""
	for _, sess := range sessions {
		if sess.ProjectID != currentProject {
			if currentProject != "" {
				if _, err := fmt.Fprintln(out); err != nil {
					return err
				}
			}
			currentProject = sess.ProjectID
			if _, err := fmt.Fprintf(out, "%s:\n", currentProject); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(out, "  %s", sess.ID); err != nil {
			return err
		}
		parts := orchestratorLineParts(sess)
		if len(parts) > 0 {
			if _, err := fmt.Fprintf(out, "  %s", strings.Join(parts, "  ")); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}
	return nil
}

func orchestratorLineParts(sess sessionDTO) []string {
	parts := []string{}
	if !sess.Activity.LastActivityAt.IsZero() {
		parts = append(parts, "("+formatSessionAge(time.Since(sess.Activity.LastActivityAt))+")")
	}
	if sess.Status != "" {
		parts = append(parts, "["+sess.Status+"]")
	}
	if sess.IsTerminated {
		parts = append(parts, "terminated")
	}
	return parts
}
