# kubevirt-console-ui Helm chart

Deploys the **OpenShift Virtualization console UI stack** without the HCO/CNV backend operands (KubeVirt, CDI, and so on). Resource names and wiring match what the [HyperConverged Cluster Operator](https://github.com/kubevirt/hyperconverged-cluster-operator) reconciles for UI (`controllers/handlers/kubevirtConsolePlugin.go`).

## What gets installed

| Resource | Name (defaults) |
|----------|-------------------|
| Deployment | `kubevirt-console-plugin`, `kubevirt-apiserver-proxy` |
| Service | `kubevirt-console-plugin-service`, `kubevirt-apiserver-proxy-service` |
| ConfigMap | `nginx-conf`, `kubevirt-user-settings`, `kubevirt-ui-features` |
| ConsolePlugin | `kubevirt-plugin` |
| Role / RoleBinding | `kubevirt-ui-config-reader` (authenticated users → UI configmaps) |

Optional: NetworkPolicies (when `networkPolicy.enabled`), Route for external proxy access (`route.enabled`), post-install Console operator plugin registration (`console.registerPlugin`).

## Prerequisites

- OpenShift cluster with the Console operator and `ConsolePlugin` CRD
- Namespace (default `openshift-cnv`) — typically already present for virtualization
- **KubeVirt control plane** reachable via the Kubernetes API (this chart does not install it)
- Avoid duplicate installs: if CNV/HCO already manages the UI in `openshift-cnv`, use OLM upgrades instead of this chart

## Install

```bash
helm upgrade --install kubevirt-ui ./helm/kubevirt-console-ui \
  --namespace openshift-cnv \
  --set plugin.image=quay.io/kubevirt-ui/kubevirt-plugin:<tag> \
  --set proxy.image=quay.io/kubevirt-ui/kubevirt-apiserver-proxy:<tag>
```

Pin images to the same versions shipped with your CNV release (see HCO `deploy/images.env`: `KUBEVIRT_CONSOLE_PLUGIN_IMAGE`, `KUBEVIRT_CONSOLE_PROXY_IMAGE`).

### Common values

| Value | Purpose |
|-------|---------|
| `plugin.image` / `proxy.image` | Container images (required) |
| `plugin.replicas` / `proxy.replicas` | HA replica count (HCO uses 2 on highly available infra) |
| `networkPolicy.enabled` | Deploy console + proxy network policies (HCO when `DEPLOY_NETWORK_POLICIES`) |
| `route.enabled` + `route.host` | Public Route for apiserver-proxy (local/off-cluster console) |
| `console.registerPlugin` | Helm hook patches `consoles.operator.openshift.io/cluster` |
| `hcoLabels.enabled` | Label resources like HCO (`app: kubevirt-hyperconverged`, and so on) |

## Uninstall

```bash
helm uninstall kubevirt-ui -n openshift-cnv
```

Remove `kubevirt-plugin` from the cluster console plugins list if no other installation provides it:

```bash
oc patch consoles.operator.openshift.io cluster --type=json \
  -p='[{"op": "remove", "path": "/spec/plugins/<index>"}]'
```

(Use `oc get consoles.operator.openshift.io cluster -o yaml` to find the plugin index.)
