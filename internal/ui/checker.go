package ui

import (
	"os/exec"
	"strings"
)

// CheckOpencode 检测 opencode 是否已安装。
// 返回 (installed bool, version string)，version 在无法获取时为空字符串。
func CheckOpencode() (bool, string) {
	path, err := exec.LookPath("opencode")
	if err != nil || path == "" {
		return false, ""
	}

	// 尝试获取版本号
	out, err := exec.Command("opencode", "--version").Output()
	if err != nil {
		// LookPath 已确认存在，只是拿不到版本
		return true, ""
	}

	version := strings.TrimSpace(string(out))
	return true, version
}
