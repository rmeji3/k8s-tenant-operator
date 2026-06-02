package controller

import (
	"context"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *TenantReconciler) ReconcileIngress(ctx context.Context, tenant *tenantv1.Tenant) error {
	// add ingress for tenant
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Spec.Subdomain + "-ingress",
			Namespace: tenant.Spec.Subdomain,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		// everything inside here is the DESIRED state.
		// runs before both create and update.
		pathType := networkingv1.PathTypePrefix
		ingress.Spec = networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: tenant.Spec.Subdomain + ".rafamejia.me",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: tenant.Spec.Subdomain + "-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		return nil
	})

	return err
}
