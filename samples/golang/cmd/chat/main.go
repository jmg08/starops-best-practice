// 交互式对话示例
// 用法: go run ./cmd/chat/
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vibeops/samples/golang/internal/client"
)

func main() {
	fmt.Println("🚀 VibeOps Chat 交互式对话")
	fmt.Println(strings.Repeat("=", 60))

	// 加载配置
	cfg, err := client.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		printEnvHelp()
		os.Exit(1)
	}

	printConfig(cfg)

	// 创建客户端
	agentClient, err := client.NewAgentClient(cfg)
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 创建会话
	fmt.Println("📝 创建会话...")
	threadID, err := agentClient.CreateThread(ctx)
	if err != nil {
		fmt.Printf("❌ 创建会话失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 会话创建成功, ThreadID: %s\n\n", threadID)

	// 创建事件打印器
	printer := client.NewEventPrinter(true, true)

	// 交互循环
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("👤 请输入消息 (输入 'quit' 退出): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n👋 再见!")
				break
			}
			fmt.Printf("❌ 读取输入失败: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println("👋 再见!")
			break
		}

		fmt.Printf("\n🤖 发送消息: %s\n", input)
		fmt.Println(strings.Repeat("-", 60))

		events := agentClient.Chat(ctx, threadID, input)
		eventIndex := 0
		for event := range events {
			eventIndex++
			printer.PrintEvent(event, eventIndex)
		}

		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()
	}
}

func printConfig(cfg *client.Config) {
	fmt.Printf("📋 配置信息:\n")
	fmt.Printf("  - Workspace: %s\n", cfg.Workspace)
	fmt.Printf("  - Endpoint: %s\n", cfg.Endpoint)
	fmt.Printf("  - Region: %s\n", cfg.Region)
	fmt.Printf("  - Employee: %s\n\n", cfg.EmployeeName)
}

func printEnvHelp() {
	fmt.Println("\n请设置以下环境变量:")
	fmt.Println("  - VIBEOPS_WORKSPACE: 工作空间")
	fmt.Println("  - VIBEOPS_ENDPOINT: API端点")
	fmt.Println("  - VIBEOPS_REGION: 区域 (可选)")
	fmt.Println("  - ALIBABA_CLOUD_ACCESS_KEY_ID: Access Key ID")
	fmt.Println("  - ALIBABA_CLOUD_ACCESS_KEY_SECRET: Access Key Secret")
	fmt.Println("  - VIBEOPS_EMPLOYEE_NAME: 数字员工名称 (可选)")
}
