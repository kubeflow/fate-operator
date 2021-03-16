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

package fatecluster

import (
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	"github.com/go-logr/logr"
)

type KubefateClient struct {
	ApiVersion string
	Domain     string
	Username   string
	Password   string
	Log        *logr.Logger
	LastError  error
}

type FateCluster struct {
	Spec        *FateSpec
	Status      *modules.Cluster
	LastJob     *modules.Job
	KubefateJob []*modules.Job
}

type FateSpec struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	Cover        bool   `json:"cover"`
	Data         []byte `json:"data"`
}
