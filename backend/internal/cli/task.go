package cli

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newTaskCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "Transition orchestrator-owned task completion"}
	cmd.AddCommand(newTaskDoneCommand(ctx))
	return cmd
}

func newTaskDoneCommand(ctx *commandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "done <worker-id>",
		Short: "Mark a fully finalized task done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestratorID := strings.TrimSpace(os.Getenv("THANOS_SESSION_ID"))
			if orchestratorID == "" {
				return usageError{errors.New("task done must run inside a managed orchestrator session")}
			}
			endpoint := "orchestrators/" + url.PathEscape(orchestratorID) + "/finalizations/" + url.PathEscape(args[0])
			var out sessionFinalizationResponse
			if err := ctx.postJSON(cmd.Context(), endpoint, advanceFinalizationRequest{State: "done"}, &out); err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), out)
		},
	}
}
