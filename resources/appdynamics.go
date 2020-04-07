/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type AppDynamics struct{}

func (AppDynamics) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (AppDynamics) Versions(source map[string]interface{}) (map[Version]string, error) {
	t, ok := source["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type must be specified")
	}

	uri := "https://download.appdynamics.com/download/downloadfilelatest"

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	var raw []struct {
		DownloadPath string `json:"download_path"`
		FileType     string `json:"filetype"`
		Version      string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	cp := regexp.MustCompile(`^([\d]+)\.([\d]+)\.([\d]+)[.-]?(.*)`)
	versions := make(map[Version]string, 1)
	for _, r := range raw {
		if t == r.FileType {
			if p := cp.FindStringSubmatch(r.Version); p != nil {
				ref := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
				if p[4] != "" {
					ref = fmt.Sprintf("%s-%s", ref, p[4])
				}

				versions[Version(ref)] = r.DownloadPath
			}
			break
		}
	}

	return versions, nil
}

func (AppDynamics) Modify(request *http.Request, source map[string]interface{}) (*http.Request, error) {
	u, ok := source["username"].(string)
	if !ok {
		return nil, fmt.Errorf("username must be specified")
	}

	p, ok := source["password"].(string)
	if !ok {
		return nil, fmt.Errorf("password must be specified")
	}

	payload := struct {
		Password string   `json:"password"`
		Scopes   []string `json:"scopes"`
		Username string   `json:"username"`
	}{
		Password: p,
		Scopes:   []string{"download"},
		Username: u,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal payload\n%w", err)
	}

	uri := "https://identity.msrv.saas.appdynamics.com/v2.0/oauth/token"
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	request.Header.Add("Authorization", fmt.Sprintf("%s %s", raw.TokenType, raw.AccessToken))

	return request, nil
}