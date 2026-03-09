//go:build !windows && !darwin && !linux

package input

import "fmt"

type terminalState struct{}

func makeTerminalRaw() (*terminalState, error) {
	return nil, fmt.Errorf("该平台不支持 raw terminal，使用数字输入模式")
}

func restoreTerminal(_ *terminalState) {}
