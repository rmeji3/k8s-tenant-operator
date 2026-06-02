package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
)

func (r *TenantReconciler) setReadyCondition(ctx context.Context, tenant *tenantv1.Tenant, status metav1.ConditionStatus, reason, message string) error {
	meta.SetStatusCondition(&tenant.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  status,
		Reason:  reason,
		Message: message,
	})
	return r.Status().Update(ctx, tenant)
}
