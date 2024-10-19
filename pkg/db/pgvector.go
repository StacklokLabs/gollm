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

// Package db provides database-related functionality for the application.
// It includes implementations for vector storage and retrieval using PostgreSQL
// with the pgvector extension, enabling efficient similarity search operations
// on high-dimensional vector data.

package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// PGVector represents a connection to a PostgreSQL database with pgvector extension.
// It provides methods for storing and querying vector embeddings.
type PGVector struct {
	conn *pgxpool.Pool
}

// Close closes the PostgreSQL connection pool.
func (pg *PGVector) Close() {
	pg.conn.Close()
}

// NewPGVector creates a new PGVector instance with a connection to the PostgreSQL database.
//
// Parameters:
//   - connString: A string containing the connection details for the PostgreSQL database.
//
// Returns:
//   - A pointer to a new PGVector instance.
//   - An error if the connection fails, nil otherwise.
func NewPGVector(connString string) (*PGVector, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &PGVector{conn: pool}, nil
}

// SaveEmbedding stores a document embedding and associated metadata in the database.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - docID: A unique identifier for the document.
//   - embedding: A slice of float32 values representing the document's embedding.
//   - metadata: A map of additional information associated with the document.
//
// Returns:
//   - An error if the saving operation fails, nil otherwise.
//
// SaveEmbeddings stores a document embedding and associated metadata in the PostgreSQL database, implementing the VectorDatabase interface.
func (pg *PGVector) SaveEmbeddings(ctx context.Context, docID string, embedding []float32, metadata map[string]interface{}) error {
	vector := pgvector.NewVector(embedding)

	// Determine the table based on the embedding length
	var query string
	switch len(embedding) {
	case 1536:
		query = `INSERT INTO openai_embeddings (doc_id, embedding, metadata) VALUES ($1, $2, $3)`
	case 1024:
		query = `INSERT INTO ollama_embeddings (doc_id, embedding, metadata) VALUES ($1, $2, $3)`
	default:
		return fmt.Errorf("unsupported embedding length: %d", len(embedding))
	}

	// Execute the query to insert the vector into the database
	_, err := pg.conn.Exec(ctx, query, docID, vector, metadata)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	return nil
}

// QueryRelevantDocuments retrieves the most relevant documents from the database based on the given embedding.
// It uses cosine similarity to find the closest matches and returns a slice of Document structs.
//
// Parameters:
//   - ctx: The context for the database query.
//   - embedding: A slice of float32 values representing the query embedding.
//
// Returns:
//   - A slice of Document structs containing the most relevant documents.
//   - An error if the query fails or if there's an issue scanning the results.
func (pg *PGVector) QueryRelevantDocuments(ctx context.Context, embedding []float32, backend string) ([]Document, error) {
	// Convert embedding to the required format
	vector := pgvector.NewVector(embedding)

	// Query similar vectors based on cosine similarity or any distance metric supported by pgvector.
	var query string
	switch backend {
	case "openai":
		query = `
			SELECT doc_id, metadata
			FROM openai_embeddings
			ORDER BY embedding <-> $1
			LIMIT 5
		`
	case "ollama":
		query = `
			SELECT doc_id, metadata
			FROM ollama_embeddings
			ORDER BY embedding <-> $1
			LIMIT 5
		`
	default:
		return nil, fmt.Errorf("unsupported backend: %s", backend)
	}
	rows, err := pg.conn.Query(ctx, query, vector)

	if err != nil {
		return nil, fmt.Errorf("failed to query relevant documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// ConvertMetadata converts a map of string keys and string values to a map of string keys and interface{} values.
// This is useful when working with metadata that needs to be stored in a more flexible format.
func ConvertMetadata(metadata map[string]string) map[string]interface{} {
	converted := make(map[string]interface{})
	for k, v := range metadata {
		converted[k] = v
	}
	return converted
}

// ConvertEmbeddingToPGVector converts a slice of float32 values representing an embedding
// into a string format compatible with PostgreSQL's vector type. The resulting string
// is a comma-separated list of values enclosed in curly braces, with each value
// formatted to 6 decimal places of precision.
func ConvertEmbeddingToPGVector(embedding []float32) string {
	var strValues []string
	for _, val := range embedding {
		strValues = append(strValues, fmt.Sprintf("%.6f", val)) // Use "%.6f" for precision
	}
	return fmt.Sprintf("{%s}", strings.Join(strValues, ","))
}

// InsertDocument inserts a document into the PGVector store, implementing the VectorDatabase interface.
func (pg *PGVector) InsertDocument(ctx context.Context, content string, embedding []float32) error {
	// Generate a unique document ID (for simplicity, using UUID)
	docID := fmt.Sprintf("doc-%s", uuid.New().String())

	// Create metadata
	metadata := map[string]interface{}{
		"content": content,
	}

	// Save the document and its embedding into the vector store
	err := pg.SaveEmbeddings(ctx, docID, embedding, metadata)
	if err != nil {
		return fmt.Errorf("error saving embedding: %v", err)
	}
	return nil
}
