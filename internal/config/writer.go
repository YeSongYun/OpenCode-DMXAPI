package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Writer 配置文件写入器
type Writer struct{}

// NewWriter 创建新的写入器
func NewWriter() *Writer {
	return &Writer{}
}

// WriteConfig 写入 opencode.json 配置文件
func (w *Writer) WriteConfig(config *OpenCodeConfig) (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	// 确保目录存在
	if err := EnsureDir(configPath); err != nil {
		return "", err
	}

	// 备份现有配置
	if err := w.backupIfExists(configPath); err != nil {
		// 备份失败不阻止写入，只打印警告
		fmt.Printf("警告: 备份现有配置失败: %v\n", err)
	}

	// 合并现有配置（如果存在）
	existingConfig := w.readExistingConfig(configPath)
	if existingConfig != nil {
		config = w.mergeConfig(existingConfig, config)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %w", err)
	}

	return configPath, nil
}

// WriteAuth 写入 auth.json 认证文件
func (w *Writer) WriteAuth(authConfig AuthConfig) (string, error) {
	authPath, err := GetAuthPath()
	if err != nil {
		return "", err
	}

	// 确保目录存在
	if err := EnsureDir(authPath); err != nil {
		return "", err
	}

	// 读取并合并现有认证配置
	existingAuth := w.readExistingAuth(authPath)
	if existingAuth != nil {
		for k, v := range authConfig {
			existingAuth[k] = v
		}
		authConfig = existingAuth
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(authConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化认证配置失败: %w", err)
	}

	// 写入文件（使用更严格的权限）
	if err := os.WriteFile(authPath, data, 0600); err != nil {
		return "", fmt.Errorf("写入认证文件失败: %w", err)
	}

	return authPath, nil
}

// backupIfExists 如果文件存在则创建备份
func (w *Writer) backupIfExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	// 创建备份文件名
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(dir, fmt.Sprintf("%s.backup.%s", base, timestamp))

	// 读取原文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取原文件失败: %w", err)
	}

	// 写入备份
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("创建备份失败: %w", err)
	}

	fmt.Printf("已备份现有配置到: %s\n", backupPath)
	return nil
}

// readExistingConfig 读取现有的 opencode.json 配置
func (w *Writer) readExistingConfig(filePath string) *OpenCodeConfig {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var config OpenCodeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}

	return &config
}

// readExistingAuth 读取现有的 auth.json 配置
func (w *Writer) readExistingAuth(filePath string) AuthConfig {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var auth AuthConfig
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil
	}

	return auth
}

// mergeConfig 合并配置（保留现有的其他提供者）
func (w *Writer) mergeConfig(existing, new *OpenCodeConfig) *OpenCodeConfig {
	if existing.Provider == nil {
		existing.Provider = make(map[string]Provider)
	}

	// 合并新的提供者配置
	for k, v := range new.Provider {
		existing.Provider[k] = v
	}

	return existing
}
