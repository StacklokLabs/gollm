// Copyright 2024 Stacklok, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trusty

import (
	"encoding/json"
	"fmt"
	"github.com/stackloklabs/gollm/pkg/backend"
	"io"
	"net/http"
)

// Tool returns a backend.Tool object that can be used to interact with the trusty tool.
func Tool() backend.Tool {
	return backend.Tool{
		Type: "function",
		Function: backend.ToolFunction{
			Name:        "trusty",
			Description: "Evaluate the trustworthiness of a package",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"package_name": map[string]any{
						"type":        "string",
						"description": "The name of the package",
					},
					"ecosystem": map[string]any{
						"type":        "string",
						"description": "The ecosystem of the package",
						"enum":        []string{"npm", "pypi", "crates", "maven", "go"},
						"default":     "pypi",
					},
				},
				"required": []string{"package_name", "ecosystem"},
			},
			Wrapper: trustyReportWrapper,
		},
	}
}

func trustyReportWrapper(params map[string]any) (string, error) {
	packageName, ok := params["package_name"].(string)
	if !ok {
		return "", fmt.Errorf("package_name must be a string")
	}
	ecosystem, ok := params["ecosystem"].(string)
	if !ok {
		ecosystem = "PyPi"
	}
	return trustyReport(packageName, ecosystem)
}

func trustyReport(packageName string, ecosystem string) (string, error) {
	url := fmt.Sprintf("https://api.trustypkg.dev/v1/report?package_name=%s&package_type=%s", packageName, ecosystem)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var prettyJSON map[string]interface{}
	err = json.Unmarshal(body, &prettyJSON)
	if err != nil {
		return "", err
	}

	// Convert the JSON back to string
	jsonString, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}
