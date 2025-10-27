---
sidebar_position: 4
---

# Requesting Access

## Using kubectl

```sh
kubectl apply -f - <<EOF
apiVersion: access.antware.xyz/v1alpha1
kind: AccessRequest
metadata:
  namespace: example-ns
  name: accessrequest-sample
spec:
  subject: user1
  duration: "5m"
  justification: "This is a sample request"
  permissions: 
    - apiGroups: [""]
      resources: ["pods"]
      verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
EOF
```

## Using kubectl-access plugin

```sh
kubectl access request -n example-ns --subject "user1" --permissions "get,list,watch,create,update,patch,delete:pods"
```