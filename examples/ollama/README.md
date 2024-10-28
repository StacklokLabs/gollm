# Ollama Example Project

This project demonstrates how to use gorag's OpenAI's API backend.

This code demonstrates using Ollama embeddings and generation models, along with
how RAG overides LLM knowledge by changing an established fact, already learned
from the LLMs dataset during the previous training. This provides insight into
how RAG can be used to populate new knowledge context, post knowedge cut-off.

Should the demo work correctly, you will see the RAG Content, will claim the 
moon landings occured in 2023 (without RAG the LLM will state the correct date
of 1969)

![alt text](image.png)

## Setup

The two main objects needed for Ollama usage, are a generation model
and embeddings model.

This demo uses `mxbai-embed-large` and `llama3`.


## Usage

Run the main go file

```bash
go run examples/ollama/main.go
```