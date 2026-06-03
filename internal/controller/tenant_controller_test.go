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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tenantv1 "github.com/rmeji3/k8s-tenant-operator/api/v1"
)

var _ = Describe("Tenant Controller", func() {
	Context("When reconciling a Tenant", func() {
		const (
			resourceName = "test-resource"
			subdomain    = "acme"
			appImage     = "nginx:latest"
		)
		var replicas int32 = 3

		ctx := context.Background()

		tenantKey := types.NamespacedName{Name: resourceName, Namespace: "default"}

		BeforeEach(func() {
			By("creating a Tenant with a valid spec")
			tenant := &tenantv1.Tenant{}
			err := k8sClient.Get(ctx, tenantKey, tenant)
			if err != nil {
				resource := &tenantv1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: tenantv1.TenantSpec{
						Subdomain: subdomain,
						AppImage:  appImage,
						Replicas:  &replicas,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("cleaning up the Tenant")
			tenant := &tenantv1.Tenant{}
			if err := k8sClient.Get(ctx, tenantKey, tenant); err == nil {
				// envtest has no garbage-collection controller, so strip the
				// finalizer before deleting to avoid a stuck object.
				tenant.SetFinalizers(nil)
				Expect(k8sClient.Update(ctx, tenant)).To(Succeed())
				Expect(k8sClient.Delete(ctx, tenant)).To(Succeed())
			}
		})

		It("provisions all child resources and marks the Tenant Ready", func() {
			reconciler := &TenantReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			By("reconciling the Tenant")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: tenantKey})
			Expect(err).NotTo(HaveOccurred())

			By("creating the tenant namespace")
			ns := &corev1.Namespace{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: subdomain}, ns)).To(Succeed())

			By("creating the Secret")
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: subdomain + "-secret", Namespace: subdomain}, secret)).To(Succeed())

			By("creating the Deployment with the right image and replica count")
			deployment := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: subdomain + "-deployment", Namespace: subdomain}, deployment)).To(Succeed())
			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(appImage))
			Expect(deployment.Spec.Replicas).NotTo(BeNil())
			Expect(*deployment.Spec.Replicas).To(Equal(replicas))

			By("creating the Service")
			service := &corev1.Service{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: subdomain + "-service", Namespace: subdomain}, service)).To(Succeed())

			By("creating the Ingress")
			ingress := &networkingv1.Ingress{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: subdomain + "-ingress", Namespace: subdomain}, ingress)).To(Succeed())

			By("setting a Ready condition on the Tenant status")
			tenant := &tenantv1.Tenant{}
			Expect(k8sClient.Get(ctx, tenantKey, tenant)).To(Succeed())
			cond := apimeta.FindStatusCondition(tenant.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		})
	})
})
