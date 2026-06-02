package controller

import (
	"context"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *TenantReconciler) ReconcileDelete(ctx context.Context, tenant *tenantv1.Tenant) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenant.Spec.Subdomain,
		},
	}
	err := r.Delete(ctx, ns)
	if apierrors.IsNotFound(err) {
		// already deleted, ignore
		return nil
	} else if err != nil {
		return err
	}
	return nil
}
