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
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0

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
