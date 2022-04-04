package check

import (
	"fmt"

	"github.com/ory/keto/internal/check"

	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"

	"github.com/ory/keto/cmd/client"
)

type checkOutput check.RESTResponse

func (o *checkOutput) String() string {
	if o.Allowed {
		return "Allowed\n"
	}
	return "Denied\n"
}

const FlagMaxDepth = "max-depth"

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <subject> <relation> <namespace> <object>",
		Short: "Check whether a subject has a relation on an object",
		Long:  "Check whether a subject has a relation on an object. This method resolves subject sets and subject set rewrites.",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := client.FromCmd(cmd, client.ModeReadOnly, cmd.Context())
			if err != nil {
				return err
			}
			defer conn.Close()

			maxDepth, err := cmd.Flags().GetInt32(FlagMaxDepth)
			if err != nil {
				return err
			}

			result, err := conn.Check(args[0], args[1], args[2], args[3], maxDepth)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not make request: %s\n", err)
				return err
			}

			cmdx.PrintJSONAble(cmd, &checkOutput{Allowed: result})
			return nil
		},
	}

	client.RegisterRemoteURLFlags(cmd.Flags())
	cmdx.RegisterFormatFlags(cmd.Flags())
	cmd.Flags().Int32P(FlagMaxDepth, "d", 0, "Maximum depth of the search tree. If the value is less than 1 or greater than the global max-depth then the global max-depth will be used instead.")

	return cmd
}

func RegisterCommandsRecursive(parent *cobra.Command) {
	parent.AddCommand(newCheckCmd())
}
