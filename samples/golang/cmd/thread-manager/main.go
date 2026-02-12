// 会话管理工具 - 演示 CMS SDK 的其他接口
// 用法:
//
//	go run ./cmd/thread-manager/ list              # 列出会话
//	go run ./cmd/thread-manager/ get <thread-id>   # 获取会话详情
//	go run ./cmd/thread-manager/ messages <thread-id>  # 列出会话消息
//	go run ./cmd/thread-manager/ delete <thread-id>    # 删除会话
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vibeops/samples/golang/internal/client"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// 加载配置
	cfg, err := client.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 创建客户端
	agentClient, err := client.NewAgentClient(cfg)
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch command {
	case "list":
		listThreads(ctx, agentClient)
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("❌ 请指定 thread-id")
			os.Exit(1)
		}
		getThread(ctx, agentClient, os.Args[2])
	case "messages":
		if len(os.Args) < 3 {
			fmt.Println("❌ 请指定 thread-id")
			os.Exit(1)
		}
		listMessages(ctx, agentClient, os.Args[2])
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("❌ 请指定 thread-id")
			os.Exit(1)
		}
		deleteThread(ctx, agentClient, os.Args[2])
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("会话管理工具 - CMS SDK 接口演示")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("\n用法:")
	fmt.Println("  go run ./cmd/thread-manager/ <command> [args]")
	fmt.Println("\n命令:")
	fmt.Println("  list                  列出所有会话")
	fmt.Println("  get <thread-id>       获取会话详情")
	fmt.Println("  messages <thread-id>  列出会话消息")
	fmt.Println("  delete <thread-id>    删除会话")
}

func listThreads(ctx context.Context, c *client.AgentClient) {
	fmt.Println("📋 会话列表")
	fmt.Println(strings.Repeat("-", 100))

	threads, total, err := c.ListThreads(ctx, 20)
	if err != nil {
		fmt.Printf("❌ 获取会话列表失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 个会话:\n\n", total)

	if len(threads) == 0 {
		fmt.Println("  (无会话)")
		return
	}

	fmt.Printf("%-40s %-25s %-10s %s\n", "Thread ID", "标题", "状态", "创建时间")
	fmt.Println(strings.Repeat("-", 100))

	for _, t := range threads {
		title := t.Title
		if len(title) > 23 {
			title = title[:23] + "..."
		}
		fmt.Printf("%-40s %-25s %-10s %s\n", t.ThreadID, title, t.Status, t.CreateTime)
	}
}

func getThread(ctx context.Context, c *client.AgentClient, threadID string) {
	fmt.Printf("📋 会话详情: %s\n", threadID)
	fmt.Println(strings.Repeat("-", 60))

	detail, err := c.GetThread(ctx, threadID)
	if err != nil {
		fmt.Printf("❌ 获取会话详情失败: %v\n", err)
		return
	}

	fmt.Printf("  Thread ID: %s\n", detail.ThreadID)
	fmt.Printf("  标题: %s\n", detail.Title)
	fmt.Printf("  状态: %s\n", detail.Status)
	fmt.Printf("  创建时间: %s\n", detail.CreateTime)
	fmt.Printf("  更新时间: %s\n", detail.UpdateTime)
}

func listMessages(ctx context.Context, c *client.AgentClient, threadID string) {
	fmt.Printf("💬 会话消息: %s\n", threadID)
	fmt.Println(strings.Repeat("-", 80))

	messages, err := c.GetThreadData(ctx, threadID, 50)
	if err != nil {
		fmt.Printf("❌ 获取消息列表失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 条消息:\n\n", len(messages))

	for i, m := range messages {
		roleIcon := "👤"
		if m.Role == "assistant" {
			roleIcon = "🤖"
		} else if m.Role == "system" {
			roleIcon = "⚙️"
		}

		content := m.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		fmt.Printf("[%d] %s %s\n", i+1, roleIcon, m.Role)
		if content != "" {
			fmt.Printf("    %s\n", content)
		}
		fmt.Printf("    时间戳: %s\n\n", m.Timestamp)
	}
}

func deleteThread(ctx context.Context, c *client.AgentClient, threadID string) {
	fmt.Printf("🗑️ 删除会话: %s\n", threadID)

	err := c.DeleteThread(ctx, threadID)
	if err != nil {
		fmt.Printf("❌ 删除会话失败: %v\n", err)
		return
	}

	fmt.Println("✅ 会话已删除")
}
