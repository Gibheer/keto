package expand

import (
	"fmt"

	"github.com/ory/x/flagx"

	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"

	"github.com/ory/keto/cmd/client"
)

const FlagMaxDepth = "max-depth"

func NewExpandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expand <relation> <namespace> <object>",
		Short: "Expand a subject set",
		Long:  "Expand a subject set into a tree of subjects.",
		Args:  cobra.ExactArgs(3),
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
			tree, err := conn.Expand(args[0], args[1], args[2], maxDepth)
			if err != nil {
				return err
			}

			cmdx.PrintJSONAble(cmd, tree)
			switch flagx.MustGetString(cmd, cmdx.FlagFormat) {
			case string(cmdx.FormatDefault), "":
				if tree == nil && !flagx.MustGetBool(cmd, cmdx.FlagQuiet) {
					_, _ = fmt.Fprint(cmd.OutOrStdout(), "Got an empty tree. This probably means that the requested relation tuple is not present in Keto.")
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}

	client.RegisterRemoteURLFlags(cmd.Flags())
	cmdx.RegisterJSONFormatFlags(cmd.Flags())
	cmdx.RegisterNoiseFlags(cmd.Flags())
	cmd.Flags().Int32P(FlagMaxDepth, "d", 0, "Maximum depth of the tree to be returned. If the value is less than 1 or greater than the global max-depth then the global max-depth will be used instead.")

	return cmd
}

func RegisterCommandsRecursive(parent *cobra.Command) {
	parent.AddCommand(NewExpandCmd())
}
