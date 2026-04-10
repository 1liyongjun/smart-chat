package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"

	"smart-chat/internal/store"
)

// mockChatModel 实现了 model.ChatModel 接口
type mockChatModel struct {
	generateResp *mockMessage
	generateErr  error
	streamResp   interface{}
	streamErr    error
}

type mockMessage struct {
	Content string
	Role    string
}

func (m *mockChatModel) Generate(ctx context.Context, messages []*struct {
	Role    string
	Content string
}, opts ...interface{}) (msg *struct {
	Content string
	Role    string
}, err error) {
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return m.generateResp, nil
}

func (m *mockChatModel) Stream(ctx context.Context, messages []*struct {
	Role    string
	Content string
}, opts ...interface{}) (interface{}, error) {
	if m.streamErr != nil {
		return nil, m.streamErr
	}
	return m.streamResp, nil
}

func TestHealth(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_health_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	// health 不需要 chatModel
	h := &Handler{
		store:     s,
		chatModel: nil,
	}

	app.Get("/api/health", h.Health)

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["status"] != "ok" {
		t.Errorf("Expected status=ok, got %v", result["status"])
	}
	if result["service"] != "smart-chat" {
		t.Errorf("Expected service=smart-chat, got %v", result["service"])
	}
}

func TestAddKnowledgeHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_add_kb_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	h := &Handler{store: s}
	app.Post("/api/knowledge", h.AddKnowledge)

	t.Run("ValidRequest", func(t *testing.T) {
		body := `{"question": "你好", "answer": "你好！"}`
		req := httptest.NewRequest("POST", "/api/knowledge", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		if result["message"] != "添加成功" {
			t.Errorf("Expected message='添加成功', got %v", result["message"])
		}
	})

	t.Run("MissingQuestion", func(t *testing.T) {
		body := `{"answer": "只有回答"}`
		req := httptest.NewRequest("POST", "/api/knowledge", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := `{invalid json}`
		req := httptest.NewRequest("POST", "/api/knowledge", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestListKnowledgeHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_list_kb_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	// 先添加一些数据
	s.AddKnowledge("问题1", "回答1", "tech", []string{})
	s.AddKnowledge("问题2", "回答2", "tech", []string{})
	s.AddKnowledge("问题3", "回答3", "general", []string{})

	h := &Handler{store: s}
	app.Get("/api/knowledge", h.ListKnowledge)

	t.Run("ListAll", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/knowledge", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		if int(result["total"].(float64)) != 3 {
			t.Errorf("Expected total=3, got %v", result["total"])
		}
		items := result["items"].([]interface{})
		if len(items) != 3 {
			t.Errorf("Expected 3 items, got %d", len(items))
		}
	})

	t.Run("FilterByCategory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/knowledge?category=tech", nil)
		resp, _ := app.Test(req)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		if int(result["total"].(float64)) != 2 {
			t.Errorf("Expected tech total=2, got %v", result["total"])
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/knowledge?page=1&page_size=2", nil)
		resp, _ := app.Test(req)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		items := result["items"].([]interface{})
		if len(items) != 2 {
			t.Errorf("Expected page_size=2, got %d", len(items))
		}
		if int(result["page"].(float64)) != 1 {
			t.Errorf("Expected page=1, got %v", result["page"])
		}
	})
}

func TestUpdateKnowledgeHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_update_kb_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	// 先添加一条数据
	id, _ := s.AddKnowledge("原始问题", "原始回答", "test", []string{})

	h := &Handler{store: s}
	app.Put("/api/knowledge/:id", h.UpdateKnowledge)

	t.Run("ValidUpdate", func(t *testing.T) {
		body := `{"question": "更新后问题", "answer": "更新后回答"}`
		req := httptest.NewRequest("PUT", "/api/knowledge/"+formatInt64(id), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// 验证更新成功
		items, _, _ := s.ListKnowledge("", 1, 10)
		if items[0].Question != "更新后问题" {
			t.Errorf("Question not updated: got '%s'", items[0].Question)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		body := `{"question": "q", "answer": "a"}`
		req := httptest.NewRequest("PUT", "/api/knowledge/invalid", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestDeleteKnowledgeHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_del_kb_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	// 先添加一条数据
	id, _ := s.AddKnowledge("待删除", "将被删除", "test", []string{})

	h := &Handler{store: s}
	app.Delete("/api/knowledge/:id", h.DeleteKnowledge)

	t.Run("ValidDelete", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/knowledge/"+formatInt64(id), nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// 验证已删除
		_, total, _ := s.ListKnowledge("", 1, 10)
		if total != 0 {
			t.Errorf("Expected 0 items after delete, got %d", total)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/knowledge/invalid", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestGetConversationHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_get_conv_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	// 先创建对话
	convID := "test_conv_001"
	s.SaveMessage(convID, "visitor1", "user", "你好")

	h := &Handler{store: s}
	app.Get("/api/conversations/:id", h.GetConversation)

	t.Run("ValidGet", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations/"+convID, nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		if result["visitor_id"] != "visitor1" {
			t.Errorf("Expected visitor_id=visitor1, got %v", result["visitor_id"])
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations/nonexistent", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
}

func TestGetStatsHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_stats_handler_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	s.AddKnowledge("Q1", "A1", "test", []string{})
	s.SaveMessage("c1", "v1", "user", "msg1")

	h := &Handler{store: s}
	app.Get("/api/stats", h.GetStats)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var stats store.Stats
	json.Unmarshal(respBody, &stats)

	if stats.KnowledgeCount != 1 {
		t.Errorf("Expected knowledge_count=1, got %d", stats.KnowledgeCount)
	}
	if stats.ConversationCount != 1 {
		t.Errorf("Expected conversation_count=1, got %d", stats.ConversationCount)
	}
}

func TestListConversationsHandler(t *testing.T) {
	app := fiber.New()

	tmpFile := "/tmp/test_list_conv_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, _ := store.NewStore(tmpFile)
	defer s.Close()

	s.SaveMessage("conv_a", "v1", "user", "A")
	s.SaveMessage("conv_b", "v2", "user", "B")

	h := &Handler{store: s}
	app.Get("/api/conversations", h.ListConversations)

	t.Run("ListAll", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		if int(result["total"].(float64)) != 2 {
			t.Errorf("Expected total=2, got %v", result["total"])
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations?page=1&page_size=1", nil)
		resp, _ := app.Test(req)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		items := result["items"].([]interface{})
		if len(items) != 1 {
			t.Errorf("Expected 1 item per page, got %d", len(items))
		}
	})
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

func formatInt64(n int64) string {
	return string(rune('0'+n%10)) // 简单转换，适用于个位数
}
