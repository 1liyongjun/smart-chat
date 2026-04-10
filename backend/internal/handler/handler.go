package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gofiber/fiber/v2"

	"smart-chat/internal/eino"
	"smart-chat/internal/store"
)

// Handler API处理器
type Handler struct {
	store      *store.Store
	chatModel  model.ChatModel
	apiKey     string
	baseURL    string
	modelName  string
}

// NewHandler 创建处理器
func NewHandler(s *store.Store, apiKey, baseURL, modelName string) (*Handler, error) {
	ctx := context.Background()

	chatModel, err := eino.NewMiniMaxChatModel(ctx, &eino.MiniMaxConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model failed: %w", err)
	}

	return &Handler{
		store:     s,
		chatModel: chatModel,
		apiKey:    apiKey,
		baseURL:   baseURL,
		modelName: modelName,
	}, nil
}

// QAItem 知识库条目请求
type QAItem struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message       string `json:"message"`
	VisitorID     string `json:"visitor_id"`
	ConversationID string `json:"conversation_id"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Answer         string   `json:"answer"`
	ConversationID string   `json:"conversation_id"`
	Sources        []string `json:"sources"`
}

// Health 健康检查
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "smart-chat"})
}

// === 知识库管理 ===

// AddKnowledge 添加知识库条目
func (h *Handler) AddKnowledge(c *fiber.Ctx) error {
	var item QAItem
	if err := c.BodyParser(&item); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if item.Question == "" || item.Answer == "" {
		return fiber.NewError(fiber.StatusBadRequest, "question and answer are required")
	}

	category := item.Category
	if category == "" {
		category = "general"
	}

	id, err := h.store.AddKnowledge(item.Question, item.Answer, category, item.Tags)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"id": id, "message": "添加成功"})
}

// ListKnowledge 获取知识库列表
func (h *Handler) ListKnowledge(c *fiber.Ctx) error {
	category := c.Query("category")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.store.ListKnowledge(category, page, pageSize)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateKnowledge 更新知识库条目
func (h *Handler) UpdateKnowledge(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	var item QAItem
	if err := c.BodyParser(&item); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if item.Question == "" || item.Answer == "" {
		return fiber.NewError(fiber.StatusBadRequest, "question and answer are required")
	}

	if err := h.store.UpdateKnowledge(id, item.Question, item.Answer, item.Category, item.Tags); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

// DeleteKnowledge 删除知识库条目
func (h *Handler) DeleteKnowledge(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	if err := h.store.DeleteKnowledge(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "删除成功"})
}

// === 聊天 ===

// Chat 聊天接口
func (h *Handler) Chat(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Message == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message is required")
	}

	visitorID := req.VisitorID
	if visitorID == "" {
		visitorID = "anonymous"
	}

	conversationID := req.ConversationID
	if conversationID == "" {
		conversationID = fmt.Sprintf("conv_%d", ctxHash(c))
	}

	// 保存用户消息
	h.store.SaveMessage(conversationID, visitorID, "user", req.Message)

	// 搜索知识库
	kbItems, err := h.store.SearchKnowledge(req.Message, 3)
	if err != nil {
		kbItems = nil
	}

	// 构建上下文和系统提示
	var context strings.Builder
	var sources []string

	if len(kbItems) > 0 {
		context.WriteString("知识库内容：\n")
		for i, item := range kbItems {
			context.WriteString(fmt.Sprintf("\n【相关问答 %d】\n问题: %s\n回答: %s\n", i+1, item.Question, item.Answer))
			sources = append(sources, fmt.Sprintf("Q: %s", item.Question))
		}
	}

	systemPrompt := fmt.Sprintf(`你是一个专业的智能客服助手。请根据提供的知识库内容回答用户问题。

%s

回答要求：
1. 如果知识库中有相关内容，优先使用知识库内容回答
2. 如果知识库中没有相关信息，礼貌地说明暂时无法回答，并引导用户联系人工客服
3. 回答要简洁、专业、友好
4. 不要编造知识库中没有的信息`, context.String())

	// 调用AI
	messages := []*schema.Message{
		{Role: schema.RoleSystem, Content: systemPrompt},
		{Role: schema.RoleUser, Content: req.Message},
	}

	resp, err := h.chatModel.Generate(c.Context(), messages)
	if err != nil {
		// 如果API调用失败，使用备用方案
		answer := h.fallbackAnswer(req.Message, kbItems)
		h.store.SaveMessage(conversationID, visitorID, "assistant", answer)
		return c.JSON(ChatResponse{
			Answer:         answer,
			ConversationID: conversationID,
			Sources:        sources,
		})
	}

	answer := resp.Content
	h.store.SaveMessage(conversationID, visitorID, "assistant", answer)

	return c.JSON(ChatResponse{
		Answer:         answer,
		ConversationID: conversationID,
		Sources:        sources,
	})
}

// fallbackAnswer 备用回答
func (h *Handler) fallbackAnswer(query string, kbItems []store.KnowledgeItem) string {
	if len(kbItems) > 0 {
		return kbItems[0].Answer
	}
	return "抱歉，AI服务暂时不可用，请稍后重试或联系人工客服。"
}

// ctxHash 生成简单的对话ID哈希
func ctxHash(c *fiber.Ctx) int64 {
	return int64(len(c.Path()))
}

// GetConversation 获取对话历史
func (h *Handler) GetConversation(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	conv, err := h.store.GetConversation(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "对话不存在")
	}

	return c.JSON(conv)
}

// ListConversations 获取对话列表
func (h *Handler) ListConversations(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))

	if page < 1 {
		page = 1
	}

	convs, total, err := h.store.ListConversations(page, pageSize)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"items": convs,
		"total": total,
	})
}

// GetStats 获取统计数据
func (h *Handler) GetStats(c *fiber.Ctx) error {
	stats, err := h.store.GetStats()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(stats)
}

// === 流式聊天 ===

// StreamChat 流式聊天
func (h *Handler) StreamChat(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Message == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message is required")
	}

	visitorID := req.VisitorID
	if visitorID == "" {
		visitorID = "anonymous"
	}

	conversationID := req.ConversationID
	if conversationID == "" {
		conversationID = fmt.Sprintf("conv_%d", len(req.Message))
	}

	// 保存用户消息
	h.store.SaveMessage(conversationID, visitorID, "user", req.Message)

	// 搜索知识库
	kbItems, _ := h.store.SearchKnowledge(req.Message, 3)

	var context strings.Builder
	var sources []string

	if len(kbItems) > 0 {
		context.WriteString("知识库内容：\n")
		for i, item := range kbItems {
			context.WriteString(fmt.Sprintf("\n【相关问答 %d】\n问题: %s\n回答: %s\n", i+1, item.Question, item.Answer))
			sources = append(sources, fmt.Sprintf("Q: %s", item.Question))
		}
	}

	systemPrompt := fmt.Sprintf(`你是一个专业的智能客服助手。请根据提供的知识库内容回答用户问题。

%s

回答要求：
1. 如果知识库中有相关内容，优先使用知识库内容回答
2. 如果知识库中没有相关信息，礼貌地说明暂时无法回答
3. 回答要简洁、专业、友好`, context.String())

	messages := []*schema.Message{
		{Role: schema.RoleSystem, Content: systemPrompt},
		{Role: schema.RoleUser, Content: req.Message},
	}

	stream, err := h.chatModel.Stream(c.Context(), messages)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer stream.Close()

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bytes.Buffer) error {
		var fullAnswer strings.Builder

		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			fullAnswer.WriteString(msg.Content)

			// 发送SSE格式
			data, _ := json.Marshal(fiber.Map{
				"content": msg.Content,
			})
			w.Write([]byte("data: " + string(data) + "\n\n"))
		}

		// 保存完整回复
		go h.store.SaveMessage(conversationID, visitorID, "assistant", fullAnswer.String())

		w.Write([]byte("data: [DONE]\n\n"))
		return nil
	})

	return nil
}

// ProxyChat 代理到MiniMax API（用于开发调试）
func (h *Handler) ProxyChat(c *fiber.Ctx) error {
	if h.apiKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "API key not configured")
	}

	body := c.Body()
	url := h.baseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(c.Context(), "POST", url, bytes.NewReader(body))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	c.Set("Content-Type", "application/json")
	return c.Send(body)
}
