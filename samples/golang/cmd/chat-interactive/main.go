// 交互事件处理示例
// 演示如何处理 CMS Agent 返回的各类交互事件
// 用法: go run ./cmd/chat-interactive/
//
// 本示例展示:
// - 处理 user_ack (用户确认) 事件
// - 处理 user_select (用户选择) 事件
// - 处理 user_input (用户输入) 事件
// - 使用交互响应恢复对话
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vibeops/samples/golang/internal/client"
	"github.com/vibeops/samples/golang/types"
)

// 预设问题，可能触发交互事件
var presetQuestions = []struct {
	ID          string
	Description string
	Question    string
}{
	{
		ID:          "1",
		Description: "查询 SLS 日志（可能触发 logstore 选择）",
		Question:    "查询最近一小时的错误日志",
	},
	{
		ID:          "2",
		Description: "执行危险操作（可能触发确认）",
		Question:    "删除所有过期的告警规则",
	},
	{
		ID:          "3",
		Description: "查询指标（可能触发指标选择）",
		Question:    "查询 ECS 实例的 CPU 使用率",
	},
	{
		ID:          "4",
		Description: "模糊查询（可能触发输入补充）",
		Question:    "查询服务的性能数据",
	},
	{
		ID:          "5",
		Description: "多选项查询（可能触发选择）",
		Question:    "帮我分析系统问题",
	},
}

func main() {
	fmt.Println("🚀 VibeOps Chat 交互事件处理示例")
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

	// 创建简洁打印器
	simplePrinter := client.NewSimplePrinter()

	// 创建交互处理器 (60秒超时)
	interactiveHandler := client.NewInteractiveHandler(agentClient, 60*time.Second)

	// 打印帮助
	printHelp()

	// 交互循环
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n👤 请输入 (help 查看帮助): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n👋 再见!")
				break
			}
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch input {
		case "quit", "exit":
			fmt.Println("👋 再见!")
			return
		case "help":
			printHelp()
			continue
		case "list":
			printPresetQuestions()
			continue
		case "1", "2", "3", "4", "5":
			// 使用预设问题
			for _, q := range presetQuestions {
				if q.ID == input {
					input = q.Question
					fmt.Printf("📝 使用预设问题: %s\n", input)
					break
				}
			}
		}

		fmt.Println(strings.Repeat("-", 60))

		// 发送消息并处理响应
		simplePrinter.Reset()
		events := agentClient.Chat(ctx, threadID, input)
		processEventsWithInteraction(ctx, events, simplePrinter, interactiveHandler, threadID)

		// 输出最终文本
		finalText := simplePrinter.GetFinalText()
		if finalText != "" {
			fmt.Printf("\n📄 回复:\n%s\n", finalText)
		}

		fmt.Println(strings.Repeat("=", 60))
	}
}

// processEventsWithInteraction 处理事件流，包括交互事件
func processEventsWithInteraction(
	ctx context.Context,
	events <-chan *client.ChatEvent,
	printer *client.SimplePrinter,
	handler *client.InteractiveHandler,
	threadID string,
) {
	for event := range events {
		if event.Error != nil {
			fmt.Printf("❌ 错误: %v\n", event.Error)
			continue
		}

		// 提取文本
		text := printer.ProcessEvent(event)
		if text != "" {
			fmt.Print(text)
		}

		// 检查是否有交互事件需要处理
		if event.Body != nil && len(event.Body.Messages) > 0 {
			for _, msg := range event.Body.Messages {
				msgJSON, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				var messageItem types.MessageItem
				if err := json.Unmarshal(msgJSON, &messageItem); err != nil {
					continue
				}

				// 提取交互事件
				interactiveEvents := client.ExtractInteractiveEvents(&messageItem)
				for _, interactiveEvent := range interactiveEvents {
					fmt.Println("\n" + strings.Repeat("*", 60))
					fmt.Println("🔔 检测到交互事件!")
					fmt.Printf("   类型: %s\n", interactiveEvent.Type)
					fmt.Println(strings.Repeat("*", 60))

					// 处理交互事件
					response, err := handler.HandleEvent(ctx, interactiveEvent)
					if err != nil {
						fmt.Printf("❌ 处理交互事件失败: %v\n", err)
						continue
					}

					fmt.Printf("\n✅ 交互响应:\n")
					fmt.Printf("   ID: %s\n", response.InteractionID)
					fmt.Printf("   类型: %s\n", response.Type)
					fmt.Printf("   响应: %v\n", response.Response)

					// 使用交互响应恢复对话
					fmt.Println("\n📤 恢复对话...")
					resumeEvents := handler.ResumeChat(ctx, threadID, response)

					// 递归处理恢复后的事件流
					processEventsWithInteraction(ctx, resumeEvents, printer, handler, threadID)
				}
			}
		}
	}
}

func printHelp() {
	fmt.Println("\n📖 命令帮助:")
	fmt.Println("  help  - 显示帮助")
	fmt.Println("  list  - 显示预设问题列表")
	fmt.Println("  1-5   - 使用预设问题（可能触发交互事件）")
	fmt.Println("  quit  - 退出")
	fmt.Println("  <msg> - 发送自定义消息")
}

func printPresetQuestions() {
	fmt.Println("\n📋 预设问题（可能触发交互事件）:")
	fmt.Println(strings.Repeat("-", 60))
	for _, q := range presetQuestions {
		fmt.Printf("  [%s] %s\n", q.ID, q.Description)
		fmt.Printf("      问题: %s\n", q.Question)
	}
	fmt.Println(strings.Repeat("-", 60))
}

func printConfig(cfg *client.Config) {
	fmt.Printf("📋 配置:\n")
	fmt.Printf("  Workspace: %s\n", cfg.Workspace)
	fmt.Printf("  Employee: %s\n\n", cfg.EmployeeName)
}

func printEnvHelp() {
	fmt.Println("\n请设置环境变量:")
	fmt.Println("  VIBEOPS_WORKSPACE, VIBEOPS_ENDPOINT")
	fmt.Println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")
}
