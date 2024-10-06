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

package backend

// FunctionCall represents a call to a function.
type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Result    any            `json:"result"`
}

// ToolCall represents a call to a tool.
type ToolCall struct {
	Function FunctionCall `json:"function"`
}

// ConversationResponse represents a response from the AI in a conversation.
type ConversationResponse struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

// Message represents a single message in a conversation.
type Message map[string]any

// Conversation represents a series of messages between the user and the AI.
type Conversation struct {
	Messages []Message
	// ToolRegistry is a map of tool names to their corresponding wrapper functions.
	Tools *ToolRegistry
}

// NewConversation creates a new Conversation instance.
func NewConversation() *Conversation {
	return &Conversation{
		Messages: make([]Message, 0),
		Tools:    NewToolRegistry(),
	}
}

func (c *Conversation) addMessage(role string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["role"] = role
	c.Messages = append(c.Messages, fields)
}

func (c *Conversation) addContentMessage(role, content string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["role"] = role
	fields["content"] = content
	c.Messages = append(c.Messages, fields)
}

// AddUserMessage adds a message from the user to the conversation.
func (c *Conversation) AddUserMessage(content string, fields map[string]any) {
	c.addContentMessage("user", content, fields)
}

// AddSystemMessage adds a message from the user to the conversation.
func (c *Conversation) AddSystemMessage(content string, fields map[string]any) {
	c.addContentMessage("system", content, fields)
}

// AddToolCall allows the conversation to record a tool invocation.
func (c *Conversation) AddToolCall(content string, fields map[string]any) {
	c.addContentMessage("tool", content, fields)
}

// AddAssistantMessage adds a message from the assistant to the conversation.
func (c *Conversation) AddAssistantMessage(content string, fields map[string]any) {
	c.addContentMessage("assistant", content, fields)
}

// MessageMap returns the conversation's messages as a list of maps.
func (c *Conversation) MessageMap() ([]map[string]any, error) {
	messageList := make([]map[string]any, 0, len(c.Messages))
	for _, message := range c.Messages {
		messageList = append(messageList, message)
	}

	return messageList, nil
}
