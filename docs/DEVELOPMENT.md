# Development

## Helm chart

```bash
helm lint ./helm/kubevirt-console-ui
helm template kubevirt-ui ./helm/kubevirt-console-ui \
  --set plugin.image=quay.io/kubevirt-ui/kubevirt-plugin:<tag> \
  --set proxy.image=quay.io/kubevirt-ui/kubevirt-apiserver-proxy:<tag>
```

Smoke install on OpenShift: see [helm/kubevirt-console-ui/README.md](../helm/kubevirt-console-ui/README.md).

## Operator (coming)

Follow [ROADMAP.md — Phase 1](ROADMAP.md#phase-1--scaffold--minimal-reconcile-24-weeks). API types live under `api/v1alpha1/`; Helm parity helpers under `internal/defaults` and `internal/labels`.

```bash
go test ./...
```

Local controller workflow will mirror [kube-gateway-operator](https://github.com/kubevirt-ui/kube-gateway-operator) (`make install`, `make run`) once the manager scaffold lands.

## Cluster prerequisites

Console and cluster requirements: [ROADMAP.md — Prerequisites](ROADMAP.md#prerequisites-cluster-and-console).
