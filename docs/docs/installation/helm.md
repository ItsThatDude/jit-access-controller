---
sidebar_position: 1
description: Installing JITAccess via Helm
---

# Install via Helm

## Installation

Add the helm repository:

```sh
helm repo add itsthatdude https://itsthatdude.github.io/helm-charts/
```

Install the helm chart:

```sh
kubectl create namespace jitaccess-system
helm install jitaccess itsthatdude/jitaccess -n jitaccess-system
```
