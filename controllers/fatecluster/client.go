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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"io/ioutil"
	"net/http"
	"time"
)

func NewKubefateClient(ApiVersion, Domain, Username, Password string, Log *logr.Logger) (*KubefateClient, error) {
	// check kubefate
	kfc := &KubefateClient{
		ApiVersion: ApiVersion,
		Domain:     Domain,
		Username:   Username,
		Password:   Password,
		Log:        Log,
	}
	if !kfc.CheckClient() {
		err := errors.New("check kubefate error")
		return nil, err
	}
	return kfc, nil
}

func (kfc *KubefateClient) CheckClient() bool {
	retry := 3
	for retry < 0 {
		// check kubefate
		_, err := kfc.GetVersion()
		if err != nil {
			retry--
		}
		time.Sleep(time.Second * 2)
	}

	if retry < 0 {
		return false
	}

	return true
}

func (kfc *KubefateClient) GetVersion() (string, error) {
	log := *kfc.Log
	//var serviceVersion, err = func() (string, error) {

	serviceUrl := kfc.Domain
	apiVersion := kfc.ApiVersion
	Url := "http://" + serviceUrl + "/" + apiVersion + "/" + "version"
	body := bytes.NewReader(nil)
	log.V(1).Info("request info", "url", Url)
	request, err := http.NewRequest("GET", Url, body)
	if err != nil {
		log.Error(err, "New  request error")
		return "", err
	}

	loginUrl := "http://" + serviceUrl + "/" + apiVersion + "/user/login"
	token, err := getToken(loginUrl, kfc.Username, kfc.Password)
	if err != nil {
		log.Error(err, "get Token error")
		return "", err
	}
	log.V(1).Info("get Token error", "token", token)
	Authorization := fmt.Sprintf("Bearer %s", token)

	request.Header.Add("Authorization", Authorization)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(err, "http request error")
		return "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "read resp body error")
		return "", err
	}
	log.V(1).Info("request success", "body", string(respBody))

	type VersionResultMsg struct {
		Msg     string
		Version string
	}

	VersionResult := new(VersionResultMsg)

	err = json.Unmarshal(respBody, &VersionResult)
	if err != nil {
		log.Error(err, "Unmarshal resp body error")
		return "", err
	}

	return VersionResult.Version, nil
}

type Request struct {
	Type string
	Path string
	Body []byte
}

type Response struct {
	Code int
	Body []byte
}

func (kfc *KubefateClient) TrySend(r *Request) (*Response, error) {
	log := *kfc.Log
	retry := 5
	for retry > 0 {
		log.V(1).Info("retry", "retry", retry)
		resp, err := kfc.Send(r)
		if err != nil {
			log.V(1).Info("request error", "Type", r.Type, "Path", r.Path, "err", err)
			retry--
			time.Sleep(time.Second * 2)
			continue
		}
		if resp.Code != 200 {
			log.V(1).Info("request code", "Type", r.Type, "Path", r.Path, "respCode", resp.Code, "respBody", string(resp.Body))
			retry--
			time.Sleep(time.Second * 2)
			continue
		}
		if resp.Code == 200 {
			return resp, nil
		}
	}
	return nil, errors.New("maximum retries exceeded")
}

func (kfc *KubefateClient) Send(r *Request) (*Response, error) {
	log := *kfc.Log
	serviceUrl := kfc.Domain
	apiVersion := kfc.ApiVersion
	Url := "http://" + serviceUrl + "/" + apiVersion + "/" + r.Path
	body := bytes.NewReader(r.Body)
	log.V(1).Info("request info", "url", Url, "type", r.Type, "body", string(r.Body))
	request, err := http.NewRequest(r.Type, Url, body)
	if err != nil {
		log.Error(err, "New  request error")
		return nil, err
	}
	loginUrl := "http://" + serviceUrl + "/" + apiVersion + "/user/login"
	token, err := getToken(loginUrl, kfc.Username, kfc.Password)
	if err != nil {
		log.Error(err, "get Token error")
		return nil, err
	}
	Authorization := fmt.Sprintf("Bearer %s", token)

	request.Header.Add("Authorization", Authorization)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(err, "http request error")
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "read resp body error")
		return nil, err
	}

	return &Response{
		Code: resp.StatusCode,
		Body: respBody,
	}, nil
}

type Result struct {
	Data []*string
	Msg  string
}

func (r *Response) Unmarshal() *Result {
	res := new(Result)
	_ = json.Unmarshal(r.Body, &res)
	return res
}
