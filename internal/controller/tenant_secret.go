package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
)

func (r *TenantReconciler) ReconcileSecret(ctx context.Context, tenant *tenantv1.Tenant) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Spec.Subdomain + "-secret",
			Namespace: tenant.Spec.Subdomain,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		// everything inside here is the DESIRED state.
		// runs before both create and update.
		secret.Labels = tenantLabels(tenant)
		secret.Data = map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("password"),
		}
		return nil
	})

	return err
}
