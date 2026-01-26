package config

import (
	"encoding/json"
	"os"
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

	// 查找所有 dmxapi-* provider（兼容新旧格式）
	var models []string
	var url, apiKey string

	for key, provider := range config.Provider {
		if key == "dmxapi" || strings.HasPrefix(key, "dmxapi-") {
			for modelName := range provider.Models {
				models = append(models, modelName)
			}
			if url == "" {
				url = provider.Options.BaseURL
				apiKey = provider.Options.APIKey
			}
		}
	}

	if len(models) == 0 {
		return nil
	}

	// 移除 /v1 后缀
	if len(url) > 3 && url[len(url)-3:] == "/v1" {
		url = url[:len(url)-3]
	}

	return &ExistingConfig{
		URL:    url,
		APIKey: apiKey,
		Models: models,
	}
}

// HasExistingConfig 检查是否存在现有配置
func (r *Reader) HasExistingConfig() bool {
	return r.ReadExistingConfig() != nil
}

// MaskAPIKey 遮蔽 API Key，只显示前4位和后4位
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "**********"
	}
	return apiKey[:4] + "**********" + apiKey[len(apiKey)-4:]
}
