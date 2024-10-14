// Copyright 2024 Stacklok, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
type OllamaBackend struct {
	Model   string
	Client  *http.Client
	BaseURL string
}

// Response represents the structure of the response received from the Ollama API.
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

// OllamaEmbeddingResponse represents the response from the Ollama API for embeddings.
type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllamaBackend creates a new OllamaBackend instance.
func NewOllamaBackend(baseURL, model string, timeout time.Duration) *OllamaBackend {
	return &OllamaBackend{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Generate produces a response from the Ollama API based on the given structured prompt.
//
// Parameters:
//   - ctx: The context for the API request, which can be used for cancellation.
//   - prompt: A structured prompt containing messages and parameters.
//
// Returns:
//   - A string containing the generated response from the Ollama model.
//   - An error if the API request fails or if there's an issue processing the response.
func (o *OllamaBackend) Generate(ctx context.Context, prompt *Prompt) (string, error) {
	url := o.BaseURL + generateEndpoint

	// Concatenate the messages into a single prompt string
	var promptText string
	for _, message := range prompt.Messages {
		// Append role and content into one string (adjust formatting as needed)
		promptText += message.Role + ": " + message.Content + "\n"
	}

	// Construct the request body with concatenated prompt
	reqBody := map[string]interface{}{
		"model":             o.Model,
		"prompt":            promptText, // Use concatenated string
		"max_tokens":        prompt.Parameters.MaxTokens,
		"temperature":       prompt.Parameters.Temperature,
		"top_p":             prompt.Parameters.TopP,
		"frequency_penalty": prompt.Parameters.FrequencyPenalty,
		"presence_penalty":  prompt.Parameters.PresencePenalty,
		"stream":            false, // Explicitly set stream to false
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to generate response from Ollama: "+
				"status code %d, response: %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	var result Response
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf(
			"failed to generate embeddings from Ollama: "+
				"status code %d, response: %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	var result OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Embedding, nil
}
