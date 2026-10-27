# Helm chart ↔ operator parity matrix

Track reconciliation parity between `helm/kubevirt-console-ui` templates and the operator controller. Update this table as operands land in Phase 1–2.

| Helm template | Resource(s) | Operator (Phase) | Notes |
|---------------|-------------|------------------|-------|
| `namespace.yaml` | Namespace | TBD | `createNamespace` |
| `priorityclass.yaml` | PriorityClass | TBD | Cluster-scoped |
| `serviceaccount-plugin.yaml` | ServiceAccount | TBD | |
| `serviceaccount-proxy.yaml` | ServiceAccount | TBD | |
| `configmap-nginx.yaml` | ConfigMap | TBD | `nginx-conf` |
| `configmap-user-settings.yaml` | ConfigMap | TBD | |
| `configmap-ui-features.yaml` | ConfigMap | TBD | |
| `deployment-plugin.yaml` | Deployment | TBD | SCC `restricted-v2` |
| `deployment-proxy.yaml` | Deployment | TBD | |
| `service-plugin.yaml` | Service | TBD | Serving cert annotation |
| `service-proxy.yaml` | Service | TBD | Serving cert annotation |
| `networkpolicy-plugin.yaml` | NetworkPolicy | TBD | Console ingress from `openshift-console` |
| `networkpolicy-proxy.yaml` | NetworkPolicy | TBD | Egress to API |
| `route-proxy.yaml` | Route | TBD | Optional |
| `consoleplugin.yaml` | ConsolePlugin | TBD | `console.openshift.io` |
| `console-register-plugin.yaml` | Console cluster patch | TBD | Helm hook today |
| `role-config-reader.yaml` | Role | TBD | |
| `rolebinding-config-reader.yaml` | RoleBinding | TBD | |

**Legend:** TBD → not implemented in operator yet.
