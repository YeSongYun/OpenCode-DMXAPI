//go:build windows

package input

// selectMenuImpl Windows 下回退到数字输入（无 raw terminal 支持）
func selectMenuImpl(prompt string, options []string) (int, error) {
	return selectMenuFallback(prompt, options)
}
