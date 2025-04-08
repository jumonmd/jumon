// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ChatCompletionRequest struct {
	Messages []ChatMessage `json:"messages"`
}

type ChatCompletionChoice struct {
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	Choices []ChatCompletionChoice `json:"choices"`
}

type MockOpenAIServer struct {
	Server *httptest.Server
	URL    string
}

// NewMockOpenAIServer creates a new mock OpenAI API server.
// It extracts the last word from the input as a name and returns "Hello! {name}. How can I help you.".
func NewMockOpenAIServer() *MockOpenAIServer {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle chat completion endpoint
		if r.URL.Path == "/v1/chat/completions" {
			var req ChatCompletionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if len(req.Messages) == 0 {
				http.Error(w, "no messages in request", http.StatusBadRequest)
				return
			}

			resp := ChatCompletionResponse{
				Choices: []ChatCompletionChoice{
					{
						Message: ChatMessage{
							Role:    "ai",
							Content: "hello",
						},
						FinishReason: "stop",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))

	return &MockOpenAIServer{
		Server: server,
		URL:    server.URL + "/v1",
	}
}

func (m *MockOpenAIServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}
