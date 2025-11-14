package commands

import (
	"fmt"

	"github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"github.com/itsthatdude/jit-access-controller/internal/plugin/common"
	"github.com/spf13/cobra"
)

func newResponseCommand(action string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   action + " [<request_name>]",
		Short: fmt.Sprintf("%s an access request", action),
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Printf("Please select a request to %s:\n", action)
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

func NewApproveCmd() *cobra.Command {
	return newResponseCommand("approve")
}

func NewRejectCmd() *cobra.Command {
	return newResponseCommand("reject")
}
