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
	Id   int32  `json:"id,omitempty"`
	IP   string `json:"ip,omitempty"`
	Port int32  `json:"port,omitempty"`
}

type PartyList struct {
	PartyID   int32  `json:"partyId,omitempty"`
	PartyIP   string `json:"partyIp,omitempty"`
	PartyPort int32  `json:"partyPort,omitempty"`
}

type NodeSelector map[string]string

type Rollsite struct {
	Type         string       `json:"type,omitempty"`
	NodePort     int32        `json:"nodePort,omitempty"`
	Exchange     *Exchange    `json:"exchange,omitempty"`
	PartyList    []PartyList  `json:"partyList,omitempty"`
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
}

type Lbrollsite struct {
	Type         string       `json:"type,omitempty"`
	NodePort     int32        `json:"nodePort,omitempty"`
	Size         string       `json:"size,omitempty"`
	Exchange     *Exchange    `json:"exchange,omitempty"`
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
	StorageClass             string `json:"storageClass,omitempty"`
	AccessMode               string `json:"accessMode,omitempty"`
	List                     []List `json:"list,omitempty"`
}
type PythonSpark struct {
	Master       string `json:"master,omitempty"`
	Home         string `json:"home,omitempty"`
	CoresPerNode string `json:"cores_per_node,omitempty"`
	Nodes        string `json:"nodes,omitempty"`
}
type PythonHdfs struct {
	NameNode   string `json:"name_node,omitempty"`
	PathPrefix string `json:"path_prefix,omitempty"`
}
type PythonRabbitmq struct {
	Host       string `json:"host,omitempty"`
	MngPort    string `json:"mng_port,omitempty"`
	Port       string `json:"port,omitempty"`
	User       string `json:"user,omitempty"`
	Password   string `json:"password,omitempty"`
	RouteTable string `json:"route_table,omitempty"`
}
type PythonNginx struct {
	Host     string `json:"host,omitempty"`
	HTTPPort string `json:"http_port,omitempty"`
	GrpcPort string `json:"grpc_port,omitempty"`
}

type Python struct {
	Type         string          `json:"type,omitempty"`
	HTTPNodePort int32           `json:"httpNodePort,omitempty"`
	GrpcNodePort int32           `json:"grpcNodePort,omitempty"`
	NodeSelector NodeSelector    `json:"nodeSelector,omitempty"`
	EnabledNN    bool            `json:"enabledNN,omitempty"`
	Spark        *PythonSpark    `json:"spark,omitempty"`
	Hdfs         *PythonHdfs     `json:"hdfs,omitempty"`
	Rabbitmq     *PythonRabbitmq `json:"rabbitmq,omitempty"`
	Nginx        *PythonNginx    `json:"nginx,omitempty"`
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

type ImagePullSecrets struct {
	Name string `json:"name,omitempty"`
}
type Host struct {
	Fateboard  string `json:"fateboard,omitempty"`
	Client     string `json:"client,omitempty"`
	SparkUI    string `json:"sparkUI,omitempty"`
	RabbitmqUI string `json:"rabbitmqUI,omitempty"`
}
type Master struct {
	Image        string       `json:"Image,omitempty"`
	ImageTag     string       `json:"ImageTag,omitempty"`
	Replicas     int32        `json:"replicas,omitempty"`
	CPU          string       `json:"cpu,omitempty"`
	Memory       string       `json:"memory,omitempty"`
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	Type         string       `json:"type,omitempty"`
}
type Worker struct {
	Image        string       `json:"Image,omitempty"`
	ImageTag     string       `json:"ImageTag,omitempty"`
	Replicas     int32        `json:"replicas,omitempty"`
	CPU          string       `json:"cpu,omitempty"`
	Memory       string       `json:"memory,omitempty"`
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	Type         string       `json:"type,omitempty"`
}
type Spark struct {
	Master Master `json:"Master,omitempty"`
	Worker Worker `json:"Worker,omitempty"`
}
type Namenode struct {
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	Type         string       `json:"type,omitempty"`
}
type Datanode struct {
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	Type         string       `json:"type,omitempty"`
}

type Hdfs struct {
	Namenode *Namenode `json:"namenode,omitempty"`
	Datanode *Datanode `json:"datanode,omitempty"`
}

type Proxy struct {
	Host         string `json:"host,omitempty"`
	HTTPNodePort int32  `json:"httpNodePort,omitempty"`
	GrpcNodePort int32  `json:"grpcNodePort,omitempty"`
}

type Fateflow struct {
	Host         string `json:"host,omitempty"`
	HTTPNodePort int32  `json:"httpNodePort,omitempty"`
	GrpcNodePort int32  `json:"grpcNodePort,omitempty"`
}

type PartyInfo struct {
	Proxy    []Proxy    `json:"proxy,omitempty"`
	Fateflow []Fateflow `json:"fateflow,omitempty"`
}

type NGRouteTable map[string]PartyInfo

type Nginx struct {
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	Type         string       `json:"type,omitempty"`
	HTTPNodePort int32        `json:"httpNodePort,omitempty"`
	GrpcNodePort int32        `json:"grpcNodePort,omitempty"`
	RouteTable   NGRouteTable `json:"route_table,omitempty"`
}

type Party struct {
	Host string `json:"host,omitempty"`
	Port int32  `json:"port,omitempty"`
}

type RBRouteTable map[string]Party

type Rabbitmq struct {
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	Type         string       `json:"type,omitempty"`
	NodePort     int32        `json:"nodePort,omitempty"`
	DefaultUser  string       `json:"default_user,omitempty"`
	DefaultPass  string       `json:"default_pass,omitempty"`
	User         string       `json:"user,omitempty"`
	Password     string       `json:"password,omitempty"`
	RouteTable   RBRouteTable `json:"route_table,omitempty"`
}

// ClusterSpec
type ClusterSpec struct {
	Name                  string             `json:"name,omitempty"`
	Namespace             string             `json:"namespace,omitempty"`
	ChartName             string             `json:"chartName"`
	ChartVersion          string             `json:"chartVersion"`
	PartyID               int32              `json:"partyId"`
	Registry              string             `json:"registry,omitempty"`
	ImageTag              string             `json:"imageTag,omitempty"`
	PullPolicy            string             `json:"pullPolicy,omitempty"`
	ImagePullSecrets      []ImagePullSecrets `json:"imagePullSecrets,omitempty"`
	Persistence           bool               `json:"persistence,omitempty"`
	Istio                 Istio              `json:"istio,omitempty"`
	Modules               []string           `json:"modules,omitempty"`
	Backend               string             `json:"backend,omitempty"`
	Host                  *Host              `json:"bachostkend,omitempty"`
	Rollsite              *Rollsite          `json:"rollsite,omitempty"`
	Lbrollsite            *Lbrollsite        `json:"lbrollsite,omitempty"`
	Nodemanager           *Nodemanager       `json:"nodemanager"`
	Python                *Python            `json:"python,omitempty"`
	Mysql                 *Mysql             `json:"mysql,omitempty"`
	ExternalMysqlIP       string             `json:"externalMysqlIp,omitempty"`
	ExternalMysqlPort     string             `json:"externalMysqlPort,omitempty"`
	ExternalMysqlDatabase string             `json:"externalMysqlDatabase,omitempty"`
	ExternalMysqlUser     string             `json:"externalMysqlUser,omitempty"`
	ExternalMysqlPassword string             `json:"externalMysqlPassword,omitempty"`
	ServingIP             string             `json:"servingIp,omitempty"`
	ServingPort           int32              `json:"servingPort,omitempty"`
	Spark                 *Spark             `json:"spark,omitempty"`
	Hdfs                  *Hdfs              `json:"hdfs,omitempty"`
	Nginx                 *Nginx             `json:"nginx,omitempty"`
	Rabbitmq              *Rabbitmq          `json:"rabbitmq,omitempty"`
}
