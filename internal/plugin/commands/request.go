package commands

import (
	"context"
	"fmt"

	"antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/common"
	plugin "antware.xyz/jitaccess/internal/plugin/common"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

func NewRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "Request access",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := plugin.GetRuntimeClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			rules := plugin.ParsePermissions(permissions)

			if scope == plugin.SCOPE_CLUSTER {
				req := &v1alpha1.ClusterAccessRequest{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "access-request-",
					},
					Spec: v1alpha1.ClusterAccessRequestSpec{
						AccessRequestBaseSpec: v1alpha1.AccessRequestBaseSpec{
							Role:          rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: roleKindStr, Name: role},
							Permissions:   rules,
							Duration:      duration,
							Justification: justification,
						},
					},
				}

				if subject != "" {
					req.Spec.Subject = subject
				}

				if err := cli.Create(ctx, req); err != nil {
					return err
				}
				fmt.Printf("ClusterAccessRequest created: %s\n", req.Name)
			} else {
				req := &v1alpha1.AccessRequest{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "access-request-",
						Namespace:    namespace,
					},
					Spec: v1alpha1.AccessRequestSpec{
						AccessRequestBaseSpec: v1alpha1.AccessRequestBaseSpec{
							Role:          rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: roleKindStr, Name: role},
							Permissions:   rules,
							Duration:      duration,
							Justification: justification,
						},
					},
				}

				if subject != "" {
					req.Spec.Subject = subject
				}

				if err := cli.Create(ctx, req); err != nil {
					return err
				}
				fmt.Printf("AccessRequest created: %s/%s\n", namespace, req.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace for the access request")
	cmd.Flags().StringVar(&scope, "scope", "namespace", "Scope of the request (namespace|cluster)")
	cmd.Flags().StringVar(&role, "role", "", "Role to request")
	cmd.Flags().StringVar(&roleKindStr, "roleKind", common.RoleKindRole, "Role kind (Role|ClusterRole)")
	cmd.Flags().StringSliceVar(&permissions, "permissions", []string{}, "List of permissions (verbs:resources)")
	cmd.Flags().StringVar(&duration, "duration", "1h", "Duration in seconds for the access")
	cmd.Flags().StringVar(&justification, "justification", "", "Justification for the request")
	cmd.Flags().StringVar(&subject, "subject", "", "Requesting subject (e.g. username)")

	return cmd
}
