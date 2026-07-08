// 交互式对话示例
// Interactive Chat Example
//
// 用法 / Usage: go run ./cmd/chat/
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vibeops/samples/golang/internal/client"
	"github.com/vibeops/samples/golang/types"
)

func main() {
	var simulateError bool
	flag.BoolVar(&simulateError, "simulate-error", false, "模拟网络断连，测试重试逻辑")
	flag.Parse()

	fmt.Println("🚀 VibeOps Chat")
	fmt.Println(strings.Repeat("=", 60))

	// 加载配置
	cfg, err := client.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		fmt.Println("\n请设置环境变量:")
		fmt.Println("  VIBEOPS_ENDPOINT")
		fmt.Println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")
		os.Exit(1)
	}

	fmt.Printf("📋 Employee: %s\n\n", cfg.EmployeeName)

	if simulateError {
		cfg.SimulateNetworkError = true
		fmt.Println("⚠️  已启用网络断连模拟，将在收到首个事件后触发重试")
	}

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
	interactiveHandler := client.NewInteractiveHandler(agentClient, 0)

	// 交互循环
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("👤 请输入 (quit 退出): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n 再见!")
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

		processChatEvents(events, printer, interactiveHandler, ctx, threadID)

		fmt.Println()
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()
	}
}

// processChatEvents 处理 SSE 事件流，支持交互事件检测和流恢复
func processChatEvents(
	events <-chan *client.ChatEvent,
	printer *client.SimplePrinter,
	handler *client.InteractiveHandler,
	ctx context.Context,
	threadID string,
) {
	for events != nil {
		event, ok := <-events
		if !ok {
			break
		}

		if event.Error != nil {
			fmt.Printf("❌ 错误: %v\n", event.Error)
			continue
		}

		// 正常输出（先输出，确保交互事件内容可见）
		text := printer.ProcessEvent(event)
		if text != "" {
			fmt.Print(text)
		}

		// 检测交互事件（在输出之后，确保用户看到交互内容）
		interactiveResp := extractChatInteractiveEvent(event, handler)
		if interactiveResp != nil {
			variables := map[string]any{}
			events = handler.ResumeChat(ctx, threadID, interactiveResp, variables)
			continue
		}
	}
}

// extractChatInteractiveEvent 从 ChatEvent 中检测交互事件并处理用户响应
func extractChatInteractiveEvent(event *client.ChatEvent, handler *client.InteractiveHandler) *client.InteractiveResponse {
	if event.Body == nil || event.Body.Messages == nil {
		return nil
	}

	var body struct {
		Messages []types.MessageItem `json:"messages"`
	}
	if err := json.Unmarshal([]byte(event.RawJSON), &body); err != nil {
		return nil
	}

	for _, msg := range body.Messages {
		for _, evt := range msg.Events {
			if evt.Type != types.EventTypeInteractive {
				continue
			}

			resp, err := handler.HandleEvent(context.Background(), evt, msg.CallID)
			if err != nil {
				fmt.Printf("⚠️ 交互处理失败: %v\n", err)
				return nil
			}
			if resp == nil {
				return nil
			}

			return resp
		}
	}
	return nil
}