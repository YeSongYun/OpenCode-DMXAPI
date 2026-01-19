package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfigMode 配置模式
type ConfigMode int

const (
	ConfigModeFull      ConfigMode = 1 // 完整配置
	ConfigModeModelOnly ConfigMode = 2 // 仅配置模型
)

// UserInput 用户输入的配置信息
type UserInput struct {
	URL    string   // API URL (如 https://www.dmxapi.cn)
	APIKey string   // API密钥
	Models []string // 模型名称列表
}

// Collector 用户输入收集器
type Collector struct {
	reader *bufio.Reader
}

// NewCollector 创建新的输入收集器
func NewCollector() *Collector {
	return &Collector{
		reader: bufio.NewReader(os.Stdin),
	}
}

// CollectConfigMode 收集配置模式选择
func (c *Collector) CollectConfigMode() (ConfigMode, error) {
	fmt.Print("请输入选项 (1 或 2): ")

	input, err := c.reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("读取选项失败: %w", err)
	}

	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return ConfigModeFull, nil
	case "2":
		return ConfigModeModelOnly, nil
	default:
		return 0, fmt.Errorf("无效的选项，请输入 1 或 2")
	}
}

// CollectURL 收集URL输入
func (c *Collector) CollectURL() (string, error) {
	fmt.Println("请输入 DMXAPI URL")
	fmt.Println("示例: https://www.dmxapi.cn")
	fmt.Print("URL: ")

	url, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("读取URL失败: %w", err)
	}

	url = strings.TrimSpace(url)

	// 如果用户没有输入，使用默认值
	if url == "" {
		url = "https://www.dmxapi.cn"
		fmt.Printf("使用默认值: %s\n", url)
	}

	// 移除末尾的斜杠
	url = strings.TrimSuffix(url, "/")

	return url, nil
}

// CollectAPIKey 收集API Key输入
func (c *Collector) CollectAPIKey() (string, error) {
	fmt.Println("请输入 API Key")
	fmt.Println("获取地址: https://www.dmxapi.cn/token")
	fmt.Print("API Key: ")

	apiKey, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("读取API Key失败: %w", err)
	}

	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return "", fmt.Errorf("API Key 不能为空")
	}

	return apiKey, nil
}

// CollectModels 收集模型名称输入
func (c *Collector) CollectModels() ([]string, error) {
	fmt.Println("请输入模型名称（多个模型用逗号分隔）")
	fmt.Println("可用模型列表: https://www.dmxapi.cn/rmb")
	fmt.Println("示例: claude-opus-4-5-20251101,DeepSeek-V3.2-Fast")
	fmt.Print("模型: ")

	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取模型失败: %w", err)
	}

	line = strings.TrimSpace(line)

	if line == "" {
		return nil, fmt.Errorf("至少需要指定一个模型")
	}

	// 分割并清理模型名称
	parts := strings.Split(line, ",")
	var models []string
	for _, p := range parts {
		model := strings.TrimSpace(p)
		if model != "" {
			models = append(models, model)
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("至少需要指定一个模型")
	}

	return models, nil
}

// CollectAll 收集所有必要的输入
func (c *Collector) CollectAll() (*UserInput, error) {
	url, err := c.CollectURL()
	if err != nil {
		return nil, err
	}

	apiKey, err := c.CollectAPIKey()
	if err != nil {
		return nil, err
	}

	models, err := c.CollectModels()
	if err != nil {
		return nil, err
	}

	return &UserInput{
		URL:    url,
		APIKey: apiKey,
		Models: models,
	}, nil
}
