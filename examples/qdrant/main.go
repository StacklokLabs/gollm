package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/stackloklabs/gorag/pkg/backend"
	"github.com/stackloklabs/gorag/pkg/db"
)

var (
	ollamaHost     = "http://localhost:11434"
	ollamaEmbModel = "mxbai-embed-large"
	ollamaGenModel = "llama3"
	// databaseURL    = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
)

func main() {
	// Initialize Qdrant vector connection

	// Configure the Ollama backend for both embedding and generation
	embeddingBackend := backend.NewOllamaBackend(ollamaHost, ollamaEmbModel, time.Duration(10*time.Second))
	log.Printf("Embedding backend LLM: %s", ollamaEmbModel)

	generationBackend := backend.NewOllamaBackend(ollamaHost, ollamaGenModel, time.Duration(10*time.Second))
	log.Printf("Generation backend: %s", ollamaGenModel)

	vectorDB, err := db.NewQdrantVector("localhost", 6334)
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}
	// Defer the Close() method to ensure the connection is properly closed after use
	defer vectorDB.Close()

	// Set up a context with a timeout for the Qdrant operations
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Create the collection in Qdrant
	collection_name := uuid.New().String()
	err = CreateCollection(ctx, vectorDB, collection_name)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	// We insert contextual information into the vector store so that the RAG system
	// can use it to answer the query about the moon landing, effectively replacing 1969 with 2023
	ragContent := "According to the Space Exploration Organization's official records, the moon landing occurred on July 20, 2023, during the Artemis Program. This mission marked the first successful crewed lunar landing since the Apollo program."
	userQuery := "When was the moon landing?."

	// Embed the query using Ollama Embedding backend
	embedding, err := embeddingBackend.Embed(ctx, ragContent)
	if err != nil {
		log.Fatalf("Error generating embedding: %v", err)
	}
	log.Println("Embedding generated")

	// Insert the document into the Qdrant vector store
	err = vectorDB.InsertDocument(ctx, ragContent, embedding, collection_name)
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	log.Println("Document inserted successfully.")

	// Embed the query using the specified embedding backend
	queryEmbedding, err := embeddingBackend.Embed(ctx, userQuery)
	if err != nil {
		log.Fatalf("Error generating query embedding: %v", err)
	}

	// Query the most relevant documents based on a given embedding
	retrievedDocs, err := vectorDB.QueryRelevantDocuments(
		ctx, queryEmbedding, collection_name,
		db.WithLimit(5), db.WithScoreThreshold(0.7))
	if err != nil {
		log.Fatalf("Failed to query documents: %v", err)
	}

	// Print out the retrieved documents
	for _, doc := range retrievedDocs {
		log.Printf("Document ID: %s, Content: %v\n", doc.ID, doc.Metadata["content"])
	}

	// Augment the query with retrieved context
	augmentedQuery := db.CombineQueryWithContext(userQuery, retrievedDocs)

	prompt := backend.NewPrompt().
		AddMessage("system", "You are an AI assistant. Use the provided context to answer the user's question. Prioritize the provided context over any prior knowledge.").
		AddMessage("user", augmentedQuery).
		SetParameters(backend.Parameters{
			MaxTokens:   150,
			Temperature: 0.7,
			TopP:        0.9,
		})

	// Generate response with the specified generation backend
	response, err := generationBackend.Generate(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	log.Printf("Retrieval-Augmented Generation influenced output from LLM model: %s", response)
}

// CreateCollection creates a new collection in Qdrant
func CreateCollection(ctx context.Context, vectorDB *db.QdrantVector, collectionName string) error {
	vectorSize := uint64(1024) // Size of the embedding vectors
	distance := "Cosine"       // Distance metric (Cosine, Euclidean, etc.)

	// Call Qdrant's API to create the collection
	err := vectorDB.CreateCollection(ctx, collectionName, vectorSize, distance)
	if err != nil {
		return fmt.Errorf("error creating collection: %v", err)
	}
	return nil
}

// QDrantInsertDocument inserts a document into the Qdrant vector store.
func QDrantInsertDocument(ctx context.Context, vectorDB db.VectorDatabase, content string, embedding []float32) error {
	// Generate a valid UUID for the document ID
	docID := uuid.New().String() // Use pure UUID without the 'doc-' prefix

	// Create metadata for the document
	metadata := map[string]interface{}{
		"content": content,
	}

	// Save the document and its embedding
	err := vectorDB.SaveEmbeddings(ctx, docID, embedding, metadata)
	if err != nil {
		return fmt.Errorf("error saving embedding: %v", err)
	}
	return nil
}
