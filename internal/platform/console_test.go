package platform

import (
	"context"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsOpenShift_crdPresent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiextensionsv1.AddToScheme(scheme)

	crd := &apiextensionsv1.CustomResourceDefinition{Name: ConsolePluginCRDName}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()

	ok, err := IsOpenShift(context.Background(), c)
	if err != nil || !ok {
		t.Fatalf("expected OpenShift when CRD exists, err=%v ok=%v", err, ok)
	}
}

func TestIsOpenShift_crdMissing(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiextensionsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	ok, err := IsOpenShift(context.Background(), c)
	if err != nil || ok {
		t.Fatalf("expected not OpenShift when CRD is absent, err=%v ok=%v", err, ok)
	}
}
