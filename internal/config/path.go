package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigPath 返回 opencode.json 配置文件的路径
// Windows: C:\Users\<用户>\.config\opencode\opencode.json
// macOS/Linux: ~/.config/opencode/opencode.json
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "opencode")
	return filepath.Join(configDir, "opencode.json"), nil
}

// GetAuthPath 返回 auth.json 认证文件的路径
// Windows: C:\Users\<用户>\.local\share\opencode\auth.json
// macOS/Linux: ~/.local/share/opencode/auth.json
func GetAuthPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}

	authDir := filepath.Join(homeDir, ".local", "share", "opencode")
	return filepath.Join(authDir, "auth.json"), nil
}

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	return nil
}

// GetOS 返回当前操作系统名称
func GetOS() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}
