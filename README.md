# Gollm: Go Interface for LLM development with RAG 📜

[![Go Report Card](https://goreportcard.com/badge/github.com/stackloklabs/gollm)](https://goreportcard.com/report/github.com/stackloklabs/gollm)
[![License](https://img.shields.io/github/license/stackloklabs/gollm)](LICENSE)

Gollm is a Go library that provides an easy interface to interact with Large 
Language Model backends including [Ollama](https://ollama.com) and [OpenAI](https://openai.com), along with an embeddings interface for RAG (currently with Postgres pgvector).


## 🌟 Features

- **Interact with Ollama & OpenAI:** Generate responses from multiple AI backends.
- **RAG / Embeddings Generation:** Generate text embeddings store / load to a vector database for RAG.

---

## 🚀 Getting Started


### 1. Installation

First, make sure you have Go installed. Then, add gollm to your project:

```bash
go get github.com/stackloklabs/gollm
```

##  2. Setting Up Ollama

You'll need to have an Ollama server running and accessible.

Install Ollama Server: Download the server from the [official Ollama website](https://ollama.com/download)

Pull and run a model

```bash
ollama run llama3
```

Ollama should run on port `11434` and `localhost`, if you change this, don't
forget to update your config.

## 3. OpenAI

You'll need an OpenAI API key to use the OpenAI backend.

## 4. Configuration

Currently Postgres is supported, and the database should be created before
running the application, with the schena provided in `db/init.sql`

Should you wish, the docker-compose will automate the setup of the database.

# 🛠️ Usage

Best bet is to see `/examples/*` for reference, this explains how to use
the library with examples for generation, embeddings and implementing RAG.

There are currently two backend systems supported, Ollama and OpenAI, with
the ability to generate embeddings for RAG.

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
```

A database is also required, we have support for PostGres with pgvector. See `/examples/*` 
for reference.

# 📝 Contributing

We welcome contributions! Please submit a pull request or raise an issue if
you want to see something included or hit a bug.
