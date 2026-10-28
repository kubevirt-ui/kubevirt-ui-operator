package defaults

import consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"

// Apply fills zero values to match helm/kubevirt-console-ui/values.yaml.
func Apply(spec *consoleuiv1alpha1.KubeVirtConsoleUISpec) {
	defaultTrueBools(spec)

	if spec.Namespace == "" {
		spec.Namespace = "openshift-cnv"
	}
	if spec.HCOLabels.App == "" {
		spec.HCOLabels.App = "kubevirt-hyperconverged"
	}
	if spec.HCOLabels.ManagedBy == "" {
		spec.HCOLabels.ManagedBy = "hco-operator"
	}
	if spec.HCOLabels.PartOf == "" {
		spec.HCOLabels.PartOf = "hyperconverged-cluster"
	}
	if spec.PriorityClass.Name == "" {
		spec.PriorityClass.Name = "kubevirt-cluster-critical"
	}
	if spec.PriorityClass.Value == 0 {
		spec.PriorityClass.Value = 1000000000
	}
	if spec.PriorityClass.Description == "" {
		spec.PriorityClass.Description = "Priority class for critical KubeVirt and virtualization UI components"
	}
	if spec.PriorityClass.PreemptionPolicy == "" {
		spec.PriorityClass.PreemptionPolicy = "PreemptLowerPriority"
	}
	if spec.Plugin.Name == "" {
		spec.Plugin.Name = "kubevirt-console-plugin"
	}
	if spec.Plugin.Port == 0 {
		spec.Plugin.Port = 9443
	}
	if spec.Plugin.Replicas == 0 {
		spec.Plugin.Replicas = 1
	}
	if spec.Plugin.ServiceName == "" {
		spec.Plugin.ServiceName = "kubevirt-console-plugin-service"
	}
	if spec.Plugin.ServingCertSecret == "" {
		spec.Plugin.ServingCertSecret = "plugin-serving-cert"
	}
	if spec.Plugin.PriorityClassName == "" {
		spec.Plugin.PriorityClassName = spec.PriorityClass.Name
	}
	if spec.Proxy.Name == "" {
		spec.Proxy.Name = "kubevirt-apiserver-proxy"
	}
	if spec.Proxy.Port == 0 {
		spec.Proxy.Port = 8080
	}
	if spec.Proxy.Replicas == 0 {
		spec.Proxy.Replicas = 1
	}
	if spec.Proxy.ServiceName == "" {
		spec.Proxy.ServiceName = "kubevirt-apiserver-proxy-service"
	}
	if spec.Proxy.ServingCertSecret == "" {
		spec.Proxy.ServingCertSecret = "console-proxy-serving-cert"
	}
	if spec.Proxy.PriorityClassName == "" {
		spec.Proxy.PriorityClassName = spec.PriorityClass.Name
	}
	if spec.ConsolePlugin.Name == "" {
		spec.ConsolePlugin.Name = "kubevirt-plugin"
	}
	if spec.ConsolePlugin.DisplayName == "" {
		spec.ConsolePlugin.DisplayName = "Kubevirt Console Plugin"
	}
	if spec.ConsolePlugin.ProxyAlias == "" {
		spec.ConsolePlugin.ProxyAlias = "kubevirt-apiserver-proxy"
	}
	if spec.Config.NginxConfigMapName == "" {
		spec.Config.NginxConfigMapName = "nginx-conf"
	}
	if spec.Config.UserSettingsConfigMapName == "" {
		spec.Config.UserSettingsConfigMapName = "kubevirt-user-settings"
	}
	if spec.Config.FeaturesConfigMapName == "" {
		spec.Config.FeaturesConfigMapName = "kubevirt-ui-features"
	}
	applyFeatureDefaults(&spec.Config.Features)
	if spec.Route.Name == "" {
		spec.Route.Name = "kubevirt-apiserver-proxy"
	}
	if spec.RBAC.ConfigReaderRoleName == "" {
		spec.RBAC.ConfigReaderRoleName = "kubevirt-ui-config-reader"
	}
	if spec.RBAC.ConfigReaderRoleBindingName == "" {
		spec.RBAC.ConfigReaderRoleBindingName = "kubevirt-ui-config-reader-rolebinding"
	}
}

// defaultTrueBools sets chart defaults that are true when omitted (nil pointer).
func defaultTrueBools(spec *consoleuiv1alpha1.KubeVirtConsoleUISpec) {
	if spec.HCOLabels.Enabled == nil {
		spec.HCOLabels.Enabled = consoleuiv1alpha1.Ptr(true)
	}
	if spec.PriorityClass.Create == nil {
		spec.PriorityClass.Create = consoleuiv1alpha1.Ptr(true)
	}
	if spec.Console.RegisterPlugin == nil {
		spec.Console.RegisterPlugin = consoleuiv1alpha1.Ptr(true)
	}
	if spec.NetworkPolicy.Enabled == nil {
		spec.NetworkPolicy.Enabled = consoleuiv1alpha1.Ptr(true)
	}
	if spec.Route.Enabled == nil {
		spec.Route.Enabled = consoleuiv1alpha1.Ptr(true)
	}
}

func applyFeatureDefaults(f *consoleuiv1alpha1.FeaturesSpec) {
	if f.DisabledGuestSystemLogsAccess == "" {
		f.DisabledGuestSystemLogsAccess = "false"
	}
	if f.KubevirtApiserverProxy == "" {
		f.KubevirtApiserverProxy = "true"
	}
	if f.LoadBalancerEnabled == "" {
		f.LoadBalancerEnabled = "true"
	}
	if f.NodePortEnabled == "" {
		f.NodePortEnabled = "false"
	}
}
