package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// isManagedProviderKey 判断 key 是否属于本工具管理的 dmxapi 命名空间。
// 仅这些 key 在写入新配置时会被清理；用户手动添加的其他 provider 不受影响。
func isManagedProviderKey(key string) bool {
	return key == "dmxapi" || strings.HasPrefix(key, "dmxapi-")
}

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

	// 原子写入（tmp + rename），避免进程中断留下半截 JSON
	// 注意：Windows 会忽略 Unix 权限位（0600），Windows 权限警告已在 EnsureDir 中统一输出
	if err := writeFileAtomic(configPath, data, 0600); err != nil {
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

	// 读取并合并现有认证配置。清理本工具管理的 dmxapi/dmxapi-* 命名空间，
	// 避免切换模型组合后旧 provider 的 key 残留在 auth.json 中。
	existingAuth := w.readExistingAuth(authPath)
	if existingAuth != nil {
		for k := range existingAuth {
			if isManagedProviderKey(k) {
				delete(existingAuth, k)
			}
		}
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

	// 原子写入（tmp + rename）
	// 注意：Windows 会忽略 Unix 权限位（0600），Windows 权限警告已在 EnsureDir 中统一输出
	if err := writeFileAtomic(authPath, data, 0600); err != nil {
		return "", fmt.Errorf("写入认证文件失败: %w", err)
	}

	return authPath, nil
}

// writeFileAtomic 先写到同目录下的临时文件再 rename 到目标路径，避免进程中断留下半截文件。
// 同目录 rename 在主流文件系统上是原子操作。
// 使用 os.CreateTemp 生成唯一后缀，避免并发或上次崩溃残留的 ".tmp" 冲突。
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	f, err := os.CreateTemp(dir, base+".tmp.*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	cleanup := func() { _ = os.Remove(tmp) }
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		cleanup()
		return err
	}
	if err := f.Chmod(perm); err != nil && runtime.GOOS != "windows" {
		_ = f.Close()
		cleanup()
		return err
	}
	if err := f.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		cleanup()
		return err
	}
	return nil
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

	// 合并：先清理 existing 中已经不在新配置里的本工具管理 provider；
	// 对新配置中的每个 managed provider 与 existing 同名 provider 做深度合并，
	// 保留用户在 model 层手动添加的自定义字段（如 reasoning/temperature/tools）。
	// 其他用户自定义 provider 保留不动。
	if newProvider, ok := newMap["provider"]; ok {
		existingProvider, _ := existing["provider"].(map[string]interface{})
		if existingProvider == nil {
			existingProvider = make(map[string]interface{})
		}
		np, _ := newProvider.(map[string]interface{})
		if np == nil {
			np = make(map[string]interface{})
		}
		// 清理 existing 中不在新配置里的 managed provider（切换模型组合时旧 provider 不残留）
		for k := range existingProvider {
			if isManagedProviderKey(k) {
				if _, keep := np[k]; !keep {
					delete(existingProvider, k)
				}
			}
		}
		// 写入或合并新 provider
		for k, v := range np {
			if isManagedProviderKey(k) {
				newP, _ := v.(map[string]interface{})
				oldP, _ := existingProvider[k].(map[string]interface{})
				if newP != nil && oldP != nil {
					existingProvider[k] = mergeManagedProvider(oldP, newP)
					continue
				}
			}
			existingProvider[k] = v
		}
		existing["provider"] = existingProvider
	}

	return existing, nil
}

// mergeManagedProvider 将新写入的 managed provider 与已有同名 provider 做深度合并：
// - npm / name / options 直接以新值覆盖（本工具权威字段）
// - models：按模型名合并；保留 existing model 内除 name 外的字段（用户自定义如
//   reasoning/temperature/tools），删除 existing 中不在新模型列表的模型
// - provider 顶层其他用户字段保留不动
func mergeManagedProvider(oldP, newP map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(oldP))
	for k, v := range oldP {
		out[k] = v
	}
	for k, v := range newP {
		if k == "models" {
			continue
		}
		out[k] = v
	}

	newModels, _ := newP["models"].(map[string]interface{})
	oldModels, _ := oldP["models"].(map[string]interface{})
	if newModels == nil {
		if oldModels != nil {
			out["models"] = oldModels
		}
		return out
	}
	mergedModels := make(map[string]interface{}, len(newModels))
	for name, nv := range newModels {
		if oldM, ok := oldModels[name].(map[string]interface{}); ok {
			merged := make(map[string]interface{}, len(oldM))
			for k, v := range oldM {
				merged[k] = v
			}
			if nm, ok := nv.(map[string]interface{}); ok {
				for k, v := range nm {
					merged[k] = v
				}
			}
			mergedModels[name] = merged
		} else {
			mergedModels[name] = nv
		}
	}
	out["models"] = mergedModels
	return out
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
// 使用 os.ReadDir + 前缀匹配（而非 filepath.Glob），避免 XDG_*_HOME 路径中包含
// "[" / "?" 等 glob 元字符时被错误解释。
func (w *Writer) pruneOldBackups(filePath string) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	prefix := base + ".backup."
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, filepath.Join(dir, name))
		}
	}
	if len(matches) <= maxBackupsPerFile {
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
