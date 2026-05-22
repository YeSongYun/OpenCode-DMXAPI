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
// 遵循 XDG Base Directory 规范：优先使用 $XDG_CONFIG_HOME，回退到 ~/.config。
// 注意：opencode 主程序在 Windows 上同样使用 ~/.config（而非 %APPDATA%），
// 因此本工具保持一致，无需针对 Windows 做特殊处理。
// 参考：https://github.com/sst/opencode/issues/6156
func GetConfigPath() (string, error) {
	base, err := xdgBase("XDG_CONFIG_HOME", ".config")
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "opencode", "opencode.json"), nil
}

// GetAuthPath 返回 auth.json 认证文件的路径
// 遵循 XDG：优先 $XDG_DATA_HOME，回退 ~/.local/share
func GetAuthPath() (string, error) {
	base, err := xdgBase("XDG_DATA_HOME", filepath.Join(".local", "share"))
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "opencode", "auth.json"), nil
}

// xdgBase 返回 XDG 基目录：若环境变量非空则使用之，否则回退到 ~/<fallback>
func xdgBase(envVar, fallback string) (string, error) {
	if v := os.Getenv(envVar); v != "" {
		return v, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	return filepath.Join(homeDir, fallback), nil
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

