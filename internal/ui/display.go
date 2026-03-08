package ui

import (
	"fmt"
	"os"
	"runtime"
)

// ANSI颜色代码
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
)

// 判断是否支持颜色输出
func supportsColor() bool {
	// 遵循 NO_COLOR 标准 (https://no-color.org/)：设置此变量则禁用颜色
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if runtime.GOOS == "windows" {
		// Windows Terminal 启动时注入 WT_SESSION，是支持 ANSI 最可靠的标志
		if os.Getenv("WT_SESSION") != "" {
			return true
		}
		// ANSICON 终端模拟器
		if os.Getenv("ANSICON") != "" {
			return true
		}
		// ConEmu / Cmder
		if os.Getenv("ConEmuANSI") == "ON" {
			return true
		}
		// 旧版 CMD：不支持 ANSI
		return false
	}
	return true
}

// IsLegacyWindowsCMD 判断当前是否运行在不支持 ANSI 的旧版 Windows CMD 中
func IsLegacyWindowsCMD() bool {
	return runtime.GOOS == "windows" && !supportsColor()
}

// colorize 添加颜色
func colorize(color, text string) string {
	if !supportsColor() {
		return text
	}
	return color + text + ColorReset
}

// PrintBanner 打印程序横幅
func PrintBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║           DMXAPI OpenCode 配置工具                        ║
║                                                           ║
║   本工具帮助您快速配置 OpenCode 使用 DMXAPI 服务         ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(colorize(ColorCyan, banner))
}

// PrintStep 打印步骤信息
func PrintStep(step int, message string) {
	fmt.Printf("%s %s\n", colorize(ColorBlue, fmt.Sprintf("[步骤 %d]", step)), message)
}

// PrintSuccess 打印成功信息
func PrintSuccess(message string) {
	fmt.Println(colorize(ColorGreen, "✓ "+message))
}

// PrintError 打印错误信息
func PrintError(message string) {
	fmt.Println(colorize(ColorRed, "✗ "+message))
}

// PrintInfo 打印提示信息
func PrintInfo(message string) {
	fmt.Println(colorize(ColorYellow, "→ "+message))
}

// PrintWarning 打印警告信息
func PrintWarning(message string) {
	fmt.Println(colorize(ColorYellow, "⚠ "+message))
}

// PrintDivider 打印分隔线
func PrintDivider() {
	fmt.Println("─────────────────────────────────────────────────────────")
}

// PrintSystemInfo 打印系统信息
func PrintSystemInfo() {
	fmt.Printf("操作系统: %s (%s)\n", runtime.GOOS, runtime.GOARCH)
}

// PrintComplete 打印完成信息
func PrintComplete() {
	complete := `
╔═══════════════════════════════════════════════════════════╗
║                    配置完成！                             ║
║                                                           ║
║   现在可以运行 'opencode' 启动程序                       ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(colorize(ColorGreen, complete))
}

// PrintConfigModeHeader 打印配置模式选择标题
func PrintConfigModeHeader() {
	fmt.Println()
	fmt.Println(colorize(ColorCyan, "=== 配置模式选择 ==="))
	fmt.Println()
	fmt.Println("检测到现有配置，请选择配置模式:")
	fmt.Println()
	fmt.Println("  [1] 完整配置 - 重新配置所有选项 (URL, API Key, 模型)")
	fmt.Println("  [2] 仅配置模型 - 保留现有 URL 和 API Key，只修改模型列表")
	fmt.Println()
}

// PrintExistingConfigInfo 显示当前配置信息
func PrintExistingConfigInfo(url, maskedAPIKey string, models []string) {
	fmt.Println()
	fmt.Println(colorize(ColorCyan, "--- 当前配置信息 ---"))
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  API Key: %s\n", maskedAPIKey)
	fmt.Printf("  当前模型: %v\n", models)
	fmt.Println()
}

// PrintModelOnlyModeInfo 打印仅模型模式提示
func PrintModelOnlyModeInfo() {
	fmt.Println()
	PrintInfo("您选择了「仅配置模型」模式")
}
