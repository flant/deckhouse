---
title: "The operator-prometheus module: configuration"
---

The module does not require configuration.

## Parameters

* `nodeSelector` — the same as in the pods' `spec.nodeSelector` parameter in Kubernetes;
    * If the parameter is omitted of `false`, it will be determined [automatically](../../#advanced-scheduling).
* `tolerations` — the same as in the pods' `spec.tolerations` parameter in Kubernetes;
    * If the parameter is omitted of `false`, it will be determined [automatically](../../#advanced-scheduling).
