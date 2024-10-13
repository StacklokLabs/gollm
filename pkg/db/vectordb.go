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
)

// Document represents a single document in the vector database.
// It contains a unique identifier and associated metadata.
type Document struct {
	ID       string
	Metadata map[string]interface{}
}

// VectorDatabase is the interface that both QdrantVector and PGVector implement
type VectorDatabase interface {
	InsertDocument(ctx context.Context, content string, embedding []float32) error
	QueryRelevantDocuments(ctx context.Context, embedding []float32, backend string) ([]Document, error)
	SaveEmbeddings(ctx context.Context, docID string, embedding []float32, metadata map[string]interface{}) error
}
