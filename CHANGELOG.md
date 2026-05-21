# 更新日志

本文件记录 opencode_dmxapi 各版本变更，遵循 [Conventional Commits](https://www.conventionalcommits.org/) 与 `cliff.toml` 分组约定。



## 🐛 修复
- ✅ 确保更新日志中的日期准确可靠，修正了之前版本的日期错误

## 📚 文档
- 📝 已更新至 v2.0.7 的更新记录 (2026-05-21)

## v2.0.7 (2026-05-21)
### ✨ 新功能
- 发布新版本时，系统会自动用AI优化更新说明，让描述更友好易懂
### 🐛 修复
- 修复了版本更新记录无法正常保存到代码仓库的问题
### 📚 文档
- 增加了从v1.0到v2.0所有版本的完整更新历史汇总

## [v2.0.6] - 2026-03-10

### ✨ 新功能

- darwin→macos 重命名，README/YML 安装说明详细化（v2.0.5）
  - 构建产物文件名 `darwin-{amd64,arm64}` → `macos-{amd64,arm64}`（.cnb.yml & release.yml）
  - macOS 安装说明拆分 Intel/Apple Silicon，各加 `xattr` 隔离属性移除命令
  - Linux 安装说明补充 ARM64 说明
  - README.md 新增"系统要求"章节、按平台展开"快速开始"、新增"常见问题"章节
- 检查更新改用 CNB API（替换 GitHub API）
- 添加检查更新功能（GitHub Releases API）
  - 新增 `internal/ui/updater.go`：异步 HTTP 检测最新 Release
  - `display.go` 新增 `PrintUpdateNotice` 打印新版本提示
  - `main.go` 在 PrintBanner 后启动异步检查，非阻塞展示结果
- 检测到现有配置时先展示当前配置信息
  - `PrintExistingConfigInfo` 标题改为"检测到现有 DMXAPI 配置"
  - 模型列表用逗号分隔替代 `%v` 格式
  - 在配置模式选择菜单前展示配置信息
- 用 ASCII Art block 字符替换 PrintBanner 图标
  - 新增 DMXAPI 六字 block 字符 ASCII Art（██ 风格）
  - 新增 `Version` 常量，方便后续统一修改
  - 副标题动态获取 `runtime.GOOS/GOARCH`
- Windows 下实现真正的上下键交互菜单
  - 通过 `syscall` 调用 Windows API（kernel32.dll）替代原数字输入 fallback
  - `GetStdHandle / SetConsoleMode` 进入原始输入模式
  - 对 stdout 开启 `ENABLE_VIRTUAL_TERMINAL_PROCESSING`（Windows 10+）
  - `ReadConsoleInputW` 捕获 `VK_UP/VK_DOWN/VK_RETURN/Ctrl+C`
- 配置模式改为上下键交互选择菜单
- 启动时检测 opencode 是否已安装；未安装时提示访问 https://opencode.ai 并退出
- 补全 4 种 provider 的联通性测试接口
  - Google Gemini：`/v1beta/models/<model>:generateContent`，key 通过 URL 参数传递
  - OpenAI Responses：`/v1/responses`，`input` 字段，检查 HTTP 200
  - Anthropic / OpenAI：保持原有逻辑

### 🐛 修复

- 修复 7 个多平台兼容性问题
  - `display.go`：新增 `supportsUnicode()` 检测（Windows 检查代码页 65001），为旧版 CMD 提供 ASCII 降级映射（✓→[OK] ✗→[X] 等）
  - `display.go`：`supportsColor()` 新增 COLORTERM、Git Bash/MSYS2、xterm/cygwin TERM 前缀检测
  - `display_windows.go`：首次调用时主动启用 `ENABLE_VIRTUAL_TERMINAL_PROCESSING`，sync.Once 缓存检测结果
  - `collector.go`：huh 调用增加 `isTTYError()` 判断，失败时 fallback 到 bufio 文本输入与数字选择
  - `collector.go`：使用 `mattn/go-isatty` 检测 TTY，非 TTY 环境直接走 fallback
  - `path.go + writer.go`：Windows 权限警告移至 `EnsureDir`，sync.Once 保证只输出一次
- 消除 `bufio.Reader` 与 raw mode 的 stdin 争抢，修复 macOS 方向键
  - 移除包级 `stdinReader`，避免启动时即缓冲 `os.Stdin` 数据
  - `selectMenuImpl` 进入 raw mode 后先 `drainStdin()` 清除残留
  - `makeTerminalRaw` 设置后读回 termios 验证 VMIN/VTIME 正确性
- 重新设计 CLI 界面风格并修复 macOS 终端方向键失效
  - 移除 box drawing 字符（╔═║╚），改用 emoji+文字的现代简洁风格
  - VMIN/VTIME 从超时模式（0/1）改为阻塞模式（1/0），ESC 序列改逐字节读取
- baseURL 规范化支持 `/v1beta` 等版本后缀
  - 新增 `NormalizeBaseURL`，正则去除末尾版本路径后缀（`/v1`、`/v1beta`、`/v1beta1`），再按 provider 拼接
- Google provider baseURL 使用 `/v1beta` 而非 `/v1`（匹配 `@ai-sdk/google` SDK 预期）
- Gemini API URL 路径从 `v1` 改为 `v1beta`，支持更多模型和最新功能
- 修复 URL 双重拼接、API 响应大小限制、配置合并等 7 个问题
  - `NewTester` 和配置摘要中规范化 URL，避免 `/v1` 双重拼接
  - `opencode --version` 使用 `exec.CommandContext` 加 10 秒超时
  - API 响应读取使用 `io.LimitReader` 限制最大 1MB
  - `WriteAuth` 写入前先备份现有认证文件
  - 配置合并使用 `map[string]interface{}` 保留未知 JSON 字段
- 修复终端阻塞、URL 重复拼接和配置文件权限三个高优先级 bug
  - `menu_term_darwin.go`：VMIN=0/VTIME=1 超时模式，避免单独按 ESC 时 read 永久阻塞
  - `config.go`：使用 `TrimSuffix` 防止 URL 已含 `/v1` 时重复拼接为 `/v1/v1`
  - `writer.go`：配置文件和备份文件权限从 0644 改为 0600，防止 API Key 泄露
- 全面代码审查和多平台兼容性修复
  - `api/test.go`：Google Gemini 测试加 `Authorization: Bearer` 头，URL 改用 `/v1/models/`
  - OpenAI Responses API 测试增加响应体解析和 output 非空验证
  - `GeminiResponse.Candidates` 改为具名结构体 `GeminiCandidate`，新增 `OpenAIResponsesResponse`
  - `menu_term_*.go`：VMIN=1,VTIME=0 改为 VMIN=0,VTIME=1（100ms 超时），修复孤立 ESC 阻塞
- Claude 模型使用 `/v1/messages` 接口测试联通性
  - 通过 `config.ClassifyModel` 区分 provider，claude 走 Anthropic Messages API（含 `max_tokens`）

### ♻️ 重构

- 修复 URL 规范化、Google model URL 编码，删除 3 个死代码函数
  - `reader.go`：用 `NormalizeBaseURL()` 替换手写的 `/v1` 后缀截断逻辑
  - `test.go`：Google API URL 中对 model 名做 `url.PathEscape`，防止含 `/` 的模型名破坏 URL 路径
  - 删除死代码 `PrintSystemInfo()` / `HasExistingConfig()` / `GetOS()`
- 用 `charmbracelet/huh` 替代手写终端控制代码
  - 删除 6 个平台相关的 `menu*.go` 文件（手写 termios/raw mode）
  - 重写 `collector.go`：Select/Input 基于 huh，内置验证和密码模式
  - `display.go`：`PrintStep` 改为 `[N/M]` 格式
  - 彻底解决 macOS 方向键失效问题

### 🔧 杂项

- 版本号更新至 v2.0.4 / v2.0.6

## [v2.0.3] - 2026-03-09

### 🐛 修复

- 联通性测试改用用户选择的模型，而非硬编码 `claude-haiku-4-5`
  - `TestConnection()` 增加 model 参数
  - 全量配置流程中，将"配置模型"移到"测试连接"之前
  - 测试时传入 `models[0]`，确保测试与实际使用模型一致
- 修复 ANSI 兼容性、模型分类、清理死代码等问题
  - `ui/display.go`：`PrintStep/Success/Error/Info/Warning` 改为通过 `colorize()` 输出，旧版 Windows CMD 不再显示 ANSI 乱码
  - `config.go`：`ClassifyModel()` 增加 o1/o3/o4 系列推理模型检测，路由到 `@ai-sdk/openai`（responses 格式）
  - `go.mod`：版本从不存在的 1.25.4 修正为与 CI 一致的 1.24
  - `internal/api/test.go`：测试模型从 `claude-opus-4-5-20251101` 改为更稳定的 `claude-haiku-4-5`
  - 删除从未被调用的 `CollectAll()`
  - `main.go`：输入验证失败时提示重新输入，而非直接退出

### 🔧 杂项

- 从 git 移除 `dmxapi-config.exe` 并完善 `.gitignore`
  - `git rm --cached` 移除已追踪的二进制
  - 添加 `*.exe` 通配规则
  - 补充各平台二进制文件忽略规则（`opencode-dmxapi-*` 等）

## [v2.0.2] - 2026-03-08

### 🐛 修复

- 修复 Windows 多平台兼容性问题
  - `ui/display.go`：实现真正的 `supportsColor()` 检测（NO_COLOR、WT_SESSION、ANSICON、ConEmuANSI），旧版 CMD 返回 false 避免乱码；导出 `IsLegacyWindowsCMD()`
  - `config/writer.go`：为 `auth.json` 添加 Windows 权限警告（Windows 忽略 0600 权限位）
  - `main.go`：检测旧版 CMD 并提示 `chcp 65001`，在 `init()` 中早于业务逻辑
- gpt-5 系列模型使用 responses 格式（`@ai-sdk/openai`）

## [v2.0.1] - 2026-01-26

### ✨ 新功能

- 更新日志显示完整 commit body 信息
  - 在 `cliff.toml` 中添加 `commit.body` 显示逻辑，Release Notes 包含详细描述（如子项列表）
- 实现基于模型名称路由的多 provider 配置
  - `claude-*` → `@ai-sdk/anthropic`（dmxapi-anthropic）
  - `gemini-*` → `@ai-sdk/google`（dmxapi-google）
  - 其他 → `@ai-sdk/openai-compatible`（dmxapi-openai）
  - 支持多 provider 认证配置，兼容读取新旧配置格式

### ♻️ 重构

- 移除配置中的 `$schema` 字段
  - 从 `OpenCodeConfig` 结构体移除 `Schema` 字段
  - 移除 `NewDMXAPIConfig` / `mergeConfigs` 中的相关赋值与合并逻辑

### 📚 文档

- 重写 README，突出智能模型路由和多 Provider 配置
  - 添加徽章提升专业感
  - 新增智能模型路由说明、配置模式说明
  - 更新配置示例为多 Provider 格式，使用表格提高可读性
  - 添加 OpenCode 文档链接

### 🔧 杂项

- 将 `dmxapi-config.exe` 添加到 `.gitignore`，排除根目录编译生成的二进制

## [v1.2.6] - 2026-01-19

### ✨ 新功能

- 添加配置模式选择功能
  - 支持"完整配置"和"仅配置模型"两种模式
  - 检测到现有配置时允许用户选择配置模式
  - 显示当前配置信息并支持部分更新

## [v1.2.5] - 2026-01-15

### 👷 CI

- 优化发布流程并添加安装说明

## [v1.2.4] - 2026-01-15

### 👷 CI

- 优化发布流程并添加安装说明

## [v1.2.3] - 2026-01-15

### 👷 CI

- 添加 GitHub Actions 发布工作流
  - 跨平台构建和发布工作流，支持 Windows、Linux、macOS 多种架构
  - 自动生成变更日志并创建 GitHub Release

## [v1.2.1] - 2026-01-15

### 👷 CI

- 将 Go 构建镜像从 1.21 升级到 latest

## [v1.2.0] - 2026-01-15

### ✨ 新功能

- 添加云原生构建配置和更新日志生成配置

### 🔧 杂项

- 在 `.gitignore` 中添加 `nul` 文件
- 更新 `.gitignore` 忽略根目录二进制文件

## [v1.0.0] - 2026-01-12

### ✨ 新功能

- 添加 DMXAPI 配置工具
  - 用户输入收集、配置验证、配置文件生成和认证管理
  - 跨平台配置文件路径处理，友好的命令行交互界面
- 添加 API 连接测试功能
  - 在配置流程中增加 API 连接测试步骤，验证 API 密钥和 URL 的有效性

### ♻️ 重构

- 将 `dmxapi-config` 目录结构调整为根目录

### 📚 文档

- 添加 DMXAPI 配置工具的 README 文档
- 更新示例模型名称和连接测试信息

### 🔧 杂项

- 添加 `jihua` 到 `.gitignore`

---

> 本日志根据 git 历史汇总。CI 在每次发布时通过 [git-cliff](https://github.com/orhun/git-cliff) 自动生成最新版本的 Release Notes（见 `.cnb.yml`）。
