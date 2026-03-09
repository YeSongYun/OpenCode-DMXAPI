package ui

import (
	"fmt"
	"os"
	"runtime"
)

// Version 版本号常量，方便后续修改
const Version = "1.0.0"

// dmxapiASCIIArt DMXAPI 的 block 字符 ASCII Art
const dmxapiASCIIArt = `
 ██████╗ ███╗   ███╗██╗  ██╗ █████╗ ██████╗ ██╗
 ██╔══██╗████╗ ████║╚██╗██╔╝██╔══██╗██╔══██╗██║
 ██║  ██║██╔████╔██║ ╚███╔╝ ███████║██████╔╝██║
 ██║  ██║██║╚██╔╝██║ ██╔██╗ ██╔══██║██╔═══╝ ██║
 ██████╔╝██║ ╚═╝ ██║██╔╝ ██╗██║  ██║██║     ██║
 ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝
`

// ANSI颜色代码
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
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
	fmt.Print(colorize(ColorBold+ColorCyan, dmxapiASCIIArt))
	fmt.Printf(" %s · %s\n", colorize(ColorBold, "OpenCode 配置工具"), colorize(ColorCyan, "让 AI 触手可及"))
	fmt.Printf(" %s\n", colorize(ColorDim, fmt.Sprintf("v%s / %s/%s", Version, runtime.GOOS, runtime.GOARCH)))
	fmt.Println()
}

// PrintStep 打印步骤信息，格式为 [N/M]
func PrintStep(step, total int, message string) {
	fmt.Printf("\n%s %s\n", colorize(ColorCyan, fmt.Sprintf("[%d/%d]", step, total)), message)
}

// PrintSuccess 打印成功信息
func PrintSuccess(message string) {
	fmt.Println(colorize(ColorGreen, "  ✓ "+message))
}

// PrintError 打印错误信息
func PrintError(message string) {
	fmt.Println(colorize(ColorRed, "  ✗ "+message))
}

// PrintInfo 打印提示信息
func PrintInfo(message string) {
	fmt.Println(colorize(ColorYellow, "  → "+message))
}

// PrintWarning 打印警告信息
func PrintWarning(message string) {
	fmt.Println(colorize(ColorYellow, "  ⚠ "+message))
}

// PrintDivider 打印分隔线
func PrintDivider() {
	fmt.Println()
}

// PrintSystemInfo 打印系统信息
func PrintSystemInfo() {
	fmt.Printf("  系统: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// PrintComplete 打印完成信息
func PrintComplete() {
	fmt.Println()
	fmt.Println(colorize(ColorGreen, "  ✓ 配置完成！"))
	fmt.Printf("  %s\n", colorize(ColorDim, "运行 'opencode' 启动程序"))
	fmt.Println()
}

// PrintConfigModeHeader 打印配置模式选择标题
func PrintConfigModeHeader() {
	fmt.Println()
	fmt.Printf("  %s %s\n", colorize(ColorCyan, "⚙"), "检测到现有配置，请选择配置模式：")
	fmt.Println()
}

// PrintExistingConfigInfo 显示当前配置信息
func PrintExistingConfigInfo(url, maskedAPIKey string, models []string) {
	fmt.Println()
	fmt.Printf("  %s 当前配置\n", colorize(ColorCyan, "ℹ"))
	fmt.Printf("    URL:    %s\n", url)
	fmt.Printf("    Key:    %s\n", maskedAPIKey)
	fmt.Printf("    模型:   %v\n", models)
	fmt.Println()
}

// PrintModelOnlyModeInfo 打印仅模型模式提示
func PrintModelOnlyModeInfo() {
	fmt.Println()
	PrintInfo("仅配置模型模式")
}
