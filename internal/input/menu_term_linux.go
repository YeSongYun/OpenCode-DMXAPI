//go:build linux

package input

import (
	"syscall"
	"unsafe"
)

// Linux 终端 ioctl 编号（来自 linux/termios.h）
const (
	ioctlGetTermios uintptr = 0x5401 // TCGETS
	ioctlSetTermios uintptr = 0x5402 // TCSETS
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

	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
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
