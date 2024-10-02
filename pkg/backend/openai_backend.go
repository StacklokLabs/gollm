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
// It contains configuration details and methods for making API requests.
type OpenAIBackend struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
	BaseURL    string
}

// OpenAIEmbeddingResponse represents the structure of the response received from the OpenAI API
// for an embedding request. It contains the generated embeddings, usage statistics, and other
// metadata related to the API call.
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

// NewOpenAIBackend creates and returns a new OpenAIBackend instance.
//
// Parameters:
//   - apiKey: The API key for authenticating with the OpenAI API.
//   - model: The name of the OpenAI model to use for generating responses.
//
// Returns:
//   - A pointer to a new OpenAIBackend instance configured with the provided API key and model.
func NewOpenAIBackend(apiKey, model string) *OpenAIBackend {
	return &OpenAIBackend{
		APIKey:     apiKey,
		Model:      model,
		HTTPClient: http.DefaultClient,
		BaseURL:    "https://api.openai.com",
	}
}

// OpenAIResponse represents the structure of the response received from the OpenAI API
// for a chat completion request. It contains information about the generated text,
// usage statistics, and other metadata related to the API call.
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

// Generate sends a prompt to the OpenAI API and returns the generated response.
//
// Parameters:
//   - ctx: The context for the API request, which can be used for cancellation.
//   - prompt: The input text prompt for which to generate a response.
//
// Returns:
//   - A string containing the generated response from the OpenAI model.
//   - An error if the API request fails or if there's an issue processing the response.
func (o *OpenAIBackend) Generate(ctx context.Context, prompt string) (string, error) {
	url := o.BaseURL + "/v1/chat/completions"
	reqBody := map[string]interface{}{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {

		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to generate response from OpenAI: "+
			"status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Choices[0].Message.Content, nil
}

// Embed generates an embedding vector for the given text using the OpenAI API.
//
// Parameters:
//   - ctx: The context for the API request, which can be used for cancellation.
//   - text: The input text to be embedded.
//
// Returns:
//   - A slice of float32 values representing the embedding vector.
//   - An error if the API request fails or if there's an issue processing the response.
func (o *OpenAIBackend) Embed(ctx context.Context, text string) ([]float32, error) {
	url := o.BaseURL + "/v1/embeddings"
	reqBody := map[string]interface{}{
		"model": o.Model,
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
		return nil, fmt.Errorf("failed to generate embedding from OpenAI: "+
			"status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data[0].Embedding, nil
}
