package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
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
	// 规范化 baseURL：去掉末尾的版本路径后缀（/v1、/v1beta 等），避免后续拼接路径时产生重复
	baseURL = config.NormalizeBaseURL(baseURL)
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

// GeminiRequest Google Generative Language API 请求结构
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent Gemini 内容结构
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart Gemini 消息片段
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiCandidate Gemini 候选结果
type GeminiCandidate struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

// GeminiResponse Google Generative Language API 响应结构
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// OpenAIResponsesRequest OpenAI Responses API 请求结构
type OpenAIResponsesRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// OpenAIResponsesResponse OpenAI Responses API 响应结构
type OpenAIResponsesResponse struct {
	Output []struct {
		Type string `json:"type"`
	} `json:"output"`
	Error *APIError `json:"error,omitempty"`
}

// TestConnection 测试API连接
// 使用用户指定的 model 发送一个简单请求，验证 API Key 和 URL 是否有效
func (t *Tester) TestConnection(model string) error {
	switch config.ClassifyModel(model) {
	case config.ProviderAnthropic:
		return t.testAnthropicConnection(model)
	case config.ProviderGoogle:
		return t.testGoogleConnection(model)
	case config.ProviderOpenAIResponses:
		return t.testOpenAIResponsesConnection(model)
	default:
		return t.testOpenAIConnection(model)
	}
}

// doRequest 发送 POST JSON 请求，返回状态码与响应体（限制 1MiB）。
// 请求体序列化、Bearer 认证、超时、响应读取的公共逻辑集中在此。
func (t *Tester) doRequest(reqURL string, body any) (int, []byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return 0, nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return 0, nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("读取响应失败: %w", err)
	}
	return resp.StatusCode, raw, nil
}

// testAnthropicConnection 使用 Anthropic Messages API 测试连接
func (t *Tester) testAnthropicConnection(model string) error {
	status, body, err := t.doRequest(t.baseURL+"/v1/messages", AnthropicRequest{
		Model:     model,
		MaxTokens: 10,
		Messages:  []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		var resp AnthropicResponse
		if json.Unmarshal(body, &resp) == nil && resp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", status, resp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", status, string(body))
	}

	var resp AnthropicResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if len(resp.Content) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何内容")
	}
	return nil
}

// testGoogleConnection 使用 Google Generative Language API 测试连接
// URL 格式与 opencode 配置中 @ai-sdk/google baseURL(url+"/v1beta") 一致：
// {baseURL}/v1beta/models/{model}:generateContent
func (t *Tester) testGoogleConnection(model string) error {
	reqURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", t.baseURL, url.PathEscape(model))
	status, body, err := t.doRequest(reqURL, GeminiRequest{
		Contents: []GeminiContent{{Parts: []GeminiPart{{Text: "Hi"}}}},
	})
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		var resp GeminiResponse
		if json.Unmarshal(body, &resp) == nil && resp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", status, resp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", status, string(body))
	}

	var resp GeminiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何 candidates")
	}
	return nil
}

// testOpenAIResponsesConnection 使用 OpenAI Responses API 测试连接
func (t *Tester) testOpenAIResponsesConnection(model string) error {
	status, body, err := t.doRequest(t.baseURL+"/v1/responses", OpenAIResponsesRequest{
		Model: model,
		Input: "Hi",
	})
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		var resp OpenAIResponsesResponse
		if json.Unmarshal(body, &resp) == nil && resp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", status, resp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", status, string(body))
	}

	var resp OpenAIResponsesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if len(resp.Output) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何输出")
	}
	return nil
}

// testOpenAIConnection 使用 OpenAI Chat Completions API 测试连接
func (t *Tester) testOpenAIConnection(model string) error {
	status, body, err := t.doRequest(t.baseURL+"/v1/chat/completions", ChatRequest{
		Model:    model,
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		var resp ChatResponse
		if json.Unmarshal(body, &resp) == nil && resp.Error != nil {
			return fmt.Errorf("API错误 (%d): %s", status, resp.Error.Message)
		}
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", status, string(body))
	}

	var resp ChatResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("API响应无效：没有返回任何内容")
	}
	return nil
}
