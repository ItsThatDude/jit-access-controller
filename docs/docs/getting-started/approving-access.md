---
sidebar_position: 5
---

# Approving Access

## Using kubectl

```sh
kubectl apply -f - <<EOF
apiVersion: access.antware.xyz/v1alpha1
kind: AccessResponse
metadata:
  namespace: example-ns
  name: accessresponse-sample
spec:
  requestRef: accessrequest-sample
  response: (Approved|Denied)
EOF
```

## Using kubectl-access plugin

You can use the kubectl-access binary to approve requests.

>If you don't provide the request name, you will be prompted to select a request.

```sh
kubectl access (approve|reject) -n example-ns [accessrequest-sample]
```
