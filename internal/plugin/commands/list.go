package commands

import (
	"context"
	"fmt"

	"github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"github.com/itsthatdude/jit-access-controller/internal/plugin/common"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var listScope string
var listNamespace string

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIT access requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := common.GetRuntimeClient()
			if err != nil {
				return err
			}
			ctx := context.Background()

			if listScope == common.SCOPE_CLUSTER {
				reqList := &v1alpha1.ClusterAccessRequestList{}
				if err := cli.List(ctx, reqList); err != nil {
					return err
				}

				if len(reqList.Items) == 0 {
					fmt.Println("No Access Requests are pending.")
					return nil
				}

				fmt.Printf("ClusterAccessRequests:\n")
				for _, r := range reqList.Items {
					state := r.Status.State

					fmt.Printf("  - Name: %s\n", r.Name)
					fmt.Printf("    Subject: %s\n", r.Spec.Subject)
					fmt.Printf("    State: %s\n", r.Status.State)
					if state != v1alpha1.RequestStateApproved {
						fmt.Printf("    Expires: %s\n", r.Status.RequestExpiresAt.String())
					}
					fmt.Printf("    Justification: %s\n", r.Spec.Justification)
				}
			} else {
				reqList := &v1alpha1.AccessRequestList{}
				listOpts := &client.ListOptions{Namespace: listNamespace}
				if err := cli.List(ctx, reqList, listOpts); err != nil {
					return err
				}

				if len(reqList.Items) == 0 {
					fmt.Println("No Access Requests are pending.")
					return nil
				}

				fmt.Printf("AccessRequests in namespace %s:\n", listNamespace)
				for _, r := range reqList.Items {
					state := r.Status.State

					fmt.Printf("  - Name: %s\n", r.Name)
					fmt.Printf("    Subject: %s\n", r.Spec.Subject)
					fmt.Printf("    State: %s\n", r.Status.State)
					if state != v1alpha1.RequestStateApproved {
						fmt.Printf("    Expires: %s\n", r.Status.RequestExpiresAt.String())
					}
					fmt.Printf("    Justification: %s\n", r.Spec.Justification)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&listScope, "scope", "namespace", "Scope to list (namespace|cluster)")
	cmd.Flags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace for listing requests")

	return cmd
}
