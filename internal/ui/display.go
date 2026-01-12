package ui

import (
	"fmt"
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
	// Windows CMD 默认不支持ANSI颜色，但Windows Terminal和PowerShell Core支持
	// 这里简化处理，始终使用颜色
	return true
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
	fmt.Printf("%s[步骤 %d]%s %s\n", ColorBlue, step, ColorReset, message)
}

// PrintSuccess 打印成功信息
func PrintSuccess(message string) {
	fmt.Printf("%s✓ %s%s\n", ColorGreen, message, ColorReset)
}

// PrintError 打印错误信息
func PrintError(message string) {
	fmt.Printf("%s✗ %s%s\n", ColorRed, message, ColorReset)
}

// PrintInfo 打印提示信息
func PrintInfo(message string) {
	fmt.Printf("%s→ %s%s\n", ColorYellow, message, ColorReset)
}

// PrintWarning 打印警告信息
func PrintWarning(message string) {
	fmt.Printf("%s⚠ %s%s\n", ColorYellow, message, ColorReset)
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
