package commands

import (
	"fmt"

	"antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/plugin/common"
	"github.com/spf13/cobra"
)

func NewApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [<request_name>]",
		Short: "Approve an access request",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println("Please select a request to approve:")
				selection, err := common.HandleRequestSelection(scope, namespace)
				if err != nil {
					return fmt.Errorf("an error occurred while requesting a selection: %s", err)
				}
				if selection == nil {
					return fmt.Errorf("the selected request was nil")
				}
				return common.CreateResponse(scope, namespace, selection.GetName(), v1alpha1.ResponseStateApproved)
			} else {
				name := args[0]
				return common.CreateResponse(scope, namespace, name, v1alpha1.ResponseStateApproved)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace for the access request")
	cmd.Flags().StringVar(&scope, "scope", "namespace", "Scope of the request (namespace|cluster)")

	return cmd
}
