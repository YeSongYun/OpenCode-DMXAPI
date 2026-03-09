package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dmxapi-config/internal/api"
	"dmxapi-config/internal/auth"
	"dmxapi-config/internal/config"
	"dmxapi-config/internal/input"
	"dmxapi-config/internal/ui"
)

func init() {
	// 在旧版 Windows CMD（不支持 ANSI）中提示切换到 UTF-8 代码页
	if ui.IsLegacyWindowsCMD() {
		fmt.Println("提示: 检测到旧版 Windows CMD，如出现中文或字符乱码，请先运行:")
		fmt.Println("      chcp 65001")
		fmt.Println()
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
	ui.PrintBanner()

	// 启动异步检查更新（在后续操作耗时期间并行进行 HTTP 请求）
	updateCh := ui.CheckForUpdateAsync()

	ui.PrintDivider()

	// 检测 opencode 是否已安装
	installed, version := ui.CheckOpencode()
	if installed {
		if version != "" {
			ui.PrintSuccess(fmt.Sprintf("已检测到 opencode 已安装（版本：%s）", version))
		} else {
			ui.PrintSuccess("已检测到 opencode 已安装")
		}
		fmt.Println()
	} else {
		ui.PrintWarning("未检测到 opencode，请先安装后再使用本工具")
		ui.PrintInfo("官网地址：https://opencode.ai")
		waitForExit()
		os.Exit(1)
	}

	// 展示更新提示（非阻塞：已有结果就显示，网络慢则跳过）
	select {
	case result := <-updateCh:
		if result.HasUpdate {
			ui.PrintUpdateNotice(result.LatestVersion, result.DownloadURL)
		}
	default:
		// 结果尚未返回，继续不等待
	}

	collector := input.NewCollector()
	reader := config.NewReader()
	existingConfig := reader.ReadExistingConfig()

	if existingConfig != nil {
		ui.PrintExistingConfigInfo(existingConfig.URL, config.MaskAPIKey(existingConfig.APIKey), existingConfig.Models)
		ui.PrintConfigModeHeader()

		mode, err := collector.CollectConfigMode()
		if err != nil {
			ui.PrintError(fmt.Sprintf("选择配置模式失败: %v", err))
			waitForExit()
			os.Exit(1)
		}

		ui.PrintDivider()

		if mode == input.ConfigModeModelOnly {
			runModelOnlyConfiguration(collector, existingConfig)
		} else {
			runFullConfiguration(collector)
		}
	} else {
		runFullConfiguration(collector)
	}

	waitForExit()
}

// runFullConfiguration 运行完整配置流程（6步）
func runFullConfiguration(collector *input.Collector) {
	// [1/6] 配置URL
	ui.PrintStep(1, 6, "配置 DMXAPI URL")
	url, err := collector.CollectURL()
	if err != nil {
		ui.PrintError(fmt.Sprintf("读取URL失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("URL 已设置: %s", url))
	fmt.Println()

	// [2/6] 配置API Key
	ui.PrintStep(2, 6, "配置 API Key")
	apiKey, err := collector.CollectAPIKey()
	if err != nil {
		ui.PrintError(fmt.Sprintf("读取API Key失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess("API Key 已设置")
	fmt.Println()

	// [3/6] 配置模型
	ui.PrintStep(3, 6, "配置模型")
	models, err := collector.CollectModels()
	if err != nil {
		ui.PrintError(fmt.Sprintf("读取模型失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("已添加 %d 个模型", len(models)))
	fmt.Println()

	// [4/6] 测试API连接
	ui.PrintStep(4, 6, "测试 API 连接")
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

	// [5/6] 配置认证信息
	ui.PrintStep(5, 6, "配置认证信息")
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

	// [6/6] 生成配置文件
	ui.PrintStep(6, 6, "生成配置文件")
	writer := config.NewWriter()
	configPath, err := writer.WriteConfig(cfg)
	if err != nil {
		ui.PrintError(fmt.Sprintf("写入配置失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("配置文件已生成: %s", configPath))
	fmt.Println()

	ui.PrintDivider()
	ui.PrintComplete()

	fmt.Println("  配置摘要:")
	fmt.Printf("    URL     %s\n", config.NormalizeBaseURL(url))
	fmt.Printf("    模型    %s\n", strings.Join(models, ", "))
	fmt.Printf("    配置    %s\n", configPath)
	fmt.Printf("    认证    %s\n", authPath)
	fmt.Println()
	fmt.Println("  运行 'opencode' 启动程序")
}

// runModelOnlyConfiguration 运行仅配置模型流程（3步）
func runModelOnlyConfiguration(collector *input.Collector, existing *config.ExistingConfig) {
	ui.PrintModelOnlyModeInfo()

	// [1/3] 配置模型
	ui.PrintStep(1, 3, "配置模型")
	models, err := collector.CollectModels()
	if err != nil {
		ui.PrintError(fmt.Sprintf("读取模型失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("已添加 %d 个模型", len(models)))
	fmt.Println()

	ui.PrintDivider()
	ui.PrintInfo("正在写入配置文件...")
	fmt.Println()

	// [2/3] 更新认证信息
	ui.PrintStep(2, 3, "更新认证信息")
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

	// [3/3] 生成配置文件
	ui.PrintStep(3, 3, "生成配置文件")
	writer := config.NewWriter()
	configPath, err := writer.WriteConfig(cfg)
	if err != nil {
		ui.PrintError(fmt.Sprintf("写入配置失败: %v", err))
		waitForExit()
		os.Exit(1)
	}
	ui.PrintSuccess(fmt.Sprintf("配置文件已生成: %s", configPath))
	fmt.Println()

	ui.PrintDivider()
	ui.PrintComplete()

	fmt.Println("  配置摘要:")
	fmt.Printf("    URL     %s\n", config.NormalizeBaseURL(existing.URL))
	fmt.Printf("    模型    %s\n", strings.Join(models, ", "))
	fmt.Printf("    配置    %s\n", configPath)
	fmt.Printf("    认证    %s\n", authPath)
	fmt.Println()
	fmt.Println("  运行 'opencode' 启动程序")
}
