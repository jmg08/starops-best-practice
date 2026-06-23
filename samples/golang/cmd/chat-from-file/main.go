// 从文件加载请求 - 批量处理 requests 目录下的请求文件
// Load Requests from File - Batch process request files in requests directory
//
// 用法 / Usage:
//
//	go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json
//	go run ./cmd/chat-from-file/ -dir ../../requests/cms/              # 处理目录下所有文件
//	go run ./cmd/chat-from-file/ -file entity.json -simple             # 简洁模式
//
// 功能 / Features:
//   - 从 JSON 文件加载请求参数
//   - 支持批量处理目录下所有 JSON 文件
//   - 自动输出日志到 output 目录
//   - 支持简洁模式输出
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vibeops/samples/golang/internal/client"
	"github.com/vibeops/samples/golang/types"
)

// RequestFile JSON 请求文件结构
type RequestFile struct {
	Region              string         `json:"region"`
	DigitalEmployeeName string         `json:"digitalEmployeeName"`
	ThreadId            string         `json:"threadId,omitempty"`
	Action              string         `json:"action"`
	Messages            []RequestMsg   `json:"messages"`
	Variables           map[string]any `json:"variables"`
}

// RequestMsg 请求消息
type RequestMsg struct {
	Role     string           `json:"role"`
	Contents []RequestContent `json:"contents"`
}

// RequestContent 消息内容
type RequestContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

var (
	filePath   = flag.String("file", "", "请求 JSON 文件路径")
	dirPath    = flag.String("dir", "", "请求文件目录，处理目录下所有 JSON 文件")
	simpleMode = flag.Bool("simple", false, "简洁模式，只输出最终文本")
	outputDir  = flag.String("output", "../../requests/output", "输出目录")
)

func main() {
	flag.Parse()

	if *filePath == "" && *dirPath == "" {
		printUsage()
		os.Exit(1)
	}

	fmt.Println("🚀 VibeOps Chat - 从文件加载请求")
	fmt.Println(strings.Repeat("=", 60))

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

	// 确保输出目录存在
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("⚠️ 创建输出目录失败: %v\n", err)
	}

	// 处理文件
	if *dirPath != "" {
		processDirectory(agentClient, *dirPath)
	} else {
		processFile(agentClient, *filePath)
	}
}

func processDirectory(agentClient *client.AgentClient, dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		fmt.Printf("❌ 读取目录失败: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("⚠️ 目录中没有 JSON 文件: %s\n", dir)
		return
	}

	fmt.Printf("📁 找到 %d 个请求文件\n\n", len(files))

	for i, file := range files {
		fmt.Printf("━━━ [%d/%d] %s ━━━\n", i+1, len(files), filepath.Base(file))
		processFile(agentClient, file)
		fmt.Println()
	}

	fmt.Printf("✅ 处理完成，共 %d 个文件\n", len(files))
}

func processFile(agentClient *client.AgentClient, file string) {
	// 加载请求文件
	reqFile, err := loadRequestFile(file)
	if err != nil {
		fmt.Printf("❌ 加载文件失败: %v\n", err)
		return
	}

	// 获取消息
	message := extractMessage(reqFile)
	if message == "" {
		fmt.Printf("⚠️ 文件中没有消息内容\n")
		return
	}

	fmt.Printf("📄 文件: %s\n", filepath.Base(file))
	fmt.Printf("💬 消息: %s\n", truncate(message, 60))

	ctx := context.Background()

	// 创建会话
	threadID, err := agentClient.CreateThread(ctx)
	if err != nil {
		fmt.Printf("❌ 创建会话失败: %v\n", err)
		return
	}

	// 创建输出文件
	outputFile := createOutputFile(file)
	defer outputFile.Close()

	// 写入请求信息
	writeOutput(outputFile, fmt.Sprintf("# Request: %s\n", filepath.Base(file)))
	writeOutput(outputFile, fmt.Sprintf("# Time: %s\n", time.Now().Format(time.RFC3339)))
	writeOutput(outputFile, fmt.Sprintf("# ThreadID: %s\n", threadID))
	writeOutput(outputFile, fmt.Sprintf("# Message: %s\n\n", message))

	fmt.Println(strings.Repeat("-", 60))

	// 发送请求
	startTime := time.Now()

	// 处理响应
	var simplePrinter *client.SimplePrinter
	var eventPrinter *client.EventPrinter
	if *simpleMode {
		simplePrinter = client.NewSimplePrinter()
	} else {
		eventPrinter = client.NewEventPrinter(false, true)
	}

	// 初始化交互处理器
	interactiveHandler := client.NewInteractiveHandler(agentClient, 0)

	events := agentClient.ChatWithVariables(ctx, threadID, message, reqFile.Variables)
	processEvents(events, simplePrinter, eventPrinter, interactiveHandler, outputFile, ctx, threadID, reqFile.Variables)

	elapsed := time.Since(startTime)
	fmt.Println()

	// 写入最终结果
	if *simpleMode && simplePrinter != nil {
		finalText := simplePrinter.GetFinalText()
		writeOutput(outputFile, fmt.Sprintf("\n# Final Result:\n%s\n", finalText))
		fmt.Printf("📄 最终文本:\n%s\n", finalText)
	}

	writeOutput(outputFile, fmt.Sprintf("\n# Duration: %v\n", elapsed))
	fmt.Printf("⏱️  耗时: %v\n", elapsed)
	fmt.Printf("📁 输出: %s\n", outputFile.Name())
}

// processEvents 处理 SSE 事件流，支持交互事件检测和流恢复
func processEvents(
	events <-chan *client.ChatEvent,
	simplePrinter *client.SimplePrinter,
	eventPrinter *client.EventPrinter,
	handler *client.InteractiveHandler,
	outputFile *os.File,
	ctx context.Context,
	threadID string,
	variables map[string]any,
) {
	eventIndex := 0
	for events != nil {
		event, ok := <-events
		if !ok {
			break
		}
		eventIndex++

		if event.Error != nil {
			fmt.Printf("❌ 错误: %v\n", event.Error)
			writeOutput(outputFile, fmt.Sprintf("[ERROR] %v\n", event.Error))
			continue
		}

		// 写入原始事件
		if event.RawJSON != "" {
			writeOutput(outputFile, fmt.Sprintf("[EVENT %d]\n%s\n\n", eventIndex, event.RawJSON))
		}

		// 检测交互事件
		interactiveResp := extractInteractiveEvent(event, handler)
		if interactiveResp != nil {
			fmt.Printf("\n🔄 检测到交互事件，等待用户响应...\n")
			events = handler.ResumeChat(ctx, threadID, interactiveResp, variables)
			eventIndex = 0
			continue
		}

		// 正常输出
		if simplePrinter != nil {
			text := simplePrinter.ProcessEvent(event)
			if text != "" {
				fmt.Print(text)
			}
		} else {
			eventPrinter.PrintEvent(event, eventIndex)
		}

		if event.IsDone {
			break
		}
	}
}

// extractInteractiveEvent 从 ChatEvent 中检测交互事件并处理用户响应
// 返回 nil 表示没有交互事件
func extractInteractiveEvent(event *client.ChatEvent, handler *client.InteractiveHandler) *client.InteractiveResponse {
	if event.Body == nil || event.Body.Messages == nil {
		return nil
	}

	// 解析 RawJSON 以获取结构化的 message 数据
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

			// 处理交互事件
			resp, err := handler.HandleEvent(context.Background(), evt)
			if err != nil {
				fmt.Printf("⚠️ 交互处理失败: %v\n", err)
				return nil
			}
			if resp == nil {
				return nil
			}

			// 填充 callId（从 MessageItem 获取）
			resp.InteractionID = msg.CallID
			return resp
		}
	}
	return nil
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

func extractMessage(req *RequestFile) string {
	if len(req.Messages) > 0 && len(req.Messages[0].Contents) > 0 {
		return req.Messages[0].Contents[0].Value
	}
	return ""
}

func createOutputFile(inputFile string) *os.File {
	baseName := strings.TrimSuffix(filepath.Base(inputFile), ".json")
	timestamp := time.Now().Format("20060102-150405")
	outputPath := filepath.Join(*outputDir, fmt.Sprintf("%s-%s.log", baseName, timestamp))

	f, err := os.Create(outputPath)
	if err != nil {
		// 如果创建失败，返回一个空的文件句柄
		return nil
	}
	return f
}

func writeOutput(f *os.File, content string) {
	if f != nil {
		f.WriteString(content)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  go run ./cmd/chat-from-file/ -file <path>   处理单个文件")
	fmt.Println("  go run ./cmd/chat-from-file/ -dir <path>    处理目录下所有 JSON 文件")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json")
	fmt.Println("  go run ./cmd/chat-from-file/ -dir ../../requests/cms/")
	fmt.Println("  go run ./cmd/chat-from-file/ -file entity.json -simple")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -file     请求 JSON 文件路径")
	fmt.Println("  -dir      请求文件目录")
	fmt.Println("  -simple   简洁模式，只输出最终文本")
	fmt.Println("  -output   输出目录 (默认: output)")
}
