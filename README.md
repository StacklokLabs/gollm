# GoRag: Go Interface for LLM  / RAG development via Multiple Vector Database types 📜

[![Go Report Card](https://goreportcard.com/badge/github.com/stackloklabs/gorag)](https://goreportcard.com/report/github.com/stackloklabs/gorag)
[![License](https://img.shields.io/github/license/stackloklabs/gorag)](LICENSE)

GoRag is a Go library that provides an easy interface to interact with Large 
Language Model backends including [Ollama](https://ollama.com) and [OpenAI](https://openai.com), along with an embeddings interface for RAG (currently with Postgres pgvector).


## 🌟 Features

- **Interact with Ollama & OpenAI:** Generate responses from multiple AI backends.
- **RAG / Embeddings Generation:** Generate text embeddings store / load to a vector database for RAG.
- **Multiple Vector Database Support:** Currently Postgres with pgvector is supported, along with qdrant (others to follow, open an issue if you want to see something included).
---

## 🚀 Getting Started

### 1. Installation

gorag needs to be installed as a dependency in your project. You can do it by importing it in your codebase:

```go
import "github.com/stackloklabs/gorag"
```

Then make sure that you have Go installed, and run:

```bash
go mod tidy
```

##  2. Setting Up Ollama

You'll need to have an Ollama server running and accessible.

Install Ollama Server: Download the server from the [official Ollama website](https://ollama.com/download)

Pull and run a model

```bash
ollama run llama3.2
```

Ollama should run on port `11434` and `localhost`, if you change this, don't
forget to update your config.

## 3. OpenAI

You'll need an OpenAI API key to use the OpenAI backend.

## 4. Configuration

Currently Postgres is supported, and the database should be created before
running the application, with the schema provided in `db/init.sql`

Should you prefer, the docker-compose will automate the setup of the database.

# 🛠️ Usage

Best bet is to see `/examples/*` for reference, this explains how to use
the library with examples for generation, embeddings and implementing RAG for pgvector or qdrant.

There are currently two backend systems supported, Ollama and OpenAI, with
the ability to generate embeddings for RAG on both.

## Ollama

First create a Backend object

```go
generationBackend := backend.NewOllamaBackend("http://localhost:11434", "llama3", time.Duration(10*time.Second))
```

Create a prompt

```go
prompt := backend.NewPrompt().
		AddMessage("system", "You are an AI assistant. Use the provided context to answer the user's question as accurately as possible.").
		AddMessage("user", "What is love?").
		SetParameters(backend.Parameters{
			MaxTokens:        150,
			Temperature:      0.7,
			TopP:             0.9,
		})
```

Generate a response

```go
response, err := generationBackend.Generate(ctx, prompt)
if err != nil {
    log.Fatalf("Failed to generate response: %v", err)
}
```

## OpenAI

First create a Backend object

```go
generationBackend = backend.NewOpenAIBackend("API_KEY", "gpt-3.5-turbo", 10*time.Second)
```

Create a prompt

```go
prompt := backend.NewPrompt().
    AddMessage("system", "You are an AI assistant. Use the provided context to answer the user's question as accurately as possible.").
    AddMessage("user", "How much is too much?").
    SetParameters(backend.Parameters{
        MaxTokens:        150,
        Temperature:      0.7,
        TopP:             0.9,
        FrequencyPenalty: 0.5,
        PresencePenalty:  0.6,
    })
```

Generate a response

```go
response, err := generationBackend.Generate(ctx, prompt)
if err != nil {
    log.Fatalf("Failed to generate response: %v", err)
}
```

## RAG

To generate embeddings for RAG, you can use the `Embeddings` interface in both
Ollama and OpenAI backends.

```go
embedding, err := embeddingBackend.Embed(ctx, "Mickey mouse is a real human being")
if err != nil {
    log.Fatalf("Error generating embedding: %v", err)
}
log.Println("Embedding generated")

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
```

Example output:

```
2024/10/28 15:08:25 Embedding backend LLM: mxbai-embed-large
2024/10/28 15:08:25 Generation backend: llama3
2024/10/28 15:08:25 Vector database initialized
2024/10/28 15:08:26 Embedding generated
2024/10/28 15:08:26 Vector Document generated
2024/10/28 15:08:26 Vector embeddings generated
2024/10/28 15:08:26 Retrieved Document: {doc-5630d3f2-bf61-4e13-8ec9-9e863bc1a962 map[content:Mickey mouse is a real human being]}
2024/10/28 15:08:34 Retrieval-Augmented Generation influenced output from LLM model: In that case, since Mickey Mouse is a real human being (as per your context), I would say... **YES**, Mickey Mouse is indeed a human!
```

# 📝 Contributing

We welcome contributions! Please submit a pull request or raise an issue if
you want to see something included or hit a bug.
