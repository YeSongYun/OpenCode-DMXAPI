package config

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

// ExistingConfig 表示已存在的配置信息
type ExistingConfig struct {
	URL    string   // API URL
	APIKey string   // API Key
	Models []string // 模型列表
}

// Reader 配置读取器
type Reader struct{}

// NewReader 创建新的配置读取器
func NewReader() *Reader {
	return &Reader{}
}

// ReadExistingConfig 读取现有的 DMXAPI 配置
// 如果配置不存在或读取失败，返回 nil
// 支持新旧两种格式（单 dmxapi 或多 dmxapi-* provider）
func (r *Reader) ReadExistingConfig() *ExistingConfig {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var config OpenCodeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}

	// 查找所有 dmxapi-* provider（兼容新旧格式）。
	// 按 key 字典序遍历，保证 URL/APIKey 取值在多 provider 场景下稳定，
	// 不受 map 随机迭代顺序影响。
	var models []string
	var url, apiKey string

	keys := make([]string, 0, len(config.Provider))
	for key := range config.Provider {
		if key == "dmxapi" || strings.HasPrefix(key, "dmxapi-") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		provider := config.Provider[key]
		for modelName := range provider.Models {
			models = append(models, modelName)
		}
		if url == "" && provider.Options.BaseURL != "" {
			url = provider.Options.BaseURL
		}
		if apiKey == "" && provider.Options.APIKey != "" {
			apiKey = provider.Options.APIKey
		}
	}

	if len(models) == 0 {
		return nil
	}

	sort.Strings(models)
	url = NormalizeBaseURL(url)

	return &ExistingConfig{
		URL:    url,
		APIKey: apiKey,
		Models: models,
	}
}

// MaskAPIKey 遮蔽 API Key，只显示前4位和后4位
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "**********"
	}
	return apiKey[:4] + "**********" + apiKey[len(apiKey)-4:]
}
