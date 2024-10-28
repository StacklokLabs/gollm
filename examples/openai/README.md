# OpenAI Example Project

This project demonstrates how to use gorag's OpenAI's API backend.

This code demonstrates using OpenAI embeddings and generation models, along with
how RAG overides LLM knowledge by changing an established fact, already learned
from the LLMs dataset during the previous training. This provides insight into
how RAG can be used to populate new knowledge context, post knowedge cut-off.

Should the demo work correctly, you will see the RAG Content, will claim the 
moon landings occured in 2023 (without RAG the LLM will state the correct date
of 1969)

![alt text](image.png)

## Setup

The three main objects needed for OpenAI usage, are an API key, generation model
and embeddings model.

This demo uses `text-embedding-ada-002` and `gpt-3.5-turbo`.

The API key should be exported from your environment variables

```bash
export OPENAI_API_KEY="MY_KEY"
```


## Usage

Run the main go file

```bash
go run examples/openai/main.go
```