# DMXAPI 配置工具

> 一键配置 OpenCode 使用 DMXAPI 服务

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](https://github.com/dmxapi/opencode_dmxapi)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 功能特点

- **智能模型路由** - 根据模型名称自动选择最佳 SDK（Claude → Anthropic SDK, Gemini → Google SDK, 其他 → OpenAI 兼容）
- **双模式配置** - 完整配置（URL + API Key + 模型）或仅模型配置（快速更换模型）
- **API 连接验证** - 自动测试 API Key 有效性
- **安全备份** - 自动备份现有配置文件
- **配置合并** - 智能合并现有配置，保留自定义设置
- **跨平台支持** - Windows / macOS / Linux

## 系统要求

| 平台 | 要求 |
|------|------|
| Windows | Windows 10 x64 及以上 |
| macOS | macOS 11 (Big Sur) 及以上，支持 Intel 与 Apple Silicon |
| Linux | glibc 2.17+，支持 x64 与 ARM64 |

> **注意**：macOS 用户首次运行未签名二进制需执行 `xattr -dr com.apple.quarantine <文件名>` 移除 Gatekeeper 隔离属性。

## 下载安装

| 平台 | 文件名 | 架构 |
|------|--------|------|
| Windows | `opencode-dmxapi-<版本>-windows-amd64.exe` | x64 |
| macOS (Intel) | `opencode-dmxapi-<版本>-macos-amd64` | x64 |
| macOS (Apple Silicon) | `opencode-dmxapi-<版本>-macos-arm64` | ARM64 |
| Linux (x64) | `opencode-dmxapi-<版本>-linux-amd64` | x64 |
| Linux (ARM64) | `opencode-dmxapi-<版本>-linux-arm64` | ARM64 |

## 快速开始

### Windows

```powershell
# 下载后直接运行
.\opencode-dmxapi-<版本>-windows-amd64.exe
```

### macOS (Intel)

```bash
chmod +x opencode-dmxapi-<版本>-macos-amd64
# 移除 Gatekeeper 隔离属性（首次运行需要）
xattr -dr com.apple.quarantine opencode-dmxapi-<版本>-macos-amd64
./opencode-dmxapi-<版本>-macos-amd64
```

### macOS (Apple Silicon)

```bash
chmod +x opencode-dmxapi-<版本>-macos-arm64
# 移除 Gatekeeper 隔离属性（首次运行需要）
xattr -dr com.apple.quarantine opencode-dmxapi-<版本>-macos-arm64
./opencode-dmxapi-<版本>-macos-arm64
```

### Linux

```bash
chmod +x opencode-dmxapi-<版本>-linux-amd64
./opencode-dmxapi-<版本>-linux-amd64
# ARM64 用户将 amd64 替换为 arm64
```

运行后按提示操作：
1. 选择配置模式（如存在现有配置）
2. 输入 **DMXAPI URL**（默认: https://www.dmxapi.cn）
3. 输入 **API Key**（从 https://www.dmxapi.cn/token 获取）
4. 输入 **模型名称**（多个用逗号分隔）
5. 程序自动测试连接并生成配置文件
6. 运行 `opencode` 启动程序

## 配置模式

当检测到现有配置时，程序提供两种模式：

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| **完整配置** | 重新配置 URL、API Key 和模型 | 首次配置或需要更换账号 |
| **仅模型配置** | 保留现有 URL 和 API Key，只更新模型 | 快速切换或添加模型 |

## 智能模型路由

程序根据模型名称前缀自动路由到最优 SDK：

| 模型前缀 | Provider | SDK | 示例 |
|----------|----------|-----|------|
| `claude*` | dmxapi-anthropic | @ai-sdk/anthropic | claude-opus-4-5-20251101, claude-sonnet-4-20250514 |
| `gemini*` | dmxapi-google | @ai-sdk/google | gemini-2.5-pro, gemini-2.5-flash |
| `gpt-5*` | dmxapi-openai-responses | @ai-sdk/openai | gpt-5, gpt-5.2, gpt-5-mini, gpt-5-turbo |
| 其他 | dmxapi-openai | @ai-sdk/openai-compatible | DeepSeek-V3, gpt-4o, o1 |

**示例：** 配置 `claude-opus-4-5-20251101,gemini-2.5-pro,DeepSeek-V3` 将自动创建 3 个 Provider。

## 配置文件

### 文件位置

| 文件 | Windows | macOS/Linux |
|------|---------|-------------|
| opencode.json | `C:\Users\<用户>\.config\opencode\opencode.json` | `~/.config/opencode/opencode.json` |
| auth.json | `C:\Users\<用户>\.local\share\opencode\auth.json` | `~/.local/share/opencode/auth.json` |

### opencode.json 示例

多 Provider 配置格式（自动生成）：

```json
{
  "provider": {
    "dmxapi-anthropic": {
      "npm": "@ai-sdk/anthropic",
      "name": "DMXAPI Claude",
      "options": {
        "baseURL": "https://www.dmxapi.cn/v1",
        "apiKey": "sk-xxx"
      },
      "models": {
        "claude-opus-4-5-20251101": { "name": "claude-opus-4-5-20251101" }
      }
    },
    "dmxapi-google": {
      "npm": "@ai-sdk/google",
      "name": "DMXAPI Gemini",
      "options": {
        "baseURL": "https://www.dmxapi.cn/v1",
        "apiKey": "sk-xxx"
      },
      "models": {
        "gemini-2.5-pro": { "name": "gemini-2.5-pro" }
      }
    },
    "dmxapi-openai": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "DMXAPI OpenAI",
      "options": {
        "baseURL": "https://www.dmxapi.cn/v1",
        "apiKey": "sk-xxx"
      },
      "models": {
        "DeepSeek-V3": { "name": "DeepSeek-V3" }
      }
    }
  }
}
```

### auth.json 示例

```json
{
  "dmxapi-anthropic": {
    "type": "api",
    "key": "sk-xxx"
  },
  "dmxapi-google": {
    "type": "api",
    "key": "sk-xxx"
  },
  "dmxapi-openai": {
    "type": "api",
    "key": "sk-xxx"
  }
}
```

## 从源码构建

```bash
# 克隆项目
git clone https://cnb.cool/dmxapi/opencode_dmxapi.git
cd opencode_dmxapi

# 构建当前平台
go build -o dmxapi-config .

# 跨平台构建
# Windows
GOOS=windows GOARCH=amd64 go build -o dmxapi-config-windows.exe .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o dmxapi-config-macos-amd64 .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dmxapi-config-macos-arm64 .

# Linux
GOOS=linux GOARCH=amd64 go build -o dmxapi-config-linux .
```

## 常见问题

### macOS 提示"无法验证开发者"

macOS Gatekeeper 会阻止运行未经 Apple 公证的二进制文件。解决方法：

```bash
# 方法一：命令行移除隔离属性（推荐）
xattr -dr com.apple.quarantine opencode-dmxapi-<版本>-macos-arm64

# 方法二：系统设置手动允许
# 前往"系统设置 > 隐私与安全性"，找到被阻止的程序，点击"仍要打开"
```

### API Key 无效或连接失败

- 确认 API Key 以 `sk-` 开头，从 https://www.dmxapi.cn/token 获取
- 检查网络是否能访问 https://www.dmxapi.cn
- 如使用自定义 URL，确保末尾不含多余斜杠

### 配置文件在哪里？

| 文件 | Windows | macOS / Linux |
|------|---------|---------------|
| `opencode.json` | `C:\Users\<用户>\.config\opencode\opencode.json` | `~/.config/opencode/opencode.json` |
| `auth.json` | `C:\Users\<用户>\.local\share\opencode\auth.json` | `~/.local/share/opencode/auth.json` |

备份文件保存在同目录下，后缀为 `.bak.<时间戳>`。

## 相关链接

| 资源 | 链接 |
|------|------|
| DMXAPI 官网 | https://www.dmxapi.cn |
| 获取 API Key | https://www.dmxapi.cn/token |
| 可用模型列表 | https://www.dmxapi.cn/rmb |
| OpenCode 文档 | https://opencode.ai |
