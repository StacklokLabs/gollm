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
	"github.com/qdrant/go-client/qdrant"
)

// QdrantVector represents a connection to Qdrant.
type QdrantVector struct {
	client *qdrant.Client
}

// Close closes the Qdrant client connection.
func (qv *QdrantVector) Close() {
	qv.client.Close()
}

// NewQdrantVector initializes a connection to Qdrant.
//
// Parameters:
//   - address: The Qdrant server address (e.g., "localhost").
//   - port: The port Qdrant is running on (e.g., 6333).
//
// Returns:
//   - A pointer to a new QdrantVector instance.
//   - An error if the connection fails, nil otherwise.
func NewQdrantVector(address string, port int) (*QdrantVector, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: address,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}
	return &QdrantVector{client: client}, nil
}

// SaveEmbedding stores an embedding and metadata in Qdrant.
//
// Parameters:
//   - ctx: Context for the operation.
//   - docID: A unique identifier for the document.
//   - embedding: A slice of float32 values representing the document's embedding.
//   - metadata: A map of additional information associated with the document.
//
// Returns:
//   - An error if the saving operation fails, nil otherwise.
func (qv *QdrantVector) SaveEmbedding(ctx context.Context, docID string, embedding []float32, metadata map[string]interface{}) error {
	point := &qdrant.PointStruct{
		Id:      qdrant.NewID(docID),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(metadata),
	}

	waitUpsert := true
	_, err := qv.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: "gollm", // Replace with actual collection name
		Wait:           &waitUpsert,
		Points:         []*qdrant.PointStruct{point},
	})
	if err != nil {
		return fmt.Errorf("failed to insert point: %w", err)
	}
	return nil
}

// QueryRelevantDocuments retrieves the most relevant documents based on a given embedding.
//
// Parameters:
//   - ctx: The context for the query.
//   - embedding: The query embedding.
//   - limit: The number of documents to return.
//
// Returns:
//   - A slice of QDrantDocument structs containing the most relevant documents.
//   - An error if the query fails.
func (qv *QdrantVector) QueryRelevantDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
	limitUint := uint64(limit) // Convert limit to uint64
	query := &qdrant.QueryPoints{
		CollectionName: "gollm", // Replace with actual collection name
		Query:          qdrant.NewQuery(embedding...),
		Limit:          &limitUint,
		WithPayload:    qdrant.NewWithPayloadInclude("content"),
	}

	response, err := qv.client.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search points: %w", err)
	}

	var docs []Document
	for _, point := range response {
		var docID string
		if numericID := point.Id.GetNum(); numericID != 0 {
			docID = fmt.Sprintf("%d", numericID) // Numeric ID
		} else {
			docID = point.Id.GetUuid() // UUID
		}
		metadata := convertPayloadToMap(point.Payload)
		doc := Document{
			ID:       docID,
			Metadata: metadata,
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// convertPayloadToMap converts a Qdrant Payload (map[string]*qdrant.Value) into a map[string]interface{}.
func convertPayloadToMap(payload map[string]*qdrant.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range payload {
		switch v := value.Kind.(type) {
		case *qdrant.Value_StringValue:
			result[key] = v.StringValue
		case *qdrant.Value_BoolValue:
			result[key] = v.BoolValue
		case *qdrant.Value_DoubleValue:
			result[key] = v.DoubleValue
		case *qdrant.Value_ListValue:
			var list []interface{}
			for _, item := range v.ListValue.Values {
				switch itemVal := item.Kind.(type) {
				case *qdrant.Value_StringValue:
					list = append(list, itemVal.StringValue)
				case *qdrant.Value_BoolValue:
					list = append(list, itemVal.BoolValue)
				case *qdrant.Value_DoubleValue:
					list = append(list, itemVal.DoubleValue)
				}
			}
			result[key] = list
		default:
			result[key] = nil
		}
	}
	return result
}

// QDrantInsertDocument inserts a document into the Qdrant vector store.
//
// Parameters:
//   - ctx: Context for the operation.
//   - vectorDB: A QdrantVector instance.
//   - content: The document content to be inserted.
//   - embedding: The embedding vector for the document.
//
// Returns:
//   - An error if the operation fails, nil otherwise.
func (qv *QdrantVector) InsertDocument(ctx context.Context, vectorDB *QdrantVector, content string, embedding []float32) error {
	// Generate a unique document ID (e.g., using UUID)
	docID := fmt.Sprintf("doc-%s", qdrant.NewID(""))

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

// CreateCollection creates a new collection in Qdrant
func (qv *QdrantVector) CreateCollection(ctx context.Context, collectionName string, vectorSize uint64, distance string) error {
	// Create the collection
	err := qv.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine, // Example: Cosine distance
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	return nil
}
