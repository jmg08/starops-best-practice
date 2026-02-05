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

	cms "github.com/alibabacloud-go/cms-20240330/v6/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"

	"github.com/vibeops/samples/golang/types"
)

// =================================================================================
// 配置结构
// =================================================================================

// Config 应用配置
type Config struct {
	Workspace       string
	Endpoint        string
	Region          string
	AccessKeyID     string
	AccessKeySecret string
	EmployeeName    string // 数字员工名称
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{
		Workspace:       os.Getenv("VIBEOPS_WORKSPACE"),
		Endpoint:        os.Getenv("VIBEOPS_ENDPOINT"),
		Region:          os.Getenv("VIBEOPS_REGION"),
		AccessKeyID:     os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
		EmployeeName:    os.Getenv("VIBEOPS_EMPLOYEE_NAME"),
	}

	// 验证必需的配置
	var missingVars []string
	if cfg.Workspace == "" {
		missingVars = append(missingVars, "VIBEOPS_WORKSPACE")
	}
	if cfg.Endpoint == "" {
		missingVars = append(missingVars, "VIBEOPS_ENDPOINT")
	}
	if cfg.AccessKeyID == "" {
		missingVars = append(missingVars, "ALIBABA_CLOUD_ACCESS_KEY_ID")
	}
	if cfg.AccessKeySecret == "" {
		missingVars = append(missingVars, "ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("缺少必需的环境变量: %s", strings.Join(missingVars, ", "))
	}

	// 设置默认值
	if cfg.EmployeeName == "" {
		cfg.EmployeeName = "default"
	}
	if cfg.Region == "" {
		cfg.Region = "cn-hangzhou"
	}

	return cfg, nil
}

// =================================================================================
// 客户端
// =================================================================================

// AgentClient Agent客户端
type AgentClient struct {
	client *cms.Client
	config *Config
}

// NewAgentClient 创建Agent客户端
func NewAgentClient(cfg *Config) (*AgentClient, error) {
	openApiConfig := &openapi.Config{}
	openApiConfig.SetAccessKeyId(cfg.AccessKeyID)
	openApiConfig.SetAccessKeySecret(cfg.AccessKeySecret)
	openApiConfig.SetEndpoint(cfg.Endpoint)
	openApiConfig.SetSignatureVersion("v3")

	client, err := cms.NewClient(openApiConfig)
	if err != nil {
		return nil, fmt.Errorf("创建CMS客户端失败: %w", err)
	}

	return &AgentClient{
		client: client,
		config: cfg,
	}, nil
}

// CreateThread 创建会话
func (c *AgentClient) CreateThread(ctx context.Context) (string, error) {
	req := &cms.CreateThreadRequest{}
	req.SetTitle(fmt.Sprintf("Chat-%d", time.Now().Unix()))

	variables := &cms.CreateThreadRequestVariables{}
	variables.SetWorkspace(c.config.Workspace)
	req.SetVariables(variables)

	resp, err := c.client.CreateThread(dara.String(c.config.EmployeeName), req)
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}

	if resp.Body == nil || resp.Body.ThreadId == nil {
		return "", fmt.Errorf("无效响应: 缺少ThreadID")
	}

	return dara.StringValue(resp.Body.ThreadId), nil
}

// =================================================================================
// 聊天事件
// =================================================================================

// ChatEvent 聊天事件
type ChatEvent struct {
	Body       *cms.CreateChatResponseBody
	RawJSON    string // 原始JSON
	StatusCode int32
	IsDone     bool
	Error      error
}

// Chat 开始SSE对话
func (c *AgentClient) Chat(ctx context.Context, threadID, message string) <-chan *ChatEvent {
	events := make(chan *ChatEvent, 10)

	go func() {
		defer close(events)

		// 构建消息内容
		content := &cms.CreateChatRequestMessagesContents{}
		content.SetType("text")
		content.SetValue(message)

		msg := &cms.CreateChatRequestMessages{}
		msg.SetRole("user")
		msg.SetContents([]*cms.CreateChatRequestMessagesContents{content})

		// 构建 variables - 关键配置，包含 region 等信息
		variables := map[string]interface{}{
			"workspace": c.config.Workspace,
			"region":    c.config.Region,
			"language":  "zh",
			"timeZone":  "Asia/Shanghai",
			"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
		}

		req := &cms.CreateChatRequest{}
		req.SetAction("create")
		req.SetThreadId(threadID)
		req.SetDigitalEmployeeName(c.config.EmployeeName)
		req.SetMessages([]*cms.CreateChatRequestMessages{msg})
		req.SetVariables(variables)

		yield := make(chan *cms.CreateChatResponse, 10)
		yieldErr := make(chan error, 1)
		runtime := &dara.RuntimeOptions{}
		runtime.SetConnectTimeout(30000) // 30秒连接超时
		runtime.SetReadTimeout(300000)   // 5分钟读取超时

		go c.client.CreateChatWithSSECtx(ctx, req, nil, runtime, yield, yieldErr)

		for {
			select {
			case resp, ok := <-yield:
				if !ok {
					events <- &ChatEvent{IsDone: true}
					return
				}

				// 获取原始JSON
				rawJSON := ""
				if resp.Body != nil {
					jsonBytes, err := json.Marshal(resp.Body)
					if err == nil {
						rawJSON = string(jsonBytes)
					}
				}

				event := &ChatEvent{
					Body:       resp.Body,
					RawJSON:    rawJSON,
					StatusCode: dara.Int32Value(resp.StatusCode),
				}

				if isDoneMessage(resp.Body) {
					event.IsDone = true
				}

				events <- event

				if event.IsDone {
					return
				}

			case err := <-yieldErr:
				if err != nil {
					events <- &ChatEvent{Error: err}
				}
				return

			case <-ctx.Done():
				events <- &ChatEvent{Error: ctx.Err()}
				return
			}
		}
	}()

	return events
}

// isDoneMessage 检查是否是结束消息
func isDoneMessage(body *cms.CreateChatResponseBody) bool {
	if body == nil {
		return false
	}

	for _, msg := range body.Messages {
		if msg.Type != nil && *msg.Type == "done" {
			return true
		}
	}

	return false
}

// =================================================================================
// 事件解析和打印
// =================================================================================

// EventPrinter 事件打印器
type EventPrinter struct {
	printRawBody   bool
	printParsed    bool
	printSeparator bool
}

// NewEventPrinter 创建事件打印器
func NewEventPrinter(printRawBody, printParsed bool) *EventPrinter {
	return &EventPrinter{
		printRawBody:   printRawBody,
		printParsed:    printParsed,
		printSeparator: true,
	}
}

// PrintEvent 打印事件
func (p *EventPrinter) PrintEvent(event *ChatEvent, eventIndex int) {
	if event.Error != nil {
		fmt.Printf("\n❌ 错误: %v\n", event.Error)
		return
	}

	if event.IsDone && event.Body == nil {
		fmt.Println("\n✅ 对话完成")
		return
	}

	if event.Body == nil {
		return
	}

	if p.printSeparator {
		fmt.Printf("\n%s 事件 #%d %s\n", strings.Repeat("=", 30), eventIndex, strings.Repeat("=", 30))
	}

	// 打印原始Body
	if p.printRawBody && event.RawJSON != "" {
		fmt.Println("\n📦 原始 Body:")
		prettyJSON, err := prettyPrintJSON(event.RawJSON)
		if err == nil {
			fmt.Println(prettyJSON)
		} else {
			fmt.Println(event.RawJSON)
		}
	}

	// 解析并打印事件详情
	if p.printParsed {
		p.printParsedEvent(event)
	}
}

// printParsedEvent 打印解析后的事件
func (p *EventPrinter) printParsedEvent(event *ChatEvent) {
	if event.Body == nil {
		return
	}

	fmt.Println("\n📋 解析详情:")

	// 遍历消息
	for _, msg := range event.Body.Messages {
		// 尝试解析为内部消息格式
		msgJSON, _ := json.Marshal(msg)
		// 打印原始消息
		fmt.Printf("  原始消息: %s\n", string(msgJSON))
		var messageItem types.MessageItem
		if err := json.Unmarshal(msgJSON, &messageItem); err != nil {
			fmt.Printf("  解析消息失败: %v\n", err)
			continue
		}

		p.printMessageItem(&messageItem)
	}
}

// printMessageItem 打印消息条目
func (p *EventPrinter) printMessageItem(item *types.MessageItem) {
	// 打印基本信息
	if item.Role != "" {
		fmt.Printf("  📌 角色: %s\n", item.Role)
	}
	if item.CallID != "" {
		fmt.Printf("  🔗 CallID: %s\n", item.CallID)
	}
	if item.ParentCallID != "" {
		fmt.Printf("  🔗 ParentCallID: %s\n", item.ParentCallID)
	}

	// 打印内容
	if len(item.Contents) > 0 {
		fmt.Println("  📝 内容:")
		for i, content := range item.Contents {
			fmt.Printf("    [%d] 类型: %s\n", i, content.Type)
			if content.Value != "" {
				// 截断过长的内容
				value := content.Value
				if len(value) > 200 {
					value = value[:200] + "..."
				}
				fmt.Printf("        值: %s\n", value)
			}
			if content.Append {
				fmt.Printf("        追加: true\n")
			}
			if content.LastChunk {
				fmt.Printf("        最后块: true\n")
			}
		}
	}

	// 打印工具调用
	if len(item.Tools) > 0 {
		fmt.Println("  🔧 工具调用:")
		for i, tool := range item.Tools {
			fmt.Printf("    [%d] 名称: %s, 状态: %s\n", i, tool.Name, tool.Status)
			if tool.ToolCallID != "" {
				fmt.Printf("        ToolCallID: %s\n", tool.ToolCallID)
			}
			if tool.Arguments != nil {
				argsJSON, _ := json.Marshal(tool.Arguments)
				argsStr := string(argsJSON)
				if len(argsStr) > 200 {
					argsStr = argsStr[:200] + "..."
				}
				fmt.Printf("        参数: %s\n", argsStr)
			}
			if tool.ArgumentsDelta != "" {
				delta := tool.ArgumentsDelta
				if len(delta) > 100 {
					delta = delta[:100] + "..."
				}
				fmt.Printf("        参数增量: %s\n", delta)
			}
			if len(tool.Contents) > 0 {
				fmt.Println("        结果:")
				for _, c := range tool.Contents {
					if c.Value != "" {
						val := c.Value
						if len(val) > 150 {
							val = val[:150] + "..."
						}
						fmt.Printf("          - %s: %s\n", c.Type, val)
					}
				}
			}
		}
	}

	// 打印Agent调用
	if len(item.Agents) > 0 {
		fmt.Println("  🤖 Agent调用:")
		for i, agent := range item.Agents {
			fmt.Printf("    [%d] 名称: %s, 状态: %s\n", i, agent.Name, agent.Status)
			if agent.CallID != "" {
				fmt.Printf("        CallID: %s\n", agent.CallID)
			}
		}
	}

	// 打印事件
	if len(item.Events) > 0 {
		fmt.Println("  📢 事件:")
		for i, evt := range item.Events {
			fmt.Printf("    [%d] 类型: %s\n", i, evt.Type)
			if evt.Payload != nil {
				p.printEventPayload(evt)
			}
		}
	}

	// 打印产物
	if len(item.Artifacts) > 0 {
		fmt.Println("  📦 产物:")
		for i, artifact := range item.Artifacts {
			artJSON, _ := json.Marshal(artifact)
			artStr := string(artJSON)
			if len(artStr) > 200 {
				artStr = artStr[:200] + "..."
			}
			fmt.Printf("    [%d] %s\n", i, artStr)
		}
	}
}

// printEventPayload 打印事件负载
func (p *EventPrinter) printEventPayload(evt *types.ItemEvent) {
	payloadJSON, _ := json.Marshal(evt.Payload)

	switch evt.Type {
	case types.EventTypeThinking:
		var thinking types.ItemThinkingPayload
		if err := json.Unmarshal(payloadJSON, &thinking); err == nil && thinking.ReasoningDelta != "" {
			delta := thinking.ReasoningDelta
			if len(delta) > 100 {
				delta = delta[:100] + "..."
			}
			fmt.Printf("        思考: %s\n", delta)
		}

	case types.EventTypeError:
		var errPayload types.ItemErrorPayload
		if err := json.Unmarshal(payloadJSON, &errPayload); err == nil {
			fmt.Printf("        错误码: %s\n", errPayload.Code)
			fmt.Printf("        消息: %s\n", errPayload.Message)
			if errPayload.Suggestion != "" {
				fmt.Printf("        建议: %s\n", errPayload.Suggestion)
			}
		}

	case types.EventTypeTaskFinished:
		var finished types.ItemTaskFinishedPayload
		if err := json.Unmarshal(payloadJSON, &finished); err == nil {
			fmt.Printf("        成功: %v\n", finished.Success)
			if finished.Statistics != nil {
				fmt.Printf("        耗时: %dms\n", finished.Statistics.Duration/1000000)
			}
		}

	case types.EventTypeInteractive:
		var interactive types.ItemInteractivePayload
		if err := json.Unmarshal(payloadJSON, &interactive); err == nil {
			fmt.Printf("        交互类型: %s\n", interactive.InteractiveType)
		}

	case types.EventTypeThreadTitleUpdated:
		var title types.ItemThreadTitleUpdatedPayload
		if err := json.Unmarshal(payloadJSON, &title); err == nil {
			fmt.Printf("        新标题: %s\n", title.Title)
		}

	default:
		// 通用打印
		payloadStr := string(payloadJSON)
		if len(payloadStr) > 200 {
			payloadStr = payloadStr[:200] + "..."
		}
		fmt.Printf("        负载: %s\n", payloadStr)
	}
}

// prettyPrintJSON 格式化JSON输出
func prettyPrintJSON(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	prettyBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(prettyBytes), nil
}

// =================================================================================
// 主程序
// =================================================================================

func main() {
	fmt.Println("🚀 VibeOps Chat 示例程序")
	fmt.Println(strings.Repeat("=", 60))

	// 加载配置
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		fmt.Println("\n请设置以下环境变量:")
		fmt.Println("  - VIBEOPS_WORKSPACE: 工作空间")
		fmt.Println("  - VIBEOPS_ENDPOINT: API端点")
		fmt.Println("  - VIBEOPS_REGION: 区域 (可选, 默认cn-hangzhou)")
		fmt.Println("  - ALIBABA_CLOUD_ACCESS_KEY_ID: Access Key ID")
		fmt.Println("  - ALIBABA_CLOUD_ACCESS_KEY_SECRET: Access Key Secret")
		fmt.Println("  - VIBEOPS_EMPLOYEE_NAME: 数字员工名称 (可选)")
		os.Exit(1)
	}

	fmt.Printf("📋 配置信息:\n")
	fmt.Printf("  - Workspace: %s\n", cfg.Workspace)
	fmt.Printf("  - Endpoint: %s\n", cfg.Endpoint)
	fmt.Printf("  - Region: %s\n", cfg.Region)
	fmt.Printf("  - Employee: %s\n", cfg.EmployeeName)
	fmt.Println()

	// 创建客户端
	client, err := NewAgentClient(cfg)
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 创建会话
	fmt.Println("📝 创建会话...")
	threadID, err := client.CreateThread(ctx)
	if err != nil {
		fmt.Printf("❌ 创建会话失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 会话创建成功, ThreadID: %s\n", threadID)
	fmt.Println()

	// 创建事件打印器
	printer := NewEventPrinter(true, true)

	// 开始对话循环
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

		// 发送消息并处理响应
		events := client.Chat(ctx, threadID, input)
		eventIndex := 0

		for event := range events {
			eventIndex++
			printer.PrintEvent(event, eventIndex)
		}

		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()
	}
}
