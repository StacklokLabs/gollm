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

package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	generateEndpoint = "/api/generate"
	embedEndpoint    = "/api/embeddings"
	defaultTimeout   = 30 * time.Second
)

// OllamaBackend represents a backend for interacting with the Ollama API.
// It holds the necessary configuration for making API requests.
type OllamaBackend struct {
	Model   string
	Client  *http.Client
	BaseURL string
}

// Response represents the structure of the response received from the Ollama API.
// It contains information about the generated content, model details, and performance metrics.
type Response struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	DoneReason         string `json:"done_reason"`
	Context            []int  `json:"context"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}

// OllamaEmbeddingResponse represents the structure of the response received from the Ollama API for embeddings.
type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllamaBackend creates and returns a new OllamaBackend instance.
// It takes a base URL and a model name as parameters.
func NewOllamaBackend(baseURL, model string) *OllamaBackend {
	return &OllamaBackend{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Generate produces a response from the Ollama API based on the given prompt.
// It sends a request to the Ollama generate endpoint and returns the response.
func (o *OllamaBackend) Generate(ctx context.Context, prompt string) (*Response, error) {
	url := o.BaseURL + generateEndpoint
	reqBody := map[string]interface{}{
		"model":  o.Model,
		"prompt": prompt,
		"stream": false,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to generate response from Ollama: status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Embed generates embeddings for the given input text using the Ollama API.
func (o *OllamaBackend) Embed(ctx context.Context, input string) ([]float32, error) {
	url := o.BaseURL + embedEndpoint
	reqBody := map[string]interface{}{
		"model":  o.Model,
		"prompt": input,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to generate embeddings from Ollama: status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Embedding, nil
}
