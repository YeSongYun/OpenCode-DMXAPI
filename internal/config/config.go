package config

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// APIFormat 是上游 API 的协议格式。ZCode 使用前三种，本工具额外支持 Google。
type APIFormat string

const (
	APIFormatAnthropic       APIFormat = "anthropic-messages"
	APIFormatGoogle          APIFormat = "google-generative-ai"
	APIFormatOpenAIChat      APIFormat = "openai-chat-completions"
	APIFormatOpenAIResponses APIFormat = "openai-responses"
)

// ProviderType 保留旧调用方的类型名称；新的代码应优先使用 APIFormat。
type ProviderType = APIFormat

const (
	ProviderAnthropic       = APIFormatAnthropic
	ProviderGoogle          = APIFormatGoogle
	ProviderOpenAI          = APIFormatOpenAIChat
	ProviderOpenAIResponses = APIFormatOpenAIResponses
)

type ProviderInfo struct {
	Format APIFormat
	ID     string
	NPM    string
	Name   string
}

var providerInfos = map[APIFormat]ProviderInfo{
	APIFormatAnthropic: {
		Format: APIFormatAnthropic,
		ID:     "dmxapi-anthropic",
		NPM:    "@ai-sdk/anthropic",
		Name:   "DMXAPI Claude",
	},
	APIFormatGoogle: {
		Format: APIFormatGoogle,
		ID:     "dmxapi-google",
		NPM:    "@ai-sdk/google",
		Name:   "DMXAPI Gemini",
	},
	APIFormatOpenAIChat: {
		Format: APIFormatOpenAIChat,
		ID:     "dmxapi-openai",
		NPM:    "@ai-sdk/openai-compatible",
		Name:   "DMXAPI OpenAI Chat",
	},
	APIFormatOpenAIResponses: {
		Format: APIFormatOpenAIResponses,
		ID:     "dmxapi-openai-responses",
		NPM:    "@ai-sdk/openai",
		Name:   "DMXAPI OpenAI Responses",
	},
}

// GetProviderInfo 根据 API 格式返回 provider 元信息。
func GetProviderInfo(format APIFormat) ProviderInfo {
	if info, ok := providerInfos[format]; ok {
		return info
	}
	return providerInfos[APIFormatOpenAIChat]
}

// ModelMatch 是模型匹配结果。Reason 便于 CLI 或测试解释匹配来源。
type ModelMatch struct {
	Model  string
	Format APIFormat
	Reason string
}

type modelRule struct {
	format  APIFormat
	reason  string
	pattern *regexp.Regexp
}

var modelRules = []modelRule{
	{
		format:  APIFormatAnthropic,
		reason:  "Anthropic/Claude 模型",
		pattern: regexp.MustCompile("^(?:claude|anthropic)(?:[-_:.\\/]|$)"),
	},
	{
		format:  APIFormatGoogle,
		reason:  "Google/Gemini 模型",
		pattern: regexp.MustCompile("^(?:gemini|gemma|google)(?:[-_:.\\/]|$)"),
	},
	{
		format:  APIFormatOpenAIResponses,
		reason:  "OpenAI Responses 模型",
		pattern: regexp.MustCompile("^(?:gpt-[56](?:[-_:.\\/]|$)|gpt-4\\.1(?:[-_:.\\/]|$)|o[134](?:[-_:.\\/]|$)|codex(?:[-_:.\\/]|$)|grok(?:[-_:.\\/]|$))"),
	},
}

// namespaceFormat 将显式 provider/model 前缀映射为 API 格式。
// 未知命名空间不参与强制匹配，以免破坏合法的 provider/model 模型 ID。
func namespaceFormat(namespace string) (APIFormat, bool) {
	switch strings.ToLower(strings.TrimSpace(namespace)) {
	case "anthropic", "claude":
		return APIFormatAnthropic, true
	case "google", "google-ai", "gemini", "gemma":
		return APIFormatGoogle, true
	case "openai", "openai-responses", "responses", "codex", "xai", "x-ai":
		return APIFormatOpenAIResponses, true
	case "chat", "openai-chat", "openai-compatible", "completions", "zai", "bigmodel":
		return APIFormatOpenAIChat, true
	default:
		return "", false
	}
}

func splitModelNamespace(model string) (namespace, local string, ok bool) {
	parts := strings.SplitN(model, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", model, false
	}
	if _, known := namespaceFormat(parts[0]); !known {
		return "", model, false
	}
	return parts[0], parts[1], true
}

// MatchModel 按显式命名空间、内置规则和默认兜底匹配 API 格式。
// 模型 ID 不会被改写，匹配只决定 provider 和请求协议。
func MatchModel(modelName string) ModelMatch {
	model := strings.TrimSpace(modelName)
	lower := strings.ToLower(model)
	local := lower

	if namespace, _, ok := splitModelNamespace(lower); ok {
		if format, known := namespaceFormat(namespace); known {
			return ModelMatch{Model: model, Format: format, Reason: "provider 命名空间 " + namespace}
		}
	}

	for _, rule := range modelRules {
		if rule.pattern.MatchString(local) {
			return ModelMatch{Model: model, Format: rule.format, Reason: rule.reason}
		}
	}
	if strings.HasSuffix(local, "-cc") {
		return ModelMatch{Model: model, Format: APIFormatAnthropic, Reason: "模型名以 -cc 结尾"}
	}
	return ModelMatch{Model: model, Format: APIFormatOpenAIChat, Reason: "未知模型使用 OpenAI Chat 兼容格式"}
}

// ClassifyModel 保留旧函数名，返回模型对应的 API 格式。
func ClassifyModel(modelName string) ProviderType {
	return MatchModel(modelName).Format
}

// APIFormatName 返回适合 CLI 展示的 API 格式名称。
func APIFormatName(format APIFormat) string {
	switch format {
	case APIFormatAnthropic:
		return "Anthropic Messages"
	case APIFormatGoogle:
		return "Google Generative AI"
	case APIFormatOpenAIResponses:
		return "OpenAI Responses"
	default:
		return "OpenAI Chat Completions"
	}
}

var (
	versionSuffixPattern  = regexp.MustCompile("/v\\d+(?:beta\\d*)?/?$")
	knownEndpointSuffixes = []string{
		"/chat/completions",
		"/responses",
		"/messages",
	}
)

func stripEndpointPath(path string) string {
	path = strings.TrimRight(path, "/")
	lower := strings.ToLower(path)
	for _, suffix := range knownEndpointSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return strings.TrimRight(path[:len(path)-len(suffix)], "/")
		}
	}
	if idx := strings.LastIndex(lower, "/models/"); idx >= 0 && strings.HasSuffix(lower, ":generatecontent") {
		return strings.TrimRight(path[:idx], "/")
	}
	return path
}

func hasVersionSuffix(path string) bool {
	return versionSuffixPattern.MatchString(strings.TrimRight(path, "/"))
}

func cleanEndpointURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return strings.TrimRight(rawURL, "/")
	}
	parsed.Path = stripEndpointPath(parsed.Path)
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/")
}

// NormalizeBaseURL 清理用户输入中的 endpoint 和版本后缀，返回基础地址。
func NormalizeBaseURL(rawURL string) string {
	rawURL = cleanEndpointURL(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return strings.TrimRight(versionSuffixPattern.ReplaceAllString(rawURL, ""), "/")
	}
	parsed.Path = strings.TrimRight(versionSuffixPattern.ReplaceAllString(parsed.Path, ""), "/")
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/")
}

func addDefaultVersion(parsed *url.URL, format APIFormat) {
	path := strings.TrimRight(parsed.Path, "/")
	if hasVersionSuffix(path) {
		parsed.Path = path
		return
	}
	switch format {
	case APIFormatAnthropic:
		// Anthropic 的自定义根（如 /api/anthropic）本身就是完整 API 根。
		if path == "" || path == "/" {
			parsed.Path = path + "/v1"
		}
	case APIFormatGoogle:
		parsed.Path = path + "/v1beta"
	default:
		parsed.Path = path + "/v1"
	}
}

// BuildProviderBaseURL 生成写入 OpenCode provider.options.baseURL 的 API 根地址。
func BuildProviderBaseURL(rawURL string, format APIFormat) string {
	cleaned := cleanEndpointURL(rawURL)
	parsed, err := url.Parse(cleaned)
	if err != nil {
		return strings.TrimRight(cleaned, "/")
	}
	addDefaultVersion(parsed, format)
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/")
}

func appendURLPath(baseURL, suffix string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return strings.TrimRight(baseURL, "/") + suffix
	}
	query := parsed.RawQuery
	parsed.RawQuery = ""
	prefix := strings.TrimRight(parsed.String(), "/")
	result := prefix + suffix
	if query != "" {
		result += "?" + query
	}
	return result
}

// BuildAPIEndpoint 生成连接测试所使用的完整 endpoint。
func BuildAPIEndpoint(rawURL string, format APIFormat, model string) string {
	baseURL := BuildProviderBaseURL(rawURL, format)
	switch format {
	case APIFormatAnthropic:
		return appendURLPath(baseURL, "/messages")
	case APIFormatGoogle:
		return appendURLPath(baseURL, "/models/"+url.PathEscape(model)+":generateContent")
	case APIFormatOpenAIResponses:
		return appendURLPath(baseURL, "/responses")
	default:
		return appendURLPath(baseURL, "/chat/completions")
	}
}

// OpenCodeConfig 表示 opencode.json 配置文件结构。
type OpenCodeConfig struct {
	Provider map[string]Provider `json:"provider"`
}

type Provider struct {
	NPM     string           `json:"npm"`
	Name    string           `json:"name"`
	Options ProviderOptions  `json:"options"`
	Models  map[string]Model `json:"models"`
}

type ProviderOptions struct {
	BaseURL string `json:"baseURL"`
	APIKey  string `json:"apiKey"`
}

type Model struct {
	Name string `json:"name"`
}

// NewDMXAPIConfig 创建 DMXAPI 配置（基于模型名称自动匹配 API 格式）。
func NewDMXAPIConfig(rawURL, apiKey string, models []string) *OpenCodeConfig {
	modelGroups := make(map[APIFormat]map[string]Model)
	for _, model := range models {
		if strings.TrimSpace(model) == "" {
			continue
		}
		format := MatchModel(model).Format
		if modelGroups[format] == nil {
			modelGroups[format] = make(map[string]Model)
		}
		modelGroups[format][model] = Model{Name: model}
	}

	providers := make(map[string]Provider)
	for format, modelMap := range modelGroups {
		info := GetProviderInfo(format)
		providers[info.ID] = Provider{
			NPM:  info.NPM,
			Name: info.Name,
			Options: ProviderOptions{
				BaseURL: BuildProviderBaseURL(rawURL, format),
				APIKey:  apiKey,
			},
			Models: modelMap,
		}
	}
	return &OpenCodeConfig{Provider: providers}
}

// GetRepresentativeModels 返回每种协议的首个模型，供完整配置流程逐协议测试。
func GetRepresentativeModels(models []string) []string {
	firstByFormat := make(map[APIFormat]string)
	for _, model := range models {
		format := MatchModel(model).Format
		if _, exists := firstByFormat[format]; !exists {
			firstByFormat[format] = model
		}
	}
	order := []APIFormat{
		APIFormatAnthropic,
		APIFormatGoogle,
		APIFormatOpenAIResponses,
		APIFormatOpenAIChat,
	}
	result := make([]string, 0, len(firstByFormat))
	for _, format := range order {
		if model, ok := firstByFormat[format]; ok {
			result = append(result, model)
		}
	}
	return result
}

type AuthConfig map[string]AuthEntry

type AuthEntry struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

func NewAuthConfig(providerIDs []string, apiKey string) AuthConfig {
	authConfig := make(AuthConfig)
	for _, id := range providerIDs {
		authConfig[id] = AuthEntry{Type: "api", Key: apiKey}
	}
	return authConfig
}

func GetProviderIDs(config *OpenCodeConfig) []string {
	var ids []string
	for id := range config.Provider {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
