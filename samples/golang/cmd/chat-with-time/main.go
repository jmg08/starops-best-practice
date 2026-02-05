// 带时间参数的对话示例
// 演示如何构建包含时间范围的 userContext
// 用法: go run ./cmd/chat-with-time/ -message "最近有什么异常"
// 用法: go run ./cmd/chat-with-time/ -from 1770274812 -to 1770275712 -message "这段时间有问题吗"
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vibeops/samples/golang/internal/client"
	"github.com/vibeops/samples/golang/types"
)

var (
	fromTime = flag.Int64("from", 0, "查询起始时间 (Unix时间戳，秒)，默认15分钟前")
	toTime   = flag.Int64("to", 0, "查询结束时间 (Unix时间戳，秒)，默认当前时间")
	message  = flag.String("message", "最近有什么异常吗", "发送的消息内容")
)

func main() {
	flag.Parse()

	fmt.Println("🚀 VibeOps Chat - 带时间参数示例")
	fmt.Println(strings.Repeat("=", 60))

	// 处理时间参数
	now := time.Now().Unix()
	from := *fromTime
	to := *toTime

	if from == 0 {
		from = now - 15*60 // 默认15分钟前
	}
	if to == 0 {
		to = now
	}

	fmt.Printf("📅 时间范围:\n")
	fmt.Printf("  - FromTime: %d (%s)\n", from, time.Unix(from, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("  - ToTime:   %d (%s)\n\n", to, time.Unix(to, 0).Format("2006-01-02 15:04:05"))

	// 加载配置
	cfg, err := client.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 构建 userContext
	variables, err := buildVariablesWithTimeRange(cfg.Workspace, cfg.Region, from, to)
	if err != nil {
		fmt.Printf("❌ 构建 variables 失败: %v\n", err)
		os.Exit(1)
	}

	// 打印构建的 userContext
	if uc, ok := variables["userContext"].(string); ok {
		fmt.Printf("📋 构建的 userContext:\n%s\n\n", prettyJSON(uc))
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
	fmt.Printf("✅ 会话创建成功, ThreadID: %s\n\n", threadID)

	fmt.Printf("🤖 发送消息: %s\n", *message)
	fmt.Println(strings.Repeat("-", 60))

	// 发送请求
	printer := client.NewEventPrinter(true, true)
	events := agentClient.ChatWithVariables(ctx, threadID, *message, variables)
	eventIndex := 0

	for event := range events {
		eventIndex++
		printer.PrintEvent(event, eventIndex)
	}

	fmt.Println(strings.Repeat("=", 60))
}

// buildVariablesWithTimeRange 构建包含时间范围的 variables
func buildVariablesWithTimeRange(workspace, region string, fromTime, toTime int64) (map[string]interface{}, error) {
	// 构建 metadata 上下文
	contexts := []types.UserContext{
		{
			Type: types.UserContextTypeMetadata,
			Data: types.MetadataUserData{
				FromTime: fromTime,
				ToTime:   toTime,
			},
		},
	}

	userContextJSON, err := json.Marshal(contexts)
	if err != nil {
		return nil, fmt.Errorf("序列化 userContext 失败: %w", err)
	}

	return map[string]interface{}{
		"workspace":   workspace,
		"region":      region,
		"language":    "zh",
		"timeZone":    "Asia/Shanghai",
		"timeStamp":   fmt.Sprintf("%d", time.Now().Unix()),
		"userContext": string(userContextJSON),
	}, nil
}

func prettyJSON(jsonStr string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr
	}
	prettyBytes, _ := json.MarshalIndent(data, "", "  ")
	return string(prettyBytes)
}
