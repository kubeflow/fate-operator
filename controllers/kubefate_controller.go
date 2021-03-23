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
	"reflect"
	"time"

	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1beta1 "github.com/kubeflow/fate-operator/api/v1beta1"
)

// KubefateReconciler reconciles a Kubefate object
type KubefateReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	kubefateFinalizer string = "finalizers.app.kubefate.net"
)

// +kubebuilder:rbac:groups=app.kubefate.net,resources=kubefates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kubefate.net,resources=kubefates/status,verbs=get;update;patch

func (r *KubefateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := r.Log.WithValues("namespace", req.Namespace, "name", req.Name)

	log.Info("Starting reconcile loop")
	defer log.Info("Finish reconcile loop")

	var kubefate appv1beta1.Kubefate
	if err := r.Get(ctx, req.NamespacedName, &kubefate); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch kubefate")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if kubefate.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(kubefate.ObjectMeta.Finalizers, kubefateFinalizer) {
			r.Log.Info(fmt.Sprintf("AddFinalizer for %v", req.NamespacedName))
			kubefate.ObjectMeta.Finalizers = append(kubefate.ObjectMeta.Finalizers, kubefateFinalizer)
			if err := r.Update(context.Background(), &kubefate); err != nil {
				r.Recorder.Event(&kubefate, corev1.EventTypeWarning, "Adding finalizer", fmt.Sprintf("Failed to add finalizer: %s", err))
				return ctrl.Result{}, err
			}
			r.Recorder.Event(&kubefate, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		}
	} else {
		if containsString(kubefate.ObjectMeta.Finalizers, kubefateFinalizer) {
			if err := r.deleteExternalResources(&kubefate); err != nil {
				r.Recorder.Event(&kubefate, corev1.EventTypeWarning, "deleting object", fmt.Sprintf("Failed to delete object: %s", err))
				return ctrl.Result{}, err
			}

			kubefate.ObjectMeta.Finalizers = removeString(kubefate.ObjectMeta.Finalizers, kubefateFinalizer)
			if err := r.Update(context.Background(), &kubefate); err != nil {
				r.Recorder.Event(&kubefate, corev1.EventTypeWarning, "deleting finalizer", fmt.Sprintf("Failed to delete finalizer: %s", err))
				return ctrl.Result{}, err
			}
		}
		r.Recorder.Event(&kubefate, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	ok, err := r.kfApply(&kubefate)

	// Make the current Kubefate as default if kfApply is successed.
	if err != nil {
		r.Recorder.Event(&kubefate, corev1.EventTypeWarning, "Applied", fmt.Sprintf("Failed to apply object: %s", err))
		return ctrl.Result{}, err
	}

	if err := r.Update(ctx, &kubefate); err != nil {
		log.Error(err, "update kubefate")
		return ctrl.Result{}, err
	}

	if !ok {
		r.Recorder.Event(&kubefate, corev1.EventTypeNormal, "Applied", "Object is applied")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, nil
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 60 * time.Second,
	}, nil
}

func (r *KubefateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.Kubefate{}).
		Complete(r)
}
func (r *KubefateReconciler) deleteExternalResources(kubefate *appv1beta1.Kubefate) error {
	//
	log := r.Log
	KubefateResources := NewKubefate(kubefate)
	if err := r.Get(context.Background(), client.ObjectKey{
		Namespace: KubefateResources.ingress.Namespace,
		Name:      KubefateResources.ingress.Name,
	}, KubefateResources.ingress); err == nil {
		err := r.Delete(context.Background(), KubefateResources.ingress)
		if err != nil {
			return err
		}
		log.Info("Kubefate ingress deleted.")
	}

	if err := r.Get(context.Background(), client.ObjectKey{
		Namespace: KubefateResources.mariadbService.Namespace,
		Name:      KubefateResources.mariadbService.Name,
	}, KubefateResources.mariadbService); err == nil {
		err = r.Delete(context.Background(), KubefateResources.mariadbService)
		if err != nil {
			return err
		}
		log.Info("Kubefate mariadbService deleted.")
	}

	if err := r.Get(context.Background(), client.ObjectKey{
		Namespace: KubefateResources.kubefateService.Namespace,
		Name:      KubefateResources.kubefateService.Name,
	}, KubefateResources.kubefateService); err == nil {
		err = r.Delete(context.Background(), KubefateResources.kubefateService)
		if err != nil {
			return err
		}
		log.Info("Kubefate kubefateService deleted.")
	}

	if err := r.Get(context.Background(), client.ObjectKey{
		Namespace: KubefateResources.mariadbDeploy.Namespace,
		Name:      KubefateResources.mariadbDeploy.Name,
	}, KubefateResources.mariadbDeploy); err == nil {
		err = r.Delete(context.Background(), KubefateResources.mariadbDeploy)
		if err != nil {
			return err
		}
		log.Info("Kubefate mariadbDeploy deleted.")
	}

	if err := r.Get(context.Background(), client.ObjectKey{
		Namespace: KubefateResources.kubefateDeploy.Namespace,
		Name:      KubefateResources.kubefateDeploy.Name,
	}, KubefateResources.kubefateDeploy); err == nil {
		err = r.Delete(context.Background(), KubefateResources.kubefateDeploy)
		if err != nil {
			return err
		}
		log.Info("Kubefate kubefateDeploy deleted.")
	}

	log.Info("Kubefate deleted.")
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

type Kubefate struct {
	mariadbService  *corev1.Service
	kubefateService *corev1.Service
	mariadbDeploy   *appsv1.Deployment
	kubefateDeploy  *appsv1.Deployment
	ingress         *networkingv1beta1.Ingress
}

const (
	NamespaceDefault = "kube-fate"
	PREFIX           = "kubefate"
)

func NewKubefate(kubefate *appv1beta1.Kubefate) *Kubefate {

	namespace := kubefate.Namespace
	if namespace == "" {
		namespace = NamespaceDefault
	}

	name := kubefate.Name
	if name == "" {
		name = randomStringWithCharset(10, charset)
	}

	var image = kubefate.Spec.Image
	if kubefate.Spec.Image == "" {
		image = "federatedai/kubefate:v1.4.0"
	}

	for _, v := range []string{"FATECLOUD_MONGO_USERNAME", "FATECLOUD_MONGO_PASSWORD", "FATECLOUD_USER_USERNAME", "FATECLOUD_USER_PASSWORD"} {
		if !IsExitEnv(kubefate.Spec.Config, v) {
			kubefate.Spec.Config = append(kubefate.Spec.Config, corev1.EnvVar{
				Name:  v,
				Value: "admin",
			})
		}
	}

	if !IsExitEnv(kubefate.Spec.Config, "FATECLOUD_MONGO_DATABASE") {
		kubefate.Spec.Config = append(kubefate.Spec.Config, corev1.EnvVar{
			Name:  "FATECLOUD_MONGO_DATABASE",
			Value: "KubeFate",
		})
	}

	var kubefateServiceDeploy = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubefate-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "kubefateService", "deployer": "fate-operator", "name": name},
		},
		Spec: appsv1.DeploymentSpec{

			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"fate": "kubefate", "apps": "kubefateService", "name": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"fate": "kubefate", "apps": "kubefateService", "deployer": "fate-operator", "name": name},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: kubefate.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:  "kubefate",
							Image: fmt.Sprintf("%s", image),
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							Env: append([]corev1.EnvVar{
								{Name: "FATECLOUD_DB_TYPE", Value: "mysql"},
								{Name: "FATECLOUD_DB_HOST", Value: fmt.Sprintf("%s-mariadb-%s", PREFIX, name)},
								{Name: "FATECLOUD_DB_PORT", Value: "3306"},
								{Name: "FATECLOUD_DB_NAME", Value: "kube_fate"},
								{Name: "FATECLOUD_SERVER_ADDRESS", Value: "0.0.0.0"},
								{Name: "FATECLOUD_SERVER_PORT", Value: "8080"},
							}, kubefate.Spec.Config...),
						},
					},
				},
			},
		},
	}

	MariadbDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "mariadb", "deployer": "fate-operator", "name": name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { var a *int32; var i int32 = 1; a = &i; return a }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"fate": "kubefate", "apps": "mariadb", "name": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"fate": "kubefate", "apps": "mariadb", "deployer": "fate-operator", "name": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mariadb",
							Image: "mariadb:10",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 3306},
							},
							Env: func() []corev1.EnvVar {
								env := make([]corev1.EnvVar, 0)
								for _, v := range kubefate.Spec.Config {
									if v.Name == "MYSQL_USER" {
										env = append(env, corev1.EnvVar{Name: "MYSQL_USER", Value: v.Value, ValueFrom: v.ValueFrom})
									}
									if v.Name == "MYSQL_PASSWORD" {
										env = append(env, corev1.EnvVar{Name: "MYSQL_PASSWORD", Value: v.Value, ValueFrom: v.ValueFrom})
									}
								}
								env = append(env, corev1.EnvVar{Name: "MYSQL_ALLOW_EMPTY_PASSWORD", Value: "1"})
								env = append(env, corev1.EnvVar{Name: "MYSQL_DATABASE", Value: "kube_fate"})
								return env
							}(),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mariadb-data",
									MountPath: "/var/lib/mysql",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name:         "mariadb-data",
							VolumeSource: kubefate.Spec.VolumeSource,
						},
					},
				},
			},
		},
	}

	var kubefateService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubefate-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "kubefateService", "deployer": "fate-operator", "name": name},
		},
		Spec: corev1.ServiceSpec{
			SessionAffinity: corev1.ServiceAffinityNone,
			Ports: []corev1.ServicePort{
				{
					Name:     "8080",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
						StrVal: "",
					},
				},
			},
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"fate": "kubefate", "apps": "kubefateService", "name": name},
		},
	}

	var MariadbService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "mariadb", "deployer": "fate-operator", "name": name},
		},
		Spec: corev1.ServiceSpec{
			SessionAffinity: corev1.ServiceAffinityNone,
			Ports: []corev1.ServicePort{
				{
					Name:     "3306",
					Port:     3306,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3306,
						StrVal: "",
					},
				},
			},
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"fate": "kubefate", "apps": "mariadb", "name": name},
		},
	}

	var kubefateIngress = &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubefate-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "kubefate", "deployer": "fate-operator", "name": name},
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: kubefate.Spec.IngressDomain,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: fmt.Sprintf("%s-kubefate-%s", PREFIX, name),
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 8080,
											StrVal: "",
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

	return &Kubefate{
		kubefateService: kubefateService,
		mariadbService:  MariadbService,
		kubefateDeploy:  kubefateServiceDeploy,
		mariadbDeploy:   MariadbDeploy,
		ingress:         kubefateIngress,
	}
}

func IsExitEnv(env []corev1.EnvVar, s string) bool {
	for _, v := range env {
		if v.Name == s {
			if v.Value != "" || v.ValueFrom != nil {
				return true
			}
		}
	}
	return false
}

func (r *KubefateReconciler) kfApply(kubefate *appv1beta1.Kubefate) (bool, error) {

	ctx := context.Background()
	log := r.Log

	kf := NewKubefate(kubefate)

	kfgot := new(Kubefate)

	// kubefate kubefateDeploy
	kfgot.kubefateDeploy = &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.kubefateDeploy.Namespace,
		Name:      kf.kubefateDeploy.Name,
	}, kfgot.kubefateDeploy); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get kubefateDeploy")
			return false, err
		}

		// kubefateDeploy IsNotFound
		log.Info("IsNotFound kubefateDeploy", "Name", kf.kubefateDeploy.Name)
		err = r.Create(ctx, kf.kubefateDeploy)
		if err != nil {
			log.Error(err, "create kubefateDeploy")
			return false, err
		}
		r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesDeployed", "kubefate deployment is Deployed")
	} else {

		if !reflect.DeepEqual(kf.kubefateDeploy.Spec.Template.Spec.Containers[0].Image, kfgot.kubefateDeploy.Spec.Template.Spec.Containers[0].Image) ||
			!reflect.DeepEqual(kf.kubefateDeploy.Spec.Template.Spec.Containers[0].Env, kfgot.kubefateDeploy.Spec.Template.Spec.Containers[0].Env) ||
			!reflect.DeepEqual(kf.kubefateDeploy.Spec.Template.Spec.ServiceAccountName, kfgot.kubefateDeploy.Spec.Template.Spec.ServiceAccountName) {
			//ffmt.Print("kfgot kubefateDeploy ", kfgot.kubefateDeploy.Spec)
			//ffmt.Print("kf kubefateDeploy ", kf.kubefateDeploy.Spec)
			log.Info("Updating kubefateDeploy", "name", kf.kubefateDeploy.Name)
			err := r.Update(ctx, kf.kubefateDeploy)
			if err != nil {
				log.Error(err, "update kubefateDeploy")
				return false, err
			}
			r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesUpdated", "kubefate deployment is updated")
		}
	}

	kubefate.Status.KubefateDeploy = kf.kubefateDeploy.Name

	// kubefate mariadbDeploy
	kfgot.mariadbDeploy = &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.mariadbDeploy.Namespace,
		Name:      kf.mariadbDeploy.Name,
	}, kfgot.mariadbDeploy); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get mariadbDeploy")
			return false, err
		}

		// mariadbDeploy IsNotFound
		log.Info("IsNotFound mariadbDeploy", "Name", kf.mariadbDeploy.Name)
		err = r.Create(ctx, kf.mariadbDeploy)
		if err != nil {
			log.Error(err, "create mariadbDeploy")
			return false, err
		}
		r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesDeployed", "mariadb deployment is deployed")
	} else {

		if !reflect.DeepEqual(kf.mariadbDeploy.Spec.Template.Spec.Containers[0].Env, kfgot.mariadbDeploy.Spec.Template.Spec.Containers[0].Env) {
			//ffmt.Print("kfgot mariadbDeploy ", kfgot.mariadbDeploy.Spec)
			//ffmt.Print("kf mariadbDeploy ", kf.mariadbDeploy.Spec)
			log.Info("Updating mariadbDeploy", "name", kf.mariadbDeploy.Name)
			err := r.Update(ctx, kf.mariadbDeploy)
			if err != nil {
				log.Error(err, "update mariadbDeploy")
				return false, err
			}
			r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesUpdated", "mariadb deployment is updated")
		}
	}

	kubefate.Status.MariadbDeploy = kf.mariadbDeploy.Name
	//ffmt.Println("kfgot kubefateDeploy", kfgot.mariadbDeploy)

	//kubefate kubefateService
	kfgot.kubefateService = &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.kubefateService.Namespace,
		Name:      kf.kubefateService.Name,
	}, kfgot.kubefateService); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get kubefateService")
			return false, err
		}

		// kubefateService IsNotFound
		log.Info("IsNotFound kubefateService", "Name", kf.kubefateService.Name)
		err = r.Create(ctx, kf.kubefateService)
		if err != nil {
			log.Error(err, "create kubefateService")
			return false, err
		}
		r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesDeployed", "kubefate service is deployed")
	} else {
		kf.kubefateService.Spec.ClusterIP = kfgot.kubefateService.Spec.ClusterIP
		if !reflect.DeepEqual(kf.kubefateService.Spec, kfgot.kubefateService.Spec) {
			//ffmt.Print("kfgot kubefateService ", kfgot.kubefateService.Spec)
			//ffmt.Print("kf kubefateService ", kf.kubefateService.Spec)
			log.Info("Updating kubefateService", "name", kf.kubefateService.Name)
			err := r.Update(ctx, kf.kubefateService)
			if err != nil {
				log.Error(err, "update kubefateService")
				return false, err
			}
			r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesUpdated", "kubefate service is updated")
		}
	}
	kubefate.Status.KubefateService = kf.kubefateService.Name
	//ffmt.Println("kfgot kubefateService", kfgot.kubefateService)

	// kubefate mariadbService
	kfgot.mariadbService = &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.mariadbService.Namespace,
		Name:      kf.mariadbService.Name,
	}, kfgot.mariadbService); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get mariadbService")
			return false, err
		}

		// mariadbService IsNotFound
		log.Info("IsNotFound mariadbService", "Name", kf.mariadbService.Name)
		err = r.Create(ctx, kf.mariadbService)
		if err != nil {
			log.Error(err, "create mariadbService")
			return false, err
		}
		r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesDeployed", "mariadb service is deployed")
	} else {
		kf.mariadbService.Spec.ClusterIP = kfgot.mariadbService.Spec.ClusterIP
		if !reflect.DeepEqual(kf.mariadbService.Spec, kfgot.mariadbService.Spec) {
			//ffmt.Print("kfgot mariadbService ", kfgot.mariadbService.Spec)
			//ffmt.Print("kf mariadbService ", kf.mariadbService.Spec)
			log.Info("Updating mariadbService", "name", kf.mariadbService.Name)
			err := r.Update(ctx, kf.mariadbService)
			if err != nil {
				log.Error(err, "update mariadbService")
				return false, err
			}
			r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesUpdated", "kubefate service is updated")
		}
	}
	kubefate.Status.MariadbService = kf.mariadbService.Name
	//ffmt.Println("kfgot kubefateService", kfgot.mariadbService)

	// kubefate ingress
	kfgot.ingress = &networkingv1beta1.Ingress{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.ingress.Namespace,
		Name:      kf.ingress.Name,
	}, kfgot.ingress); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get ingress")
			return false, err
		}

		// ingress IsNotFound
		log.Info("IsNotFound ingress", "Name", kf.ingress.Name)
		err = r.Create(ctx, kf.ingress)
		if err != nil {
			log.Error(err, "create ingress")
			return false, err
		}
		r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesDeployed", "kubefate ingress is deployed")
	} else {

		if !reflect.DeepEqual(kf.ingress.Spec, kfgot.ingress.Spec) {
			//ffmt.Print("kfgot ingress ", kfgot.ingress.Spec)
			//ffmt.Print("kf ingress ", kf.ingress.Spec)
			log.Info("Updating ingress", "name", kf.ingress.Name)
			err := r.Update(ctx, kf.ingress)
			if err != nil {
				log.Error(err, "update ingress")
				return false, err
			}
			r.Recorder.Event(kubefate, corev1.EventTypeNormal, "ResourcesUpdated", "kubefate service is updated")
		}
	}
	kubefate.Status.Ingress = kf.ingress.Name

	if kfgot.kubefateDeploy.Status.UnavailableReplicas != 0 {
		kubefate.Status.Status = appv1beta1.Pending
		log.Info("Kubefate has not been deployed successfully")
		return false, nil
	}

	if kfgot.mariadbDeploy.Status.UnavailableReplicas != 0 {
		kubefate.Status.Status = appv1beta1.Pending
		log.Info("MariadbDB has not been deployed successfully")
		return false, nil

	}

	if kubefate.Status.Status == "" {
		kubefate.Status.Status = appv1beta1.Pending
		return false, nil
	}

	if kubefate.Status.Status == appv1beta1.Pending {
		kubefate.Status.Status = appv1beta1.Running
	}

	log.Info("kubefate apply success", "kubefate name", kubefate.Name)
	return true, nil
}
