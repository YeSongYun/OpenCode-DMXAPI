# DMXAPI 配置工具

一键配置 OpenCode 使用 DMXAPI 服务的命令行工具。

## 功能

- 交互式配置向导
- API 连接测试（自动验证 API Key 是否有效）
- 自动生成 opencode.json 配置文件
- 自动写入 auth.json 认证信息
- 跨平台支持（Windows/macOS/Linux）

## 下载

| 平台 | 文件名 |
|------|--------|
| Windows | `dmxapi-config-windows.exe` |
| macOS (Intel) | `dmxapi-config-macos-amd64` |
| macOS (Apple Silicon) | `dmxapi-config-macos-arm64` |
| Linux | `dmxapi-config-linux` |

## 使用方法

1. 下载对应平台的可执行文件
2. 运行程序
3. 按提示输入：
   - **DMXAPI URL**（默认: https://www.dmxapi.cn）
   - **API Key**（从 https://www.dmxapi.cn/token 获取）
   - **模型名称**（多个用逗号分隔，如: `claude-opus-4-5-20251101,DeepSeek-V3.2-Fast`）
4. 程序会自动测试连接并生成配置文件
5. 运行 `opencode` 启动程序

## 配置流程

```
[步骤 1] 配置 DMXAPI URL
[步骤 2] 配置 API Key
[步骤 3] 测试 API 连接
[步骤 4] 配置模型
[步骤 5] 配置认证信息
[步骤 6] 生成配置文件
```

## 配置文件位置

| 文件 | Windows | macOS/Linux |
|------|---------|-------------|
| opencode.json | `C:\Users\<用户>\.config\opencode\opencode.json` | `~/.config/opencode/opencode.json` |
| auth.json | `C:\Users\<用户>\.local\share\opencode\auth.json` | `~/.local/share/opencode/auth.json` |

## 生成的配置格式

**opencode.json**:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "dmxapi": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "DMXAPI",
      "options": {
        "baseURL": "https://www.dmxapi.cn/v1",
        "apiKey": "sk-xxx"
      },
      "models": {
        "claude-opus-4-5-20251101": { "name": "claude-opus-4-5-20251101" },
        "DeepSeek-V3.2-Fast": { "name": "DeepSeek-V3.2-Fast" }
      }
    }
  }
}
```

**auth.json**:
```json
{
  "dmxapi": {
    "type": "api",
    "key": "sk-xxx"
  }
}
```

## 从源码构建

```bash
# 克隆项目
cd dmxapi-config

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

- DMXAPI 官网: https://www.dmxapi.cn
- 获取 API Key: https://www.dmxapi.cn/token
- 可用模型列表: https://www.dmxapi.cn/rmb
