package input

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateURL 验证URL格式
func ValidateURL(input string) error {
	if input == "" {
		return fmt.Errorf("URL不能为空")
	}

	parsed, err := url.ParseRequestURI(input)
	if err != nil {
		return fmt.Errorf("URL格式无效: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL必须以 http:// 或 https:// 开头")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL必须包含有效的主机名")
	}

	return nil
}

// ValidateAPIKey 验证API Key格式
func ValidateAPIKey(key string) error {
	if key == "" {
		return fmt.Errorf("API Key不能为空")
	}

	// API Key 通常以 sk- 开头，但不强制要求
	if len(key) < 8 {
		return fmt.Errorf("API Key 长度过短")
	}

	// 检查是否包含空格或特殊字符
	if strings.ContainsAny(key, " \t\n\r") {
		return fmt.Errorf("API Key 不能包含空格或换行符")
	}

	return nil
}

// ValidateModels 验证模型列表
func ValidateModels(models []string) error {
	if len(models) == 0 {
		return fmt.Errorf("至少需要指定一个模型")
	}

	for _, model := range models {
		if strings.TrimSpace(model) == "" {
			return fmt.Errorf("模型名称不能为空")
		}
	}

	return nil
}

// ValidateInput 验证所有输入
func ValidateInput(input *UserInput) error {
	if err := ValidateURL(input.URL); err != nil {
		return fmt.Errorf("URL验证失败: %w", err)
	}

	if err := ValidateAPIKey(input.APIKey); err != nil {
		return fmt.Errorf("API Key验证失败: %w", err)
	}

	if err := ValidateModels(input.Models); err != nil {
		return fmt.Errorf("模型验证失败: %w", err)
	}

	return nil
}
