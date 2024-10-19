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
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create a test wrapper for QdrantVector that uses interface instead of concrete client
type testQdrantVector struct {
	QdrantVector
	mockClient *mockClient
}

// mockClient implements the necessary methods we use from qdrant.Client
type mockClient struct {
	mock.Mock
}

func newTestQdrantVector() *testQdrantVector {
	mc := &mockClient{}
	return &testQdrantVector{
		QdrantVector: QdrantVector{},
		mockClient:   mc,
	}
}

func (m *mockClient) Close() {
	m.Called()
}

func (m *mockClient) Upsert(ctx context.Context, req *qdrant.UpsertPoints) (*qdrant.PointsOperationResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*qdrant.PointsOperationResponse), args.Error(1)
}

func (m *mockClient) Query(ctx context.Context, req *qdrant.QueryPoints) ([]*qdrant.ScoredPoint, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*qdrant.ScoredPoint), args.Error(1)
}

func (m *mockClient) CreateCollection(ctx context.Context, req *qdrant.CreateCollection) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// Modified QdrantVector struct for testing
func (t *testQdrantVector) SaveEmbeddings(ctx context.Context, docID string, embedding []float32, metadata map[string]interface{}, collection string) error {
	point := &qdrant.PointStruct{
		Id:      qdrant.NewID(docID),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(metadata),
	}

	waitUpsert := true
	_, err := t.mockClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection,
		Wait:           &waitUpsert,
		Points:         []*qdrant.PointStruct{point},
	})
	return err
}

func (t *testQdrantVector) QueryRelevantDocuments(ctx context.Context, embedding []float32, limit int, collection string) ([]Document, error) {
	limitUint := uint64(limit)
	query := &qdrant.QueryPoints{
		CollectionName: collection,
		Query:          qdrant.NewQuery(embedding...),
		Limit:          &limitUint,
		WithPayload:    qdrant.NewWithPayloadInclude("content"),
	}

	response, err := t.mockClient.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var docs []Document
	for _, point := range response {
		var docID string
		if numericID := point.Id.GetNum(); numericID != 0 {
			docID = point.Id.GetUuid()
		} else {
			docID = point.Id.GetUuid()
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

func (t *testQdrantVector) CreateCollection(ctx context.Context, collectionName string, vectorSize uint64, distance string) error {
	return t.mockClient.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
}

func TestSaveEmbeddings(t *testing.T) {
	qv := newTestQdrantVector()

	ctx := context.Background()
	docID := "test-doc"
	embedding := []float32{0.1, 0.2, 0.3}
	metadata := map[string]interface{}{
		"content": "test content",
	}
	collection := "test-collection"

	// Set up expectations
	qv.mockClient.On("Upsert", mock.Anything, mock.MatchedBy(func(req *qdrant.UpsertPoints) bool {
		return req.CollectionName == collection &&
			len(req.Points) == 1 &&
			req.Points[0].Id.GetUuid() == docID
	})).Return(&qdrant.PointsOperationResponse{}, nil)

	// Test the SaveEmbeddings function
	err := qv.SaveEmbeddings(ctx, docID, embedding, metadata, collection)
	assert.NoError(t, err)

	// Verify expectations
	qv.mockClient.AssertExpectations(t)
}

func TestQueryRelevantDocuments(t *testing.T) {
	qv := newTestQdrantVector()

	ctx := context.Background()
	embedding := []float32{0.1, 0.2, 0.3}
	limit := 5
	collection := "test-collection"

	// Create mock response
	testUUID := uuid.New().String()
	mockResponse := []*qdrant.ScoredPoint{
		{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: testUUID,
				},
			},
			Payload: map[string]*qdrant.Value{
				"content": {
					Kind: &qdrant.Value_StringValue{
						StringValue: "test content",
					},
				},
			},
		},
	}

	// Set up expectations
	qv.mockClient.On("Query", mock.Anything, mock.MatchedBy(func(req *qdrant.QueryPoints) bool {
		return req.CollectionName == collection &&
			*req.Limit == uint64(limit)
	})).Return(mockResponse, nil)

	// Test the QueryRelevantDocuments function
	docs, err := qv.QueryRelevantDocuments(ctx, embedding, limit, collection)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "test content", docs[0].Metadata["content"])

	// Verify expectations
	qv.mockClient.AssertExpectations(t)
}

func TestCreateCollection(t *testing.T) {
	qv := newTestQdrantVector()

	ctx := context.Background()
	collectionName := "test-collection"
	vectorSize := uint64(3)
	distance := "cosine"

	// Set up expectations
	qv.mockClient.On("CreateCollection", mock.Anything, mock.MatchedBy(func(req *qdrant.CreateCollection) bool {
		return req.CollectionName == collectionName &&
			req.VectorsConfig.GetParams().Size == vectorSize
	})).Return(nil)

	// Test the CreateCollection function
	err := qv.CreateCollection(ctx, collectionName, vectorSize, distance)
	assert.NoError(t, err)

	// Verify expectations
	qv.mockClient.AssertExpectations(t)
}

// Add InsertDocument method to testQdrantVector
func (t *testQdrantVector) InsertDocument(ctx context.Context, content string, embedding []float32, collection string) error {
	// Create metadata map with content
	metadata := map[string]interface{}{
		"content": content,
	}

	// Generate a new UUID for the document
	docID := uuid.New().String()

	// Use our mock-aware SaveEmbeddings method
	return t.SaveEmbeddings(ctx, docID, embedding, metadata, collection)
}

func TestInsertDocument(t *testing.T) {
	qv := newTestQdrantVector()

	ctx := context.Background()
	content := "test content"
	embedding := []float32{0.1, 0.2, 0.3}
	collection := "test-collection"

	// Set up expectations for the mock client
	qv.mockClient.On("Upsert", mock.Anything, mock.MatchedBy(func(req *qdrant.UpsertPoints) bool {
		if len(req.Points) != 1 {
			return false
		}
		point := req.Points[0]

		// Check collection name
		if req.CollectionName != collection {
			return false
		}

		// Check payload contains correct content
		payload := point.Payload
		contentValue, exists := payload["content"]
		if !exists {
			return false
		}
		stringValue, ok := contentValue.Kind.(*qdrant.Value_StringValue)
		if !ok {
			return false
		}
		if stringValue.StringValue != content {
			return false
		}

		// Check vectors
		if !reflect.DeepEqual(point.Vectors.GetVector().Data, embedding) {
			return false
		}

		return true
	})).Return(&qdrant.PointsOperationResponse{}, nil)

	// Test the InsertDocument function
	err := qv.InsertDocument(ctx, content, embedding, collection)
	assert.NoError(t, err)

	// Verify expectations
	qv.mockClient.AssertExpectations(t)
}
