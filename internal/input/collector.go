package input

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// ConfigMode 配置模式
type ConfigMode int

const (
	ConfigModeFull      ConfigMode = 1 // 完整配置
	ConfigModeModelOnly ConfigMode = 2 // 仅配置模型
)

// Collector 用户输入收集器
type Collector struct{}

// NewCollector 创建新的输入收集器
func NewCollector() *Collector {
	return &Collector{}
}

// CollectConfigMode 收集配置模式选择
func (c *Collector) CollectConfigMode() (ConfigMode, error) {
	var mode ConfigMode
	err := huh.NewSelect[ConfigMode]().
		Title("请选择配置模式").
		Options(
			huh.NewOption("完整配置 - 重新配置所有选项", ConfigModeFull),
			huh.NewOption("仅配置模型 - 保留现有 URL 和 API Key", ConfigModeModelOnly),
		).
		Value(&mode).
		Run()
	if err != nil {
		return 0, mapError(err)
	}
	return mode, nil
}

// CollectURL 收集URL输入
func (c *Collector) CollectURL() (string, error) {
	var rawURL string
	err := huh.NewInput().
		Title("请输入 DMXAPI URL").
		Description("留空使用默认值: https://www.dmxapi.cn").
		Placeholder("https://www.dmxapi.cn").
		Validate(func(s string) error {
			if s == "" {
				return nil // 允许空值，后续填默认值
			}
			return ValidateURL(s)
		}).
		Value(&rawURL).
		Run()
	if err != nil {
		return "", mapError(err)
	}
	if rawURL == "" {
		rawURL = "https://www.dmxapi.cn"
	}
	return strings.TrimSuffix(rawURL, "/"), nil
}

// CollectAPIKey 收集API Key输入
func (c *Collector) CollectAPIKey() (string, error) {
	var apiKey string
	err := huh.NewInput().
		Title("请输入 API Key").
		Description("获取地址: https://www.dmxapi.cn/token").
		EchoMode(huh.EchoModePassword).
		Validate(ValidateAPIKey).
		Value(&apiKey).
		Run()
	if err != nil {
		return "", mapError(err)
	}
	return apiKey, nil
}

// CollectModels 收集模型名称输入
func (c *Collector) CollectModels() ([]string, error) {
	var line string
	err := huh.NewInput().
		Title("请输入模型名称，多个用逗号分隔").
		Description("可用模型: https://www.dmxapi.cn/rmb").
		Placeholder("claude-opus-4-5-20251101,DeepSeek-V3.2-Fast").
		Validate(func(s string) error {
			return ValidateModels(parseModels(s))
		}).
		Value(&line).
		Run()
	if err != nil {
		return nil, mapError(err)
	}
	return parseModels(line), nil
}

// parseModels 解析逗号分隔的模型名称
func parseModels(s string) []string {
	var models []string
	for _, p := range strings.Split(s, ",") {
		if m := strings.TrimSpace(p); m != "" {
			models = append(models, m)
		}
	}
	return models
}

// mapError 将 huh 的 Ctrl+C 错误映射为友好消息
func mapError(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return fmt.Errorf("用户取消")
	}
	return err
}
