---
slug: /
sidebar_position: 1
---

# Introduction

## Overview

The **Just-in-time Access Controller** enables users to request **just-in-time (JIT) access** to Kubernetes resources.

Users submit either a:

- **`AccessRequest`** – for namespace-scoped access  
- **`ClusterAccessRequest`** – for cluster-wide access  

In each request, the user specifies the **roles** or **permissions** they require.  

An approver then submits either a:  

- **`AccessResponse`** – for namespace-scoped access  
- **`ClusterAccessResponse`** – for cluster-wide access

The controller evaluates the requests/responses against configured policies:

- **`AccessPolicy`** – defines rules for namespace-scoped access requests  
- **`ClusterAccessPolicy`** – defines rules for cluster-scoped access requests

If the responses fulful the required number of approvals, the controller creates a **`AccessGrant`** object.  
The **`AccessGrant`** is then reconciled and creates the requested Kubernetes RBAC objects:

- A **ClusterRole** or **Role** for adhoc permission requests
- A **ClusterRoleBinding** or **RoleBinding** for both role and adhoc permission requests

Each request must also define a **Duration** which must be within the maximum configured in the policy.  
Once the duration expires, the controller automatically revokes access by removing any created roles and bindings.
