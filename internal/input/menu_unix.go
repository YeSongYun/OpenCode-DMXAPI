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

// selectMenuImpl Unix/macOS 交互实现：syscall raw mode + 方向键读取
func selectMenuImpl(prompt string, options []string) (int, error) {
	state, err := makeTerminalRaw()
	if err != nil {
		// 终端不支持 raw mode，fallback 到数字输入
		return selectMenuFallback(prompt, options)
	}
	defer restoreTerminal(state)

	selected := 0
	renderMenu(prompt, options, selected, true)

	buf := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(buf); err != nil {
			return 0, fmt.Errorf("读取输入失败: %w", err)
		}

		switch buf[0] {
		case '\r', '\n': // Enter 确认选择
			return selected, nil
		case 3: // Ctrl+C
			return 0, fmt.Errorf("用户取消")
		case 27: // ESC —— 读后续 2 字节判断是否为方向键
			seq := make([]byte, 2)
			n, _ := os.Stdin.Read(seq)
			if n >= 2 && seq[0] == '[' {
				switch seq[1] {
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
}
