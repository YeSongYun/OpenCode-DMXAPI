//go:build !windows

package ui

// tryEnableVT 在非 Windows 平台上无需操作，终端默认支持 ANSI
func tryEnableVT() bool {
	return true
}

// getConsoleOutputCP 在非 Windows 平台上始终返回 UTF-8 代码页编号
func getConsoleOutputCP() uint32 {
	return 65001
}
