/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenant.rafaelmejia.me,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.rafaelmejia.me,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tenant.rafaelmejia.me,resources=tenants/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// TODO(user): your logic here
	tenant := &tenantv1.Tenant{}
	if err := r.Get(ctx, req.NamespacedName, tenant); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizerName := "tenant.rafaelmejia.me/finalizer"

	if tenant.DeletionTimestamp.IsZero() {
		// not being deleted: add finalizer if missing
		if !controllerutil.ContainsFinalizer(tenant, finalizerName) {
			controllerutil.AddFinalizer(tenant, finalizerName)
			if err := r.Update(ctx, tenant); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// being deleted: run cleanup, remove finalizer
		if err := r.ReconcileDelete(ctx, tenant); err != nil {
			log.Error(err, "Failed to reconcile delete", "tenant", tenant.Name)
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(tenant, finalizerName)
		if err := r.Update(ctx, tenant); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling tenant", "tenant", tenant.Name)

	// If a namespace with this name is still terminating from a previous delete,
	// wait for it to fully disappear before recreating anything. Creating child
	// resources in a Terminating namespace fails or gets garbage-collected.
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: tenant.Spec.Subdomain}, ns); err == nil {
		if !ns.DeletionTimestamp.IsZero() {
			log.Info("Namespace is still terminating, requeueing", "namespace", tenant.Spec.Subdomain)
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
	} else if !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// create namespace for tenant
	if err := r.ReconcileNamespace(ctx, tenant); err != nil {
		log.Error(err, "Failed to reconcile namespace", "tenant", tenant.Name)
		return ctrl.Result{}, err
	}

	// get secret for tenant
	if err := r.ReconcileSecret(ctx, tenant); err != nil {
		log.Error(err, "Failed to reconcile secret", "tenant", tenant.Name)
		return ctrl.Result{}, err
	}

	// create deployment for tenant
	if err := r.ReconcileDeployment(ctx, tenant); err != nil {
		log.Error(err, "Failed to reconcile deployment", "tenant", tenant.Name)
		return ctrl.Result{}, err
	}

	// create service for tenant
	if err := r.ReconcileService(ctx, tenant); err != nil {
		log.Error(err, "Failed to reconcile service", "tenant", tenant.Name)
		return ctrl.Result{}, err
	}

	// add ingress for tenant
	if err := r.ReconcileIngress(ctx, tenant); err != nil {
		log.Error(err, "Failed to reconcile ingress", "tenant", tenant.Name)
		return ctrl.Result{}, err
	}

	if err := r.setReadyCondition(ctx, tenant, metav1.ConditionTrue, "Provisioned", "All tenant resources are ready"); err != nil {
		log.Error(err, "Failed to set ready condition", "tenant", tenant.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1.Tenant{}).
		// Watch the resources we provision so manual drift (e.g. someone
		// deletes a Deployment) triggers a reconcile of the owning Tenant.
		// We use Watches + a label-based mapping instead of Owns() because
		// our children live in a different namespace than the Tenant, so
		// owner references aren't allowed.
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(r.mapToTenant)).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.mapToTenant)).
		Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(r.mapToTenant)).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(r.mapToTenant)).
		Watches(&networkingv1.Ingress{}, handler.EnqueueRequestsFromMapFunc(r.mapToTenant)).
		Named("tenant").
		Complete(r)
}
