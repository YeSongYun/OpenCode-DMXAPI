//go:build windows

package ui

import (
	"os"
	"sync"
	"syscall"
	"unsafe"
)

const _ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procGetConsoleOutputCP = kernel32.NewProc("GetConsoleOutputCP")

	vtOnce    sync.Once
	vtEnabled bool
)

// tryEnableVT 尝试启用 Windows 虚拟终端处理（ANSI 颜色支持）
// 适用于 Windows 10 v1511+，包含 PowerShell 和 Windows Terminal
// 使用 sync.Once 缓存结果，避免重复 syscall 开销
func tryEnableVT() bool {
	vtOnce.Do(func() {
		handle := syscall.Handle(os.Stdout.Fd())
		var mode uint32
		// GetConsoleMode 获取当前控制台模式
		r, _, _ := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
		if r == 0 {
			return // 句柄无效（非控制台，如管道输出）
		}
		// SetConsoleMode 追加 ENABLE_VIRTUAL_TERMINAL_PROCESSING 标志
		r, _, _ = procSetConsoleMode.Call(uintptr(handle), uintptr(mode|_ENABLE_VIRTUAL_TERMINAL_PROCESSING))
		vtEnabled = r != 0
	})
	return vtEnabled
}

// getConsoleOutputCP 获取当前控制台输出代码页
// UTF-8 代码页为 65001
func getConsoleOutputCP() uint32 {
	r, _, _ := procGetConsoleOutputCP.Call()
	return uint32(r)
}
