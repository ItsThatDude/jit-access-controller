package common

import (
	"context"
	"fmt"

	"antware.xyz/kairos/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createResponse creates a AccessResponse or ClusterAccessResponse
func CreateResponse(scope string, namespace string, requestName string, state v1alpha1.ResponseState) error {
	cli, err := GetRuntimeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	if scope == SCOPE_CLUSTER {
		resp := &v1alpha1.ClusterAccessResponse{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "response-",
			},
			Spec: v1alpha1.AccessResponseSpec{
				RequestRef: requestName,
				Response:   state,
			},
		}
		if err := cli.Create(ctx, resp); err != nil {
			return err
		}
		fmt.Printf("ClusterAccessResponse created for request %s\n", requestName)
	} else {
		resp := &v1alpha1.AccessResponse{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "response-",
				Namespace:    namespace,
			},
			Spec: v1alpha1.AccessResponseSpec{
				RequestRef: requestName,
				Response:   state,
			},
		}
		if err := cli.Create(ctx, resp); err != nil {
			return err
		}
		fmt.Printf("AccessResponse created for request %s/%s\n", namespace, requestName)
	}
	return nil
}
