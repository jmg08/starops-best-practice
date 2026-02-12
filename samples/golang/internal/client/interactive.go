// Package client 提供 VibeOps Agent 客户端的公共实现
package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vibeops/samples/golang/types"
)

// InteractiveHandler 交互事件处理器
// 用于处理用户确认、选择和输入等交互事件
type InteractiveHandler struct {
	client  *AgentClient
	timeout time.Duration
	reader  io.Reader
	writer  io.Writer
}

// InteractiveResponse 交互响应
// 包含用户对交互事件的响应信息
type InteractiveResponse struct {
	InteractionID string                 `json:"interactionId"`
	Type          types.InteractionType  `json:"type"`
	Response      map[string]interface{} `json:"response"`
}

// NewInteractiveHandler 创建交互处理器
// client: Agent 客户端
// timeout: 用户响应超时时间，0 表示不设置超时
func NewInteractiveHandler(client *AgentClient, timeout time.Duration) *InteractiveHandler {
	return &InteractiveHandler{
		client:  client,
		timeout: timeout,
		reader:  os.Stdin,
		writer:  os.Stdout,
	}
}

// SetIO 设置输入输出流
// 用于测试或自定义输入输出
func (h *InteractiveHandler) SetIO(reader io.Reader, writer io.Writer) {
	h.reader = reader
	h.writer = writer
}

// HandleEvent 处理交互事件
// 根据事件类型分发到对应的处理方法
func (h *InteractiveHandler) HandleEvent(ctx context.Context, event *types.ItemEvent) (*InteractiveResponse, error) {
	if event == nil {
		return nil, NewSDKError(ErrCodeParseError, "事件为空")
	}

	if event.Type != types.EventTypeInteractive {
		return nil, NewSDKError(ErrCodeParseError, fmt.Sprintf("不支持的事件类型: %s", event.Type))
	}

	// 解析交互负载
	payload, err := h.parseInteractivePayload(event.Payload)
	if err != nil {
		return nil, err
	}

	// 根据交互类型分发处理
	switch payload.InteractiveType {
	case types.InteractionTypeUserAck:
		return h.HandleUserAck(ctx, payload)
	case types.InteractionTypeUserSelect:
		return h.HandleUserSelect(ctx, payload)
	case types.InteractionTypeUserInput:
		return h.HandleUserInput(ctx, payload)
	default:
		return nil, NewSDKError(ErrCodeParseError, fmt.Sprintf("不支持的交互类型: %s", payload.InteractiveType))
	}
}

// HandleUserAck 处理用户确认
// 显示确认提示并等待用户响应
func (h *InteractiveHandler) HandleUserAck(ctx context.Context, payload *types.ItemInteractivePayload) (*InteractiveResponse, error) {
	if payload == nil {
		return nil, NewSDKError(ErrCodeParseError, "交互负载为空")
	}

	// 获取交互 ID
	interactionID := h.getInteractionID(payload)

	// 显示确认提示
	title := h.getTitle(payload)
	description := h.getDescription(payload)

	h.printf("\n🔔 确认请求\n")
	if title != "" {
		h.printf("   标题: %s\n", title)
	}
	if description != "" {
		h.printf("   描述: %s\n", description)
	}
	h.printf("   请输入 [y/yes] 确认，[n/no] 取消: ")

	// 读取用户输入
	input, err := h.readInputWithTimeout(ctx)
	if err != nil {
		return nil, err
	}

	// 解析用户响应
	input = strings.TrimSpace(strings.ToLower(input))
	confirmed := input == "y" || input == "yes" || input == "是" || input == ""

	return &InteractiveResponse{
		InteractionID: interactionID,
		Type:          types.InteractionTypeUserAck,
		Response: map[string]interface{}{
			"confirmed": confirmed,
		},
	}, nil
}

// HandleUserSelect 处理用户选择
// 显示选项列表并捕获用户选择
func (h *InteractiveHandler) HandleUserSelect(ctx context.Context, payload *types.ItemInteractivePayload) (*InteractiveResponse, error) {
	if payload == nil {
		return nil, NewSDKError(ErrCodeParseError, "交互负载为空")
	}

	// 获取交互 ID
	interactionID := h.getInteractionID(payload)

	// 显示选择提示
	title := h.getTitle(payload)
	h.printf("\n📋 请选择\n")
	if title != "" {
		h.printf("   标题: %s\n", title)
	}

	// 获取选项列表
	options := h.getOptions(payload)
	if len(options) == 0 {
		return nil, NewSDKError(ErrCodeParseError, "没有可选项")
	}

	// 显示选项
	h.printf("   选项:\n")
	for i, opt := range options {
		label := h.getOptionLabel(opt, i)
		h.printf("   [%d] %s\n", i+1, label)
	}
	h.printf("   请输入选项编号 (1-%d): ", len(options))

	// 读取用户输入
	input, err := h.readInputWithTimeout(ctx)
	if err != nil {
		return nil, err
	}

	// 解析用户选择
	input = strings.TrimSpace(input)
	selectedIndex, err := strconv.Atoi(input)
	if err != nil || selectedIndex < 1 || selectedIndex > len(options) {
		return nil, NewSDKError(ErrCodeParseError, fmt.Sprintf("无效的选择: %s，请输入 1-%d 之间的数字", input, len(options)))
	}

	// 获取选中的选项
	selectedOption := options[selectedIndex-1]

	return &InteractiveResponse{
		InteractionID: interactionID,
		Type:          types.InteractionTypeUserSelect,
		Response: map[string]interface{}{
			"selectedIndex": selectedIndex - 1, // 0-based index
			"selectedValue": selectedOption,
		},
	}, nil
}

// HandleUserInput 处理用户输入
// 提示用户输入并提交响应
func (h *InteractiveHandler) HandleUserInput(ctx context.Context, payload *types.ItemInteractivePayload) (*InteractiveResponse, error) {
	if payload == nil {
		return nil, NewSDKError(ErrCodeParseError, "交互负载为空")
	}

	// 获取交互 ID
	interactionID := h.getInteractionID(payload)

	// 显示输入提示
	title := h.getTitle(payload)
	description := h.getDescription(payload)
	placeholder := h.getPlaceholder(payload)

	h.printf("\n✏️  请输入\n")
	if title != "" {
		h.printf("   标题: %s\n", title)
	}
	if description != "" {
		h.printf("   描述: %s\n", description)
	}
	if placeholder != "" {
		h.printf("   提示: %s\n", placeholder)
	}
	h.printf("   请输入内容: ")

	// 读取用户输入
	input, err := h.readInputWithTimeout(ctx)
	if err != nil {
		return nil, err
	}

	input = strings.TrimSpace(input)

	return &InteractiveResponse{
		InteractionID: interactionID,
		Type:          types.InteractionTypeUserInput,
		Response: map[string]interface{}{
			"value": input,
		},
	}, nil
}

// ResumeChat 使用交互响应恢复对话
// 将用户的交互响应发送给 Agent 以继续对话
func (h *InteractiveHandler) ResumeChat(ctx context.Context, threadID string, response *InteractiveResponse) <-chan *ChatEvent {
	if h.client == nil {
		events := make(chan *ChatEvent, 1)
		events <- &ChatEvent{Error: NewSDKError(ErrCodeClientCreate, "客户端未初始化")}
		close(events)
		return events
	}

	if response == nil {
		events := make(chan *ChatEvent, 1)
		events <- &ChatEvent{Error: NewSDKError(ErrCodeParseError, "交互响应为空")}
		close(events)
		return events
	}

	// 构建恢复消息
	// 将交互响应序列化为 JSON 作为消息内容
	responseJSON, err := json.Marshal(response)
	if err != nil {
		events := make(chan *ChatEvent, 1)
		events <- &ChatEvent{Error: NewSDKErrorWithCause(ErrCodeParseError, "序列化交互响应失败", err)}
		close(events)
		return events
	}

	// 构建包含交互响应的变量
	variables := map[string]interface{}{
		"workspace":         h.client.config.Workspace,
		"region":            h.client.config.Region,
		"language":          "zh",
		"timeZone":          "Asia/Shanghai",
		"timeStamp":         fmt.Sprintf("%d", time.Now().Unix()),
		"interactionId":     response.InteractionID,
		"interactionType":   string(response.Type),
		"interactionResult": response.Response,
	}

	// 使用交互响应作为消息恢复对话
	message := fmt.Sprintf("[交互响应] %s", string(responseJSON))

	return h.client.ChatWithVariables(ctx, threadID, message, variables)
}

// =================================================================================
// 辅助方法
// =================================================================================

// parseInteractivePayload 解析交互负载
func (h *InteractiveHandler) parseInteractivePayload(payload interface{}) (*types.ItemInteractivePayload, error) {
	if payload == nil {
		return nil, NewSDKError(ErrCodeParseError, "交互负载为空")
	}

	// 尝试直接类型断言
	if p, ok := payload.(*types.ItemInteractivePayload); ok {
		return p, nil
	}

	// 尝试 JSON 序列化/反序列化
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, NewSDKErrorWithCause(ErrCodeParseError, "序列化交互负载失败", err)
	}

	var interactivePayload types.ItemInteractivePayload
	if err := json.Unmarshal(payloadJSON, &interactivePayload); err != nil {
		return nil, NewSDKErrorWithCause(ErrCodeParseError, "解析交互负载失败", err)
	}

	return &interactivePayload, nil
}

// readInputWithTimeout 带超时的读取用户输入
func (h *InteractiveHandler) readInputWithTimeout(ctx context.Context) (string, error) {
	// 创建输入通道
	inputChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		reader := bufio.NewReader(h.reader)
		input, err := reader.ReadString('\n')
		if err != nil {
			errChan <- err
			return
		}
		inputChan <- input
	}()

	// 根据是否设置超时选择等待方式
	if h.timeout > 0 {
		select {
		case input := <-inputChan:
			return input, nil
		case err := <-errChan:
			return "", NewSDKErrorWithCause(ErrCodeParseError, "读取输入失败", err)
		case <-time.After(h.timeout):
			return "", NewSDKError(ErrCodeInteractiveTimeout, fmt.Sprintf("用户响应超时 (%v)", h.timeout)).
				WithContext("timeout", h.timeout.String())
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return "", NewSDKErrorWithCause(ErrCodeInteractiveTimeout, "上下文超时", ctx.Err())
			}
			return "", NewSDKErrorWithCause(ErrCodeCancelled, "操作已取消", ctx.Err())
		}
	}

	// 无超时，只等待输入或上下文取消
	select {
	case input := <-inputChan:
		return input, nil
	case err := <-errChan:
		return "", NewSDKErrorWithCause(ErrCodeParseError, "读取输入失败", err)
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return "", NewSDKErrorWithCause(ErrCodeInteractiveTimeout, "上下文超时", ctx.Err())
		}
		return "", NewSDKErrorWithCause(ErrCodeCancelled, "操作已取消", ctx.Err())
	}
}

// printf 格式化输出
func (h *InteractiveHandler) printf(format string, args ...interface{}) {
	fmt.Fprintf(h.writer, format, args...)
}

// getInteractionID 从负载中获取交互 ID
func (h *InteractiveHandler) getInteractionID(payload *types.ItemInteractivePayload) string {
	if payload.Meta != nil {
		if id, ok := payload.Meta["id"].(string); ok {
			return id
		}
		if id, ok := payload.Meta["interactionId"].(string); ok {
			return id
		}
	}
	// 生成默认 ID
	return fmt.Sprintf("interaction_%d", time.Now().UnixNano())
}

// getTitle 从负载中获取标题
func (h *InteractiveHandler) getTitle(payload *types.ItemInteractivePayload) string {
	if payload.Meta != nil {
		if title, ok := payload.Meta["title"].(string); ok {
			return title
		}
	}
	return ""
}

// getDescription 从负载中获取描述
func (h *InteractiveHandler) getDescription(payload *types.ItemInteractivePayload) string {
	if payload.Meta != nil {
		if desc, ok := payload.Meta["description"].(string); ok {
			return desc
		}
		if desc, ok := payload.Meta["desc"].(string); ok {
			return desc
		}
	}
	return ""
}

// getPlaceholder 从负载中获取占位符提示
func (h *InteractiveHandler) getPlaceholder(payload *types.ItemInteractivePayload) string {
	if payload.Meta != nil {
		if placeholder, ok := payload.Meta["placeholder"].(string); ok {
			return placeholder
		}
	}
	return ""
}

// getOptions 从负载中获取选项列表
func (h *InteractiveHandler) getOptions(payload *types.ItemInteractivePayload) []map[string]interface{} {
	// 优先从 Data 字段获取
	if len(payload.Data) > 0 {
		return payload.Data
	}

	// 尝试从 Meta 中获取
	if payload.Meta != nil {
		if options, ok := payload.Meta["options"].([]interface{}); ok {
			result := make([]map[string]interface{}, 0, len(options))
			for _, opt := range options {
				if optMap, ok := opt.(map[string]interface{}); ok {
					result = append(result, optMap)
				}
			}
			return result
		}
	}

	return nil
}

// getOptionLabel 获取选项的显示标签
func (h *InteractiveHandler) getOptionLabel(option map[string]interface{}, index int) string {
	// 尝试获取 label 字段
	if label, ok := option["label"].(string); ok {
		return label
	}
	// 尝试获取 name 字段
	if name, ok := option["name"].(string); ok {
		return name
	}
	// 尝试获取 title 字段
	if title, ok := option["title"].(string); ok {
		return title
	}
	// 尝试获取 value 字段
	if value, ok := option["value"].(string); ok {
		return value
	}
	// 默认显示选项编号
	return fmt.Sprintf("选项 %d", index+1)
}

// =================================================================================
// 便捷方法
// =================================================================================

// IsInteractiveEvent 检查事件是否为交互事件
func IsInteractiveEvent(event *types.ItemEvent) bool {
	return event != nil && event.Type == types.EventTypeInteractive
}

// ExtractInteractiveEvents 从消息项中提取交互事件
func ExtractInteractiveEvents(item *types.MessageItem) []*types.ItemEvent {
	if item == nil || len(item.Events) == 0 {
		return nil
	}

	var interactiveEvents []*types.ItemEvent
	for _, evt := range item.Events {
		if IsInteractiveEvent(evt) {
			interactiveEvents = append(interactiveEvents, evt)
		}
	}
	return interactiveEvents
}
