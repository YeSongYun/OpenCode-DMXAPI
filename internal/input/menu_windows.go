//go:build windows

package input

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle      = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode    = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode    = kernel32.NewProc("SetConsoleMode")
	procReadConsoleInputW = kernel32.NewProc("ReadConsoleInputW")
)

const (
	// GetStdHandle 参数：Windows DWORD -10 / -11
	stdInputHandle  = uintptr(^uint32(9))  // 0xFFFFFFF6 = STD_INPUT_HANDLE
	stdOutputHandle = uintptr(^uint32(10)) // 0xFFFFFFF5 = STD_OUTPUT_HANDLE

	// SetConsoleMode 输入模式标志（应用于 stdin）
	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004

	// SetConsoleMode 输出模式标志（应用于 stdout，与输入标志数值独立）
	enableVTProcessing = 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING

	// INPUT_RECORD 事件类型
	keyEvent = 0x0001

	// 虚拟键码
	vkUp     = 0x26 // VK_UP
	vkDown   = 0x28 // VK_DOWN
	vkReturn = 0x0D // VK_RETURN

	// 控制键状态掩码
	leftCtrlPressed  = 0x0008
	rightCtrlPressed = 0x0004
)

// inputRecord 对应 Windows INPUT_RECORD（共 20 字节）
// EventType(2) + 对齐填充(2) + Event union(16)
type inputRecord struct {
	EventType uint16
	_         [2]byte  // 结构对齐填充
	Event     [16]byte // union 取最大子结构 KEY_EVENT_RECORD
}

// keyEventRecord 对应 Windows KEY_EVENT_RECORD（16 字节）
type keyEventRecord struct {
	KeyDown         int32  // BOOL (4)
	RepeatCount     uint16 // (2)
	VirtualKeyCode  uint16 // (2)
	VirtualScanCode uint16 // (2)
	Char            uint16 // WCHAR union (2)
	ControlKeyState uint32 // (4)
}

// winConsoleState 保存控制台原始模式，用于退出时恢复
type winConsoleState struct {
	hIn     syscall.Handle
	hOut    syscall.Handle
	inMode  uint32
	outMode uint32
}

func setupWindowsConsole() (*winConsoleState, error) {
	const invalidHandle = ^uintptr(0) // INVALID_HANDLE_VALUE

	r, _, err := procGetStdHandle.Call(stdInputHandle)
	if r == 0 || r == invalidHandle {
		return nil, fmt.Errorf("获取 stdin 句柄失败: %w", err)
	}
	hIn := syscall.Handle(r)

	r, _, err = procGetStdHandle.Call(stdOutputHandle)
	if r == 0 || r == invalidHandle {
		return nil, fmt.Errorf("获取 stdout 句柄失败: %w", err)
	}
	hOut := syscall.Handle(r)

	var inMode, outMode uint32
	if r, _, e := procGetConsoleMode.Call(uintptr(hIn), uintptr(unsafe.Pointer(&inMode))); r == 0 {
		return nil, fmt.Errorf("GetConsoleMode(stdin) 失败: %w", e)
	}
	if r, _, e := procGetConsoleMode.Call(uintptr(hOut), uintptr(unsafe.Pointer(&outMode))); r == 0 {
		return nil, fmt.Errorf("GetConsoleMode(stdout) 失败: %w", e)
	}

	state := &winConsoleState{hIn: hIn, hOut: hOut, inMode: inMode, outMode: outMode}

	// 原始输入模式：禁用行缓冲、回显、信号处理
	newInMode := inMode &^ uint32(enableLineInput|enableEchoInput|enableProcessedInput)
	if r, _, e := procSetConsoleMode.Call(uintptr(hIn), uintptr(newInMode)); r == 0 {
		return nil, fmt.Errorf("SetConsoleMode(stdin) 失败: %w", e)
	}

	// 开启 ANSI/VT 输出（Windows 10+）；失败说明系统过旧，恢复 stdin 后降级
	if r, _, e := procSetConsoleMode.Call(uintptr(hOut), uintptr(outMode|enableVTProcessing)); r == 0 {
		procSetConsoleMode.Call(uintptr(hIn), uintptr(inMode))
		return nil, fmt.Errorf("开启 VT 输出失败（系统版本过低）: %w", e)
	}

	return state, nil
}

func restoreWindowsConsole(s *winConsoleState) {
	procSetConsoleMode.Call(uintptr(s.hIn), uintptr(s.inMode))
	procSetConsoleMode.Call(uintptr(s.hOut), uintptr(s.outMode))
}

// renderMenu Windows 版：ANSI 转义码原地重绘菜单（与 menu_unix.go 逻辑一致）
func renderMenu(prompt string, options []string, selected int, isFirst bool) {
	if !isFirst {
		// 向上移动光标 len(options)+1 行（1 行 prompt + N 行选项）
		fmt.Printf("\033[%dA", len(options)+1)
	}
	fmt.Printf("\033[2K\r%s\n", prompt)
	for i, opt := range options {
		if i == selected {
			fmt.Printf("\033[2K\r%s▶ %s%s\n", colorGreen, opt, colorReset)
		} else {
			fmt.Printf("\033[2K\r  %s\n", opt)
		}
	}
}

// selectMenuImpl Windows 实现：ReadConsoleInputW 捕获方向键事件
func selectMenuImpl(prompt string, options []string) (int, error) {
	s, err := setupWindowsConsole()
	if err != nil {
		// 控制台初始化失败（极旧 Windows / 非控制台环境），降级到数字输入
		return selectMenuFallback(prompt, options)
	}
	defer restoreWindowsConsole(s)

	selected := 0
	renderMenu(prompt, options, selected, true)

	var rec inputRecord
	var numRead uint32

	for {
		r, _, readErr := procReadConsoleInputW.Call(
			uintptr(s.hIn),
			uintptr(unsafe.Pointer(&rec)),
			1,
			uintptr(unsafe.Pointer(&numRead)),
		)
		if r == 0 {
			return 0, fmt.Errorf("ReadConsoleInputW 失败: %w", readErr)
		}

		if rec.EventType != keyEvent {
			continue
		}

		key := (*keyEventRecord)(unsafe.Pointer(&rec.Event[0]))
		if key.KeyDown == 0 {
			continue // 忽略键弹起事件，只处理按下事件
		}

		switch key.VirtualKeyCode {
		case vkUp:
			if selected > 0 {
				selected--
				renderMenu(prompt, options, selected, false)
			}
		case vkDown:
			if selected < len(options)-1 {
				selected++
				renderMenu(prompt, options, selected, false)
			}
		case vkReturn:
			return selected, nil
		case 0x43: // 'C' — 仅当同时按下 Ctrl 时视为 Ctrl+C
			if key.ControlKeyState&uint32(leftCtrlPressed|rightCtrlPressed) != 0 {
				return 0, fmt.Errorf("用户取消")
			}
		}
	}
}
