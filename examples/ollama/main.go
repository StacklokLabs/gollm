package main

import (
	"context"
	"log"

	"time"

	"github.com/stackloklabs/gorag/pkg/backend"
	"github.com/stackloklabs/gorag/pkg/db"
)

var (
	ollamaHost     = "http://localhost:11434"
	ollamaEmbModel = "mxbai-embed-large"
	ollamaGenModel = "llama3"
	databaseURL    = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
)

func main() {

	// Configure the Ollama backend for both embedding and generation
	embeddingBackend := backend.NewOllamaBackend(ollamaHost, ollamaEmbModel, time.Duration(10*time.Second))
	log.Printf("Embedding backend LLM: %s", ollamaEmbModel)

	generationBackend := backend.NewOllamaBackend(ollamaHost, ollamaGenModel, time.Duration(10*time.Second))
	log.Printf("Generation backend: %s", ollamaGenModel)

	// Initialize the vector database
	vectorDB, err := db.NewPGVector(databaseURL)
	if err != nil {
		log.Fatalf("Error initializing vector database: %v", err)
	}
	log.Println("Vector database initialized")

	// Make sure to close the connection when done
	defer vectorDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close the connection when done
	defer vectorDB.Close()

	// We insert contextual information into the vector store so that the RAG system
	// can use it to answer the query about the moon landing, effectively replacing 1969 with 2023
	ragContent := "According to the Space Exploration Organization's official records, the moon landing occurred on July 20, 2023, during the Artemis Program. This mission marked the first successful crewed lunar landing since the Apollo program."
	query := "When was the moon landing?."

	// Embed the query using Ollama Embedding backend
	embedding, err := embeddingBackend.Embed(ctx, ragContent)
	if err != nil {
		log.Fatalf("Error generating embedding: %v", err)
	}
	log.Println("Embedding generated")

	// Insert the document into the vector store
	err = vectorDB.InsertDocument(ctx, ragContent, embedding)
	if err != nil {
		log.Fatalf("Error inserting document: %v", err)
	}
	log.Println("Vector Document generated")

	// Embed the query using the specified embedding backend
	queryEmbedding, err := embeddingBackend.Embed(ctx, query)
	if err != nil {
		log.Fatalf("Error generating query embedding: %v", err)
	}
	log.Println("Vector embeddings generated")

	// Retrieve relevant documents for the query embedding
	retrievedDocs, err := vectorDB.QueryRelevantDocuments(ctx, queryEmbedding, "ollama")
	if err != nil {
		log.Fatalf("Error retrieving relevant documents: %v", err)
	}

	// Log the retrieved documents to see if they include the inserted content
	for _, doc := range retrievedDocs {
		log.Printf("Retrieved Document: %v", doc)
	}

	// Augment the query with retrieved context
	augmentedQuery := db.CombineQueryWithContext(query, retrievedDocs)

	prompt := backend.NewPrompt().
		AddMessage("system", "You are an AI assistant. Use the provided context to answer the user's question as accurately as possible.").
		AddMessage("user", augmentedQuery).
		SetParameters(backend.Parameters{
			MaxTokens:   150, // Supported by LLaMa
			Temperature: 0.7, // Supported by LLaMa
			TopP:        0.9, // Supported by LLaMa
		})

	// Generate response with the specified generation backend
	response, err := generationBackend.Generate(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	log.Printf("Retrieval-Augmented Generation influenced output from LLM model: %s", response)
}
