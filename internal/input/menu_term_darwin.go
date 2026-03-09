//go:build darwin

package input

import (
	"fmt"
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
	// VMIN=1, VTIME=0：阻塞模式，Read 至少等待 1 字节才返回
	// 保证每次 Read 一定返回数据，避免超时导致 ESC 序列读取不完整
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		ioctlSetTermios,
		uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}

	// 验证 raw mode 是否设置成功：读回 termios 检查关键参数
	var verify syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		ioctlGetTermios,
		uintptr(unsafe.Pointer(&verify))); errno != 0 {
		return nil, fmt.Errorf("验证 termios 失败: %w", errno)
	}
	if verify.Cc[syscall.VMIN] != 1 || verify.Cc[syscall.VTIME] != 0 {
		return nil, fmt.Errorf("termios 验证失败: VMIN=%d(期望1) VTIME=%d(期望0)",
			verify.Cc[syscall.VMIN], verify.Cc[syscall.VTIME])
	}

	return state, nil
}

func restoreTerminal(state *terminalState) {
	syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		ioctlSetTermios,
		uintptr(unsafe.Pointer(&state.saved)))
}
