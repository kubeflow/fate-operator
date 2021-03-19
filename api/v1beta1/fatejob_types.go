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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FateJobSpec defines the desired state of FateJob
type FateJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image          string                 `json:"image,omitempty"`
	FateClusterRef corev1.ObjectReference `json:"fateClusterRef"`
	JobConf        JobConf                `json:"jobConf,omitempty"`
}

type JobConf struct {
	Pipeline    string `json:"pipeline,omitempty"`
	ModulesConf string `json:"modulesConf,omitempty"`
}

// FateJobStatus defines the observed state of FateJob
type FateJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status,omitempty"`

	JobStatus batchv1.JobStatus `json:"jobStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`

// FateJob is the Schema for the fatejobs API
type FateJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FateJobSpec   `json:"spec,omitempty"`
	Status FateJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FateJobList contains a list of FateJob
type FateJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FateJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FateJob{}, &FateJobList{})
}
