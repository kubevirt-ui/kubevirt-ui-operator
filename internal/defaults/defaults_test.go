package defaults

import (
	"testing"

	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
)

func TestApply_defaultTrueBools(t *testing.T) {
	spec := &consoleuiv1alpha1.KubeVirtConsoleUISpec{}
	Apply(spec)

	if !consoleuiv1alpha1.BoolOrDefault(spec.HCOLabels.Enabled, false) {
		t.Fatal("expected hcoLabels.enabled true by default")
	}
	if !consoleuiv1alpha1.BoolOrDefault(spec.PriorityClass.Create, false) {
		t.Fatal("expected priorityClass.create true by default")
	}
	if !consoleuiv1alpha1.BoolOrDefault(spec.Console.RegisterPlugin, false) {
		t.Fatal("expected console.registerPlugin true by default")
	}
	if !consoleuiv1alpha1.BoolOrDefault(spec.NetworkPolicy.Enabled, false) {
		t.Fatal("expected networkPolicy.enabled true by default")
	}
	if !consoleuiv1alpha1.BoolOrDefault(spec.Route.Enabled, false) {
		t.Fatal("expected route.enabled true by default")
	}
}

func TestApply_explicitFalse(t *testing.T) {
	spec := &consoleuiv1alpha1.KubeVirtConsoleUISpec{
		NetworkPolicy: consoleuiv1alpha1.NetworkPolicySpec{
			Enabled: consoleuiv1alpha1.Ptr(false),
		},
	}
	Apply(spec)
	if consoleuiv1alpha1.BoolOrDefault(spec.NetworkPolicy.Enabled, true) {
		t.Fatal("expected networkPolicy.enabled to stay false")
	}
}
