package builders

import (
	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/defaults"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/operconfig"
)

// Context holds a deep copy of the CR after defaults.Apply. Operands deploy in OperandNamespace from operator config.
type Context struct {
	CR                *consoleuiv1alpha1.KubeVirtConsoleUI
	Spec              consoleuiv1alpha1.KubeVirtConsoleUISpec
	OperandNamespace  string
	HCONamespace      string
	HCOInstalled      bool
	PluginImage string
	ProxyImage  string
}

// NewContext deep-copies the CR and applies default spec values for manifest generation.
func NewContext(cr *consoleuiv1alpha1.KubeVirtConsoleUI, cfg operconfig.Config) *Context {
	copy := cr.DeepCopy() // please rename the copy var. copy is a golang builtin function. Please avoid shadowing it.
	defaults.Apply(&copy.Spec)
	return &Context{
		CR:               copy,
		Spec:             copy.Spec, // why do you need both? CR already contains it
		OperandNamespace: cfg.OperandNamespace, // all these five fields wil never changed on runtime. consider using
		                                        // cfg, as a new filed instead of these fields.
		HCONamespace:     cfg.HCONamespace,
		HCOInstalled:     cfg.HCOInstalled,
		PluginImage:      cfg.PluginImage,
		ProxyImage:       cfg.ProxyImage,
	}
}
