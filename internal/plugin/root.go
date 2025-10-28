package plugin

import (
	"os"

	"antware.xyz/kairos/internal/plugin/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "kubectl-access",
	Short:        "Manage JIT access requests",
	SilenceUsage: true,
}

func Execute() {
	rootCmd.AddCommand(commands.NewRequestCmd())
	rootCmd.AddCommand(commands.NewApproveCmd())
	rootCmd.AddCommand(commands.NewRejectCmd())
	rootCmd.AddCommand(commands.NewListCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
