package controller

import (
	"context"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *TenantReconciler) ReconcileDeployment(ctx context.Context, tenant *tenantv1.Tenant) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Spec.Subdomain + "-deployment",
			Namespace: tenant.Spec.Subdomain,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// everything inside here is the DESIRED state.
		// runs before both create and update.
		deployment.Labels = tenantLabels(tenant)
		deployment.Spec.Replicas = tenant.Spec.Replicas
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": tenant.Spec.Subdomain},
		}
		deployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": tenant.Spec.Subdomain},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app",
						Image: tenant.Spec.AppImage,
					},
				},
			},
		}
		return nil
	})
	return err
}
