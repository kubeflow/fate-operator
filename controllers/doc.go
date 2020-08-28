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

// +kubebuilder:rbac:groups=app.kubefate.net,resources=events,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="",resources=configmap,verbs=get;list;watch;create;update;patch;delete;deletecollection

// +kubebuilder:rbac:groups=apps,resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;list;watch;create;update;patch;delete;deletecollection

// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=apps,resources=replicasets/status,verbs=get;list;watch;create;update;patch;delete;deletecollection

// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch;deletecollection

// +kubebuilder:rbac:groups=extensions,resources=ingress,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=extensions,resources=ingress/status,verbs=get;update;patch;deletecollection

// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=extensions,resources=ingresses/status,verbs=get;update;patch;deletecollection

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;deletecollection
