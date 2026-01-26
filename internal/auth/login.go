package auth

import (
	"dmxapi-config/internal/config"
)

// AuthManager 认证管理器
type AuthManager struct {
	providerIDs []string // 支持多 provider
	apiKey      string
}

// NewAuthManager 创建认证管理器（支持多 provider）
func NewAuthManager(providerIDs []string, apiKey string) *AuthManager {
	return &AuthManager{
		providerIDs: providerIDs,
		apiKey:      apiKey,
	}
}

// Login 执行认证配置（直接写入auth.json）
func (a *AuthManager) Login() (string, error) {
	// 创建认证配置（支持多 provider）
	authConfig := config.NewAuthConfig(a.providerIDs, a.apiKey)

	// 写入认证文件
	writer := config.NewWriter()
	return writer.WriteAuth(authConfig)
}
