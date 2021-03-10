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

	corev1 "k8s.io/api/core/v1"
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

			kubefateKey := types.NamespacedName{
				Name:      "kubefate-for-fatejob" + "-" + randomStringWithCharset(10, charset),
				Namespace: "kube-fate",
			}

			kfCreated := &appv1beta1.Kubefate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kubefateKey.Name,
					Namespace: kubefateKey.Namespace,
				},
				Spec: appv1beta1.KubefateSpec{
					Image:              "federatedai/kubefate:v1.3.0",
					IngressDomain:      "test-fatejob.kubefate.net",
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
				_ = k8sClient.Get(context.Background(), kubefateKey, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("Deploy kubefate error, status: wat=Running  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			fateClusterKey := types.NamespacedName{
				Name:      "fatecluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "fate-9999",
			}
			fateCreated := &appv1beta1.FateCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fateClusterKey.Name,
					Namespace: fateClusterKey.Namespace,
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
				_ = k8sClient.Get(context.Background(), fateClusterKey, f)
				if f.Status.Status == appv1beta1.Creating {
					return nil
				}
				return fmt.Errorf("Creating FateCluster error, status: wat=Creating  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), fateClusterKey, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("Creating FateCluste error, status: wat=Running  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			// Create fateJob
			fateJobKey := types.NamespacedName{
				Name:      "fatejob" + "-" + randomStringWithCharset(10, charset),
				Namespace: fateClusterKey.Namespace,
			}

			fateJob := &appv1beta1.FateJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fateJobKey.Name,
					Namespace: fateJobKey.Namespace,
				},
				Spec: appv1beta1.FateJobSpec{
					FateClusterRef: corev1.ObjectReference{
						Kind:      fateCreated.Kind,
						Namespace: fateCreated.Namespace,
						Name:      fateCreated.Name,
					},
					JobConf: appv1beta1.JobConf{
						Pipeline: `{
        "components": {
          "secure_add_example_0": {
            "module": "SecureAddExample"
          }
        }
      }`,
						ModulesConf: `{
        "initiator": {
          "role": "guest",
          "party_id": 9999
        },
        "job_parameters": {
          "work_mode": 1
        },
        "role": {
          "guest": [
            9999
          ],
          "host": [
            9999
          ]
        },
        "role_parameters": {
          "guest": {
            "secure_add_example_0": {
              "seed": [
                123
              ]
            }
          },
          "host": {
            "secure_add_example_0": {
              "seed": [
                321
              ]
            }
          }
        },
        "algorithm_parameters": {
          "secure_add_example_0": {
            "partition": 10,
            "data_num": 1000
          }
        }
      }`,
					},
				},
			}

			// Create fateJob
			Expect(k8sClient.Create(context.Background(), fateJob)).Should(Succeed())

			By("Expecting Create fateJob")
			Eventually(func() error {
				f := &appv1beta1.FateJob{}
				_ = k8sClient.Get(context.Background(), fateJobKey, f)
				if f.Status.Status == appv1beta1.Running {
					return nil
				}
				return fmt.Errorf("Creating fateJob error, status: wat=Running  got=%s\n", f.Status.Status)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &appv1beta1.FateJob{}
				_ = k8sClient.Get(context.Background(), fateJobKey, f)
				if f.Status.Status == "Succeeded" {
					return nil
				}
				return fmt.Errorf("Creating fateJob error, status: wat=Succeeded  got=%s\n", f.Status.Status)
			}, time.Second*300, interval).Should(Succeed())

			// Delete fateJob
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.FateJob{}
				_ = k8sClient.Get(context.Background(), fateJobKey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.FateJob{}
				return k8sClient.Get(context.Background(), fateJobKey, f)
			}, timeout, interval).ShouldNot(Succeed())

			// Delete fateCluster
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				_ = k8sClient.Get(context.Background(), fateClusterKey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.FateCluster{}
				return k8sClient.Get(context.Background(), fateClusterKey, f)
			}, timeout, interval).ShouldNot(Succeed())

			// Delete kubefate
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				_ = k8sClient.Get(context.Background(), kubefateKey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &appv1beta1.Kubefate{}
				return k8sClient.Get(context.Background(), kubefateKey, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

	})
})
