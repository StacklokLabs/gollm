// Copyright 2024 Stacklok, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package backend

import "context"

// Backend defines the interface for interacting with various LLM backends.
type Backend interface {
	Generate(ctx context.Context, prompt *Prompt) (string, error)
	Embed(ctx context.Context, input string) ([]float32, error)
}

// Message represents a single role-based message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Parameters defines generation settings for LLM completions.
type Parameters struct {
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
}

// Prompt represents a structured prompt with role-based messages and parameters.
type Prompt struct {
	Messages   []Message  `json:"messages"`
	Parameters Parameters `json:"parameters"`
}

// NewPrompt creates and returns a new Prompt.
func NewPrompt() *Prompt {
	return &Prompt{}
}

// AddMessage adds a message with a specific role to the prompt.
func (p *Prompt) AddMessage(role, content string) *Prompt {
	p.Messages = append(p.Messages, Message{Role: role, Content: content})
	return p
}

// SetParameters sets the generation parameters for the prompt.
func (p *Prompt) SetParameters(params Parameters) *Prompt {
	p.Parameters = params
	return p
}
