package platform

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConsolePluginCRDName is the OpenShift ConsolePlugin CRD resource name.
const ConsolePluginCRDName = "consoleplugins.console.openshift.io"

// IsOpenShift reports whether the cluster exposes the OpenShift console APIs (ConsolePlugin CRD).
func IsOpenShift(ctx context.Context, c client.Reader) (bool, error) {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := c.Get(ctx, client.ObjectKey{Name: ConsolePluginCRDName}, crd); err != nil {
		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
