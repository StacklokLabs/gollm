// backend_test.go
package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	// Mock server to simulate OpenAI API
	mockResponse := OpenAIResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-3.5-turbo",
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "This is a test response.",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     5,
			CompletionTokens: 5,
			TotalTokens:      10,
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and URL path
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		// Write the mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Create an instance of OpenAIBackend with the mock server
	backend := &OpenAIBackend{
		APIKey:     "test-api-key",
		Model:      "gpt-3.5-turbo",
		HTTPClient: mockServer.Client(),
		BaseURL:    mockServer.URL,
	}

	ctx := context.Background()
	prompt := "Hello, world!"

	response, err := backend.Generate(ctx, prompt)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	// Validate the response
	if response.ID != mockResponse.ID {
		t.Errorf("Expected ID %s, got %s", mockResponse.ID, response.ID)
	}
	if response.Choices[0].Message.Content != mockResponse.Choices[0].Message.Content {
		t.Errorf("Expected content %s, got %s", mockResponse.Choices[0].Message.Content, response.Choices[0].Message.Content)
	}
}

func TestGenerateEmbedding(t *testing.T) {
	// Mock server to simulate OpenAI API
	mockResponse := OpenAIEmbeddingResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			{
				Object:    "embedding",
				Embedding: []float32{0.1, 0.2, 0.3},
				Index:     0,
			},
		},
		Model: "text-embedding-ada-002",
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and URL path
		if r.Method != "POST" || r.URL.Path != "/v1/embeddings" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		// Write the mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Create an instance of OpenAIBackend with the mock server
	backend := &OpenAIBackend{
		APIKey:     "test-api-key",
		HTTPClient: mockServer.Client(),
		BaseURL:    mockServer.URL,
	}

	ctx := context.Background()
	text := "Test embedding text."

	response, err := backend.Embed(ctx, text)
	if err != nil {
		t.Fatalf("GenerateEmbedding returned error: %v", err)
	}

	// Validate the response
	if response.Data[0].Embedding[0] != mockResponse.Data[0].Embedding[0] {
		t.Errorf("Expected embedding %v, got %v", mockResponse.Data[0].Embedding, response.Data[0].Embedding)
	}
	if response.Model != mockResponse.Model {
		t.Errorf("Expected model %s, got %s", mockResponse.Model, response.Model)
	}
}
