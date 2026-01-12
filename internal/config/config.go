package config

// OpenCodeConfig 表示 opencode.json 配置文件结构
type OpenCodeConfig struct {
	Schema   string              `json:"$schema"`
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

// NewDMXAPIConfig 创建 DMXAPI 配置
func NewDMXAPIConfig(url, apiKey string, models []string) *OpenCodeConfig {
	modelMap := make(map[string]Model)
	for _, m := range models {
		modelMap[m] = Model{Name: m}
	}

	return &OpenCodeConfig{
		Schema: "https://opencode.ai/config.json",
		Provider: map[string]Provider{
			"dmxapi": {
				NPM:  "@ai-sdk/openai-compatible",
				Name: "DMXAPI",
				Options: ProviderOptions{
					BaseURL: url + "/v1",
					APIKey:  apiKey,
				},
				Models: modelMap,
			},
		},
	}
}

// AuthConfig 表示 auth.json 认证配置
type AuthConfig map[string]AuthEntry

// AuthEntry 单个认证条目
type AuthEntry struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

// NewAuthConfig 创建认证配置
func NewAuthConfig(providerID, apiKey string) AuthConfig {
	return AuthConfig{
		providerID: AuthEntry{
			Type: "api",
			Key:  apiKey,
		},
	}
}
