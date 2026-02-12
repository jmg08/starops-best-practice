// 交互式对话示例
// Interactive Chat Example
//
// 用法 / Usage: go run ./cmd/chat/
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
	fmt.Println("🚀 VibeOps Chat")
	fmt.Println(strings.Repeat("=", 60))

	// 加载配置
	cfg, err := client.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		fmt.Println("\n请设置环境变量:")
		fmt.Println("  VIBEOPS_WORKSPACE, VIBEOPS_ENDPOINT")
		fmt.Println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")
		os.Exit(1)
	}

	fmt.Printf("📋 Workspace: %s\n", cfg.Workspace)
	fmt.Printf("📋 Employee: %s\n\n", cfg.EmployeeName)

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
	fmt.Printf("✅ ThreadID: %s\n\n", threadID)

	// 创建打印器
	printer := client.NewSimplePrinter()

	// 交互循环
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("👤 请输入 (quit 退出): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n� 再见!")
				break
			}
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

		fmt.Println(strings.Repeat("-", 60))

		// 发送消息
		printer.Reset()
		events := agentClient.Chat(ctx, threadID, input)

		for event := range events {
			if event.Error != nil {
				fmt.Printf("❌ 错误: %v\n", event.Error)
				continue
			}
			text := printer.ProcessEvent(event)
			if text != "" {
				fmt.Print(text)
			}
		}

		fmt.Println()
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()
	}
}
