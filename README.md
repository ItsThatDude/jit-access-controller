[![Lint](https://github.com/ItsThatDude/jit-access-controller/actions/workflows/lint.yml/badge.svg)](https://github.com/ItsThatDude/jit-access-controller/actions/workflows/lint.yml)  [![Tests](https://github.com/ItsThatDude/jit-access-controller/actions/workflows/test.yml/badge.svg)](https://github.com/ItsThatDude/jit-access-controller/actions/workflows/test.yml)  [![Test Chart](https://github.com/ItsThatDude/jit-access-controller/actions/workflows/test-chart.yml/badge.svg)](https://github.com/ItsThatDude/jit-access-controller/actions/workflows/test-chart.yml)

# jit-access-controller

The **JIT Access Controller** enables users to request **just-in-time (JIT) access** to Kubernetes resources.

## Description

This controller provides a mechanism for granting temporary, on-demand access to resources in Kubernetes clusters.

Users submit either a:

- **`AccessRequest`** – for namespace-scoped access  
- **`ClusterAccessRequest`** – for cluster-wide access  

In each request, the user specifies the **roles** or **permissions** they require.  

The controller evaluates the request against configured policies:

- **`AccessPolicy`** – defines rules for namespace-scoped access requests  
- **`ClusterAccessPolicy`** – defines rules for cluster-scoped access requests  

If the request matches a policy, the controller creates a **`AccessGrant`** object.  
The **`AccessGrant`** is then reconciled and creates the requested Kubernetes RBAC objects:

- A **ClusterRole** or **Role** for adhoc permission requests
- A **ClusterRoleBinding** or **RoleBinding** for both role and adhoc permission requests

Each request must also define a **Duration** which must be within the maximum configured in the policy.  
Once the duration expires, the controller automatically revokes access by removing any created roles and bindings.

### Examples
#### ns-policy.yaml
```yaml
apiVersion: access.antware.xyz/v1alpha1
kind: AccessPolicy
metadata:
  namespace: example
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
      verbs: ["get", "list", "watch"]
  maxDuration: "60m"
  requiredApprovals: 1
  approvers:
    - admin
```
#### ns-request.yaml
In the example below, we specify the subject - if this is omitted, the user submitting the request will be used.
```yaml
apiVersion: access.antware.xyz/v1alpha1
kind: AccessRequest
metadata:
  namespace: example
  name: accessrequest-sample
spec:
  subject: user1
  duration: "5m"
  justification: "This is a sample request"
  # We can specify a pre-defined role:
  role: view
  roleKind: Role # This can be either Role or ClusterRole
  # Or we can specify a list of permissions we want to request:
  permissions: 
    - apiGroups: [""]
      resources: ["pods"]
      verbs: ["get", "list", "watch"]
```
#### ns-response.yaml
For responses, the user submitting the request will be used as the approver.
```yaml
apiVersion: access.antware.xyz/v1alpha1
kind: AccessResponse
metadata:
  namespace: example
  name: accessresponse-sample
spec:
  requestRef: accessrequest-sample
  response: Approved
```

## Getting Started

Install helm repository:
```sh
helm repo add itsthatdude https://itsthatdude.github.io/helm-charts/`
```

Install the chart
```sh
kubectl create namespace jitaccess-system
helm install jitaccess -n jitaccess-system
```

Create a policy
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

Request access  

Using kubectl:
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

Using kubectl-access plugin:
```sh
kubectl-access request -n example-ns --subject "user1" --permissions "get,list,watch,create,update,patch,delete:pods"
```

Approve/Reject access
Using kubectl:
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
Using kubectl-access plugin:
```sh
kubectl-access (approve|reject) -n example-ns accessrequest-sample
```

## Development

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/jitaccess:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/jitaccess:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/jitaccess:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/jitaccess/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

