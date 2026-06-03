package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// mapToTenant maps a changed child resource back to the Tenant that owns it.
// It reads the identifying labels stamped by tenantLabels and, if present,
// returns a reconcile request for that Tenant. This is how the controller
// heals drift in child resources without using owner references (which can't
// cross namespaces).
func (r *TenantReconciler) mapToTenant(ctx context.Context, obj client.Object) []reconcile.Request {
	labels := obj.GetLabels()
	name := labels[labelTenantName]
	namespace := labels[labelTenantNamespace]

	// Not one of our managed resources — ignore it.
	if name == "" {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}
