package main

import (
	"context"

	"fmt"
	"log"
	"math"
	"time"

	"github.com/stackloklabs/gollm/pkg/backend"
	"github.com/stackloklabs/gollm/pkg/config"
)

func main() {
	cfg := config.InitializeViperConfig("config", "yaml", ".")

	// OLLAMA Example
	ollamaBackend := backend.NewOllamaBackend(cfg.Get("ollama.host"), cfg.Get("ollama.model"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ollamaResponse, err := ollamaBackend.Generate(ctx, "Hello, how are you?")
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Fatal("timeout while waiting for Ollama response")
		}
		log.Fatalf("failed to generate response: %v", err)
	}

	fmt.Printf("Model: %s\n", ollamaResponse.Model)
	fmt.Printf("Created At: %s\n", ollamaResponse.CreatedAt)
	fmt.Printf("Response: %s\n", ollamaResponse.Response)
	fmt.Printf("Done: %t\n", ollamaResponse.Done)
	fmt.Printf("Done Reason: %s\n", ollamaResponse.DoneReason)
	fmt.Printf("Total Duration: %d ms\n", ollamaResponse.TotalDuration)
	fmt.Printf("Prompt Eval Count: %d\n", ollamaResponse.PromptEvalCount)
	fmt.Printf("Eval Count: %d\n", ollamaResponse.EvalCount)

	// OpenAI Example
	openaiBackend := backend.NewOpenAIBackend(cfg.Get("openai.api_key"), cfg.Get("openai.model"))

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	openAIResponse, err := openaiBackend.Generate(ctx, "Hello, how are you?")
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Fatal("Timeout while waiting for OpenAI response")
		}
		log.Fatalf("Failed to generate response from OpenAI: %v", err)
	}

	fmt.Printf("Model: %s\n", openAIResponse.Model)
	fmt.Printf("Created At: %d\n", openAIResponse.Created)
	fmt.Printf("ID: %s\n", openAIResponse.ID)
	fmt.Printf("Object: %s\n", openAIResponse.Object)

	if len(openAIResponse.Choices) > 0 {
		fmt.Printf("Choice Index: %d\n", openAIResponse.Choices[0].Index)
		fmt.Printf("Message Role: %s\n", openAIResponse.Choices[0].Message.Role)
		fmt.Printf("Message Content: %s\n", openAIResponse.Choices[0].Message.Content)
		fmt.Printf("Finish Reason: %s\n", openAIResponse.Choices[0].FinishReason)
	}

	fmt.Printf("Prompt Tokens: %d\n", openAIResponse.Usage.PromptTokens)
	fmt.Printf("Completion Tokens: %d\n", openAIResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens: %d\n", openAIResponse.Usage.TotalTokens)

	// Create a context with a timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Text to generate embedding for
	text := "Hello, world! This is a test of the embedding generation."

	// Call the GenerateEmbedding function
	response, err := openaiBackend.Embed(ctx, text)
	if err != nil {
		fmt.Printf("Error generating embedding: %v\n", err)
		return
	}

	// Print the embedding response details
	fmt.Printf("Embedding Object: %s\n", response.Object)
	fmt.Printf("Embedding Model: %s\n", response.Model)
	fmt.Printf("Prompt Tokens: %d\n", response.Usage.PromptTokens)
	fmt.Printf("Total Tokens: %d\n", response.Usage.TotalTokens)

	// Print the first few values of the embedding vector
	if len(response.Data) > 0 {
		fmt.Printf("Embedding Object: %s\n", response.Data[0].Object)
		fmt.Printf("Embedding Index: %d\n", response.Data[0].Index)
		fmt.Printf("Embedding Vector (first 5 values): %v\n", response.Data[0].Embedding[:5])
	}

	// Example of how to use the embedding
	if len(response.Data) > 0 && len(response.Data[0].Embedding) > 0 {
		// Calculate the magnitude of the embedding vector
		magnitude := 0.0
		for _, value := range response.Data[0].Embedding {
			magnitude += float64(value * value)
		}
		magnitude = math.Sqrt(magnitude)

		fmt.Printf("Embedding Vector Magnitude: %f\n", magnitude)
	}

}
