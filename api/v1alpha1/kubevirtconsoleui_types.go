package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeReady                   = "Ready"
	ConditionTypeProgressing             = "Progressing"
	ConditionTypeDegraded                = "Degraded"
	ConditionTypePlatformRequirementsMet = "PlatformRequirementsMet"
)

// KubeVirtConsoleUISpec defines optional day-2 settings for the console UI stack.
// Operand names, images, and namespace are fixed by the operator (see OPERAND_NAMESPACE and image env vars).
// Feature toggles and hcoVersion live in the kubevirt-ui-features ConfigMap (operator-managed).
type KubeVirtConsoleUISpec struct {
	// Proxy holds optional proxy container flags (for example TLS profile args on OpenShift).
	Proxy ProxySpec `json:"proxy,omitempty"`
}

// I would say that the operator should build the argument for the proxy, because it knows it the best.
// I wouldn't let the user to do that.
type ProxySpec struct {
	// We are required to use the common TLS configurations from the "cluster" APIServer CR. In addition, we need the
	// same TLS configurations also for the UI plugin pods, and for this operator pod.
	//
	// I can think about three options to achieve that here:
	// (1) HCO is already watches the APIServer CR, and set the TLS configurations for all the rest of the CNV
	//     components. For that I think we should use the OpenShift github.com/openshift/api/config/v1.TLSSecurityProfile
	//     type, in order to use the OpenShift common language, and avoid asking the user or HCO to learn a new one.
	//
	//     So we can move the field up, to the spec field, rename it to "TLSSecurityProfile" and change the type to
	//     github.com/openshift/api/config/v1.TLSSecurityProfile; alternatively, copy the file
	//     https://github.com/openshift/api/blob/master/config/v1/types_tlssecurityprofile.go to this repo and use the
	//     local type. Then we could use many helper functions from OpenShift (github.com/openshift/library-go/pkg/crypto),
	//     to translate the TLS configurations to be used in three places: this operator pod, the ui plugin pods and the
	//     proxy pods.
	//
	// (2) Remove this field completely. Instead, watch the OpenShift's APIServer CR (named "cluster"), and use the
	//     common configuration directly, to again, configure the TLS in the same three places.
	//
	// (3) Combination of the two options above: expose the TLS configurations as in #1, but use the APIServer CR by
	//     default. So only if the field is set, we'll use it. But usually, the user will have to set only one place -
	//     the APIServer CR
	//
	// I think I would choose the second option for simplicity. #3 is a bit safer but more complex. #1 is not really an
	// option, because if the operator is deployed as standalone, there is no HCO to wath the APIServer CR and to set
	// the TLS config.
	TLSArgs []string `json:"tlsArgs,omitempty"`
}

// KubeVirtConsoleUIStatus defines observed state.
type KubeVirtConsoleUIStatus struct {
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	ConsolePluginRegistered bool `json:"consolePluginRegistered,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=kvui
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// KubeVirtConsoleUI deploys the KubeVirt OpenShift console plugin stack.
// The singleton instance must be named "cluster"; operand namespace comes from OPERAND_NAMESPACE.
type KubeVirtConsoleUI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeVirtConsoleUISpec   `json:"spec,omitempty"`
	Status KubeVirtConsoleUIStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type KubeVirtConsoleUIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeVirtConsoleUI `json:"items"`
}
