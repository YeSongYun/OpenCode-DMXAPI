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

## 下载安装

| 平台 | 文件名 | 架构 |
|------|--------|------|
| Windows | `dmxapi-config-windows.exe` | x64 |
| macOS (Intel) | `dmxapi-config-macos-amd64` | x64 |
| macOS (Apple Silicon) | `dmxapi-config-macos-arm64` | ARM64 |
| Linux | `dmxapi-config-linux` | x64 |

## 快速开始

1. 下载对应平台的可执行文件
2. 运行程序
3. 选择配置模式（如存在现有配置）
4. 按提示输入信息：
   - **DMXAPI URL**（默认: https://www.dmxapi.cn）
   - **API Key**（从 https://www.dmxapi.cn/token 获取）
   - **模型名称**（多个用逗号分隔）
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

## 相关链接

| 资源 | 链接 |
|------|------|
| DMXAPI 官网 | https://www.dmxapi.cn |
| 获取 API Key | https://www.dmxapi.cn/token |
| 可用模型列表 | https://www.dmxapi.cn/rmb |
| OpenCode 文档 | https://opencode.ai |
