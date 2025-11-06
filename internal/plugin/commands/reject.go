package commands

import (
	"fmt"

	"github.com/itsthatdude/jitaccess-controller/api/v1alpha1"
	"github.com/itsthatdude/jitaccess-controller/internal/plugin/common"
	"github.com/spf13/cobra"
)

func NewRejectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject [<request_name>]",
		Short: "Reject an access request",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println("Please select a request to reject:")
				selection, err := common.HandleRequestSelection(scope, namespace)
				if err != nil {
					return fmt.Errorf("an error occurred while requesting a selection: %s", err)
				}
				if selection == nil {
					return fmt.Errorf("the selected request was nil")
				}
				return common.CreateResponse(scope, namespace, selection.GetName(), v1alpha1.ResponseStateDenied)
			} else {
				name := args[0]
				return common.CreateResponse(scope, namespace, name, v1alpha1.ResponseStateDenied)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace for the access request")
	cmd.Flags().StringVar(&scope, "scope", "namespace", "Scope of the request (namespace|cluster)")

	return cmd
}
