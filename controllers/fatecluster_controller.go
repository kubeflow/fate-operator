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
	"encoding/json"
	"fmt"
	"k8s.io/client-go/tools/record"
	"reflect"
	"time"

	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/db"
	"github.com/go-logr/logr"
	"gopkg.in/ffmt.v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1beta1 "github.com/kubeflow/fate-operator/api/v1beta1"
	"github.com/kubeflow/fate-operator/controllers/fatecluster"
)

// FateClusterReconciler reconciles a FateCluster object
type FateClusterReconciler struct {
	client.Client
	//	Log    logr.Logger
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	fateClusterFinalizer string = "finalizers.app.kubefate.net"
)

var FateOperatorTest bool

// +kubebuilder:rbac:groups=app.kubefate.net,resources=fateclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kubefate.net,resources=fateclusters/status,verbs=get;update;patch

func (r *FateClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("namespace", req.Namespace, "name", req.Name)
	log.Info("start Reconcile")
	defer log.Info("end Reconcile")
	var fateCluster appv1beta1.FateCluster
	if err := r.Get(ctx, req.NamespacedName, &fateCluster); err != nil {
		log.Error(err, "unable to fetch fateCluster", "namespace:", req.NamespacedName)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(10).Info(ffmt.Sprint("Get fateCluster", fateCluster))
	log.V(1).Info("Get fateCluster success", "name:", req.Name, "namespace:", req.Namespace)

	if fateCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(fateCluster.ObjectMeta.Finalizers, fateClusterFinalizer) {
			r.Log.Info(fmt.Sprintf("AddFinalizer for %v", req.NamespacedName))
			fateCluster.ObjectMeta.Finalizers = append(fateCluster.ObjectMeta.Finalizers, fateClusterFinalizer)
			if err := r.Update(ctx, &fateCluster); err != nil {
				r.Recorder.Event(&fateCluster, corev1.EventTypeWarning, "Adding finalizer", fmt.Sprintf("Failed to add finalizer: %s", err))
				return ctrl.Result{}, err
			}
			r.Recorder.Event(&fateCluster, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		}
	} else {
		if containsString(fateCluster.ObjectMeta.Finalizers, fateClusterFinalizer) {
			if err := r.deleteExternalResources(&fateCluster); err != nil {
				r.Recorder.Event(&fateCluster, corev1.EventTypeWarning, "deleting object", fmt.Sprintf("Failed to delete object: %s", err))
				return ctrl.Result{}, err
			}

			fateCluster.ObjectMeta.Finalizers = removeString(fateCluster.ObjectMeta.Finalizers, fateClusterFinalizer)
			if err := r.Update(ctx, &fateCluster); err != nil {
				r.Recorder.Event(&fateCluster, corev1.EventTypeWarning, "deleting finalizer", fmt.Sprintf("Failed to delete finalizer: %s", err))
				return ctrl.Result{}, err
			}
		}
		r.Recorder.Event(&fateCluster, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	var kubefate appv1beta1.Kubefate
	if err := r.Get(ctx, types.NamespacedName{Namespace: fateCluster.Spec.Kubefate.Namespace, Name: fateCluster.Spec.Kubefate.Name}, &kubefate); err != nil {
		r.Log.Error(err, "unable to fetch kubefate", "namespace:", fateCluster.Spec.Kubefate.Namespace, "name:", fateCluster.Spec.Kubefate.Name)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}

		if fateCluster.Status.Status == db.Running_c.String() {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	}
	r.Log.V(1).Info("get kubefate success")

	if kubefate.Status.Status != appv1beta1.Running {

		if fateCluster.Status.Status == "" {
			fateCluster.Status.Status = "Pending"
			if err := r.Update(ctx, &fateCluster); err != nil {
				r.Recorder.Event(&fateCluster, corev1.EventTypeWarning, "Status", fmt.Sprintf("Failed to update status of object: %s", err))
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	ok, err := r.Apply(&fateCluster, &kubefate)
	if err != nil {
		r.Recorder.Event(&fateCluster, corev1.EventTypeWarning, "Applied", fmt.Sprintf("Failed to apply object: %s", err))
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 2 * time.Second,
		}, err
	}

	if !ok {
		r.Recorder.Event(&fateCluster, corev1.EventTypeNormal, "Applied", "Object is applied")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 60 * time.Second,
	}, nil
}

func (r *FateClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.FateCluster{}).
		Complete(r)
}

func (r *FateClusterReconciler) Apply(fateclusterCR *appv1beta1.FateCluster, kubefateCR *appv1beta1.Kubefate) (bool, error) {
	ctx := context.Background()
	log := r.Log

	serviceurl := fmt.Sprintf("%s-kubefate-%s.%s:8080", PREFIX, kubefateCR.Name, kubefateCR.Namespace)
	if FateOperatorTest {
		serviceurl = kubefateCR.Spec.IngressDomain
	}
	username := r.kubefateConfigValue(kubefateCR, "FATECLOUD_USER_USERNAME")
	password := r.kubefateConfigValue(kubefateCR, "FATECLOUD_USER_PASSWORD")

	KubefateClient, err := fatecluster.NewKubefateClient("v1", serviceurl, username, password, &log)
	if err != nil {
		log.Error(err, "new kubefateClient")
		return false, err
	}

	defer func() {
		if err := r.Update(ctx, fateclusterCR); err != nil {
			log.Error(err, "update kubefate")
			return
		}
	}()

	// Render deployment request
	fateCluster, err := CreateFateCluster(fateclusterCR)
	if err != nil {
		log.Error(err, "Create fateCluster")
		return false, err
	}

	// TODO：
	// The Reconcile of updating resource is not working well right now
	// beacuse the failure of helm install will block the further process

	// Splitting the creation into two steps
	// Skip the first step for cluster update
	// 1. Create a job and wait until it return the cluster ID
	if fateclusterCR.Status.KubefateClusterId == "" {
		// fateCluster not created
		if fateclusterCR.Status.KubefateJobId == "" {
			log.Info("Creating FATE Cluster", "Cluster Name:", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)

			job, err := KubefateClient.CreateFateCluster(fateCluster)
			if err != nil {
				return false, err
			}
			fateclusterCR.Status.Status = db.Creating_c.String()
			fateclusterCR.Status.KubefateJobId = job.Uuid

			// Sync status immediately
			if syncErr := r.Update(context.Background(), fateclusterCR); syncErr != nil {
				log.Error(err, "Failed to sync", "Cluster Name", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)
				return false, syncErr
			}
		}

		// Fetch job status
		job, jobErr := KubefateClient.GetFateClusterJob(fateclusterCR.Status.KubefateJobId)
		if jobErr != nil {
			log.Error(err, "Failed to create FATE Cluster", "Cluster Name:", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)
			return false, err
		}

		// Get job success, FATE created
		if job != nil && job.ClusterId != "" {
			// Job was failed, will not go into the step 2 anyway.
			// Unless the configuration was updated, otherwise the status of CR won't be modified
			fateclusterCR.Status.KubefateClusterId = job.ClusterId
			if err := r.Update(ctx, fateclusterCR); err != nil {
				return false, err
			}
			return false, nil
		} else {
			// The creation process is running, update cluster ID for CR
			fateclusterCR.Status.Status = db.Unavailable_c.String()
			if err := r.Update(ctx, fateclusterCR); err != nil {
				return false, err
			}
			return false, nil
		}

	} else {
		// 2. Fetch info from KubeFATE to update cluster status, never get here if step one failed
		FateClusterGot, err := KubefateClient.GetFateCluster(fateclusterCR.Status.KubefateClusterId)
		if err != nil {
			log.Error(err, "Failed to fetch cluster id due to the previous failure")
			return false, err
		}

		if FateClusterGot != nil && FateClusterGot.Status.Status == db.Deleted_c {
			log.Info("FateCluster is deleted")
			fateclusterCR.Status.KubefateClusterId = ""
			fateclusterCR.Status.KubefateJobId = ""

			if syncErr := r.Update(context.Background(), fateclusterCR); syncErr != nil {
				log.Error(err, "Failed to sync", "Cluster Name", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)
				return true, syncErr
			}
			return false, r.cleanNamespace(fateclusterCR.Namespace)
		}

		clusterStatus := FateClusterGot.Status.Status
		// Update CR status only if the status is different from KubeFATE
		if fateclusterCR.Status.Status != clusterStatus.String() {
			// Don't care about the original status since the log contains the context
			// For cluster creation, log will have content as "Creating FATE Cluster" .... "FATE Cluster is Running"
			// For cluster update, log will have content as "Updating FATE Cluster" .... "FATE Cluster is Running"
			if clusterStatus == db.Running_c {
				log.Info("FATE Cluster is Running", "Cluster Name:", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)
			}

			fateclusterCR.Status.Status = clusterStatus.String()

			// Sync update
			if syncErr := r.Update(context.Background(), fateclusterCR); syncErr != nil {
				log.Error(err, "Failed to sync", "Cluster Name", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)
				return true, syncErr
			}
		}

		// The cluster status is not completed and cannot be updated
		if fateclusterCR.Status.Status != db.Running_c.String() {
			return false, nil
		}

		// Handle the update of the FATE cluster' configuration
		// It is safe for futher update if kubefate keep the record
		if !reflect.DeepEqual(fateCluster.Spec, FateClusterGot.Spec) {
			log.Info("Update info", "fateCluster", fateCluster.Spec, "FateClusterGot", FateClusterGot.Spec)

			log.Info("Updating FATE Cluster", "Cluster Name:", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)

			job, err := KubefateClient.UpdateFateCluster(fateCluster)
			if err != nil {
				log.Error(err, "Failed to sync status")
				return false, err
			}

			fateclusterCR.Status.KubefateJobId = job.Uuid
			fateclusterCR.Status.Status = db.Updating_c.String()

			// Sync update
			if syncErr := r.Update(context.Background(), fateclusterCR); syncErr != nil {
				log.Error(err, "Failed to sync", "Cluster Name", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)
				return false, syncErr
			}
		}
	}
	log.Info("apply fateCluster done")
	return true, nil
}

func CreateFateCluster(fateclusterCR *appv1beta1.FateCluster) (*fatecluster.FateCluster, error) {
	clusterSpec := fateclusterCR.Spec.ClusterSpec

	clusterSpec.Name = fateclusterCR.Name
	clusterSpec.Namespace = fateclusterCR.Namespace
	valBJ, err := json.Marshal(clusterSpec)
	if err != nil {
		return nil, err
	}

	fateSpec := &fatecluster.FateSpec{
		Name:         fateclusterCR.Name,
		Namespace:    fateclusterCR.Namespace,
		ChartName:    clusterSpec.ChartName,
		ChartVersion: clusterSpec.ChartVersion,
		Cover:        true,
		Data:         valBJ,
	}
	return &fatecluster.FateCluster{
		Spec:   fateSpec,
		Status: new(db.Cluster),
	}, nil
}
func (r *FateClusterReconciler) deleteExternalResources(fateclusterCR *appv1beta1.FateCluster) error {
	// TODO

	// There are roughly five condictions in deletion
	// 1. with kubefate:
	//    a. delete a healthy cluster
	//    b. delete a unhealthy cluster with db record
	//    c. delete a unhealthy cluster without db record
	// 2. without kubefate:
	//    a. delete a healthy cluster
	//    b. delete a unhealthy cluster
	//
	// The most tricky one is 1.c, in this condiction, no cluster id is provided
	// to do deletion. While the failure record in Helm's cache will block the
	// further operations. So it is impossible to implement from here.
	// A potential solution is to provide the cluster id to kubefate
	// to do deletion.
	//
	// For deletion without kubefate service, just clean all stuff under the giving namespace

	ctx := context.Background()
	var kubefate appv1beta1.Kubefate
	if err := r.Get(ctx, types.NamespacedName{Namespace: fateclusterCR.Spec.Kubefate.Namespace, Name: fateclusterCR.Spec.Kubefate.Name}, &kubefate); err != nil {
		r.Log.Error(err, "unable to fetch kubefate", "namespace:", fateclusterCR.Spec.Kubefate.Namespace, "name:", fateclusterCR.Spec.Kubefate.Name)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.

		return r.cleanNamespace(fateclusterCR.Namespace)
	}
	r.Log.V(1).Info("get kubefate success")

	log := r.Log
	log.Info("Deleting cluster", "Cluster Name", fateclusterCR.Name, "Cluster Namespace", fateclusterCR.Namespace)

	kubefateCR := &kubefate

	serviceurl := fmt.Sprintf("%s-kubefate-%s.%s:8080", PREFIX, kubefateCR.Name, kubefateCR.Namespace)
	if FateOperatorTest {
		serviceurl = kubefateCR.Spec.IngressDomain
	}

	username := r.kubefateConfigValue(kubefateCR, "FATECLOUD_USER_USERNAME")
	password := r.kubefateConfigValue(kubefateCR, "FATECLOUD_USER_PASSWORD")

	KubefateClient, err := fatecluster.NewKubefateClient("v1", serviceurl, username, password, &log)
	if err != nil {
		log.Error(err, "No available kubefate service was found")
		return r.cleanNamespace(fateclusterCR.Namespace)
	}

	if fateclusterCR.Status.KubefateClusterId == "" {
		log.Info("fateCluster not created")
		return r.cleanNamespace(fateclusterCR.Namespace)
	}

	fateCluster, err := CreateFateCluster(fateclusterCR)

	fateCluster.Status.Uuid = fateclusterCR.Status.KubefateClusterId

	cluster, err := KubefateClient.GetFateCluster(fateCluster.Status.Uuid)
	if err != nil {
		log.Error(err, "Get fateCluster")
		return r.cleanNamespace(fateclusterCR.Namespace)
	}

	if cluster.Status.Status == db.Deleted_c {
		log.V(1).Info("fatecluster does not exist", "fateClusterUUID", fateCluster.Status.Uuid)
		return r.cleanNamespace(fateclusterCR.Namespace)
	}

	job, err := KubefateClient.DeleteFateCluster(fateCluster)
	if err != nil {
		log.Error(err, "delete fateCluster")
		return r.cleanNamespace(fateclusterCR.Namespace)
	}
	fateclusterCR.Status.KubefateJobId = job.Uuid

	clusterId, err := KubefateClient.CheckFateClusterJob(fateclusterCR.Status.KubefateJobId)
	if err != nil {
		log.Error(err, "check fateCluster job")
		return r.cleanNamespace(fateclusterCR.Namespace)
	}
	fateclusterCR.Status.KubefateClusterId = clusterId

	fateclusterCR.Status.Status = appv1beta1.Deleted

	log.Info("fateCluster deleted.")
	return nil
}

func (r *FateClusterReconciler) kubefateConfigValue(kubefateCR *appv1beta1.Kubefate, key string) string {
	log := r.Log

	for _, env := range kubefateCR.Spec.Config {
		if env.Name != key {
			continue
		}
		if env.Value != "" {
			return env.Value
		}

		// Print the simple value
		if env.ValueFrom != nil {
			values, err := r.getSecretRefValue(kubefateCR.Namespace, env.ValueFrom)
			if err != nil {
				log.Error(err, "getSecretRefValue error", "env.Name", env.Name)
			}
			return values
		}

		return env.Value
	}
	return ""
}

func (r *FateClusterReconciler) getSecretRefValue(namespace string, from *corev1.EnvVarSource) (string, error) {
	log := r.Log
	secret := new(corev1.Secret)
	ctx := context.Background()
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: from.SecretKeyRef.Name}, secret); err != nil {
		log.Error(err, "Get secret error")
		return "", err
	}

	if data, ok := secret.Data[from.SecretKeyRef.Key]; ok {
		return string(data), nil
	}
	return "", fmt.Errorf("key %s not found in secret %s", from.SecretKeyRef.Key, from.SecretKeyRef.Name)
}

func (r *FateClusterReconciler) cleanNamespace(namespace string) error {
	log := r.Log
	ctx := context.Background()

	log.Info("Deleting deployment")
	err := r.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "Deleting deployment")
		return err
	}
	log.Info("Deleting ingress")
	err = r.DeleteAllOf(ctx, &extensionsv1beta1.Ingress{}, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "Deleting ingress")
		return err
	}
	log.Info("Deleting service")
	serviceList := &corev1.ServiceList{}
	err = r.Client.List(ctx, serviceList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		log.Error(err, "Get service")
		return err
	}
	for _, v := range serviceList.Items {
		err = r.Delete(ctx, &v)
		if err != nil {
			log.Error(err, "Deleting service", "serviceName", v.Name)
			return err
		}
		log.Info("Deleting service", "serviceName", v.Name)
	}
	log.Info("Deleting secret")
	err = r.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "Deleting secret")
		return err
	}
	log.Info("Deleting configMap")
	err = r.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "Deleting configMap")
		return err
	}
	return nil
}
