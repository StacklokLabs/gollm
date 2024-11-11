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
	"slices"

	"github.com/google/uuid"
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
//
// SaveEmbeddings stores an embedding and metadata in Qdrant, implementing the VectorDatabase interface.
func (qv *QdrantVector) SaveEmbeddings(ctx context.Context, docID string, embedding []float32, metadata map[string]interface{}, collection string) error {
	point := &qdrant.PointStruct{
		Id:      qdrant.NewID(docID),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(metadata),
	}

	waitUpsert := true
	_, err := qv.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection, // Replace with actual collection name
		Wait:           &waitUpsert,
		Points:         []*qdrant.PointStruct{point},
	})
	if err != nil {
		return fmt.Errorf("failed to insert point: %w", err)
	}
	return nil
}

// QueryOpt represents an option for a query. This is the type that should
// be returned from query options functions.
type QueryOpt func(*qdrant.QueryPoints)

// WithLimit sets the limit of the number of documents to return in a query.
func WithLimit(limit uint64) QueryOpt {
	return func(q *qdrant.QueryPoints) {
		q.Limit = &limit
	}
}

// WithScoreThreshold sets the score threshold for a query. The higher the threshold, the more relevant the results.
func WithScoreThreshold(threshold float32) QueryOpt {
	return func(q *qdrant.QueryPoints) {
		q.ScoreThreshold = &threshold
	}
}

// RetrieveMetadata adds its arguments to the list of payload keys that are retrieved. Content is always retrieved
func RetrieveMetadata(keys ...string) QueryOpt {
	if !slices.Contains(keys, "content") {
		keys = append(keys, "content")
	}

	return func(q *qdrant.QueryPoints) {
		q.WithPayload = qdrant.NewWithPayloadInclude(keys...)
	}
}

// QueryRelevantDocuments retrieves the most relevant documents based on a given embedding.
//
// Parameters:
//   - ctx: The context for the query.
//   - embedding: The query embedding.
//   - limit: The number of documents to return.
//   - collection: The collection name to query.
//
// Returns:
//   - A slice of QDrantDocument structs containing the most relevant documents.
//   - An error if the query fails.
func (qv *QdrantVector) QueryRelevantDocuments(
	ctx context.Context, embedding []float32, collection string, queryOpts ...QueryOpt,
) ([]Document, error) {
	query := &qdrant.QueryPoints{
		CollectionName: collection, // Replace with actual collection name
		Query:          qdrant.NewQuery(embedding...),
		WithPayload:    qdrant.NewWithPayloadInclude("content"),
	}

	for _, opt := range queryOpts {
		opt(query)
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
		result[key] = convertQdrantValue(value)
	}
	return result
}

func convertQdrantValue(v *qdrant.Value) any {
	switch v := v.Kind.(type) {
	case *qdrant.Value_NullValue:
		return nil
	case *qdrant.Value_BoolValue:
		return v.BoolValue
	case *qdrant.Value_IntegerValue:
		return v.IntegerValue
	case *qdrant.Value_DoubleValue:
		return v.DoubleValue
	case *qdrant.Value_StringValue:
		return v.StringValue
	case *qdrant.Value_ListValue:
		result := make([]any, len(v.ListValue.Values))
		for i, val := range v.ListValue.Values {
			result[i] = convertQdrantValue(val)
		}
		return result
	case *qdrant.Value_StructValue:
		return convertQdrantStruct(v.StructValue)
	default:
		return nil
	}
}

func convertQdrantStruct(s *qdrant.Struct) map[string]any {
	result := make(map[string]any)
	for key, value := range s.Fields {
		result[key] = convertQdrantValue(value)
	}
	return result
}

// InsertMetadataOption represents a modifier for payload metadata.
type InsertMetadataOption func(metadata map[string]any)

// AddDocumentMetadata sets a key-value pair in the metadata. The value can be of any type.
func AddDocumentMetadata(key string, value any) InsertMetadataOption {
	return func(metadata map[string]any) {
		metadata[key] = value
	}
}

// InsertDocument inserts a document into the Qdrant vector store.
//
// Parameters:
//   - ctx: Context for the operation.
//   - vectorDB: A QdrantVector instance.
//   - content: The document content to be inserted.
//   - embedding: The embedding vector for the document.
//
// Returns:
//   - An error if the operation fails, nil otherwise.
//
// QdrantVector should implement the InsertDocument method as defined in VectorDatabase
func (qv *QdrantVector) InsertDocument(ctx context.Context, content string, embedding []float32, collection string, opts ...InsertMetadataOption) error {
	// Generate a valid UUID for the document ID
	docID := uuid.New().String() // Properly generate a UUID

	metadata := map[string]interface{}{
		"content": content,
	}

	for _, opt := range opts {
		opt(metadata)
	}

	// Call the SaveEmbeddings method to save the document and its embedding
	err := qv.SaveEmbeddings(ctx, docID, embedding, metadata, collection)
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
