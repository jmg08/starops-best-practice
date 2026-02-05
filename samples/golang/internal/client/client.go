// Package client 提供 VibeOps Agent 客户端的公共实现
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	cms "github.com/alibabacloud-go/cms-20240330/v6/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
)

// Config 应用配置
type Config struct {
	Workspace       string
	Endpoint        string
	Region          string
	AccessKeyID     string
	AccessKeySecret string
	EmployeeName    string
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

	if cfg.EmployeeName == "" {
		cfg.EmployeeName = "default"
	}
	if cfg.Region == "" {
		cfg.Region = "cn-hangzhou"
	}

	return cfg, nil
}

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

	cmsClient, err := cms.NewClient(openApiConfig)
	if err != nil {
		return nil, fmt.Errorf("创建CMS客户端失败: %w", err)
	}

	return &AgentClient{
		client: cmsClient,
		config: cfg,
	}, nil
}

// Config 获取配置
func (c *AgentClient) Config() *Config {
	return c.config
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

// ChatEvent 聊天事件
type ChatEvent struct {
	Body       *cms.CreateChatResponseBody
	RawJSON    string
	StatusCode int32
	IsDone     bool
	Error      error
}

// Chat 开始SSE对话（基础版本）
func (c *AgentClient) Chat(ctx context.Context, threadID, message string) <-chan *ChatEvent {
	variables := map[string]interface{}{
		"workspace": c.config.Workspace,
		"region":    c.config.Region,
		"language":  "zh",
		"timeZone":  "Asia/Shanghai",
		"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	return c.ChatWithVariables(ctx, threadID, message, variables)
}

// ChatWithVariables 开始SSE对话（支持自定义 variables）
func (c *AgentClient) ChatWithVariables(ctx context.Context, threadID, message string, variables map[string]interface{}) <-chan *ChatEvent {
	events := make(chan *ChatEvent, 10)

	go func() {
		defer close(events)

		content := &cms.CreateChatRequestMessagesContents{}
		content.SetType("text")
		content.SetValue(message)

		msg := &cms.CreateChatRequestMessages{}
		msg.SetRole("user")
		msg.SetContents([]*cms.CreateChatRequestMessagesContents{content})

		// 确保包含必要字段
		if variables == nil {
			variables = make(map[string]interface{})
		}
		if _, ok := variables["workspace"]; !ok {
			variables["workspace"] = c.config.Workspace
		}
		if _, ok := variables["region"]; !ok {
			variables["region"] = c.config.Region
		}
		if _, ok := variables["language"]; !ok {
			variables["language"] = "zh"
		}
		if _, ok := variables["timeZone"]; !ok {
			variables["timeZone"] = "Asia/Shanghai"
		}
		if _, ok := variables["timeStamp"]; !ok {
			variables["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
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
		runtime.SetConnectTimeout(30000)
		runtime.SetReadTimeout(300000)

		go c.client.CreateChatWithSSECtx(ctx, req, nil, runtime, yield, yieldErr)

		for {
			select {
			case resp, ok := <-yield:
				if !ok {
					events <- &ChatEvent{IsDone: true}
					return
				}

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
