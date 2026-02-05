// 从 JSON 文件加载请求的示例
// 用法: go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vibeops/samples/golang/internal/client"
)

// RequestFile JSON 请求文件结构
type RequestFile struct {
	Region              string                 `json:"region"`
	DigitalEmployeeName string                 `json:"digitalEmployeeName"`
	ThreadId            string                 `json:"threadId,omitempty"`
	Action              string                 `json:"action"`
	Messages            []RequestMessage       `json:"messages"`
	Variables           map[string]interface{} `json:"variables"`
}

// RequestMessage 请求消息
type RequestMessage struct {
	Role     string           `json:"role"`
	Contents []RequestContent `json:"contents"`
}

// RequestContent 消息内容
type RequestContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

var filePath = flag.String("file", "", "请求 JSON 文件路径 (必需)")

func main() {
	flag.Parse()

	if *filePath == "" {
		fmt.Println("❌ 请指定请求文件路径")
		fmt.Println("\n用法:")
		fmt.Println("  go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json")
		os.Exit(1)
	}

	fmt.Println("🚀 VibeOps Chat - 从文件加载请求")
	fmt.Println(strings.Repeat("=", 60))

	// 加载请求文件
	reqFile, err := loadRequestFile(*filePath)
	if err != nil {
		fmt.Printf("❌ 加载请求文件失败: %v\n", err)
		os.Exit(1)
	}

	printRequestInfo(reqFile)

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

	// 创建会话
	fmt.Println("📝 创建会话...")
	threadID, err := agentClient.CreateThread(ctx)
	if err != nil {
		fmt.Printf("❌ 创建会话失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 会话创建成功, ThreadID: %s\n\n", threadID)

	// 获取消息
	message := ""
	if len(reqFile.Messages) > 0 && len(reqFile.Messages[0].Contents) > 0 {
		message = reqFile.Messages[0].Contents[0].Value
	}

	fmt.Printf("🤖 发送消息: %s\n", message)
	fmt.Println(strings.Repeat("-", 60))

	// 发送请求
	printer := client.NewEventPrinter(true, true)
	events := agentClient.ChatWithVariables(ctx, threadID, message, reqFile.Variables)
	eventIndex := 0

	for event := range events {
		eventIndex++
		printer.PrintEvent(event, eventIndex)
	}

	fmt.Println(strings.Repeat("=", 60))
}

func loadRequestFile(filePath string) (*RequestFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var req RequestFile
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &req, nil
}

func printRequestInfo(req *RequestFile) {
	fmt.Printf("📁 请求文件: %s\n", *filePath)
	fmt.Printf("📋 请求信息:\n")
	fmt.Printf("  - Region: %s\n", req.Region)
	fmt.Printf("  - DigitalEmployeeName: %s\n", req.DigitalEmployeeName)
	fmt.Printf("  - Action: %s\n", req.Action)
	if len(req.Messages) > 0 && len(req.Messages[0].Contents) > 0 {
		fmt.Printf("  - Message: %s\n", req.Messages[0].Contents[0].Value)
	}
	if userContext, ok := req.Variables["userContext"].(string); ok {
		if len(userContext) > 80 {
			userContext = userContext[:80] + "..."
		}
		fmt.Printf("  - UserContext: %s\n", userContext)
	}
	fmt.Println()
}
