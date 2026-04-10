package store

import (
	"os"
	"testing"
)

func TestNewStore(t *testing.T) {
	// 创建临时数据库
	tmpFile := "/tmp/test_smart_chat_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	// 验证表已创建
	t.Run("TablesCreated", func(t *testing.T) {
		rows, err := s.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
		if err != nil {
			t.Fatalf("Query tables failed: %v", err)
		}
		defer rows.Close()

		tables := make(map[string]bool)
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables[name] = true
		}

		expected := []string{"knowledge_base", "conversations", "visitors"}
		for _, table := range expected {
			if !tables[table] {
				t.Errorf("Table %s not created", table)
			}
		}
	})
}

func TestAddKnowledge(t *testing.T) {
	tmpFile := "/tmp/test_knowledge_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	t.Run("AddAndRetrieve", func(t *testing.T) {
		id, err := s.AddKnowledge("你好", "你好，请问有什么可以帮您？", "greeting", []string{"问候"})
		if err != nil {
			t.Fatalf("AddKnowledge failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("Expected positive id, got %d", id)
		}

		// 验证可以查询到
		items, total, err := s.ListKnowledge("", 1, 10)
		if err != nil {
			t.Fatalf("ListKnowledge failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total=1, got %d", total)
		}
		if len(items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(items))
		}
		if items[0].Question != "你好" {
			t.Errorf("Expected question='你好', got '%s'", items[0].Question)
		}
	})

	t.Run("AddMultiple", func(t *testing.T) {
		s.AddKnowledge("问题1", "回答1", "tech", []string{"标签1"})
		s.AddKnowledge("问题2", "回答2", "tech", []string{"标签2"})
		s.AddKnowledge("问题3", "回答3", "general", []string{"标签3"})

		// 按分类查询
		techItems, total, err := s.ListKnowledge("tech", 1, 10)
		if err != nil {
			t.Fatalf("ListKnowledge failed: %v", err)
		}
		if total != 2 {
			t.Errorf("Expected tech total=2, got %d", total)
		}
		if len(techItems) != 2 {
			t.Errorf("Expected 2 tech items, got %d", len(techItems))
		}
	})
}

func TestUpdateKnowledge(t *testing.T) {
	tmpFile := "/tmp/test_update_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	id, _ := s.AddKnowledge("原始问题", "原始回答", "test", []string{})

	err = s.UpdateKnowledge(id, "更新后问题", "更新后回答", "updated", []string{"new"})
	if err != nil {
		t.Fatalf("UpdateKnowledge failed: %v", err)
	}

	items, _, _ := s.ListKnowledge("", 1, 10)
	if items[0].Question != "更新后问题" {
		t.Errorf("Question not updated: got '%s'", items[0].Question)
	}
	if items[0].Category != "updated" {
		t.Errorf("Category not updated: got '%s'", items[0].Category)
	}
}

func TestDeleteKnowledge(t *testing.T) {
	tmpFile := "/tmp/test_delete_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	_, _ := s.AddKnowledge("保留", "保留回答", "test", []string{})
	id2, _ := s.AddKnowledge("删除", "删除回答", "test", []string{})

	err = s.DeleteKnowledge(id2)
	if err != nil {
		t.Fatalf("DeleteKnowledge failed: %v", err)
	}

	items, total, _ := s.ListKnowledge("", 1, 10)
	if total != 1 {
		t.Errorf("Expected total=1 after delete, got %d", total)
	}
	if items[0].Question != "保留" {
		t.Errorf("Wrong item deleted: got '%s'", items[0].Question)
	}
}

func TestSearchKnowledge(t *testing.T) {
	tmpFile := "/tmp/test_search_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	s.AddKnowledge("如何重置密码？", "访问设置页面，点击重置密码。", "help", []string{"密码"})
	s.AddKnowledge("如何修改邮箱？", "在账户设置中修改邮箱地址。", "help", []string{"邮箱"})
	s.AddKnowledge("天气不错", "是的，今天天气很好。", "chat", []string{})

	t.Run("SearchByQuestion", func(t *testing.T) {
		results, err := s.SearchKnowledge("密码", 5)
		if err != nil {
			t.Fatalf("SearchKnowledge failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for '密码', got %d", len(results))
		}
		if len(results) > 0 && results[0].Question != "如何重置密码？" {
			t.Errorf("Wrong result: %s", results[0].Question)
		}
	})

	t.Run("SearchByAnswer", func(t *testing.T) {
		results, err := s.SearchKnowledge("账户设置", 5)
		if err != nil {
			t.Fatalf("SearchKnowledge failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for '账户设置', got %d", len(results))
		}
	})

	t.Run("NoResults", func(t *testing.T) {
		results, err := s.SearchKnowledge("不存在的查询", 5)
		if err != nil {
			t.Fatalf("SearchKnowledge failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("UsageCountIncrements", func(t *testing.T) {
		// 第一次搜索
		s.SearchKnowledge("密码", 5)
		items, _, _ := s.ListKnowledge("", 1, 10)
		count1 := items[0].UsageCount

		// 第二次搜索
		s.SearchKnowledge("密码", 5)
		items, _, _ = s.ListKnowledge("", 1, 10)
		count2 := items[0].UsageCount

		if count2 <= count1 {
			t.Errorf("UsageCount should increment: %d -> %d", count1, count2)
		}
	})
}

func TestConversations(t *testing.T) {
	tmpFile := "/tmp/test_conv_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	t.Run("SaveAndGetMessage", func(t *testing.T) {
		convID := "test_conv_001"
		s.SaveMessage(convID, "visitor1", "user", "你好")
		s.SaveMessage(convID, "visitor1", "assistant", "你好，请问有什么帮助？")

		conv, err := s.GetConversation(convID)
		if err != nil {
			t.Fatalf("GetConversation failed: %v", err)
		}
		if conv.VisitorID != "visitor1" {
			t.Errorf("Expected visitor_id=visitor1, got '%s'", conv.VisitorID)
		}
	})

	t.Run("NewConversation", func(t *testing.T) {
		convID := "new_conv_002"
		_, err := s.SaveMessage(convID, "visitor2", "user", "新对话")
		if err != nil {
			t.Fatalf("SaveMessage for new conv failed: %v", err)
		}

		conv, _ := s.GetConversation(convID)
		if conv.ID != convID {
			t.Errorf("Expected conv id=%s, got '%s'", convID, conv.ID)
		}
	})

	t.Run("ListConversations", func(t *testing.T) {
		s.SaveMessage("conv_a", "v1", "user", "A")
		s.SaveMessage("conv_b", "v2", "user", "B")

		convs, total, err := s.ListConversations(1, 10)
		if err != nil {
			t.Fatalf("ListConversations failed: %v", err)
		}
		if total < 2 {
			t.Errorf("Expected at least 2 conversations, got %d", total)
		}
		_ = convs // 验证分页
	})
}

func TestGetStats(t *testing.T) {
	tmpFile := "/tmp/test_stats_" + randomString(8) + ".db"
	defer os.Remove(tmpFile)

	s, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	// 添加一些数据
	s.AddKnowledge("Q1", "A1", "test", []string{})
	s.AddKnowledge("Q2", "A2", "test", []string{})
	s.SaveMessage("c1", "v1", "user", "msg1")
	s.SaveMessage("c2", "v2", "user", "msg2")

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.KnowledgeCount != 2 {
		t.Errorf("Expected knowledge_count=2, got %d", stats.KnowledgeCount)
	}
	if stats.ConversationCount != 2 {
		t.Errorf("Expected conversation_count=2, got %d", stats.ConversationCount)
	}
	if stats.VisitorCount != 2 {
		t.Errorf("Expected visitor_count=2, got %d", stats.VisitorCount)
	}
}

// randomString 生成随机字符串（用于临时文件）
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
