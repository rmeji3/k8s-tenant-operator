package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
)

func (r *TenantReconciler) ReconcileNamespace(ctx context.Context, tenant *tenantv1.Tenant) error {
	// create namespace for tenant
	ns := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: tenant.Spec.Subdomain}, ns)

	if apierrors.IsNotFound(err) {
		// build and create the namespace
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   tenant.Spec.Subdomain,
				Labels: tenantLabels(tenant),
			},
		}
		if err := r.Create(ctx, ns); err != nil {
			return err
		}

	} else if err != nil {
		return err
	}
	return nil
}
