// ollama_backend_test.go
package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaGenerate(t *testing.T) {
	// Mock response from Ollama API
	mockResponse := Response{
		Model:     "test-model",
		CreatedAt: time.Now().Format(time.RFC3339),
		Response:  "This is a test response from Ollama.",
		Done:      true,
	}

	// Create a mock server to simulate the Ollama API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the request
		if r.Method != http.MethodPost || r.URL.Path != generateEndpoint {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		// Check Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode the request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Optional: Validate request body contents
		if reqBody["model"] != "test-model" {
			t.Errorf("Expected model 'test-model', got '%v'", reqBody["model"])
		}
		if reqBody["prompt"] != "Hello, Ollama!" {
			t.Errorf("Expected prompt 'Hello, Ollama!', got '%v'", reqBody["prompt"])
		}
		if reqBody["stream"] != false {
			t.Errorf("Expected stream false, got '%v'", reqBody["stream"])
		}

		// Write the mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Create an instance of OllamaBackend with the mock server
	backend := &OllamaBackend{
		Model:   "test-model",
		Client:  mockServer.Client(),
		BaseURL: mockServer.URL,
	}

	ctx := context.Background()
	prompt := "Hello, Ollama!"

	response, err := backend.Generate(ctx, prompt)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	// Validate the response
	if response.Model != mockResponse.Model {
		t.Errorf("Expected model %s, got %s", mockResponse.Model, response.Model)
	}
	if response.Response != mockResponse.Response {
		t.Errorf("Expected response '%s', got '%s'", mockResponse.Response, response.Response)
	}
	if response.Done != mockResponse.Done {
		t.Errorf("Expected Done %v, got %v", mockResponse.Done, response.Done)
	}
}

func TestOllamaEmbed(t *testing.T) {
	// Mock response from Ollama API
	mockResponse := OllamaEmbeddingResponse{
		Embedding: []float32{0.1, 0.2, 0.3},
	}

	// Create a mock server to simulate the Ollama API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the request
		if r.Method != http.MethodPost || r.URL.Path != embedEndpoint {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		// Check Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode the request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Optional: Validate request body contents
		if reqBody["model"] != "test-model" {
			t.Errorf("Expected model 'test-model', got '%v'", reqBody["model"])
		}
		if reqBody["prompt"] != "Test embedding text." {
			t.Errorf("Expected prompt 'Test embedding text.', got '%v'", reqBody["prompt"])
		}

		// Write the mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Create an instance of OllamaBackend with the mock server
	backend := &OllamaBackend{
		Model:   "test-model",
		Client:  mockServer.Client(),
		BaseURL: mockServer.URL,
	}

	ctx := context.Background()
	input := "Test embedding text."

	embedding, err := backend.Embed(ctx, input)
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}

	// Validate the response
	if len(embedding) != len(mockResponse.Embedding) {
		t.Errorf("Expected embedding length %d, got %d", len(mockResponse.Embedding), len(embedding))
	}
	for i, v := range embedding {
		if v != mockResponse.Embedding[i] {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, mockResponse.Embedding[i], v)
		}
	}
}
