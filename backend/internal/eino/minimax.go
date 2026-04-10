package eino

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MiniMaxConfig MiniMax模型配置
type MiniMaxConfig struct {
	APIKey string // MiniMax API Key
	Model  string // 模型名称，默认 MiniMax-M2.7
	BaseURL string // API地址
}

// minimaxMessage MiniMax API消息格式
type minimaxMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// minimaxRequest MiniMax请求格式
type minimaxRequest struct {
	Model       string             `json:"model"`
	Messages    []minimaxMessage  `json:"messages"`
	Temperature float64            `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
}

// minimaxResponse MiniMax响应格式
type minimaxResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      minimaxMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// streamChunk 流式响应块
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// MiniMaxChatModel MiniMax聊天模型实现
type MiniMaxChatModel struct {
	config *MiniMaxConfig
	client *http.Client
}

var _ model.ChatModel = (*MiniMaxChatModel)(nil)

// NewMiniMaxChatModel 创建MiniMax聊天模型
func NewMiniMaxChatModel(ctx context.Context, config *MiniMaxConfig) (*MiniMaxChatModel, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.minimax.io/v1"
	}
	if config.Model == "" {
		config.Model = "MiniMax-M2.7"
	}

	return &MiniMaxChatModel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Generate 生成回复
func (m *MiniMaxChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	options := model.GetOptions(opts)

	req := &minimaxRequest{
		Model:       m.config.Model,
		Messages:    convertMessages(messages),
		Temperature: float64(options.Temperature),
		MaxTokens:   int(options.MaxTokens),
	}

	if options.Temperature == 0 {
		req.Temperature = 0.7
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.config.APIKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result minimaxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return &schema.Message{
		Role:    schema.RoleAssistant,
		Content: result.Choices[0].Message.Content,
	}, nil
}

// Stream 流式生成回复
func (m *MiniMaxChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	options := model.GetOptions(opts)

	req := &minimaxRequest{
		Model:       m.config.Model,
		Messages:    convertMessages(messages),
		Temperature: float64(options.Temperature),
		MaxTokens:   int(options.MaxTokens),
		Stream:     true,
	}

	if options.Temperature == 0 {
		req.Temperature = 0.7
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.config.APIKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	stream := schema.NewStreamReader[*schema.Message](func() (*schema.Message, error) {
		reader := resp.Body
		defer reader.Close()

		// 逐行读取SSE流
		buf := make([]byte, 4096)
		line := ""

		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					return nil, nil
				}
				return nil, err
			}

			line += string(buf[:n])

			// 查找完整行
			lines := strings.Split(line, "\n")
			for i := 0; i < len(lines)-1; i++ {
				l := strings.TrimSpace(lines[i])
				if strings.HasPrefix(l, "data: ") {
					data := l[6:]
					if data == "[DONE]" {
						return nil, nil
					}

					var chunk streamChunk
					if err := json.Unmarshal([]byte(data), &chunk); err != nil {
						continue
					}

					if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
						return &schema.Message{
							Role:    schema.RoleAssistant,
							Content: chunk.Choices[0].Delta.Content,
						}, nil
					}
				}
			}
			line = lines[len(lines)-1]
		}
	})

	return stream, nil
}

// convertMessages 转换消息格式
func convertMessages(in []*schema.Message) []minimaxMessage {
	out := make([]minimaxMessage, 0, len(in))
	for _, m := range in {
		role := "user"
		if m.Role == schema.RoleAssistant {
			role = "assistant"
		} else if m.Role == schema.RoleSystem {
			role = "system"
		}
		out = append(out, minimaxMessage{
			Role:    role,
			Content: m.Content,
		})
	}
	return out
}
