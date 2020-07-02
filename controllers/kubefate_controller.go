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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
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
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	kubefateFinalizer string = "finalizers.app.kubefate.net"
)

// +kubebuilder:rbac:groups=app.kubefate.net,resources=kubefates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kubefate.net,resources=kubefates/status,verbs=get;update;patch

func (r *KubefateReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("namespace", req.Namespace, "name", req.Name)

	log.Info("Starting reconcile loop")
	defer log.Info("Finish reconcile loop")

	var kubefate appv1beta1.Kubefate
	if err := r.Get(ctx, req.NamespacedName, &kubefate); err != nil {
		log.Error(err, "unable to fetch kubefate")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if kubefate.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(kubefate.ObjectMeta.Finalizers, kubefateFinalizer) {
			kubefate.ObjectMeta.Finalizers = append(kubefate.ObjectMeta.Finalizers, kubefateFinalizer)
			if err := r.Update(context.Background(), &kubefate); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(kubefate.ObjectMeta.Finalizers, kubefateFinalizer) {
			if err := r.deleteExternalResources(&kubefate); err != nil {
				return ctrl.Result{}, err
			}

			kubefate.ObjectMeta.Finalizers = removeString(kubefate.ObjectMeta.Finalizers, kubefateFinalizer)
			if err := r.Update(context.Background(), &kubefate); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	ok, err := r.kfApply(&kubefate)

	// Make the current Kubefate as default if kfApply is successed.
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Update(ctx, &kubefate); err != nil {
		log.Error(err, "update kubefate")
		return ctrl.Result{}, err
	}

	if !ok {
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
		Namespace: KubefateResources.mongoService.Namespace,
		Name:      KubefateResources.mongoService.Name,
	}, KubefateResources.mongoService); err == nil {
		err = r.Delete(context.Background(), KubefateResources.mongoService)
		if err != nil {
			return err
		}
		log.Info("Kubefate mongoService deleted.")
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
		Namespace: KubefateResources.mongoDeploy.Namespace,
		Name:      KubefateResources.mongoDeploy.Name,
	}, KubefateResources.mongoDeploy); err == nil {
		err = r.Delete(context.Background(), KubefateResources.mongoDeploy)
		if err != nil {
			return err
		}
		log.Info("Kubefate mongoDeploy deleted.")
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
	mongoService    *corev1.Service
	kubefateService *corev1.Service
	mongoDeploy     *appsv1.Deployment
	kubefateDeploy  *appsv1.Deployment
	ingress         *extensionsv1beta1.Ingress
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

	var imageVersion = kubefate.Spec.ImageVersion
	if kubefate.Spec.ImageVersion == "" {
		imageVersion = "latest"
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
							Image: fmt.Sprintf("federatedai/kubefate:%s", imageVersion),
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							Env: append([]corev1.EnvVar{
								{Name: "FATECLOUD_MONGO_URL", Value: fmt.Sprintf("%s-mongo-%s:27017", PREFIX, name)},
								{Name: "FATECLOUD_SERVER_ADDRESS", Value: "0.0.0.0"},
								{Name: "FATECLOUD_SERVER_PORT", Value: "8080"},
								{Name: "FATECLOUD_LOG_LEVEL", Value: "info"},
							}, kubefate.Spec.Config...),
						},
					},
				},
			},
		},
	}

	var MongoDeploy = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mongo-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "mongo", "deployer": "fate-operator", "name": name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { var a *int32; var i int32 = 1; a = &i; return a }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"fate": "kubefate", "apps": "mongo", "name": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"fate": "kubefate", "apps": "mongo", "deployer": "fate-operator", "name": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mongo",
							Image: "mongo",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 27017},
							},
							Env: func() []corev1.EnvVar {
								env := make([]corev1.EnvVar, 0)
								for _, v := range kubefate.Spec.Config {
									if v.Name == "FATECLOUD_MONGO_USERNAME" {
										env = append(env, corev1.EnvVar{Name: "MONGO_INITDB_ROOT_USERNAME", Value: v.Value, ValueFrom: v.ValueFrom})
									}
									if v.Name == "FATECLOUD_MONGO_PASSWORD" {
										env = append(env, corev1.EnvVar{Name: "MONGO_INITDB_ROOT_PASSWORD", Value: v.Value, ValueFrom: v.ValueFrom})
									}
								}
								return env
							}(),
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

	var MongoService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mongo-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "mongo", "deployer": "fate-operator", "name": name},
		},
		Spec: corev1.ServiceSpec{
			SessionAffinity: corev1.ServiceAffinityNone,
			Ports: []corev1.ServicePort{
				{
					Name:     "27017",
					Port:     27017,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 27017,
						StrVal: "",
					},
				},
			},
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"fate": "kubefate", "apps": "mongo", "name": name},
		},
	}

	var kubefateIngress = &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubefate-%s", PREFIX, name),
			Namespace: namespace,
			Labels:    map[string]string{"fate": "kubefate", "apps": "kubefate", "deployer": "fate-operator", "name": name},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					Host: kubefate.Spec.IngressDomain,
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
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
		mongoService:    MongoService,
		kubefateDeploy:  kubefateServiceDeploy,
		mongoDeploy:     MongoDeploy,
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
		}
	}

	kubefate.Status.KubefateDeploy = kf.kubefateDeploy.Name

	// kubefate mongoDeploy
	kfgot.mongoDeploy = &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.mongoDeploy.Namespace,
		Name:      kf.mongoDeploy.Name,
	}, kfgot.mongoDeploy); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get mongoDeploy")
			return false, err
		}

		// mongoDeploy IsNotFound
		log.Info("IsNotFound mongoDeploy", "Name", kf.mongoDeploy.Name)
		err = r.Create(ctx, kf.mongoDeploy)
		if err != nil {
			log.Error(err, "create mongoDeploy")
			return false, err
		}
	} else {

		if !reflect.DeepEqual(kf.mongoDeploy.Spec.Template.Spec.Containers[0].Env, kfgot.mongoDeploy.Spec.Template.Spec.Containers[0].Env) {
			//ffmt.Print("kfgot mongoDeploy ", kfgot.mongoDeploy.Spec)
			//ffmt.Print("kf mongoDeploy ", kf.mongoDeploy.Spec)
			log.Info("Updating mongoDeploy", "name", kf.mongoDeploy.Name)
			err := r.Update(ctx, kf.mongoDeploy)
			if err != nil {
				log.Error(err, "update mongoDeploy")
				return false, err
			}
		}
	}

	kubefate.Status.MongoDeploy = kf.mongoDeploy.Name
	//ffmt.Println("kfgot kubefateDeploy", kfgot.mongoDeploy)

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
		}
	}
	kubefate.Status.KubefateService = kf.kubefateService.Name
	//ffmt.Println("kfgot kubefateService", kfgot.kubefateService)

	// kubefate mongoService
	kfgot.mongoService = &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kf.mongoService.Namespace,
		Name:      kf.mongoService.Name,
	}, kfgot.mongoService); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "get mongoService")
			return false, err
		}

		// mongoService IsNotFound
		log.Info("IsNotFound mongoService", "Name", kf.mongoService.Name)
		err = r.Create(ctx, kf.mongoService)
		if err != nil {
			log.Error(err, "create mongoService")
			return false, err
		}
	} else {
		kf.mongoService.Spec.ClusterIP = kfgot.mongoService.Spec.ClusterIP
		if !reflect.DeepEqual(kf.mongoService.Spec, kfgot.mongoService.Spec) {
			//ffmt.Print("kfgot mongoService ", kfgot.mongoService.Spec)
			//ffmt.Print("kf mongoService ", kf.mongoService.Spec)
			log.Info("Updating mongoService", "name", kf.mongoService.Name)
			err := r.Update(ctx, kf.mongoService)
			if err != nil {
				log.Error(err, "update mongoService")
				return false, err
			}
		}
	}
	kubefate.Status.MongoService = kf.mongoService.Name
	//ffmt.Println("kfgot kubefateService", kfgot.mongoService)

	// kubefate ingress
	kfgot.ingress = &extensionsv1beta1.Ingress{}
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
		}
	}
	kubefate.Status.Ingress = kf.ingress.Name

	if kfgot.kubefateDeploy.Status.UnavailableReplicas != 0 {
		kubefate.Status.Status = appv1beta1.Pending
		log.Info("Kubefate has not been deployed successfully")
		return false, nil
	}

	if kfgot.mongoDeploy.Status.UnavailableReplicas != 0 {
		kubefate.Status.Status = appv1beta1.Pending
		log.Info("MongoDB has not been deployed successfully")
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
