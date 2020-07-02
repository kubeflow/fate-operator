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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1beta1 "github.com/kubeflow/fate-operator/api/v1beta1"
)

// FateJobReconciler reconciles a FateJob object
type FateJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	fateJobFinalizer string = "finalizers.app.kubefate.net"
)

// +kubebuilder:rbac:groups=app.kubefate.net,resources=fatejobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kubefate.net,resources=fatejobs/status,verbs=get;update;patch

func (r *FateJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	log := r.Log.WithValues("fateJob", req.NamespacedName)
	var fateJob appv1beta1.FateJob
	if err := r.Get(ctx, req.NamespacedName, &fateJob); err != nil {
		log.Error(err, "unable to fetch fateJob", "namespace:", req.NamespacedName)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(1).Info("get kubefate success")

	if fateJob.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(fateJob.ObjectMeta.Finalizers, fateJobFinalizer) {
			fateJob.ObjectMeta.Finalizers = append(fateJob.ObjectMeta.Finalizers, fateJobFinalizer)
			if err := r.Update(ctx, &fateJob); err != nil {
				log.Error(err, "Update fateJobCR  error")
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(fateJob.ObjectMeta.Finalizers, fateJobFinalizer) {
			if err := r.deleteExternalResources(&fateJob); err != nil {
				log.Error(err, "deleteExternalResources error")
				return ctrl.Result{}, err
			}

			fateJob.ObjectMeta.Finalizers = removeString(fateJob.ObjectMeta.Finalizers, fateJobFinalizer)
			if err := r.Update(ctx, &fateJob); err != nil {
				log.Error(err, "Update fateJobCR error")
				return ctrl.Result{}, err
			}
			log.V(1).Info("delete fateJob success", "fateJob name", fateJob.Name)
		}
		return ctrl.Result{}, nil
	}

	fatecluster := appv1beta1.FateCluster{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: fateJob.Namespace, Name: fateJob.Spec.FateClusterRef}, &fatecluster); err != nil {
		log.Error(err, "unable to fetch fatecluster", "namespace:", req.NamespacedName)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		if client.IgnoreNotFound(err) == nil {
			fateJob.Status.Status = "Pending"

			err := r.Update(ctx, &fateJob)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 10 * time.Second,
			}, nil
		}
		return ctrl.Result{}, err
	}

	if fatecluster.Status.Status != appv1beta1.Running {
		fateJob.Status.Status = "Pending"

		err := r.Update(ctx, &fateJob)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	ok, err := r.Apply(&fateJob)
	if err != nil {
		log.Error(err, "Apply fateCluster")
		return ctrl.Result{}, err
	}

	if !ok {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	}
	return ctrl.Result{}, nil
}

func (r *FateJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.FateJob{}).
		Complete(r)
}

func (r *FateJobReconciler) Apply(fateJobCR *appv1beta1.FateJob) (bool, error) {
	ctx := context.Background()
	log := r.Log

	fateJobGot := NewFateJob()
	err := r.Get(ctx, client.ObjectKey{
		Namespace: fateJobCR.Namespace,
		Name:      fateJobCR.Name,
	}, fateJobGot)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get fateJob")
			return false, err
		}

		// namespace IsNotFound
		fateJob := CreateFateJob(fateJobCR)
		log.Info("IsNotFound fateJob", "name", fateJobCR.Name)
		err := r.Create(ctx, fateJob)
		if err != nil {
			log.Error(err, "create fateJob")
			return false, err
		}
		//return nil
	}

	log.Info("fateJob got", "fatejob", fateJobGot.Status)

	if !reflect.DeepEqual(fateJobGot.Status, fateJobCR.Status.JobStatus) {
		fateJobCR.Status.JobStatus = fateJobGot.Status
		err := r.Update(ctx, fateJobCR)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	if IsRunning(fateJobGot) {
		fateJobCR.Status.Status = "Running"
		err := r.Update(ctx, fateJobCR)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	if fateJobGot.Status.Succeeded == 1 {
		fateJobCR.Status.Status = "Succeeded"
		err := r.Update(ctx, fateJobCR)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func IsRunning(job *batchv1.Job) bool {
	if job.Status.CompletionTime == nil {
		return true
	}
	return false
}

func (r *FateJobReconciler) deleteExternalResources(fateJobCR *appv1beta1.FateJob) error {
	ctx := context.Background()
	log := r.Log
	fateJobGot := NewFateJob()
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: fateJobCR.Namespace,
		Name:      fateJobCR.Name,
	}, fateJobGot); err == nil {
		err = r.Delete(ctx, fateJobGot)
		if err != nil {
			return err
		}
		log.Info("Kubefate mongoService deleted.")
	}
	return nil
}
func NewFateJob() *batchv1.Job {
	return &batchv1.Job{}
}
func CreateFateJob(fateJobCR *appv1beta1.FateJob) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fateJobCR.Name,
			Namespace: fateJobCR.Namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "fateJob", "deployer": "fate-operator", "name": fateJobCR.Name},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   fateJobCR.Name,
					Labels: map[string]string{"fate": "kubefate", "apps": "fateJob", "deployer": "fate-operator", "name": fateJobCR.Name},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: func() *int64 { var a *int64; var i int64 = 30; a = &i; return a }(),
					RestartPolicy:                 corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  fateJobCR.Name,
							Image: "federatedai/job:1.3.0-release",
							Command: []string{
								"/bin/bash",
								"-c",
								`   set -eux;
                                    source /data/projects/python/venv/bin/activate;
                                    python job.py '` + fateJobCR.Spec.JobConf.Pipeline + "' '" + fateJobCR.Spec.JobConf.ModulesConf + `'
                                `,
							},

							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "python-confs",
									MountPath: "/data/projects/fate/python/arch/conf/server_conf.json",
									SubPath:   "server_conf.json",
								},
							},
							Env: []corev1.EnvVar{
								{Name: "env", Value: ""},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "python-confs",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "python-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
