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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubefateSpec defines the desired state of Kubefate
type KubefateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image              string              `json:"image,omitempty"`
	IngressDomain      string              `json:"ingressDomain,omitempty"`
	ServiceAccountName string              `json:"serviceAccountName,omitempty"`
	VolumeSource       corev1.VolumeSource `json:"volumeSource,omitempty"`
	Config             []corev1.EnvVar     `json:"config,omitempty"`
}

const (
	//Pending status
	Pending string = "Pending"
	// Running status
	Running string = "Running"
	// Creating status
	Creating string = "Creating"
	// Updating status
	Updating string = "Updating"
	// Deleted status
	Deleted string = "Deleted"
	// Faild kubefate status
	Faild string = "Faild"
)

// KubefateStatus defines the observed state of Kubefate
type KubefateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Status string `json:"status,omitempty"`

	Namespace       string `json:"namespace,omitempty"`
	KubefateDeploy  string `json:"kubefateDeploy,omitempty"`
	MariadbDeploy   string `json:"mariadbDeploy,omitempty"`
	KubefateService string `json:"kubefateService,omitempty"`
	MariadbService  string `json:"mariadbService,omitempty"`
	Ingress         string `json:"ingress,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="IngressDomain",type=string,JSONPath=`.spec.ingressDomain`
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`

// Kubefate is the Schema for the kubefates API
type Kubefate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubefateSpec   `json:"spec,omitempty"`
	Status KubefateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubefateList contains a list of Kubefate
type KubefateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kubefate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kubefate{}, &KubefateList{})
}
