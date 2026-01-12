package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

// TestConnection 测试API连接
// 使用 claude-opus-4-5-20251101 模型发送一个简单请求
func (t *Tester) TestConnection() error {
	// 构造请求
	req := ChatRequest{
		Model: "claude-opus-4-5-20251101",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hi",
			},
		},
	}

	// 序列化请求
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := t.baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)

	// 发送请求
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误信息
		var chatResp ChatResponse
		if json.Unmarshal(body, &chatResp) == nil && chatResp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", resp.StatusCode, chatResp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否有选择
	if len(chatResp.Choices) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何内容")
	}

	return nil
}
