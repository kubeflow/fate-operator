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
	"io/ioutil"
	"net/http"
)

func getToken(loginUrl, username, pasword string) (string, error) {

	login := map[string]string{
		"username": username,
		"password": pasword,
	}

	loginJsonB, err := json.Marshal(login)

	body := bytes.NewReader(loginJsonB)
	Request, err := http.NewRequest("POST", loginUrl, body)
	if err != nil {
		return "", err
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(Request)
	if err != nil {
		return "", err
	}

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	Result := map[string]interface{}{}

	err = json.Unmarshal(rbody, &Result)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprint(Result["message"]))
	}

	token := fmt.Sprint(Result["token"])

	return token, nil
}
