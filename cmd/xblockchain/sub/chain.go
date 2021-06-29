package sub

import "github.com/spf13/cobra"

var (
	chainCommand = &cobra.Command{
		Use:   "chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(chainCommand)
}