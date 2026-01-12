package auth

import (
	"dmxapi-config/internal/config"
)

// AuthManager 认证管理器
type AuthManager struct {
	providerID string
	apiKey     string
}

// NewAuthManager 创建认证管理器
func NewAuthManager(providerID, apiKey string) *AuthManager {
	return &AuthManager{
		providerID: providerID,
		apiKey:     apiKey,
	}
}

// Login 执行认证配置（直接写入auth.json）
func (a *AuthManager) Login() (string, error) {
	// 创建认证配置
	authConfig := config.NewAuthConfig(a.providerID, a.apiKey)

	// 写入认证文件
	writer := config.NewWriter()
	return writer.WriteAuth(authConfig)
}
