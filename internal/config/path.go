package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// GetConfigPath 返回 opencode.json 配置文件的路径
//
// 所有平台均遵循 XDG Base Directory 规范，使用 ~/.config/opencode/opencode.json。
// 注意：opencode 主程序在 Windows 上同样使用此路径（而非 %APPDATA%），
// 因此本工具保持一致，无需针对 Windows 做特殊处理。
// 参考：https://github.com/sst/opencode/issues/6156
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

// windowsPermWarning 确保 Windows 权限提示只输出一次（问题7修复）
var windowsPermWarning sync.Once

// EnsureDir 确保目录存在，如果不存在则创建
// 在 Windows 上输出一次性提示，说明 Unix 权限位（0600/0755）不受文件系统保护
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if runtime.GOOS == "windows" {
		windowsPermWarning.Do(func() {
			fmt.Println("注意: Windows 不支持 Unix 文件权限 (0600/0755)，请确保配置文件所在目录的访问权限受限。")
		})
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
