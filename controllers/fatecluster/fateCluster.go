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
	"encoding/json"
	"errors"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/db"
	"time"
)

func (kfc *KubefateClient) GetFateCluster(uuid string) (*FateCluster, error) {
	if uuid == "" {
		return nil, errors.New("FateCluster no find")
	}

	log := *kfc.Log

	req := &Request{
		Type: "GET",
		Path: "cluster/" + uuid,
		Body: nil,
	}

	resp, err := kfc.TrySend(req)
	if err != nil {
		log.V(1).Info("error http request", "err", err.Error())
		return nil, err
	}
	if resp.Code != 200 {
		log.V(1).Info("error resp code", "err", errors.New(string(resp.Body)))
		return nil, errors.New(string(resp.Body))
	}

	log.V(1).Info("request success", "body", string(resp.Body))

	type ClusterResult struct {
		Data *db.Cluster
		Msg  string
	}

	clusterResult := new(ClusterResult)

	err = json.Unmarshal(resp.Body, &clusterResult)
	if err != nil {
		log.Error(err, "Unmarshal resp body error")
		return nil, err
	}

	log.V(1).Info("get cluster success", "clusterResult", clusterResult)

	fateSpec := &FateSpec{
		Name:         clusterResult.Data.Name,
		Namespace:    clusterResult.Data.NameSpace,
		ChartName:    clusterResult.Data.ChartName,
		ChartVersion: clusterResult.Data.ChartVersion,
		Cover:        true,
		Data:         []byte(clusterResult.Data.Values),
	}

	return &FateCluster{
		Spec:   fateSpec,
		Status: clusterResult.Data,
	}, nil

}

func (kfc *KubefateClient) CreateFateCluster(fateCluster *FateCluster) (*db.Job, error) {
	log := *kfc.Log

	log.V(1).Info("Spec", "Spec", fateCluster.Spec)

	reqbody, err := json.Marshal(fateCluster.Spec)
	if err != nil {
		return nil, err
	}

	req := &Request{
		Type: "POST",
		Path: "cluster/",
		Body: reqbody,
	}
	log.V(1).Info("req", "req", req)
	resp, err := kfc.TrySend(req)
	if err != nil {
		log.V(1).Info("error http request", "err", err.Error())
		return nil, err
	}
	if resp.Code != 200 {
		log.V(1).Info("error resp code", "err", errors.New(string(resp.Body)))
		return nil, errors.New(string(resp.Body))
	}

	type ClusterJobResult struct {
		Data *db.Job
		Msg  string
	}
	//
	clusterJobResult := new(ClusterJobResult)

	err = json.Unmarshal(resp.Body, &clusterJobResult)
	if err != nil {
		log.Error(err, "Unmarshal resp body error")
		return nil, err
	}

	log.V(1).Info("create cluster success", "clusterJobResult", clusterJobResult)
	fateCluster.LastJob = clusterJobResult.Data
	fateCluster.KubefateJob = append(fateCluster.KubefateJob, clusterJobResult.Data)
	return clusterJobResult.Data, nil
}

func (kfc *KubefateClient) CheckFateClusterJob(uuid string) (string, error) {
	log := *kfc.Log
	retry := 5
	startTime := time.Now()
	var clusterId string
	for {

		time.Sleep(5 * time.Second)
		job, err := kfc.GetFateClusterJob(uuid)
		if err != nil || job == nil {
			log.Error(err, "CheckFateClusterJob Get FateClusterJob")
			retry--
		}
		if job.Status == db.Success_j {
			clusterId = job.ClusterId
			break
		}
		if job.Status != db.Running_j && job.Status != db.Pending_j && job.Status != db.Success_j {
			return "", errors.New("job run failed, " + job.Result)
		}

		if retry < 0 {
			return "", errors.New("maximum attempts exceeded")
		}
		if time.Now().After(startTime.Add(1 * time.Hour)) {
			return "", errors.New("job run timeOut")
		}
	}

	return clusterId, nil
}

func (kfc *KubefateClient) GetFateClusterJob(uuid string) (*db.Job, error) {
	if uuid == "" {
		return nil, errors.New("FateClusterJob no find")
	}

	log := *kfc.Log

	req := &Request{
		Type: "GET",
		Path: "job/" + uuid,
		Body: nil,
	}

	resp, err := kfc.TrySend(req)
	if err != nil {
		log.V(1).Info("error http request", "err", err.Error())
		return nil, err
	}
	if resp.Code != 200 {
		log.V(1).Info("error resp code", "err", errors.New(string(resp.Body)))
		return nil, errors.New(string(resp.Body))
	}

	log.V(1).Info("request success", "body", string(resp.Body))

	type JobResult struct {
		Data *db.Job
		Msg  string
	}

	jobResult := new(JobResult)

	err = json.Unmarshal(resp.Body, &jobResult)
	if err != nil {
		log.Error(err, "Unmarshal resp body error")
		return nil, err
	}

	log.V(1).Info("get clusterJob success", "jobResult", jobResult)

	return jobResult.Data, nil

}
func (kfc *KubefateClient) UpdateFateCluster(fateCluster *FateCluster) (*db.Job, error) {

	log := *kfc.Log

	reqbody, err := json.Marshal(fateCluster.Spec)
	if err != nil {
		return nil, err
	}
	req := &Request{
		Type: "PUT",
		Path: "cluster/" + fateCluster.Status.Uuid,
		Body: reqbody,
	}

	resp, err := kfc.TrySend(req)
	if err != nil {
		log.V(1).Info("error http request", "err", err.Error())
		return nil, err
	}
	if resp.Code != 200 {
		log.V(1).Info("error resp code", "err", errors.New(string(resp.Body)))
		return nil, errors.New(string(resp.Body))
	}

	log.V(1).Info("request success", "body", string(resp.Body))

	type ClusterJobResult struct {
		Data *db.Job
		Msg  string
	}

	clusterJobResult := new(ClusterJobResult)

	err = json.Unmarshal(resp.Body, &clusterJobResult)
	if err != nil {
		log.Error(err, "Unmarshal resp body error")
		return nil, err
	}

	log.V(1).Info("get clusterJob success", "clusterResultMsg", clusterJobResult)

	fateCluster.LastJob = clusterJobResult.Data
	fateCluster.KubefateJob = append(fateCluster.KubefateJob, clusterJobResult.Data)
	return clusterJobResult.Data, nil
}

func (kfc *KubefateClient) DeleteFateCluster(fateCluster *FateCluster) (*db.Job, error) {

	log := *kfc.Log

	req := &Request{
		Type: "DELETE",
		Path: "cluster/" + fateCluster.Status.Uuid,
		Body: nil,
	}

	resp, err := kfc.TrySend(req)
	if err != nil {
		log.V(1).Info("error http request", "err", err.Error())
		return nil, err
	}
	if resp.Code != 200 {
		log.V(1).Info("error resp code", "err", errors.New(string(resp.Body)))
		return nil, errors.New(string(resp.Body))
	}

	log.V(1).Info("request success", "body", string(resp.Body))

	type ClusterJobResult struct {
		Data *db.Job
		Msg  string
	}

	clusterJobResult := new(ClusterJobResult)

	err = json.Unmarshal(resp.Body, &clusterJobResult)
	if err != nil {
		log.Error(err, "Unmarshal resp body error")
		return nil, err
	}

	log.V(1).Info("get clusterJob success", "clusterResultMsg", clusterJobResult)

	fateCluster.LastJob = clusterJobResult.Data
	fateCluster.KubefateJob = append(fateCluster.KubefateJob, clusterJobResult.Data)
	return clusterJobResult.Data, nil
}
