package manifests

import (
	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/manifests/builders"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/operconfig"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/tlsprofile"
)

// Context is the manifest reconciliation context.
type Context = builders.Context

// NewContext deep-copies the CR and pairs it with operator config and TLS settings for manifest generation.
func NewContext(cr *consoleuiv1alpha1.KubeVirtConsoleUI, cfg operconfig.Config, tls tlsprofile.Settings) *Context {
	return builders.NewContext(cr, cfg, tls)
}
