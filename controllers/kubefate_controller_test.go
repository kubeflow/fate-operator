/*
 * Copyright 2019-2020 VMware, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * you may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	appv1beta1 "github.com/kubeflow/fate-operator/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Kubefate", func() {

	const timeout = time.Second * 120
	const interval = time.Second * 5

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Create API", func() {

		It("Should create successfully", func() {

			key := types.NamespacedName{
				Name:      "kubefate" + "-" + randomStringWithCharset(10, charset),
				Namespace: "default",
			}

			created := &appv1beta1.Kubefate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
			}

			update := &appv1beta1.Kubefate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: appv1beta1.KubefateSpec{
					ImageVersion:       "v1.0.3",
					IngressDomain:      "test-kubefate.net",
					ServiceAccountName: "kubefate-admin",
					Config: []v1.EnvVar{
						{
							Name:      "FATECLOUD_MONGO_USERNAME",
							Value:     "admin",
							ValueFrom: nil,
						}, {
							Name:      "FATECLOUD_MONGO_PASSWORD",
							Value:     "admin",
							ValueFrom: nil,
						}, {
							Name:      "FATECLOUD_MONGO_DATABASE",
							Value:     "kubefate",
							ValueFrom: nil,
						}, {
							Name:      "FATECLOUD_REPO_NAME",
							Value:     "",
							ValueFrom: nil,
						}, {
							Name:      "FATECLOUD_REPO_URL",
							Value:     "",
							ValueFrom: nil,
						}, {
							Name:      "FATECLOUD_USER_USERNAME",
							Value:     "root",
							ValueFrom: nil,
						}, {
							Name:      "FATECLOUD_USER_PASSWORD",
							Value:     "root",
							ValueFrom: nil,
						}, {
							Name:  "FATECLOUD_LOG_LEVEL",
							Value: "debug",
						},
					},
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting Create")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Pending {
					return nil
				}
				return fmt.Errorf("Deploy kubefate error,  status: wat=Pending  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("Deploy kubefate error, status: wat=Running  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			// Update
			By("Expecting Update")
			updated := &appv1beta1.Kubefate{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			updated.Spec = update.Spec
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("Deploy kubefate error, status: wat=Running  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
