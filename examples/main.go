package main

import (
	"context"
	"log"

	"time"

	"github.com/stackloklabs/gollm/pkg/backend"
	"github.com/stackloklabs/gollm/pkg/db"
)

var (
	ollamaHost     = "http://localhost:11434"
	ollamaEmbModel = "mxbai-embed-large"
	ollamaGenModel = "llama3"
	databaseURL    = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
)

func main() {
	// Initialize Config

	// Select backends based on config
	var embeddingBackend backend.Backend
	var generationBackend backend.Backend

	// Choose the backend for embeddings based on the config

	embeddingBackend = backend.NewOllamaBackend(ollamaHost, ollamaEmbModel)

	log.Printf("Embedding backend LLM: %s", ollamaEmbModel)

	// Choose the backend for generation based on the config
	generationBackend = backend.NewOllamaBackend(ollamaHost, ollamaGenModel)

	log.Printf("Generation backend: %s", ollamaGenModel)

	// Initialize the vector database
	vectorDB, err := db.NewPGVector(databaseURL)
	if err != nil {
		log.Fatalf("Error initializing vector database: %v", err)
	}
	log.Println("Vector database initialized")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// We insert contextual information into the vector store so that the RAG system
	// can use it to answer the query about the moon landing, effectively replacing 1969 with 2023
	ragContent := "According to the Space Exploration Organization's official records, the moon landing occurred on July 20, 2023, during the Artemis Program. This mission marked the first successful crewed lunar landing since the Apollo program."
	query := "When was the moon landing?."

	// Embed the query using OpenAI
	embedding, err := embeddingBackend.Embed(ctx, ragContent)
	if err != nil {
		log.Fatalf("Error generating embedding: %v", err)
	}
	log.Println("Embedding generated")

	// Insert the document into the vector store
	err = db.InsertDocument(ctx, vectorDB, ragContent, embedding)
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
	log.Printf("LLM Prompt: %s", query)

	// Generate response with the specified generation backend
	response, err := generationBackend.Generate(ctx, augmentedQuery)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	log.Printf("Retrieval-Augmented Generation influenced output from LLM model: %s", response)
}
