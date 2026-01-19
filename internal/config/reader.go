package config

import (
	"encoding/json"
	"os"
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

	// 提取 dmxapi 提供者配置
	dmxapi, exists := config.Provider["dmxapi"]
	if !exists {
		return nil
	}

	// 提取模型列表
	var models []string
	for modelName := range dmxapi.Models {
		models = append(models, modelName)
	}

	// 提取 URL（移除 /v1 后缀）
	url := dmxapi.Options.BaseURL
	if len(url) > 3 && url[len(url)-3:] == "/v1" {
		url = url[:len(url)-3]
	}

	return &ExistingConfig{
		URL:    url,
		APIKey: dmxapi.Options.APIKey,
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
