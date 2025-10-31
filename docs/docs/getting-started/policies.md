---
sidebar_position: 3
description: Creating Kairos policies
---

# Policies

## Create a policy

Kairos utilizes policies to enforce who can restrict access to resources.

The following policy allows `user1` to request either the `view` role, or adhoc permissions to get and modify pods in the `example-ns` namespace.

> When requesting adhoc permissions, a role is created on-demand and then later cleaned up when the grant expires.

### Cluster Policy

A cluster-scoped policy allows a user to request `ClusterRole` &rarr; `ClusterRoleBindings`.  
This allows a user to request access to cluster-scoped CRDs and cluster-wide access to namespace scoped CRDs.

```sh
kubectl apply -f - <<EOF
apiVersion: access.antware.xyz/v1alpha1
kind: ClusterAccessPolicy
metadata:
  name: accesspolicy-sample
spec:
  subjects:
    - user1
  # allow the user to request the binding of roles
  allowedRoles:
    - view
  # allow the user to request specific permissions
  allowedPermissions:
    - apiGroups: [""]
      resources: ["pods"]
      verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  maxDuration: "60m"
  requiredApprovals: 1
  approvers:
    - admin
EOF
```

### Namespaced Policy Example

A namespace-scoped policy allows a user to request `ClusterRole`/`Role` &rarr; `RoleBindings`.  
This allows a user to request access to namespace-scoped CRDs in the requested namespace.

```sh
kubectl apply -f - <<EOF
apiVersion: access.antware.xyz/v1alpha1
kind: AccessPolicy
metadata:
  namespace: example-ns
  name: accesspolicy-sample
spec:
  subjects:
    - user1
  # allow the user to request the binding of roles
  allowedRoles:
    - view
  # allow the user to request specific permissions
  allowedPermissions:
    - apiGroups: [""]
      resources: ["pods"]
      verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  maxDuration: "60m"
  requiredApprovals: 1
  approvers:
    - admin
EOF
```