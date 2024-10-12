package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stackloklabs/gollm/pkg/db"
	"log"
	"time"
)

func main() {
	qdrantVector, err := db.NewQdrantVector("localhost", 6334)
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}
	defer qdrantVector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Create the collection
	err = CreateCollection(ctx, qdrantVector)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	// Example embedding and content
	embedding := []float32{0.05, 0.61, 0.76, 0.74}
	content := "This is a test document."

	// Insert the document into the Qdrant vector store
	err = QDrantInsertDocument(ctx, qdrantVector, content, embedding)
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	log.Println("Document inserted successfully.")

	// Query the most relevant documents based on a given embedding
	docs, err := qdrantVector.QueryRelevantDocuments(ctx, embedding, 5)
	if err != nil {
		log.Fatalf("Failed to query documents: %v", err)
	}

	// Print out the results
	for _, doc := range docs {
		log.Printf("Document ID: %s, Content: %v\n", doc.ID, doc.Metadata["content"])
	}
}

// CreateCollection creates a new collection in Qdrant
func CreateCollection(ctx context.Context, vectorDB *db.QdrantVector) error {
	collectionName := "gollm" // Replace with your collection name
	vectorSize := uint64(4)   // Size of the embedding vectors
	distance := "Cosine"      // Distance metric (Cosine, Euclidean, etc.)

	// Call Qdrant's API to create the collection
	err := vectorDB.CreateCollection(ctx, collectionName, vectorSize, distance)
	if err != nil {
		return fmt.Errorf("error creating collection: %v", err)
	}
	return nil
}

// QDrantInsertDocument inserts a document into the Qdrant vector store.
func QDrantInsertDocument(ctx context.Context, vectorDB *db.QdrantVector, content string, embedding []float32) error {
	// Generate a valid UUID for the document ID
	docID := uuid.New().String() // Use pure UUID without the 'doc-' prefix

	// Create metadata for the document
	metadata := map[string]interface{}{
		"content": content,
	}

	// Save the document and its embedding
	err := vectorDB.SaveEmbedding(ctx, docID, embedding, metadata)
	if err != nil {
		return fmt.Errorf("error saving embedding: %v", err)
	}
	return nil
}
