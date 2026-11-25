package controllers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"

	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/manifests"
)

func (r *KubeVirtConsoleUIReconciler) reconcile(
	ctx context.Context,
	cr *consoleuiv1alpha1.KubeVirtConsoleUI,
	mctx *manifests.Context,
) error {
	if err := manifests.ReconcileStack(ctx, r.Client, r.Scheme, mctx); err != nil {
		return err
	}

	registered, regErr := r.reconcileConsoleRegistration(ctx, mctx)
	cr.Status.ConsolePluginRegistered = registered
	if regErr != nil {
		log.FromContext(ctx).Error(regErr, "console plugin registration failed")
		return regErr
	}
	return nil
}
