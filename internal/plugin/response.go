package plugin

import (
	"context"
	"fmt"

	"antware.xyz/jitaccess/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespace string
	scope     string
	approver  string
)

// createResponse creates a JITAccessResponse or ClusterJITAccessResponse
func createResponse(requestName string, state v1alpha1.ResponseState) error {
	cli, err := getRuntimeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	if scope == "cluster" {
		resp := &v1alpha1.ClusterJITAccessResponse{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "response-",
			},
			Spec: v1alpha1.JITAccessResponseSpec{
				RequestRef: requestName,
				Approver:   approver,
				Response:   state,
			},
		}
		if err := cli.Create(ctx, resp); err != nil {
			return err
		}
		fmt.Printf("ClusterJITAccessResponse created for request %s by %s\n", requestName, approver)
	} else {
		resp := &v1alpha1.JITAccessResponse{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "response-",
				Namespace:    namespace,
			},
			Spec: v1alpha1.JITAccessResponseSpec{
				RequestRef: requestName,
				Approver:   approver,
				Response:   state,
			},
		}
		if err := cli.Create(ctx, resp); err != nil {
			return err
		}
		fmt.Printf("JITAccessResponse created for request %s/%s by %s\n", namespace, requestName, approver)
	}
	return nil
}
