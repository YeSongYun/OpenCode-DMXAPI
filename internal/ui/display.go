package ui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

// Version 版本号常量，方便后续修改
const Version = "2.0.6"

// dmxapiASCIIArt DMXAPI 的 block 字符 ASCII Art（需要 Unicode Box Drawing 字符支持）
const dmxapiASCIIArt = `
 ██████╗ ███╗   ███╗██╗  ██╗ █████╗ ██████╗ ██╗
 ██╔══██╗████╗ ████║╚██╗██╔╝██╔══██╗██╔══██╗██║
 ██║  ██║██╔████╔██║ ╚███╔╝ ███████║██████╔╝██║
 ██║  ██║██║╚██╔╝██║ ██╔██╗ ██╔══██║██╔═══╝ ██║
 ██████╔╝██║ ╚═╝ ██║██╔╝ ██╗██║  ██║██║     ██║
 ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝
`

// dmxapiASCIIFallback 纯 ASCII 横幅，用于旧版 Windows CMD 等不支持 Unicode 的终端
const dmxapiASCIIFallback = `
 +------------------------------------------+
 |  DMXAPI  -  OpenCode Configuration Tool  |
 +------------------------------------------+
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

// 颜色支持缓存（sync.Once 保证线程安全，避免重复检测）
var (
	colorOnce  sync.Once
	colorCache bool
)

// Unicode 支持缓存
var (
	unicodeOnce  sync.Once
	unicodeCache bool
)

// supportsColor 判断终端是否支持 ANSI 颜色输出（结果缓存，避免重复检测）
func supportsColor() bool {
	colorOnce.Do(func() {
		colorCache = detectColor()
	})
	return colorCache
}

// detectColor 实际检测终端颜色支持能力
func detectColor() bool {
	// 遵循 NO_COLOR 标准 (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// COLORTERM 表示终端明确声明支持真彩色或 256 色
	if os.Getenv("COLORTERM") != "" {
		return true
	}
	if runtime.GOOS == "windows" {
		// 问题3修复：首次调用时主动尝试启用 ENABLE_VIRTUAL_TERMINAL_PROCESSING
		// 适用于 Windows 10 v1511+ 的 PowerShell、cmd.exe 和 Windows Terminal
		if tryEnableVT() {
			return true
		}
		// Windows Terminal（WT_SESSION 由 Windows Terminal 进程注入）
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
		// 问题2修复：Git Bash / MSYS2 / Cygwin 检测
		if os.Getenv("MSYSTEM") != "" {
			return true
		}
		term := os.Getenv("TERM")
		if strings.HasPrefix(term, "xterm") || strings.Contains(term, "cygwin") {
			return true
		}
		// 旧版 CMD：不支持 ANSI
		return false
	}
	// TERM=dumb 明确表示不支持颜色（如某些 CI 环境）
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// supportsUnicode 判断终端是否支持 Unicode 字符输出（结果缓存）
// 在 Windows 上检查控制台输出代码页是否为 65001（UTF-8）
func supportsUnicode() bool {
	unicodeOnce.Do(func() {
		unicodeCache = detectUnicode()
	})
	return unicodeCache
}

// detectUnicode 实际检测终端 Unicode 支持能力
func detectUnicode() bool {
	if runtime.GOOS == "windows" {
		// 问题1修复：通过 GetConsoleOutputCP() 检查代码页
		// 65001 = UTF-8，其他值（如 936=GBK, 437=CP437）不能可靠显示 Unicode Box Drawing 字符
		return getConsoleOutputCP() == 65001
	}
	// Unix 系统：检查 locale 环境变量
	for _, env := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := os.Getenv(env); strings.Contains(strings.ToUpper(v), "UTF") {
			return true
		}
	}
	// 现代 Unix 终端默认支持 Unicode
	return true
}

// symbol 根据终端 Unicode 能力返回合适的符号
// 在不支持 Unicode 的终端（如旧版 Windows CMD）中使用 ASCII 降级符号
func symbol(unicode, ascii string) string {
	if supportsUnicode() {
		return unicode
	}
	return ascii
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
	// 问题1修复：根据终端能力选择 Unicode Art 或纯 ASCII 横幅
	if supportsUnicode() {
		fmt.Print(colorize(ColorBold+ColorCyan, dmxapiASCIIArt))
	} else {
		fmt.Print(dmxapiASCIIFallback)
	}
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
	fmt.Println(colorize(ColorGreen, "  "+symbol("✓", "[OK]")+" "+message))
}

// PrintError 打印错误信息
func PrintError(message string) {
	fmt.Println(colorize(ColorRed, "  "+symbol("✗", "[X]")+" "+message))
}

// PrintInfo 打印提示信息
func PrintInfo(message string) {
	fmt.Println(colorize(ColorYellow, "  "+symbol("→", "->")+" "+message))
}

// PrintWarning 打印警告信息
func PrintWarning(message string) {
	fmt.Println(colorize(ColorYellow, "  "+symbol("⚠", "[!]")+" "+message))
}

// PrintDivider 打印分隔线
func PrintDivider() {
	fmt.Println()
}

// PrintComplete 打印完成信息
func PrintComplete() {
	fmt.Println()
	fmt.Println(colorize(ColorGreen, "  "+symbol("✓", "[OK]")+" 配置完成！"))
	fmt.Printf("  %s\n", colorize(ColorDim, "运行 'opencode' 启动程序"))
	fmt.Println()
}

// PrintConfigModeHeader 打印配置模式选择标题
func PrintConfigModeHeader() {
	fmt.Println()
	fmt.Printf("  %s %s\n", colorize(ColorCyan, symbol("⚙", "[*]")), "请选择配置模式：")
	fmt.Println()
}

// PrintExistingConfigInfo 显示当前配置信息
func PrintExistingConfigInfo(url, maskedAPIKey string, models []string) {
	fmt.Println()
	fmt.Printf("  %s 检测到现有 DMXAPI 配置\n", colorize(ColorCyan, symbol("⚙", "[*]")))
	fmt.Printf("    %-6s  %s\n", "URL:", url)
	fmt.Printf("    %-6s  %s\n", "Key:", maskedAPIKey)
	fmt.Printf("    %-6s  %s\n", "模型:", strings.Join(models, ", "))
	fmt.Println()
}

// PrintModelOnlyModeInfo 打印仅模型模式提示
func PrintModelOnlyModeInfo() {
	fmt.Println()
	PrintInfo("仅配置模型模式")
}

// PrintUpdateNotice 打印新版本提示
func PrintUpdateNotice(latestVersion, dlURL string) {
	fmt.Printf("  %s 发现新版本 %s（当前 v%s）\n",
		colorize(ColorYellow, symbol("→", "->")),
		colorize(ColorGreen+ColorBold, "v"+latestVersion),
		Version,
	)
	fmt.Printf("    %s %s\n\n", colorize(ColorDim, "下载:"), dlURL)
}
