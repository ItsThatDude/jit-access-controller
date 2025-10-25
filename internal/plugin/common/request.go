package common

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"antware.xyz/jitaccess/api/v1alpha1"
	internal "antware.xyz/jitaccess/internal/common"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HandleRequestSelection(scope string, namespace string) (internal.AccessRequestObject, error) {
	cli, err := GetRuntimeClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	options := []internal.AccessRequestObject{}

	fieldSelector := fields.OneTermEqualSelector("status.state", string(v1alpha1.RequestStatePending))

	if scope == SCOPE_CLUSTER {
		reqList := &v1alpha1.ClusterAccessRequestList{}
		listOpts := &client.ListOptions{FieldSelector: fieldSelector}
		if err := cli.List(ctx, reqList, listOpts); err != nil {
			return nil, err
		}
		for _, r := range reqList.Items {
			options = append(options, &r)
		}
	} else {
		reqList := &v1alpha1.AccessRequestList{}
		listOpts := &client.ListOptions{Namespace: namespace, FieldSelector: fieldSelector}
		if err := cli.List(ctx, reqList, listOpts); err != nil {
			return nil, err
		}
		for _, r := range reqList.Items {
			options = append(options, &r)
		}
	}

	for i, opt := range options {
		fmt.Printf("  [%d] %s (%s)\n", i+1, opt.GetSpec().Subject, opt.GetName())
	}
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(options) {
		return nil, fmt.Errorf("an invalid selection was specified")
	}

	return options[choice-1], nil
}
