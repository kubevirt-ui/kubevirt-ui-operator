package manifests

import (
	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/manifests/builders"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/operconfig"
)

// Context is the manifest reconciliation context.
type Context = builders.Context

// NewContext deep-copies the CR and pairs it with operator config for manifest generation.
func NewContext(cr *consoleuiv1alpha1.KubeVirtConsoleUI, cfg operconfig.Config) *Context {
	return builders.NewContext(cr, cfg)
}
