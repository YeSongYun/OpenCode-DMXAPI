//go:build darwin

package input

import (
	"syscall"
	"unsafe"
)

// Darwin 终端 ioctl 编号（来自 macOS sys/ttycom.h）
const (
	ioctlGetTermios uintptr = 0x402c7413 // TIOCGETA
	ioctlSetTermios uintptr = 0x802c7414 // TIOCSETA
)

type terminalState struct {
	saved syscall.Termios
}

func makeTerminalRaw() (*terminalState, error) {
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		ioctlGetTermios,
		uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}

	state := &terminalState{saved: t}

	// 关闭 ECHO（回显）、ICANON（行缓冲）、IEXTEN（扩展处理）、ISIG（信号）
	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	// 关闭软件流控和 CR/LF 转换等输入处理
	t.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	t.Cflag |= syscall.CS8
	// VMIN=0, VTIME=1：最多等待 100ms；若无字节则 Read 返回 0
	// 用于优雅处理孤立 ESC 按键（不阻塞等待后续字节）
	t.Cc[syscall.VMIN] = 0
	t.Cc[syscall.VTIME] = 1

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		ioctlSetTermios,
		uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}

	return state, nil
}

func restoreTerminal(state *terminalState) {
	syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		ioctlSetTermios,
		uintptr(unsafe.Pointer(&state.saved)))
}
