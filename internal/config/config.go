package config

import "strings"

// ProviderType 定义 provider 类型
type ProviderType int

const (
	ProviderAnthropic ProviderType = iota
	ProviderGoogle
	ProviderOpenAI
)

// ProviderInfo 存储 provider 元信息
type ProviderInfo struct {
	ID   string
	NPM  string
	Name string
}

// GetProviderInfo 根据类型返回 provider 信息
func GetProviderInfo(pType ProviderType) ProviderInfo {
	switch pType {
	case ProviderAnthropic:
		return ProviderInfo{ID: "dmxapi-anthropic", NPM: "@ai-sdk/anthropic", Name: "DMXAPI Claude"}
	case ProviderGoogle:
		return ProviderInfo{ID: "dmxapi-google", NPM: "@ai-sdk/google", Name: "DMXAPI Gemini"}
	default:
		return ProviderInfo{ID: "dmxapi-openai", NPM: "@ai-sdk/openai-compatible", Name: "DMXAPI OpenAI"}
	}
}

// ClassifyModel 根据模型名称前缀判断 provider 类型
func ClassifyModel(modelName string) ProviderType {
	name := strings.ToLower(modelName)
	if strings.HasPrefix(name, "claude") {
		return ProviderAnthropic
	}
	if strings.HasPrefix(name, "gemini") {
		return ProviderGoogle
	}
	return ProviderOpenAI
}

// OpenCodeConfig 表示 opencode.json 配置文件结构
type OpenCodeConfig struct {
	Provider map[string]Provider `json:"provider"`
}

// Provider 表示一个API提供者配置
type Provider struct {
	NPM     string           `json:"npm"`
	Name    string           `json:"name"`
	Options ProviderOptions  `json:"options"`
	Models  map[string]Model `json:"models"`
}

// ProviderOptions 提供者选项
type ProviderOptions struct {
	BaseURL string `json:"baseURL"`
	APIKey  string `json:"apiKey"`
}

// Model 模型配置
type Model struct {
	Name string `json:"name"`
}

// NewDMXAPIConfig 创建 DMXAPI 配置（基于模型名称路由）
func NewDMXAPIConfig(url, apiKey string, models []string) *OpenCodeConfig {
	// 按 provider 类型分组模型
	modelGroups := make(map[ProviderType]map[string]Model)
	for _, m := range models {
		pType := ClassifyModel(m)
		if modelGroups[pType] == nil {
			modelGroups[pType] = make(map[string]Model)
		}
		modelGroups[pType][m] = Model{Name: m}
	}

	// 为每组模型创建对应的 provider
	providers := make(map[string]Provider)
	for pType, modelMap := range modelGroups {
		info := GetProviderInfo(pType)
		providers[info.ID] = Provider{
			NPM:  info.NPM,
			Name: info.Name,
			Options: ProviderOptions{
				BaseURL: url + "/v1",
				APIKey:  apiKey,
			},
			Models: modelMap,
		}
	}

	return &OpenCodeConfig{
		Provider: providers,
	}
}

// AuthConfig 表示 auth.json 认证配置
type AuthConfig map[string]AuthEntry

// AuthEntry 单个认证条目
type AuthEntry struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

// NewAuthConfig 创建认证配置（支持多 provider）
func NewAuthConfig(providerIDs []string, apiKey string) AuthConfig {
	authConfig := make(AuthConfig)
	for _, id := range providerIDs {
		authConfig[id] = AuthEntry{Type: "api", Key: apiKey}
	}
	return authConfig
}

// GetProviderIDs 从配置中提取所有 provider ID
func GetProviderIDs(config *OpenCodeConfig) []string {
	var ids []string
	for id := range config.Provider {
		ids = append(ids, id)
	}
	return ids
}
