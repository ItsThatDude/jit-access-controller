---
sidebar_position: 1
description: Installing Kairos via Helm
---

# Install via Helm

## Installation

Add the helm repository:

```sh
helm repo add itsthatdude https://itsthatdude.github.io/helm-charts/
```

Install the helm chart:

```sh
kubectl create namespace kairos-system
helm install kairos itsthatdude/kairos -n kairos-system
```
