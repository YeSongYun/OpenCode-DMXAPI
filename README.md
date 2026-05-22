# DMXAPI 配置工具

> 一键配置 OpenCode 使用 DMXAPI 服务

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](https://cnb.cool/dmxapi/opencode_dmxapi)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 功能特点

- **智能模型路由** — 根据模型名称自动归类到 Anthropic / Google / OpenAI Responses / OpenAI 兼容 SDK
- **双模式配置** — 完整配置（URL + API Key + 模型）或仅模型配置（保留 URL/Key，仅更换模型）
- **URL 规范化** — 自动去除用户输入末尾的 `/v1` / `/v1beta` 等版本后缀，按 provider 类型重新拼接
- **API 连接验证** — 写入配置前自动测试 API Key 与连通性
- **自动备份** — 写入前备份既有 `opencode.json` 和 `auth.json`
- **配置合并** — 保留 `opencode.json` 中已有的自定义字段
- **跨平台支持** — Windows / macOS / Linux（x64 与 ARM64）

## 系统要求

| 平台 | 要求 |
|------|------|
| Windows | Windows 10 及以上，x64 / ARM64 |
| macOS | 推荐 macOS 11 (Big Sur) 及以上，Intel / Apple Silicon |
| Linux | glibc 2.17+，x64 / ARM64 |

## 快速安装（推荐）

无需手动下载，一行命令完成安装并自动启动配置向导。

### Linux / macOS

```bash
curl -fsSL https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.sh | bash
```

### Windows PowerShell

```powershell
iwr -useb https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.ps1 | iex
```

### Windows CMD

```cmd
powershell -NoProfile -ExecutionPolicy Bypass -Command "iwr -useb https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.ps1 | iex"
```

运行后按提示操作：

1. 选择配置模式（如检测到现有配置）
2. 输入 **DMXAPI URL**（默认 `https://www.dmxapi.cn`，可省略 `/v1` 后缀）
3. 输入 **API Key**（在 https://www.dmxapi.cn/token 获取）
4. 输入 **模型名称**（多个以英文逗号分隔）
5. 工具自动测试连接、生成并写入配置文件
6. 完成后运行 `opencode` 即可使用 DMXAPI 服务

> 安装脚本会下载与当前平台匹配的发布资产（`opencode-dmxapi-<tag>-<platform>`）到临时目录并直接运行，不会驻留二进制文件。

## 配置模式

当检测到现有 `opencode.json` 时，工具提供两种模式：

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| **完整配置** | 重新配置 URL、API Key 和模型 | 首次配置或更换账号 |
| **仅模型配置** | 保留现有 URL 与 API Key，只更新模型列表 | 快速切换或追加模型 |

## 智能模型路由

工具根据模型名称自动归类到对应 SDK（见 `internal/config/config.go` 中的 `ClassifyModel`）：

| 匹配规则 | Provider ID | NPM 包 | 示例 |
|----------|-------------|--------|------|
| 以 `claude` 开头，或以 `-cc` 结尾 | `dmxapi-anthropic` | `@ai-sdk/anthropic` | `claude-opus-4-5-20251101`、`claude-sonnet-4-20250514`、`my-model-cc` |
| 以 `gemini` 开头 | `dmxapi-google` | `@ai-sdk/google` | `gemini-2.5-pro`、`gemini-2.5-flash` |
| 以 `gpt-5` 开头，或 `o1` / `o3` / `o4` 系列 | `dmxapi-openai-responses` | `@ai-sdk/openai` | `gpt-5`、`gpt-5-mini`、`o1-preview`、`o3-mini`、`o4-mini` |
| 其他 | `dmxapi-openai` | `@ai-sdk/openai-compatible` | `gpt-4o`、`DeepSeek-V3`、`qwen-max` |

**示例：** 配置 `claude-opus-4-5-20251101,gemini-2.5-pro,gpt-5-mini,DeepSeek-V3` 会同时生成 4 个 provider，并自动为 Google provider 使用 `/v1beta` 路径、为其他 provider 使用 `/v1` 路径。

## 配置文件

### 文件位置

| 文件 | Windows | macOS / Linux |
|------|---------|---------------|
| `opencode.json` | `C:\Users\<用户>\.config\opencode\opencode.json` | `~/.config/opencode/opencode.json` |
| `auth.json` | `C:\Users\<用户>\.local\share\opencode\auth.json` | `~/.local/share/opencode/auth.json` |

> 注：所有平台均遵循 XDG Base Directory 规范，包括 Windows。这与 OpenCode 主程序保持一致（详见 [sst/opencode#6156](https://github.com/sst/opencode/issues/6156)），因此本工具未使用 `%APPDATA%`。

写入前若文件已存在，会在同目录下生成备份：`<文件名>.backup.<YYYYMMDD_HHMMSS>`（例如 `opencode.json.backup.20260521_143012`）。

### opencode.json 示例

输入 URL `https://www.dmxapi.cn`（或 `https://www.dmxapi.cn/v1`，工具会自动规范化）配合模型 `claude-opus-4-5-20251101,gemini-2.5-pro,gpt-5-mini,DeepSeek-V3` 时，自动生成：

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
        "baseURL": "https://www.dmxapi.cn/v1beta",
        "apiKey": "sk-xxx"
      },
      "models": {
        "gemini-2.5-pro": { "name": "gemini-2.5-pro" }
      }
    },
    "dmxapi-openai-responses": {
      "npm": "@ai-sdk/openai",
      "name": "DMXAPI OpenAI Responses",
      "options": {
        "baseURL": "https://www.dmxapi.cn/v1",
        "apiKey": "sk-xxx"
      },
      "models": {
        "gpt-5-mini": { "name": "gpt-5-mini" }
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

注意：
- 仅当输入的模型命中对应路由规则时，才会生成相应的 provider 条目（未命中的 provider 不会出现在文件中）。
- Google provider 使用 `/v1beta`，其他 provider 使用 `/v1`，由工具自动拼接，无需手动指定。
- 现有 `opencode.json` 中其它字段（如自定义 provider、`mcp`、`agent` 等）会被保留，仅按 ID 覆盖对应的 `provider` 子项。

### auth.json 示例

与 `opencode.json` 中生成的 provider 一一对应：

```json
{
  "dmxapi-anthropic": { "type": "api", "key": "sk-xxx" },
  "dmxapi-google":    { "type": "api", "key": "sk-xxx" },
  "dmxapi-openai-responses": { "type": "api", "key": "sk-xxx" },
  "dmxapi-openai":    { "type": "api", "key": "sk-xxx" }
}
```

## 从源码构建

```bash
git clone https://cnb.cool/dmxapi/opencode_dmxapi.git
cd opencode_dmxapi

# 当前平台构建（不注入版本号则显示 "dev"）
go build -o opencode-dmxapi .

# 注入版本号（与 release 流水线一致）
go build -ldflags="-s -w -X 'dmxapi-config/internal/ui.Version=2.1.2'" -o opencode-dmxapi .

# 跨平台构建（产物命名与发布资产一致）
GOOS=windows GOARCH=amd64 go build -o opencode-dmxapi-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -o opencode-dmxapi-windows-arm64.exe .
GOOS=darwin  GOARCH=amd64 go build -o opencode-dmxapi-darwin-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o opencode-dmxapi-darwin-arm64 .
GOOS=linux   GOARCH=amd64 go build -o opencode-dmxapi-linux-amd64 .
GOOS=linux   GOARCH=arm64 go build -o opencode-dmxapi-linux-arm64 .
```

需要 Go 1.24 及以上。

## 常见问题

### macOS 提示"无法验证开发者"

macOS Gatekeeper 会阻止运行未经 Apple 公证的二进制。手动下载发布资产时可执行：

```bash
# 推荐：移除隔离属性后再运行
xattr -dr com.apple.quarantine /path/to/下载的可执行文件

# 或在"系统设置 → 隐私与安全性"中点击"仍要打开"
```

一键安装脚本会自动尝试移除隔离属性，通常无需手动处理。

### API Key 无效或连接失败

- 确认 API Key 以 `sk-` 开头，来源 https://www.dmxapi.cn/token
- 检查网络能否访问 https://www.dmxapi.cn
- 自定义 URL 时可不带尾部 `/v1`，工具会自动处理；带或不带都行
- 如使用 Gemini 模型，请确认所选模型在 DMXAPI 端可用（Google provider 使用 `/v1beta` 路径）

### 配置文件在哪里？

见上文 [文件位置](#文件位置) 表格。备份命名为 `<文件名>.backup.<YYYYMMDD_HHMMSS>`，与原文件位于同一目录。

### 误删配置怎么办？

工具不会删除备份文件。进入 `~/.config/opencode/` 或 `~/.local/share/opencode/`，找到最近的 `*.backup.*` 文件还原即可。

## 相关链接

| 资源 | 链接 |
|------|------|
| DMXAPI 官网 | https://www.dmxapi.cn |
| 获取 API Key | https://www.dmxapi.cn/token |
| 可用模型列表 | https://www.dmxapi.cn/rmb |
| OpenCode 文档 | https://opencode.ai |
| 项目仓库 | https://cnb.cool/dmxapi/opencode_dmxapi |
