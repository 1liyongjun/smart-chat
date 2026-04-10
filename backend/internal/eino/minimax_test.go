package eino

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func mockServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestNewMiniMaxChatModel(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		ctx := context.Background()
		model, err := NewMiniMaxChatModel(ctx, &MiniMaxConfig{
			APIKey: "test-key",
		})
		if err != nil {
			t.Fatalf("NewMiniMaxChatModel failed: %v", err)
		}
		if model == nil {
			t.Fatal("Model should not be nil")
		}
		if model.config.BaseURL != "https://api.minimax.io/v1" {
			t.Errorf("Expected default BaseURL, got '%s'", model.config.BaseURL)
		}
		if model.config.Model != "MiniMax-M2.7" {
			t.Errorf("Expected default Model=MiniMax-M2.7, got '%s'", model.config.Model)
		}
	})

	t.Run("CustomValues", func(t *testing.T) {
		ctx := context.Background()
		model, err := NewMiniMaxChatModel(ctx, &MiniMaxConfig{
			APIKey:  "custom-key",
			BaseURL: "https://custom.api.com/v1",
			Model:   "custom-model",
		})
		if err != nil {
			t.Fatalf("NewMiniMaxChatModel failed: %v", err)
		}
		if model.config.BaseURL != "https://custom.api.com/v1" {
			t.Errorf("Expected BaseURL='https://custom.api.com/v1', got '%s'", model.config.BaseURL)
		}
		if model.config.Model != "custom-model" {
			t.Errorf("Expected Model='custom-model', got '%s'", model.config.Model)
		}
	})
}

func TestMiniMaxChatModelGenerate(t *testing.T) {
	t.Run("SuccessfulResponse", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求
			if r.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Errorf("Expected Authorization header with Bearer token")
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type: application/json")
			}

			// 解析请求体，验证
			var req minimaxRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Model != "MiniMax-M2.7" {
				t.Errorf("Expected model=MiniMax-M2.7, got '%s'", req.Model)
			}
			if len(req.Messages) != 1 {
				t.Errorf("Expected 1 message, got %d", len(req.Messages))
			}

			// 返回模拟响应
			resp := minimaxResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1234567890,
				Model:   "MiniMax-M2.7",
				Choices: []struct {
					Index        int              `json:"index"`
					Message      minimaxMessage  `json:"message"`
					FinishReason string           `json:"finish_reason"`
				}{
					{
						Index:        0,
						Message:      minimaxMessage{Role: "assistant", Content: "你好！有什么可以帮助你的？"},
						FinishReason: "stop",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		ctx := context.Background()
		model, _ := NewMiniMaxChatModel(ctx, &MiniMaxConfig{
			APIKey:  "test-api-key",
			BaseURL: server.URL,
			Model:   "MiniMax-M2.7",
		})

		messages := []*schema.Message{
			{Role: schema.RoleUser, Content: "你好"},
		}

		result, err := model.Generate(ctx, messages)

		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result should not be nil")
		}
		if result.Content != "你好！有什么可以帮助你的？" {
			t.Errorf("Unexpected content: %s", result.Content)
		}
		if result.Role != schema.RoleAssistant {
			t.Errorf("Expected role=assistant, got '%s'", result.Role)
		}
	})

	t.Run("APIError", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
		})
		defer server.Close()

		ctx := context.Background()
		model, _ := NewMiniMaxChatModel(ctx, &MiniMaxConfig{
			APIKey:  "invalid-key",
			BaseURL: server.URL,
		})

		messages := []*schema.Message{
			{Role: schema.RoleUser, Content: "hi"},
		}
		_, err := model.Generate(ctx, messages)

		if err == nil {
			t.Fatal("Expected error for invalid API key")
		}
		if !strings.Contains(err.Error(), "API error") {
			t.Errorf("Expected API error message, got: %v", err)
		}
	})

	t.Run("EmptyResponse", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			resp := minimaxResponse{
				ID:      "chatcmpl-123",
				Choices: []struct {
					Index        int             `json:"index"`
					Message      minimaxMessage `json:"message"`
					FinishReason string          `json:"finish_reason"`
				}{}, // 空 choices
			}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		ctx := context.Background()
		model, _ := NewMiniMaxChatModel(ctx, &MiniMaxConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		messages := []*schema.Message{
			{Role: schema.RoleUser, Content: "hi"},
		}
		_, err := model.Generate(ctx, messages)

		if err == nil {
			t.Fatal("Expected error for empty response")
		}
		if !strings.Contains(err.Error(), "empty response") {
			t.Errorf("Expected 'empty response' error, got: %v", err)
		}
	})

	t.Run("ServerError", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Internal server error"}`))
		})
		defer server.Close()

		ctx := context.Background()
		model, _ := NewMiniMaxChatModel(ctx, &MiniMaxConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		messages := []*schema.Message{
			{Role: schema.RoleUser, Content: "hi"},
		}
		_, err := model.Generate(ctx, messages)

		if err == nil {
			t.Fatal("Expected error for server error")
		}
		if !strings.Contains(err.Error(), "API error: status=500") {
			t.Errorf("Expected status=500 error, got: %v", err)
		}
	})
}

func TestConvertMessages(t *testing.T) {
	t.Run("UserMessage", func(t *testing.T) {
		input := []*schema.Message{
			{Role: schema.RoleUser, Content: "Hello"},
		}
		result := convertMessages(input)

		if len(result) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(result))
		}
		if result[0].Role != "user" {
			t.Errorf("Expected role='user', got '%s'", result[0].Role)
		}
		if result[0].Content != "Hello" {
			t.Errorf("Expected content='Hello', got '%s'", result[0].Content)
		}
	})

	t.Run("SystemMessage", func(t *testing.T) {
		input := []*schema.Message{
			{Role: schema.RoleSystem, Content: "You are a helpful assistant."},
		}
		result := convertMessages(input)

		if result[0].Role != "system" {
			t.Errorf("Expected role='system', got '%s'", result[0].Role)
		}
	})

	t.Run("AssistantMessage", func(t *testing.T) {
		input := []*schema.Message{
			{Role: schema.RoleAssistant, Content: "I can help you."},
		}
		result := convertMessages(input)

		if result[0].Role != "assistant" {
			t.Errorf("Expected role='assistant', got '%s'", result[0].Role)
		}
	})

	t.Run("MultipleMessages", func(t *testing.T) {
		input := []*schema.Message{
			{Role: schema.RoleSystem, Content: "System prompt"},
			{Role: schema.RoleUser, Content: "User message"},
			{Role: schema.RoleAssistant, Content: "Assistant response"},
		}
		result := convertMessages(input)

		if len(result) != 3 {
			t.Fatalf("Expected 3 messages, got %d", len(result))
		}
		roles := []string{"system", "user", "assistant"}
		for i, expected := range roles {
			if result[i].Role != expected {
				t.Errorf("Message %d: expected role='%s', got '%s'", i, expected, result[i].Role)
			}
		}
	})
}
