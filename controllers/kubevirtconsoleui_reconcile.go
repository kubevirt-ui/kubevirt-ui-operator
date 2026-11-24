package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/log"

	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/manifests"
)

// legacyPlatformRequirementsMetCondition was set by earlier operator versions; it is removed from existing CRs.
const legacyPlatformRequirementsMetCondition = "PlatformRequirementsMet"

func (r *KubeVirtConsoleUIReconciler) reconcile(
	ctx context.Context,
	cr *consoleuiv1alpha1.KubeVirtConsoleUI,
	mctx *manifests.Context,
) error {
	meta.RemoveStatusCondition(&cr.Status.Conditions, legacyPlatformRequirementsMetCondition)

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
