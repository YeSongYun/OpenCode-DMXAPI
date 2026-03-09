//go:build !windows

package input

import (
	"fmt"
	"os"
)

// renderMenu 用 ANSI 控制码原地重绘菜单，避免滚动
func renderMenu(prompt string, options []string, selected int, isFirst bool) {
	if !isFirst {
		// 向上移动光标 len(options)+1 行（1 行 prompt + N 行选项）
		fmt.Printf("\033[%dA", len(options)+1)
	}
	// 清行并打印 prompt
	fmt.Printf("\033[2K\r%s\n", prompt)
	// 打印每个选项
	for i, opt := range options {
		if i == selected {
			fmt.Printf("\033[2K\r%s▶ %s%s\n", colorGreen, opt, colorReset)
		} else {
			fmt.Printf("\033[2K\r  %s\n", opt)
		}
	}
}

// readByte 从 stdin 读取一个字节（阻塞）
func readByte() (byte, error) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return 0, err
		}
		if n > 0 {
			return buf[0], nil
		}
	}
}

// selectMenuImpl Unix/macOS 交互实现：syscall raw mode + 方向键读取
func selectMenuImpl(prompt string, options []string) (int, error) {
	state, err := makeTerminalRaw()
	if err != nil {
		// 终端不支持 raw mode，fallback 到数字输入
		fmt.Println("\033[33m  ⚠ 终端不支持方向键选择，已切换为数字输入模式\033[0m")
		return selectMenuFallback(prompt, options)
	}
	defer restoreTerminal(state)

	selected := 0
	renderMenu(prompt, options, selected, true)

	for {
		b, err := readByte()
		if err != nil {
			return 0, fmt.Errorf("读取输入失败: %w", err)
		}

		switch b {
		case '\r', '\n': // Enter 确认选择
			return selected, nil
		case 3: // Ctrl+C
			return 0, fmt.Errorf("用户取消")
		case 27: // ESC —— 逐字节读取后续序列判断方向键
			b2, err := readByte()
			if err != nil {
				continue
			}
			if b2 != '[' {
				continue // 不是 CSI 序列，忽略
			}
			b3, err := readByte()
			if err != nil {
				continue
			}
			switch b3 {
			case 'A': // 上箭头 ESC[A
				if selected > 0 {
					selected--
					renderMenu(prompt, options, selected, false)
				}
			case 'B': // 下箭头 ESC[B
				if selected < len(options)-1 {
					selected++
					renderMenu(prompt, options, selected, false)
				}
			}
		}
	}
}
