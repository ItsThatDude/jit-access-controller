package commands

import (
	"antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/plugin/common"
	"github.com/spf13/cobra"
)

func NewRejectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject <request_name>",
		Short: "Reject an access request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return common.CreateResponse(scope, namespace, name, v1alpha1.ResponseStateDenied)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace for the access request")
	cmd.Flags().StringVar(&scope, "scope", "namespace", "Scope of the request (namespace|cluster)")

	return cmd
}
