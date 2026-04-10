package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store 数据存储
type Store struct {
	db   *sql.DB
	mu   sync.RWMutex
}

// KnowledgeItem 知识库条目
type KnowledgeItem struct {
	ID         int64     `json:"id"`
	Question   string    `json:"question"`
	Answer     string    `json:"answer"`
	Category   string    `json:"category"`
	Tags       string    `json:"tags"`
	UsageCount int       `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Conversation 对话记录
type Conversation struct {
	ID         string    `json:"id"`
	VisitorID  string    `json:"visitor_id"`
	Messages   string    `json:"messages"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Message 消息
type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// NewStore 创建存储
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		return nil, fmt.Errorf("init db failed: %w", err)
	}
	return s, nil
}

func (s *Store) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS knowledge_base (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			question TEXT NOT NULL,
			answer TEXT NOT NULL,
			category TEXT DEFAULT 'general',
			tags TEXT DEFAULT '[]',
			usage_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			visitor_id TEXT NOT NULL,
			messages TEXT DEFAULT '[]',
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS visitors (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			metadata TEXT DEFAULT '{}',
			first_visit TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_visit TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_kb_question ON knowledge_base(question)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_visitor ON conversations(visitor_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// === 知识库操作 ===

// AddKnowledge 添加知识库条目
func (s *Store) AddKnowledge(question, answer, category string, tags []string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tagsJSON, _ := json.Marshal(tags)
	result, err := s.db.Exec(
		"INSERT INTO knowledge_base (question, answer, category, tags) VALUES (?, ?, ?, ?)",
		question, answer, category, string(tagsJSON),
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// ListKnowledge 获取知识库列表
func (s *Store) ListKnowledge(category string, page, pageSize int) ([]KnowledgeItem, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rows *sql.Rows
	var err error

	offset := (page - 1) * pageSize

	if category != "" {
		rows, err = s.db.Query(
			"SELECT id, question, answer, category, tags, usage_count, created_at, updated_at FROM knowledge_base WHERE category = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?",
			category, pageSize, offset,
		)
	} else {
		rows, err = s.db.Query(
			"SELECT id, question, answer, category, tags, usage_count, created_at, updated_at FROM knowledge_base ORDER BY updated_at DESC LIMIT ? OFFSET ?",
			pageSize, offset,
		)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []KnowledgeItem
	for rows.Next() {
		var item KnowledgeItem
		if err := rows.Scan(&item.ID, &item.Question, &item.Answer, &item.Category, &item.Tags, &item.UsageCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	// 获取总数
	var total int
	s.db.QueryRow("SELECT COUNT(*) FROM knowledge_base").Scan(&total)

	return items, total, nil
}

// UpdateKnowledge 更新知识库条目
func (s *Store) UpdateKnowledge(id int64, question, answer, category string, tags []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tagsJSON, _ := json.Marshal(tags)
	_, err := s.db.Exec(
		"UPDATE knowledge_base SET question = ?, answer = ?, category = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		question, answer, category, string(tagsJSON), id,
	)
	return err
}

// DeleteKnowledge 删除知识库条目
func (s *Store) DeleteKnowledge(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM knowledge_base WHERE id = ?", id)
	return err
}

// SearchKnowledge 搜索知识库
func (s *Store) SearchKnowledge(query string, topK int) ([]KnowledgeItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pattern := "%" + query + "%"
	rows, err := s.db.Query(
		"SELECT id, question, answer, category, tags, usage_count, created_at, updated_at FROM knowledge_base WHERE question LIKE ? OR answer LIKE ? ORDER BY usage_count DESC LIMIT ?",
		pattern, pattern, topK,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []KnowledgeItem
	for rows.Next() {
		var item KnowledgeItem
		if err := rows.Scan(&item.ID, &item.Question, &item.Answer, &item.Category, &item.Tags, &item.UsageCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// 更新使用次数
	for _, item := range items {
		s.db.Exec("UPDATE knowledge_base SET usage_count = usage_count + 1 WHERE id = ?", item.ID)
	}

	return items, nil
}

// === 对话操作 ===

// SaveMessage 保存消息
func (s *Store) SaveMessage(conversationID, visitorID string, role, content string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取现有消息
	var messages string
	var existingID string
	err := s.db.QueryRow("SELECT id, messages FROM conversations WHERE id = ?", conversationID).Scan(&existingID, &messages)
	if err == sql.ErrNoRows {
		// 创建新对话
		messages = "[]"
	} else if err != nil {
		return "", err
	}

	var msgList []Message
	json.Unmarshal([]byte(messages), &msgList)

	msgList = append(msgList, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
	})

	newMessages, _ := json.Marshal(msgList)

	if existingID == "" {
		_, err = s.db.Exec(
			"INSERT INTO conversations (id, visitor_id, messages) VALUES (?, ?, ?)",
			conversationID, visitorID, string(newMessages),
		)
	} else {
		_, err = s.db.Exec(
			"UPDATE conversations SET messages = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			string(newMessages), conversationID,
		)
	}

	return conversationID, err
}

// GetConversation 获取对话
func (s *Store) GetConversation(id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var conv Conversation
	err := s.db.QueryRow(
		"SELECT id, visitor_id, messages, status, created_at, updated_at FROM conversations WHERE id = ?",
		id,
	).Scan(&conv.ID, &conv.VisitorID, &conv.Messages, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListConversations 获取对话列表
func (s *Store) ListConversations(page, pageSize int) ([]Conversation, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	offset := (page - 1) * pageSize
	rows, err := s.db.Query(
		"SELECT id, visitor_id, messages, status, created_at, updated_at FROM conversations ORDER BY updated_at DESC LIMIT ? OFFSET ?",
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.VisitorID, &conv.Messages, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, 0, err
		}
		convs = append(convs, conv)
	}

	var total int
	s.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&total)

	return convs, total, nil
}

// === 统计 ===

// Stats 统计数据
type Stats struct {
	KnowledgeCount    int `json:"knowledge_count"`
	ConversationCount int `json:"conversation_count"`
	VisitorCount      int `json:"visitor_count"`
}

// GetStats 获取统计数据
func (s *Store) GetStats() (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var stats Stats
	s.db.QueryRow("SELECT COUNT(*) FROM knowledge_base").Scan(&stats.KnowledgeCount)
	s.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&stats.ConversationCount)
	s.db.QueryRow("SELECT COUNT(DISTINCT visitor_id) FROM conversations").Scan(&stats.VisitorCount)

	return &stats, nil
}
