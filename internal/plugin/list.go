package plugin

import (
	"context"
	"fmt"

	"antware.xyz/jitaccess/api/v1alpha1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var listScope string
var listNamespace string

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIT access requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := getRuntimeClient()
			if err != nil {
				return err
			}
			ctx := context.Background()

			if listScope == SCOPE_CLUSTER {
				reqList := &v1alpha1.ClusterJITAccessRequestList{}
				if err := cli.List(ctx, reqList); err != nil {
					return err
				}
				fmt.Printf("ClusterJITAccessRequests:\n")
				for _, r := range reqList.Items {
					state := r.Status.State
					fmt.Printf("- %s : %s\n", r.Name, state)
				}
			} else {
				reqList := &v1alpha1.JITAccessRequestList{}
				listOpts := &client.ListOptions{Namespace: listNamespace}
				if err := cli.List(ctx, reqList, listOpts); err != nil {
					return err
				}
				fmt.Printf("JITAccessRequests in namespace %s:\n", listNamespace)
				for _, r := range reqList.Items {
					state := r.Status.State

					if state == v1alpha1.RequestStateApproved {
						fmt.Printf("- %s : %s (Expires %s)\n", r.Name, state, r.Status.ExpiresAt)
					} else {
						fmt.Printf("- %s : %s\n", r.Name, state)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&listScope, "scope", "namespace", "Scope to list (namespace|cluster)")
	cmd.Flags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace for listing requests")

	return cmd
}
