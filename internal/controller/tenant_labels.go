package controller

import (
	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
)

const (
	labelManagedBy       = "app.kubernetes.io/managed-by"
	labelTenantName      = "tenant.rafaelmejia.me/name"
	labelTenantNamespace = "tenant.rafaelmejia.me/namespace"
)

// tenantLabels returns the labels stamped on every resource the operator
// creates for a Tenant. The name/namespace labels let the controller map a
// changed child resource back to its owning Tenant (see the watches in
// SetupWithManager), since cross-namespace owner references aren't allowed.
func tenantLabels(tenant *tenantv1.Tenant) map[string]string {
	return map[string]string{
		labelManagedBy:       "tenant-operator",
		labelTenantName:      tenant.Name,
		labelTenantNamespace: tenant.Namespace,
	}
}
