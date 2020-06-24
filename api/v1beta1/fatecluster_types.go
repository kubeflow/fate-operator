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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FateClusterSpec defines the desired state of FateCluster
type FateClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterSpec *ClusterSpec           `json:"clusterSpec,omitempty"`
	Kubefate    KubefateNamespacedName `json:"kubefate,omitempty"`
	ClusterData string                 `json:"clusterData,omitempty"`
}

type KubefateNamespacedName struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// FateClusterStatus defines the observed state of FateCluster
type FateClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status            string `json:"status,omitempty"`
	KubefateJobId     string `json:"jobId,omitempty"`
	KubefateClusterId string `json:"clusterId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="PartyId",type=string,JSONPath=`.spec.clusterSpec.partyId`
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`

// FateCluster is the Schema for the fateclusters API
type FateCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FateClusterSpec   `json:"spec,omitempty"`
	Status FateClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FateClusterList contains a list of FateCluster
type FateClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FateCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FateCluster{}, &FateClusterList{})
}

type Istio struct {
	Enabled bool `json:"enabled,omitempty"`
}

type Exchange struct {
	IP   string `json:"ip,omitempty"`
	Port int32  `json:"port,omitempty"`
}

type PartyList struct {
	PartyID   int32  `json:"partyId,omitempty"`
	PartyIP   string `json:"partyIp,omitempty"`
	PartyPort int32  `json:"partyPort,omitempty"`
}

type NodeSelector struct {
}

type Rollsite struct {
	Type         string       `json:"type,omitempty"`
	NodePort     int32        `json:"nodePort,omitempty"`
	Exchange     Exchange     `json:"exchange,omitempty"`
	PartyList    []PartyList  `json:"partyList,omitempty"`
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
}

type List struct {
	Name                     string       `json:"name,omitempty"`
	NodeSelector             NodeSelector `json:"nodeSelector,omitempty"`
	SessionProcessorsPerNode int32        `json:"sessionProcessorsPerNode,omitempty"`
	SubPath                  string       `json:"subPath,omitempty"`
	ExistingClaim            string       `json:"existingClaim,omitempty"`
	StorageClass             string       `json:"storageClass,omitempty"`
	AccessMode               string       `json:"accessMode,omitempty"`
	Size                     string       `json:"size,omitempty"`
}

type Nodemanager struct {
	Count                    int32  `json:"count,omitempty"`
	SessionProcessorsPerNode int32  `json:"sessionProcessorsPerNode"`
	List                     []List `json:"list,omitempty"`
}

type Python struct {
	FateflowType     string `json:"fateflowType,omitempty"`
	FateflowNodePort int32  `json:"fateflowNodePort,omitempty"`
	// +kubebuilder:default:={}
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
}

type Mysql struct {
	NodeSelector  NodeSelector `json:"nodeSelector,omitempty"`
	IP            string       `json:"ip,omitempty"`
	Port          int32        `json:"port,omitempty"`
	Database      string       `json:"database,omitempty"`
	User          string       `json:"user,omitempty"`
	Password      string       `json:"password,omitempty"`
	SubPath       string       `json:"subPath,omitempty"`
	ExistingClaim string       `json:"existingClaim,omitempty"`
	StorageClass  string       `json:"storageClass,omitempty"`
	AccessMode    string       `json:"accessMode,omitempty"`
	Size          string       `json:"size,omitempty"`
}
type ClusterSpec struct {
	Name         string      `json:"name,omitempty"`
	Namespace    string      `json:"namespace,omitempty"`
	ChartName    string      `json:"chartName"`
	ChartVersion string      `json:"chartVersion"`
	PartyID      int32       `json:"partyId"`
	Registry     string      `json:"registry,omitempty"`
	PullPolicy   string      `json:"pullPolicy,omitempty"`
	Persistence  bool        `json:"persistence,omitempty"`
	Istio        Istio       `json:"istio,omitempty"`
	Modules      []string    `json:"modules,omitempty"`
	Rollsite     Rollsite    `json:"rollsite,omitempty"`
	Nodemanager  Nodemanager `json:"nodemanager"`
	Python       Python      `json:"python,omitempty"`
	Mysql        Mysql       `json:"mysql,omitempty"`
	ServingIP    string      `json:"servingIp,omitempty"`
	ServingPort  int32       `json:"servingPort,omitempty"`
}
