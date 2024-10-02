# Gollm: Go Interface for LLM development 📜

[![Go Report Card](https://goreportcard.com/badge/github.com/stackloklabs/gollm)](https://goreportcard.com/report/github.com/stackloklabs/gollm)
[![License](https://img.shields.io/github/license/stackloklabs/gollm)](LICENSE)

Gollm is a Go library that provides an easy interface to interact with Large 
Language Model backends including [Ollama](https://ollama.com) and [OpenAI](https://openai.com), along with an embeddings interface for RAG.


## 🌟 Features

- **Interact with Ollama & OpenAI:** Generate responses from multiple AI backends.
- **RAG / Embeddings Generation:** Generate text embeddings store / load to a vector database for RAG.

---

## 🚀 Getting Started


### 1. Installation

First, make sure you have Go installed. Then, add Gollm to your project:

```bash
go get github.com/stackloklabs/gollm
```


##  2. Setting Up Ollama

You'll need to have an Ollama server running and accessible.

Install Ollama Server: Download the server from the [official Ollama website](https://ollama.com/download)

Pull and run a model

```bash
ollama run qwen2.5
```

Ollama should run on port `11434` and `localhost`, if you change this, don't
forget to update your config.

## 3. OpenAI

You'll need an OpenAI API key to use the OpenAI backend, which can be be
set within the config as below.

## 4. Configuration

Gollm uses Viper to manage configuration settings.

Backends are configured for either generation or embeddings, and can be set to either Ollama or OpenAI.

For each backend Models is set. Note that for Ollama you will need to 
have this as running model, e.g. `ollama run qwen2.5` or `ollama run mxbai-embed-large`.

Finally, in the case of RAG embeddings, a database URL is required.

Currently Postgres is supported, and the database should be created before running the application, with the schena provided in `db/init.sql`

Should you wish, the docker-compose will automate the setup of the database.

```bash
cp examples/config-example.yaml ./config.yaml
```

```yaml
backend:
  embeddings: "ollama"  # or "ollama"
  generation: "ollama"  # or "openai"
ollama:
  host: "http://localhost:11434"
  gen_model: "qwen2.5"
  emb_model: "mxbai-embed-large"
openai:
  api_key: "your-key"
  gen_model: "gpt-3.5-turbo"
  emb_model: "text-embedding-ada-002"
database:
  url: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
log_level: "info"
```

# 🛠️ Usage

Best bet is to see `/examples/main.go` for reference, this explains how to use
the library with full examples for generation, embeddings and implementing RAG.


# 📝 Contributing

We welcome contributions! Please submit a pull request or raise an issue if
you want to see something included or hit a bug.
