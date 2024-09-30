# Gollm: Go Interface for LLM development 📜

[![Go Report Card](https://goreportcard.com/badge/github.com/stackloklabs/gollm)](https://goreportcard.com/report/github.com/stackloklabs/gollm)
[![License](https://img.shields.io/github/license/stackloklabs/gollm)](LICENSE)

Gollm is a Go library that provides an easy interface to interact with AI backends 
like [Ollama](https://ollama.com) and [OpenAI](https://openai.com). 

Quickly generate responses and embeddings from these AI models and integrate
them into your Go applications.

## 🌟 Features

- **Interact with Ollama & OpenAI:** Generate responses from multiple AI backends.
- **Embeddings Generation:** Generate text embeddings for your applications.

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

Ollama should run on port `11435` and `localhost`, if you change this, don't
forget to update your config.

# 3. Configuration

Gollm uses Viper to manage configuration settings.

```bash
cp examples/config-example.yaml ./config.yaml
```

```yaml
ollama:
  host: "http://localhost:11434"
  model: "your-ollama-model-name"

openai:
  api_key: "your-openai-api-key"
  model: "text-davinci-003"
```

# 🛠️ Usage

Best bet is to see `/examples/main.go` for reference

# 📋 API Reference

Initialise the config (or roll your own)

```go
cfg := config.InitializeViperConfig("config", "yaml", ".")
```

## Ollama Integration

Create Ollama Backend Instance:

```go
ollamaBackend := backend.NewOllamaBackend(cfg.Get("ollama.host"), cfg.Get("ollama.model"))
```

Generate Response:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := ollamaBackend.Generate(ctx, "Your prompt here")

fmt.Printf("Model: %s\nResponse: %s\n", response.Model, response.Response)
```

Embeddings response:

Support is also present for the Ollama's embeddings API

```go
embeddingResponse, err := ollamaBackend.Embed(ctx, "Text to generate embedding for")
```

> **Note**
> 📝 Only certain models provide an embeddings interface, see [ollama docs](https://ollama.com/blog/embedding-models) for more details

## OpenAI Integration

Create OpenAI Backend Instance:

```go
openaiBackend := backend.NewOpenAIBackend(cfg.Get("openai.api_key"), cfg.Get("openai.model"))
```

Generate Response:

```go
response, err := openaiBackend.Generate(ctx, "Your prompt here")

if len(response.Choices) > 0 {
		fmt.Printf("Choice Index: %d\n", response.Choices[0].Index)
		fmt.Printf("Message Role: %s\n", response.Choices[0].Message.Role)
		fmt.Printf("Message Content: %s\n", response.Choices[0].Message.Content)
		fmt.Printf("Finish Reason: %s\n", response.Choices[0].FinishReason)
	}
```

Embeddings response:

Support is also present for the OpenAI embeddings API

```go
embeddingResponse, err := openaiBackend.GenerateEmbedding(ctx, "Text to generate embedding for")
```

# 📝 Contributing

We welcome contributions! Please submit a pull request or raise an issue if
you want to see something included or hit a bug.
