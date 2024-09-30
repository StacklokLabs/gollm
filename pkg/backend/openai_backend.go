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
)

// OpenAIBackend represents a backend for interacting with the OpenAI API.
// It holds the necessary credentials and configuration for making API requests.
type OpenAIBackend struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
	BaseURL    string
}

// NewOpenAIBackend creates and returns a new OpenAIBackend instance.
// It takes an API key and a model name as parameters.
//
// Parameters:
//   - apiKey: A string containing the OpenAI API key for authentication.
//   - model: A string specifying the name of the OpenAI model to use.
//
// Returns:
//   - *OpenAIBackend: A pointer to the newly created OpenAIBackend instance.
func NewOpenAIBackend(apiKey, model string) *OpenAIBackend {
	return &OpenAIBackend{
		APIKey:     apiKey,
		Model:      model,
		HTTPClient: http.DefaultClient,
		BaseURL:    "https://api.openai.com",
	}
}

// OpenAIResponse represents the structure of the response received from the OpenAI API.
// It contains information about the generated content, model details, and usage statistics.
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate produces a response from the OpenAI API based on the given prompt.
// It sends a request to the OpenAI chat completions endpoint and returns the response.
//
// Parameters:
//   - ctx: A context.Context for handling timeouts and cancellations.
//   - prompt: A string containing the user's input prompt.
//
// Returns:
//   - *OpenAIResponse: A pointer to the OpenAIResponse struct containing the API's response.
//   - error: An error if the request fails or if there's an issue processing the response.
func (o *OpenAIBackend) Generate(ctx context.Context, prompt string) (*OpenAIResponse, error) {
	url := o.BaseURL + "/v1/chat/completions"
	reqBody := map[string]interface{}{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to generate response from OpenAI: status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// OpenAIEmbeddingResponse represents the structure of the response received from OpenAI's embedding API.
// It contains information about the generated embeddings, including the model used and usage statistics.
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbedding creates an embedding vector representation of the input text using OpenAI's API.
// It takes a context for cancellation and timeout, and the text to be embedded.
// The function returns an EmbeddingResponse containing the embedding vector and related information,
// or an error if the API request fails or the response cannot be processed.
func (o *OpenAIBackend) Embed(ctx context.Context, text string) (*OpenAIEmbeddingResponse, error) {
	url := o.BaseURL + "/v1/embeddings"
	reqBody := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to generate embedding from OpenAI: status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
