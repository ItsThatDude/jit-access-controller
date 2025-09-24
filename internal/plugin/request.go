package plugin

import (
	"context"
	"fmt"

	"antware.xyz/jitaccess/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

var (
	role          string
	roleKindStr   string
	permissions   []string
	duration      int64
	justification string
	subject       string
)

func newRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "Request access",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := getRuntimeClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			rules := parsePermissions(permissions)

			if scope == "cluster" {
				req := &v1alpha1.ClusterJITAccessRequest{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "access-request-",
					},
					Spec: v1alpha1.ClusterJITAccessRequestSpec{
						JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
							Subject:         subject,
							Role:            role,
							Permissions:     rules,
							DurationSeconds: duration,
							Justification:   justification,
						},
					},
				}
				if err := cli.Create(ctx, req); err != nil {
					return err
				}
				fmt.Printf("ClusterJITAccessRequest created: %s\n", req.Name)
			} else {
				req := &v1alpha1.JITAccessRequest{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "access-request-",
						Namespace:    namespace,
					},
					Spec: v1alpha1.JITAccessRequestSpec{
						JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
							Subject:         subject,
							Role:            role,
							Permissions:     rules,
							DurationSeconds: duration,
							Justification:   justification,
						},
						RoleKind: v1alpha1.RoleKind(roleKindStr),
					},
				}
				if err := cli.Create(ctx, req); err != nil {
					return err
				}
				fmt.Printf("JITAccessRequest created: %s/%s\n", namespace, req.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace for the access request")
	cmd.Flags().StringVar(&scope, "scope", "namespace", "Scope of the request (namespace|cluster)")
	cmd.Flags().StringVar(&role, "role", "", "Role to request")
	cmd.Flags().StringVar(&roleKindStr, "roleKind", "Role", "Role kind (Role|ClusterRole)")
	cmd.Flags().StringSliceVar(&permissions, "permissions", []string{}, "List of permissions (verbs:resources)")
	cmd.Flags().Int64Var(&duration, "duration", 3600, "Duration in seconds for the access")
	cmd.Flags().StringVar(&justification, "justification", "", "Justification for the request")
	cmd.Flags().StringVar(&subject, "subject", "", "Requesting subject (e.g. username)")
	_ = cmd.MarkFlagRequired("subject")

	return cmd
}
