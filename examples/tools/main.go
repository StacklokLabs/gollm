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

package main

import (
	"context"
	"github.com/stackloklabs/gollm/examples/tools/trusty"
	"log"
	"os"
	"strings"
	"time"

	"github.com/stackloklabs/gollm/pkg/backend"
)

var (
	ollamaHost     = "http://localhost:11434"
	ollamaGenModel = "qwen2.5"
	openaiModel    = "gpt-4o-mini"
)

const (
	systemMessage = `
You are a security and software sustainability expert AI assistant. Your task is to help users by evaluating software packages based on their prompt and an optional JSON summary provided by an external tool. 

- Your primary responsibility is to assess whether the package is **malicious**, **deprecated**, or **unsafe** based on the JSON input. 
- **Do not print back** the JSON for the user. Only use relevant information from the JSON for your recommendation. 
- If the package is **malicious**, **deprecated**, or **has very low score**, recommend a **safer alternative** package, and explain why it is a better choice.
- If the package is NEITHER **malicious**, **deprecated**, or **has very low score**, recommend the package.
- Do not place emphasis on metrics like **stars** or **forks** unless specifically relevant to safety or sustainability concerns.

If the user does not specify an **ecosystem** (e.g., npm, pypi, crates, Maven, Go), or a **language** politely ask the user for clarification.
- If the user asks about a python package, assume pypi.
- If the user asks about a java package, assume Maven.
- If the user asks about a javascript package, assume npm.
- If the user asks about a rust package, assume crates.
If the user doesn't specify a package manager or language, ask the user for clarification, but *do not* make assumptions about the language.

Your responses should be concise, clear, and helpful. Focus on security, safety, and active maintenance in this order.
`
	summarizeMessage = `
Summarize the tool's analysis of the package in clear, plain language for the user. 

- If the package is **malicious**, **deprecated**, or **no longer maintained**, provide a **bulleted list** of **two to three** safer alternative packages that serve the same purpose.
- If the package is **safe**, confirm the package as a **recommended option**.

Ensure the response is concise and easy to understand. Prioritize **clarity** and **helpfulness** in your explanation.
`
)

func main() {
	var generationBackend backend.Backend

	beSelection := os.Getenv("BACKEND")
	if beSelection == "" {
		log.Println("No backend selected with the BACKEND env variable. Defaulting to Ollama.")
		beSelection = "ollama"
	}
	modelSelection := os.Getenv("MODEL")
	if modelSelection == "" {
		switch beSelection {
		case "ollama":
			modelSelection = ollamaGenModel
		case "openai":
			modelSelection = openaiModel
		}
		log.Println("No model selected with the MODEL env variable. Defaulting to ", modelSelection)
	}

	switch beSelection {
	case "ollama":
		generationBackend = backend.NewOllamaBackend(ollamaHost, ollamaGenModel)
		log.Println("Using Ollama backend: ", ollamaGenModel)
	case "openai":
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			log.Fatalf("OPENAI_API_KEY is required for OpenAI backend")
		}
		generationBackend = backend.NewOpenAIBackend(openaiKey, openaiModel)
		log.Println("Using OpenAI backend: ", openaiModel)
	default:
		log.Fatalf("Unknown backend: %s", beSelection)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	userPrompt := os.Args[1:]
	if len(userPrompt) == 0 {
		log.Fatalf("Please provide a prompt")
	}

	convo := backend.NewConversation()
	convo.Tools.RegisterTool(trusty.Tool())
	// start the conversation. We add a system message to tune the output
	// and add the trusty tool to the conversation so that the model knows to call it.
	convo.AddSystemMessage(systemMessage, nil)
	convo.AddUserMessage(strings.Join(userPrompt, " "), nil)

	// generate the response
	resp, err := generationBackend.Converse(ctx, convo)
	if err != nil {
		log.Fatalf("Error generating response: %v", err)
	}

	if len(resp.ToolCalls) == 0 {
		log.Println("No tool calls in response.")
		log.Println("Response:", convo.Messages[len(convo.Messages)-1]["content"])
		return
	}

	log.Println("Tool called")

	// if there was a tool response, first just feed it back to the model so it makes sense of it
	_, err = generationBackend.Converse(ctx, convo)
	if err != nil {
		log.Fatalf("Error generating response: %v", err)
	}

	log.Println("Summarizing tool response")

	// summarize the tool response
	convo.AddSystemMessage(summarizeMessage, nil)
	_, err = generationBackend.Converse(ctx, convo)
	if err != nil {
		log.Fatalf("Error generating response: %v", err)
	}

	log.Println("Response:")
	log.Println(convo.Messages[len(convo.Messages)-1]["content"])
}
