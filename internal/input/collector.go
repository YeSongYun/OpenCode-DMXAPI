package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
)

// ConfigMode 配置模式
type ConfigMode int

const (
	ConfigModeFull      ConfigMode = 1 // 完整配置
	ConfigModeModelOnly ConfigMode = 2 // 仅配置模型
)

// TestFailedAction 表示 API 测试失败后用户选择的下一步动作
type TestFailedAction int

const (
	TestFailedActionRetry TestFailedAction = iota + 1 // 重新输入 URL/API Key/模型
	TestFailedActionForce                              // 强制写入（跳过连接测试）
	TestFailedActionAbort                              // 退出
)

// Collector 用户输入收集器
type Collector struct{}

// NewCollector 创建新的输入收集器
func NewCollector() *Collector {
	return &Collector{}
}

// isTerminal 检测标准输入是否为终端设备（问题5修复）
// 使用 go-isatty 正确处理 Unix、Windows 和 Cygwin/Git Bash 环境
func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// isTTYError 判断 huh 的错误是否与 TTY 不可用相关（问题4修复）
// 覆盖 Unix（inappropriate ioctl）、Windows 旧版（handle is invalid）等错误形式
func isTTYError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not a terminal") ||
		strings.Contains(msg, "inappropriate ioctl") ||
		strings.Contains(msg, "no such device") ||
		strings.Contains(msg, "handle is invalid") ||
		strings.Contains(msg, "operation not supported")
}

// fallbackInput 在非 TTY 环境下使用 bufio 读取一行输入
// 当 huh 不可用时（如脚本重定向、旧版 Windows）提供基础输入能力
func fallbackInput(prompt, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("  %s [默认: %s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("  %s: ", prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	// 容忍 "EOF + 非空行" 的情况：heredoc / `echo -n` 等不带尾换行的输入
	// 也能正确读到内容，而不是被当作读取失败
	if err != nil && (!errors.Is(err, io.EOF) || line == "") {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}

// fallbackSelect 在非 TTY 环境下使用数字编号替代交互式下拉菜单
func fallbackSelect(prompt string, options []string) (int, error) {
	fmt.Printf("  %s\n", prompt)
	for i, opt := range options {
		fmt.Printf("    %d) %s\n", i+1, opt)
	}
	fmt.Print("  请输入选项编号: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || line == "") {
		return 0, fmt.Errorf("读取输入失败: %w", err)
	}
	line = strings.TrimSpace(line)
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(options) {
		return 0, fmt.Errorf("无效的选项: %s（请输入 1-%d）", line, len(options))
	}
	return n - 1, nil
}

// CollectConfigMode 收集配置模式选择
func (c *Collector) CollectConfigMode() (ConfigMode, error) {
	// 问题5修复：非 TTY 环境直接使用 fallback
	if !isTerminal() {
		return c.collectConfigModeFallback()
	}
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
		// 问题4修复：huh 失败时（如旧版 Windows）fallback 到简单输入
		if errors.Is(err, huh.ErrUserAborted) {
			return 0, fmt.Errorf("用户取消")
		}
		if isTTYError(err) {
			return c.collectConfigModeFallback()
		}
		return 0, err
	}
	return mode, nil
}

func (c *Collector) collectConfigModeFallback() (ConfigMode, error) {
	idx, err := fallbackSelect("请选择配置模式", []string{
		"完整配置 - 重新配置所有选项",
		"仅配置模型 - 保留现有 URL 和 API Key",
	})
	if err != nil {
		return 0, err
	}
	return ConfigMode(idx + 1), nil
}

// CollectTestFailedAction 在 API 连接测试失败后询问用户下一步动作。
// 返回 Retry / Force / Abort 三种选项之一；用户按 Ctrl+C 视为 Abort。
func (c *Collector) CollectTestFailedAction() (TestFailedAction, error) {
	options := []string{
		"重新配置 - 重新输入 URL/API Key/模型",
		"强制写入 - 跳过测试，按当前输入继续",
		"退出",
	}
	if !isTerminal() {
		idx, err := fallbackSelect("API 测试失败，请选择下一步", options)
		if err != nil {
			return 0, err
		}
		return TestFailedAction(idx + 1), nil
	}
	var act TestFailedAction
	err := huh.NewSelect[TestFailedAction]().
		Title("API 测试失败，请选择下一步").
		Options(
			huh.NewOption(options[0], TestFailedActionRetry),
			huh.NewOption(options[1], TestFailedActionForce),
			huh.NewOption(options[2], TestFailedActionAbort),
		).
		Value(&act).
		Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return TestFailedActionAbort, nil
		}
		if isTTYError(err) {
			idx, ferr := fallbackSelect("API 测试失败，请选择下一步", options)
			if ferr != nil {
				return 0, ferr
			}
			return TestFailedAction(idx + 1), nil
		}
		return 0, err
	}
	return act, nil
}

// CollectURL 收集URL输入
func (c *Collector) CollectURL() (string, error) {
	if !isTerminal() {
		return c.collectURLFallback()
	}
	var rawURL string
	err := huh.NewInput().
		Title("请输入 DMXAPI URL").
		Description("留空使用默认值: https://www.dmxapi.cn").
		Validate(func(s string) error {
			if s == "" {
				return nil // 允许空值，后续填默认值
			}
			return ValidateURL(s)
		}).
		Value(&rawURL).
		Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", fmt.Errorf("用户取消")
		}
		if isTTYError(err) {
			return c.collectURLFallback()
		}
		return "", err
	}
	if rawURL == "" {
		rawURL = "https://www.dmxapi.cn"
	}
	return strings.TrimSuffix(rawURL, "/"), nil
}

func (c *Collector) collectURLFallback() (string, error) {
	rawURL, err := fallbackInput("请输入 DMXAPI URL（留空使用默认值 https://www.dmxapi.cn）", "https://www.dmxapi.cn")
	if err != nil {
		return "", err
	}
	if rawURL != "https://www.dmxapi.cn" {
		if err := ValidateURL(rawURL); err != nil {
			return "", err
		}
	}
	return strings.TrimSuffix(rawURL, "/"), nil
}

// CollectAPIKey 收集API Key输入
func (c *Collector) CollectAPIKey() (string, error) {
	if !isTerminal() {
		return c.collectAPIKeyFallback()
	}
	var apiKey string
	err := huh.NewInput().
		Title("请输入 API Key").
		Description("获取地址: https://www.dmxapi.cn/token").
		EchoMode(huh.EchoModePassword).
		Validate(ValidateAPIKey).
		Value(&apiKey).
		Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", fmt.Errorf("用户取消")
		}
		if isTTYError(err) {
			return c.collectAPIKeyFallback()
		}
		return "", err
	}
	return apiKey, nil
}

func (c *Collector) collectAPIKeyFallback() (string, error) {
	// 优先尝试无回显读取（Unix/Windows 真终端均支持）；失败再降级到明文输入。
	if isTerminal() {
		fmt.Print("  请输入 API Key（输入不回显）: ")
		pw, err := term.ReadPassword(os.Stdin.Fd())
		fmt.Println()
		if err == nil {
			apiKey := strings.TrimSpace(string(pw))
			if err := ValidateAPIKey(apiKey); err != nil {
				return "", err
			}
			return apiKey, nil
		}
		// 读取失败（如 Windows 旧版控制台），降级
	}
	fmt.Println("  注意: 非交互模式，API Key 将以明文显示")
	apiKey, err := fallbackInput("请输入 API Key", "")
	if err != nil {
		return "", err
	}
	if err := ValidateAPIKey(apiKey); err != nil {
		return "", err
	}
	return apiKey, nil
}

// CollectModels 收集模型名称输入
func (c *Collector) CollectModels() ([]string, error) {
	if !isTerminal() {
		return c.collectModelsFallback()
	}
	var line string
	err := huh.NewInput().
		Title("请输入模型名称，多个用逗号分隔").
		Description("可用模型: https://www.dmxapi.cn/rmb").
		Validate(func(s string) error {
			return ValidateModels(parseModels(s))
		}).
		Value(&line).
		Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, fmt.Errorf("用户取消")
		}
		if isTTYError(err) {
			return c.collectModelsFallback()
		}
		return nil, err
	}
	return parseModels(line), nil
}

func (c *Collector) collectModelsFallback() ([]string, error) {
	line, err := fallbackInput("请输入模型名称（多个用逗号分隔，如 claude-opus-4-5-20251101,DeepSeek-V3.2-Fast）", "")
	if err != nil {
		return nil, err
	}
	models := parseModels(line)
	if err := ValidateModels(models); err != nil {
		return nil, err
	}
	return models, nil
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

