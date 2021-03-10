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

var _ = Describe("FateCluster", func() {

	const timeout = time.Second * 150
	const interval = time.Second * 1

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

		It("should create an object successfully", func() {

			kfkey := types.NamespacedName{
				Name:      "kubefate-for-fatecluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "kube-fate",
			}

			kfCreated := &appv1beta1.Kubefate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kfkey.Name,
					Namespace: kfkey.Namespace,
				},
				Spec: appv1beta1.KubefateSpec{
					Image:              "federatedai/kubefate:v1.3.0",
					IngressDomain:      "kubefate.net",
					ServiceAccountName: "kubefate-admin",
					Config: []v1.EnvVar{
						{Name: "MYSQL_ALLOW_EMPTY_PASSWORD", Value: "1"},
						{Name: "MYSQL_USER", Value: "kubefate"},
						{Name: "MYSQL_PASSWORD", Value: "123456"},
						{Name: "FATECLOUD_DB_USERNAME", Value: "kubefate"},
						{Name: "FATECLOUD_DB_PASSWORD", Value: "123456"},
						{Name: "FATECLOUD_REPO_NAME", Value: "KubeFate"},
						{Name: "FATECLOUD_REPO_URL", Value: "https://federatedai.github.io/KubeFATE/"},
						{Name: "FATECLOUD_USER_USERNAME", Value: "admin"},
						{Name: "FATECLOUD_USER_PASSWORD", Value: "admin"},
						{Name: "FATECLOUD_LOG_LEVEL", Value: "debug"},
						{Name: "FATECLOUD_LOG_NOCOLOR", Value: "false"},
					},
				},
			}

			// Create  kubefate
			By("Expecting Create kubefate")
			Expect(k8sClient.Create(context.Background(), kfCreated)).Should(Succeed())
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), kfkey, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("deploy kubefate error, status: wat=Running  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			key := types.NamespacedName{
				Name:      "fatecluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "default",
			}
			fateCreated := &appv1beta1.FateCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: appv1beta1.FateClusterSpec{
					Kubefate: appv1beta1.KubefateNamespacedName{
						Name:      kfCreated.Name,
						Namespace: kfCreated.Namespace,
					},
					ClusterSpec: &appv1beta1.ClusterSpec{
						ChartName:    "fate",
						ChartVersion: "v1.4.0-a",
						PartyID:      9999,
						Modules:      []string{"rollsite", "clustermanager", "nodemanager", "mysql", "python"},
						Rollsite: appv1beta1.Rollsite{
							Type:      "NodePort",
							NodePort:  30009,
							PartyList: nil,
						},
						Nodemanager: appv1beta1.Nodemanager{
							Count:                    1,
							SessionProcessorsPerNode: 2,
						},
					},
				},
			}

			fateUpdate := &appv1beta1.FateCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: appv1beta1.FateClusterSpec{
					Kubefate: appv1beta1.KubefateNamespacedName{
						Name:      kfCreated.Name,
						Namespace: kfCreated.Namespace,
					},
					ClusterSpec: &appv1beta1.ClusterSpec{
						ChartName:    "fate",
						ChartVersion: "v1.4.0-a",
						PartyID:      10000,
						Modules:      []string{"rollsite", "clustermanager", "nodemanager", "mysql", "python", "client"},
						Rollsite: appv1beta1.Rollsite{
							Type:      "NodePort",
							NodePort:  30010,
							PartyList: nil,
						},
						Nodemanager: appv1beta1.Nodemanager{
							Count:                    3,
							SessionProcessorsPerNode: 2,
						},
					},
				},
			}

			FateOperatorTest = true
			// Create FateCluster
			Expect(k8sClient.Create(context.Background(), fateCreated)).Should(Succeed())

			By("Expecting Create fateCluster")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Creating {
					return nil
				}
				return fmt.Errorf("creating fateCluster error, status: wat=Creating  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("creating FateCluste error, status: wat=Running  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			// Update fateCluster
			By("Expecting Update")
			updated := &appv1beta1.FateCluster{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			updated.Spec = fateUpdate.Spec
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Updating {
					return nil
				}
				return fmt.Errorf("updating FateCluste error, status: wat=Updating  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("updating FateCluste error, status: wat=Running  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

			// Delete kubefate
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), kfkey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				return k8sClient.Get(context.Background(), kfkey, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("If KubeFATE does not exist, it can be deleted successfully", func() {

			kfkey := types.NamespacedName{
				Name:      "kubefate-for-fatecluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "kube-fate",
			}

			kfCreated := &appv1beta1.Kubefate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kfkey.Name,
					Namespace: kfkey.Namespace,
				},
				Spec: appv1beta1.KubefateSpec{
					Image:              "federatedai/kubefate:v1.3.0",
					IngressDomain:      "notkf.kubefate.net",
					ServiceAccountName: "kubefate-admin",
					Config: []v1.EnvVar{
						{Name: "MYSQL_ALLOW_EMPTY_PASSWORD", Value: "1"},
						{Name: "MYSQL_USER", Value: "kubefate"},
						{Name: "MYSQL_PASSWORD", Value: "123456"},
						{Name: "FATECLOUD_DB_USERNAME", Value: "kubefate"},
						{Name: "FATECLOUD_DB_PASSWORD", Value: "123456"},
						{Name: "FATECLOUD_REPO_NAME", Value: "KubeFate"},
						{Name: "FATECLOUD_REPO_URL", Value: "https://federatedai.github.io/KubeFATE/"},
						{Name: "FATECLOUD_USER_USERNAME", Value: "admin"},
						{Name: "FATECLOUD_USER_PASSWORD", Value: "admin"},
						{Name: "FATECLOUD_LOG_LEVEL", Value: "debug"},
						{Name: "FATECLOUD_LOG_NOCOLOR", Value: "false"},
					},
				},
			}

			// Create  kubefate
			By("Expecting Create kubefate")
			Expect(k8sClient.Create(context.Background(), kfCreated)).Should(Succeed())
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), kfkey, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("deploy kubefate error, status: wat=Running  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			key := types.NamespacedName{
				Name:      "fatecluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "default",
			}
			fateCreated := &appv1beta1.FateCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: appv1beta1.FateClusterSpec{
					Kubefate: appv1beta1.KubefateNamespacedName{
						Name:      kfCreated.Name,
						Namespace: kfCreated.Namespace,
					},
					ClusterSpec: &appv1beta1.ClusterSpec{
						ChartName:    "fate",
						ChartVersion: "v1.4.0-a",
						PartyID:      9999,
						Modules:      []string{"rollsite", "clustermanager", "nodemanager", "mysql", "python"},
						Rollsite: appv1beta1.Rollsite{
							Type:      "NodePort",
							NodePort:  30009,
							PartyList: nil,
						},
						Nodemanager: appv1beta1.Nodemanager{
							Count:                    1,
							SessionProcessorsPerNode: 2,
						},
					},
				},
			}

			FateOperatorTest = true
			// Create FateCluster
			Expect(k8sClient.Create(context.Background(), fateCreated)).Should(Succeed())

			By("Expecting Create fateCluster")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Creating {
					return nil
				}
				return fmt.Errorf("creating FateCluster error, status: wat=Creating  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("creating FateCluste error, status: wat=Running  got=%s", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			// Delete kubefate
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), kfkey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				return k8sClient.Get(context.Background(), kfkey, f)
			}, timeout, interval).ShouldNot(Succeed())

			// Delete FateCluster
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

		})

	})
})
