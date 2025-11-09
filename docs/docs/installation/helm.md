---
sidebar_position: 1
description: Installing Jit-access via Helm
---

# Install via Helm

## Installation

Add the helm repository:

```sh
helm repo add itsthatdude https://itsthatdude.github.io/helm-charts/
```

Install the helm chart:

```sh
kubectl create namespace jit-access-system
helm install jit-access itsthatdude/jit-access -n jit-access-system
```
