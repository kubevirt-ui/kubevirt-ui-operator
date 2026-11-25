package builders

import (
	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/operconfig"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/tlsprofile"
)

// Context holds a deep copy of the CR plus operator config. Operands deploy in OperandNamespace from operator config.
type Context struct {
	CR *consoleuiv1alpha1.KubeVirtConsoleUI
	operconfig.Config
	// TLS is the resolved cluster TLS profile applied to the proxy and plugin nginx.
	TLS tlsprofile.Settings
}

// NewContext deep-copies the CR and pairs it with operator config and TLS settings for manifest generation.
func NewContext(cr *consoleuiv1alpha1.KubeVirtConsoleUI, cfg operconfig.Config, tls tlsprofile.Settings) *Context {
	crCopy := cr.DeepCopy()
	return &Context{CR: crCopy, Config: cfg, TLS: tls}
}
