package controller

import (
	"context"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *TenantReconciler) ReconcileService(ctx context.Context, tenant *tenantv1.Tenant) error {
	// create service for tenant
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Spec.Subdomain + "-service",
			Namespace: tenant.Spec.Subdomain,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// everything inside here is the DESIRED state.
		// runs before both create and update.
		service.Spec.Selector = map[string]string{
			"app": tenant.Spec.Subdomain,
		}
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
		}
		return nil
	})
	return err
}
