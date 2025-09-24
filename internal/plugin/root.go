package plugin

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-access",
	Short: "Manage JIT access requests",
}

func Execute() {
	rootCmd.AddCommand(newRequestCmd())
	rootCmd.AddCommand(newApproveCmd())
	rootCmd.AddCommand(newRejectCmd())
	rootCmd.AddCommand(newListCmd())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
