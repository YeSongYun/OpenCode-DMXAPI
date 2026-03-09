package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

// stdinReader 包级别共享的 stdin reader，避免多个 bufio.Reader 争抢同一个 os.Stdin
var stdinReader = bufio.NewReader(os.Stdin)

// SelectMenu 显示交互式上下键选择菜单，返回选中项的索引（0-based）
func SelectMenu(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("选项不能为空")
	}
	return selectMenuImpl(prompt, options)
}

// selectMenuFallback 数字输入回退方案（用于 Windows 或终端不支持 raw mode 时）
func selectMenuFallback(prompt string, options []string) (int, error) {
	fmt.Println(prompt)
	for i, opt := range options {
		fmt.Printf("  %d. %s\n", i+1, opt)
	}

	for {
		fmt.Printf("请输入选项 (1-%d): ", len(options))
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("读取输入失败: %w", err)
		}
		line = strings.TrimSpace(line)
		for i := range options {
			if line == fmt.Sprintf("%d", i+1) {
				return i, nil
			}
		}
		fmt.Printf("无效输入，请输入 1 到 %d 之间的数字\n", len(options))
	}
}
