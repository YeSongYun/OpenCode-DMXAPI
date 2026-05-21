package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// maxBackupsPerFile 限制单个目标文件保留的备份份数，避免长期使用堆积无数 .backup.*
const maxBackupsPerFile = 5

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
	w.pruneOldBackups(configPath)

	// 合并现有配置（使用 map 保留未知字段）
	merged, err := w.mergeConfigPreservingFields(configPath, config)
	if err != nil {
		return "", fmt.Errorf("合并配置失败: %w", err)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件（配置中含 API Key，使用 0600 限制权限）
	// 注意：Windows 会忽略 Unix 权限位（0600），Windows 权限警告已在 EnsureDir 中统一输出
	if err := os.WriteFile(configPath, data, 0600); err != nil {
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

	// 备份现有认证配置
	if err := w.backupIfExists(authPath); err != nil {
		fmt.Printf("警告: 备份现有认证配置失败: %v\n", err)
	}
	w.pruneOldBackups(authPath)

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
	// 注意：Windows 会忽略 Unix 权限位（0600），Windows 权限警告已在 EnsureDir 中统一输出
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

	// 写入备份（继承原文件的严格权限）
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("创建备份失败: %w", err)
	}

	fmt.Printf("已备份现有配置到: %s\n", backupPath)
	return nil
}

// mergeConfigPreservingFields 使用 map[string]interface{} 合并配置，保留 JSON 中的所有字段。
// 读取或解析失败时返回错误，避免静默覆盖现有 opencode.json。备份已在调用前写出，
// 用户可从备份恢复后再修复源文件。
func (w *Writer) mergeConfigPreservingFields(filePath string, newConfig *OpenCodeConfig) (interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newConfig, nil
		}
		return nil, fmt.Errorf("读取现有配置失败: %w", err)
	}

	var existing map[string]interface{}
	if err := json.Unmarshal(data, &existing); err != nil {
		return nil, fmt.Errorf("解析现有配置失败（已备份原文件，请检查 %s 后重试）: %w", filePath, err)
	}

	// 将新配置序列化再反序列化为 map，以便合并
	newData, err := json.Marshal(newConfig)
	if err != nil {
		return nil, fmt.Errorf("序列化新配置失败: %w", err)
	}
	var newMap map[string]interface{}
	if err := json.Unmarshal(newData, &newMap); err != nil {
		return nil, fmt.Errorf("反序列化新配置失败: %w", err)
	}

	// 合并：新配置的 provider 覆盖到现有 map 中
	if newProvider, ok := newMap["provider"]; ok {
		existingProvider, _ := existing["provider"].(map[string]interface{})
		if existingProvider == nil {
			existingProvider = make(map[string]interface{})
		}
		if np, ok := newProvider.(map[string]interface{}); ok {
			for k, v := range np {
				existingProvider[k] = v
			}
		}
		existing["provider"] = existingProvider
	}

	return existing, nil
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

// pruneOldBackups 仅保留 filePath 对应的最近 maxBackupsPerFile 份 .backup.* 文件，
// 删除更早的备份，避免长期使用堆积无数备份占用磁盘。失败时仅输出警告。
func (w *Writer) pruneOldBackups(filePath string) {
	matches, err := filepath.Glob(filePath + ".backup.*")
	if err != nil || len(matches) <= maxBackupsPerFile {
		return
	}
	// 时间戳格式 20060102_150405 字典序即时间序，升序排序后删除最旧的若干份
	sort.Strings(matches)
	for _, p := range matches[:len(matches)-maxBackupsPerFile] {
		if err := os.Remove(p); err != nil {
			fmt.Printf("警告: 清理旧备份失败 %s: %v\n", p, err)
		}
	}
}
