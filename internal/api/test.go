package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dmxapi-config/internal/config"
)

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice 选择结构
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// APIError API错误结构
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Tester API测试器
type Tester struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewTester 创建新的API测试器
func NewTester(baseURL, apiKey string) *Tester {
	return &Tester{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// AnthropicRequest Anthropic Messages API 请求结构
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// AnthropicResponse Anthropic Messages API 响应结构
type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *AnthropicError `json:"error,omitempty"`
}

// AnthropicError Anthropic API 错误结构
type AnthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// TestConnection 测试API连接
// 使用用户指定的 model 发送一个简单请求，验证 API Key 和 URL 是否有效
func (t *Tester) TestConnection(model string) error {
	if config.ClassifyModel(model) == config.ProviderAnthropic {
		return t.testAnthropicConnection(model)
	}
	return t.testOpenAIConnection(model)
}

// testAnthropicConnection 使用 Anthropic Messages API 测试连接
func (t *Tester) testAnthropicConnection(model string) error {
	req := AnthropicRequest{
		Model:     model,
		MaxTokens: 10,
		Messages:  []Message{{Role: "user", Content: "Hi"}},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := t.baseURL + "/v1/messages"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var anthResp AnthropicResponse
		if json.Unmarshal(body, &anthResp) == nil && anthResp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", resp.StatusCode, anthResp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var anthResp AnthropicResponse
	if err := json.Unmarshal(body, &anthResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if len(anthResp.Content) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何内容")
	}

	return nil
}

// testOpenAIConnection 使用 OpenAI Chat Completions API 测试连接
func (t *Tester) testOpenAIConnection(model string) error {
	req := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: "Hi"},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := t.baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var chatResp ChatResponse
		if json.Unmarshal(body, &chatResp) == nil && chatResp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", resp.StatusCode, chatResp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何内容")
	}

	return nil
}
