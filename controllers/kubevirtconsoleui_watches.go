package controllers

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/hcoversion"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/names"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/tlsprofile"
)

func (r *KubeVirtConsoleUIReconciler) mapConsolePluginToRequests(ctx context.Context, obj client.Object) []reconcile.Request {
	if cp, ok := obj.(*consolev1.ConsolePlugin); ok && cp.GetName() == names.ConsolePluginName {
		return r.enqueueClusterCR()
	}
	return nil
}

func (r *KubeVirtConsoleUIReconciler) mapPriorityClassToRequests(ctx context.Context, obj client.Object) []reconcile.Request {
	if pc, ok := obj.(*schedulingv1.PriorityClass); ok && pc.GetName() == names.PriorityClass {
		return r.enqueueClusterCR()
	}
	return nil
}

func (r *KubeVirtConsoleUIReconciler) mapConsoleToRequests(ctx context.Context, obj client.Object) []reconcile.Request {
	if console, ok := obj.(*operatorv1.Console); ok && console.GetName() == "cluster" {
		return r.enqueueClusterCR()
	}
	return nil
}

func (r *KubeVirtConsoleUIReconciler) mapAPIServerToRequests(ctx context.Context, obj client.Object) []reconcile.Request {
	if api, ok := obj.(*configv1.APIServer); ok && api.GetName() == tlsprofile.APIServerName {
		return r.enqueueClusterCR()
	}
	return nil
}

func (r *KubeVirtConsoleUIReconciler) mapHyperConvergedToRequests(ctx context.Context, obj client.Object) []reconcile.Request {
	if _, ok := obj.(*hcoversion.HyperConverged); ok {
		return r.enqueueClusterCR()
	}
	return nil
}

func (r *KubeVirtConsoleUIReconciler) enqueueClusterCR() []reconcile.Request {
	return []reconcile.Request{{Name: names.ResourceName}}
}
