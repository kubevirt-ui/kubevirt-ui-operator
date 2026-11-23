package defaults

import consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"

// I would do it like this. First, I don't think we need to add a future placeholder in general. They are usually
// never used and get old badly. Specifically, you can add a +kubebuilder:default comment to fields in the api package
// to add the default value to the CRD, and let K8s fill the default for you.

// Apply fills unset spec fields. No-op today; kept for manifest context parity.
func Apply(spec *consoleuiv1alpha1.KubeVirtConsoleUISpec) {}
