package main

import (
	"bufio"
	"fmt"
	"os"

	"dmxapi-config/internal/api"
	"dmxapi-config/internal/auth"
	"dmxapi-config/internal/config"
	"dmxapi-config/internal/input"
	"dmxapi-config/internal/ui"
)

func init() {
	// 在旧版 Windows CMD（不支持 ANSI）中提示切换到 UTF-8 代码页
	// 避免中文和 Unicode 字符（如 ╔═║✓）显示为乱码
	if ui.IsLegacyWindowsCMD() {
		fmt.Println("提示: 检测到旧版 Windows CMD，如出现中文或字符乱码，请先运行:")
		fmt.Println("      chcp 65001")
		fmt.Println()
		// 同时提示推荐使用 Windows Terminal 以获得最佳显示效果
		fmt.Println("建议: 使用 Windows Terminal 可获得完整的颜色和字符显示支持。")
		fmt.Println()
	}
}

// waitForExit 等待用户按任意键退出
func waitForExit() {
	fmt.Println()
	fmt.Print("按 Enter 键退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	// 显示欢迎横幅
	ui.PrintBanner()
	ui.PrintDivider()

	// 创建输入收集器
	collector := input.NewCollector()

	// 检测现有配置
	reader := config.NewReader()
	existingConfig := reader.ReadExistingConfig()

	if existingConfig != nil {
		// 存在现有配置，显示选择菜单
		ui.PrintConfigModeHeader()

		mode, err := collector.CollectConfigMode()
		if err != nil {
			ui.PrintError(fmt.Sprintf("选择配置模式失败: %v", err))
			waitForExit()
			os.Exit(1)
		}

		ui.PrintDivider()

		if mode == input.ConfigModeModelOnly {
			// 仅配置模型模式
			runModelOnlyConfiguration(collector, existingConfig)
		} else {
			// 完整配置模式
			runFullConfiguration(collector)
		}
	} else {
		// 不存在配置，直接进入完整配置流程
		runFullConfiguration(collector)
	}

	// 等待用户按键退出
	waitForExit()
}

// collectWithRetry 通用重试包装：循环调用 collect 直到 validate 通过
func collectWithRetry(collect func() (string, error), validate func(string) error) (string, error) {
	for {
		val, err := collect()
		if err != nil {
			return "", err
		}
		if err := validate(val); err != nil {
			ui.PrintWarning(fmt.Sprintf("输入无效：%v，请重新输入。", err))
			continue
		}
		return val, nil
	}
}

// runFullConfiguration 运行完整配置流程
func runFullConfiguration(collector *input.Collector) {
	// 步骤1: 收集URL（验证失败时提示重试）
	ui.PrintStep(1, "配置 DMXAPI URL")
	url, err := collectWithRetry(collector.CollectURL, input.ValidateURL)
	if err != nil {
		ui.PrintError(fmt.Sprintf("读取URL失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("URL 已设置: %s", url))
	fmt.Println()

	// 步骤2: 收集API Key（验证失败时提示重试）
	ui.PrintStep(2, "配置 API Key")
	apiKey, err := collectWithRetry(collector.CollectAPIKey, input.ValidateAPIKey)
	if err != nil {
		ui.PrintError(fmt.Sprintf("读取API Key失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess("API Key 已设置")
	fmt.Println()

	// 步骤3: 收集模型名称（验证失败时提示重试）
	ui.PrintStep(3, "配置模型")
	var models []string
	for {
		var modelsErr error
		models, modelsErr = collector.CollectModels()
		if modelsErr != nil {
			ui.PrintError(fmt.Sprintf("读取模型失败: %v", modelsErr))
			waitForExit()
			os.Exit(1)
		}
		if err := input.ValidateModels(models); err != nil {
			ui.PrintWarning(fmt.Sprintf("输入无效：%v，请重新输入。", err))
			continue
		}
		break
	}
	ui.PrintSuccess(fmt.Sprintf("已添加 %d 个模型: %v", len(models), models))
	fmt.Println()

	// 步骤4: 测试API连接（使用用户配置的第一个模型）
	ui.PrintStep(4, "测试 API 连接")
	ui.PrintInfo("正在测试连接...")
	tester := api.NewTester(url, apiKey)
	if err := tester.TestConnection(models[0]); err != nil {
		ui.PrintError(fmt.Sprintf("API 连接测试失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess("API 连接测试成功！")
	fmt.Println()

	ui.PrintDivider()
	ui.PrintInfo("正在写入配置文件...")
	fmt.Println()

	// 步骤5: 写入认证配置
	ui.PrintStep(5, "配置认证信息")
	cfg := config.NewDMXAPIConfig(url, apiKey, models)
	providerIDs := config.GetProviderIDs(cfg)
	authMgr := auth.NewAuthManager(providerIDs, apiKey)
	authPath, err := authMgr.Login()
	if err != nil {
		ui.PrintError(fmt.Sprintf("认证配置失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("认证配置完成: %s", authPath))
	fmt.Println()

	// 步骤6: 生成opencode.json配置文件
	ui.PrintStep(6, "生成配置文件")
	writer := config.NewWriter()
	configPath, err := writer.WriteConfig(cfg)
	if err != nil {
		ui.PrintError(fmt.Sprintf("写入配置失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("配置文件已生成: %s", configPath))
	fmt.Println()

	// 完成
	ui.PrintDivider()
	ui.PrintComplete()

	// 显示配置摘要
	fmt.Println("配置摘要:")
	fmt.Printf("  - URL: %s/v1\n", url)
	fmt.Printf("  - 模型: %v\n", models)
	fmt.Printf("  - 配置文件: %s\n", configPath)
	fmt.Printf("  - 认证文件: %s\n", authPath)
	fmt.Println()
	fmt.Println("运行 'opencode' 启动程序")
}

// runModelOnlyConfiguration 运行仅配置模型流程
func runModelOnlyConfiguration(collector *input.Collector, existing *config.ExistingConfig) {
	ui.PrintModelOnlyModeInfo()

	// 显示当前配置信息
	ui.PrintExistingConfigInfo(existing.URL, config.MaskAPIKey(existing.APIKey), existing.Models)

	// 步骤1: 收集新模型名称（验证失败时提示重试）
	ui.PrintStep(1, "配置模型")
	var models []string
	for {
		var modelsErr error
		models, modelsErr = collector.CollectModels()
		if modelsErr != nil {
			ui.PrintError(fmt.Sprintf("读取模型失败: %v", modelsErr))
			waitForExit()
			os.Exit(1)
		}
		if err := input.ValidateModels(models); err != nil {
			ui.PrintWarning(fmt.Sprintf("输入无效：%v，请重新输入。", err))
			continue
		}
		break
	}
	ui.PrintSuccess(fmt.Sprintf("已添加 %d 个模型: %v", len(models), models))
	fmt.Println()

	ui.PrintDivider()
	ui.PrintInfo("正在写入配置文件...")
	fmt.Println()

	// 步骤2: 更新认证配置（保持不变，但确保文件存在）
	ui.PrintStep(2, "更新认证信息")
	cfg := config.NewDMXAPIConfig(existing.URL, existing.APIKey, models)
	providerIDs := config.GetProviderIDs(cfg)
	authMgr := auth.NewAuthManager(providerIDs, existing.APIKey)
	authPath, err := authMgr.Login()
	if err != nil {
		ui.PrintError(fmt.Sprintf("认证配置失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("认证配置完成: %s", authPath))
	fmt.Println()

	// 步骤3: 生成opencode.json配置文件（使用现有URL和APIKey，新模型）
	ui.PrintStep(3, "生成配置文件")
	writer := config.NewWriter()
	configPath, err := writer.WriteConfig(cfg)
	if err != nil {
		ui.PrintError(fmt.Sprintf("写入配置失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("配置文件已生成: %s", configPath))
	fmt.Println()

	// 完成
	ui.PrintDivider()
	ui.PrintComplete()

	// 显示配置摘要
	fmt.Println("配置摘要:")
	fmt.Printf("  - URL: %s/v1\n", existing.URL)
	fmt.Printf("  - 模型: %v\n", models)
	fmt.Printf("  - 配置文件: %s\n", configPath)
	fmt.Printf("  - 认证文件: %s\n", authPath)
	fmt.Println()
	fmt.Println("运行 'opencode' 启动程序")
}
