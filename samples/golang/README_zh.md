# VibeOps STAROps SDK Go 语言示例

本项目是一组示例程序，演示如何使用阿里云 STAROps SDK 与 VibeOps 数字员工进行交互。示例涵盖多种场景，包括交互式对话、多轮会话、超时控制、交互事件处理和会话管理。

## 目录

- [目录结构](#目录结构)
- [快速开始](#快速开始)
- [环境变量](#环境变量)
- [凭证管理](#凭证管理)
- [重试配置](#重试配置)
- [示例程序](#示例程序)
  - [交互式对话 (chat)](#1-交互式对话-cmdchat)
  - [简洁输出模式 (chat-simple)](#2-简洁输出模式-cmdchat-simple)
  - [超时控制 (chat-timeout)](#3-超时控制-cmdchat-timeout)
  - [交互事件处理 (chat-interactive)](#4-交互事件处理-cmdchat-interactive)
  - [多轮对话 (chat-multi-turn)](#5-多轮对话-cmdchat-multi-turn)
  - [会话管理器 (thread-manager)](#6-会话管理器-cmdthread-manager)
  - [文件加载请求 (chat-from-file)](#7-文件加载请求-cmdchat-from-file)
  - [时间参数示例 (chat-with-time)](#8-时间参数示例-cmdchat-with-time)
- [API 参考](#api-参考)
  - [客户端模块](#客户端模块)
  - [会话管理](#会话管理)
  - [交互处理器](#交互处理器)
  - [简洁打印器](#简洁打印器)
  - [事件打印器](#事件打印器)
  - [日志模块](#日志模块)
  - [响应管理器](#响应管理器)
  - [错误处理](#错误处理)
- [类型定义](#类型定义)
  - [事件类型](#事件类型)
  - [输入类型](#输入类型)
- [请求文件格式](#请求文件格式)
- [SDK 依赖](#sdk-依赖)

## 目录结构

```
samples/golang/
├── cmd/
│   ├── chat/                  # 交互式对话（基础示例）
│   ├── chat-simple/           # 简洁输出模式（仅文本）
│   ├── chat-timeout/          # 超时控制示例
│   ├── chat-interactive/      # 交互事件处理
│   ├── chat-multi-turn/       # 多轮对话
│   ├── thread-manager/        # 会话管理工具
│   ├── chat-from-file/        # 从 JSON 文件加载请求
│   └── chat-with-time/        # 时间参数示例
├── internal/
│   ├── client/                # 核心客户端实现
│   │   ├── client.go          # AgentClient 和对话方法
│   │   ├── thread.go          # 会话管理 API
│   │   ├── interactive.go     # 交互事件处理器
│   │   ├── simple_printer.go  # 简洁模式打印器
│   │   ├── printer.go         # 事件打印器
│   │   └── errors.go          # 错误定义
│   ├── logger/                # 结构化日志
│   │   └── logger.go          # 日志实现
│   └── response/              # 响应结构管理
│       └── manager.go         # 响应管理器
├── types/
│   ├── events.go              # 事件类型定义
│   └── inputs.go              # 输入参数类型
├── responses/                 # 响应结构存储
├── go.mod
├── go.sum
└── README_zh.md
```


## 快速开始

1. **克隆仓库**
   ```bash
   git clone <repository-url>
   cd samples/golang
   ```

2. **设置环境变量**
   ```bash
   export VIBEOPS_ENDPOINT="starops.cn-hongkong.aliyuncs.com"
   export VIBEOPS_REGION="cn-hongkong"
   # 以下 AK/SK 可选，如未设置则自动使用阿里云默认凭据链
   export ALIBABA_CLOUD_ACCESS_KEY_ID="your-access-key-id"
   export ALIBABA_CLOUD_ACCESS_KEY_SECRET="your-access-key-secret"
   export VIBEOPS_EMPLOYEE_NAME="default"  # 可选
   ```

3. **运行示例程序**
   ```bash
   go run ./cmd/chat/
   ```

## 环境变量

| 变量名 | 必需 | 描述 | 示例 |
|--------|------|------|------|
| `VIBEOPS_ENDPOINT` | ✅ | STAROps API 端点，格式: `starops.{region-id}.aliyuncs.com` | `starops.cn-hongkong.aliyuncs.com` |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ❌* | 阿里云 Access Key ID（如使用默认凭据链则可选） | |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ❌* | 阿里云 Access Key Secret（如使用默认凭据链则可选） | |
| `VIBEOPS_REGION` | ❌ | 地域（需与端点匹配） | `cn-hongkong` |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | 数字员工名称（默认：`default`） | `apsara-ops` |
| `VIBEOPS_MAX_RETRIES` | ❌ | SSE 连接最大重试次数（默认：10） | `10` |
| `LOG_LEVEL` | ❌ | 日志级别：debug, info, warn, error | `info` |

> *当环境变量 AK/SK 未设置时，SDK 会自动使用阿里云默认凭据链。

## 凭证管理

SDK 支持两种凭证获取方式：

### 1. 环境变量（优先级最高）

直接设置 `ALIBABA_CLOUD_ACCESS_KEY_ID` 和 `ALIBABA_CLOUD_ACCESS_KEY_SECRET` 环境变量：

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID="your-access-key-id"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="your-access-key-secret"
```

### 2. 阿里云默认凭据链（自动 fallback）

当环境变量 AK/SK 未设置时，`LoadConfigFromEnv()` 会自动 fallback 到阿里云默认凭据链（通过 `credentials-go` SDK）。

**凭据链优先级：**
1. 环境变量（`ALIBABA_CLOUD_ACCESS_KEY_ID` / `ALIBABA_CLOUD_ACCESS_KEY_SECRET`）
2. 配置文件（`~/.alibabacloud/credentials`）
3. ECS RAM 角色
4. OIDC Role SSO
5. IMDSv2

> **推荐：** 生产环境建议使用默认凭据链，避免硬编码密钥。

## 重试配置

SDK 内置了 SSE 连接断线自动重连能力，用于应对网络中断场景。

**配置项：**
- 环境变量 `VIBEOPS_MAX_RETRIES`：最大重试次数（默认：10）

**重试策略：**
- 指数退避：1s, 2s, 4s, 8s, ... 最大 30s
- 重连时使用 `action="reconnect"` 恢复会话
- 通过 timestamp 去重，避免处理重复事件

**正常结束标志：**
- 收到 `stream_done` 事件表示对话正常完成
- 如果连接在 `stream_done` 之前断开，SDK 会自动尝试重连

### 测试重试逻辑

使用 `-simulate-error` flag 模拟网络断连，验证重连机制：

```bash
go run ./cmd/chat/ -simulate-error
```

启用后行为：
1. 正常创建会话并发送消息
2. 接收到首个 SSE 事件后，主动模拟网络断连
3. 客户端输出重试日志并执行指数退避
4. 自动重连后通过 timestamp 去重，继续接收后续事件
5. 最终完整接收所有消息至 stream_done

## 示例程序

### 1. 交互式对话 (`cmd/chat`)

基础交互式对话示例，显示完整事件输出。

```bash
cd samples/golang
go run ./cmd/chat/
```

**功能特点：**
- 交互式命令行界面
- 完整事件流显示，包含原始 JSON 和解析详情
- 在同一会话中持续对话

### 2. 简洁输出模式 (`cmd/chat-simple`)

演示如何从 SSE 事件中仅提取最终文本内容，过滤掉工具调用、思考事件和中间消息。

```bash
go run ./cmd/chat-simple/
```

**功能特点：**
- 仅输出助手文本内容
- 过滤工具调用和思考事件
- 适合与其他系统集成

**输出示例：**
```
👤 请输入消息: 什么是全域智能运维平台？
🤖 发送消息: 什么是全域智能运维平台？
------------------------------------------------------------
全域智能运维平台是一项帮助您监控云资源的服务...
------------------------------------------------------------
📄 最终文本 (Final Text):
------------------------------
全域智能运维平台是一项帮助您监控云资源的服务...
```

### 3. 超时控制 (`cmd/chat-timeout`)

演示配置不同的超时时长（30秒、60秒、120秒）以及处理超时错误。

```bash
go run ./cmd/chat-timeout/
```

**功能特点：**
- 预定义超时配置（30秒、60秒、120秒）
- 支持自定义超时
- 超时错误处理，包含已用时间

**命令：**
- `help` - 显示帮助信息
- `timeout` - 显示超时配置
- `set <n>` - 切换到超时配置 n（1=30秒，2=60秒，3=120秒）
- `custom <duration>` - 使用自定义超时（如 `custom 45s`、`custom 2m`）
- `quit/exit` - 退出程序


### 4. 交互事件处理 (`cmd/chat-interactive`)

演示处理来自 STAROps Agent 的各种交互事件。

```bash
go run ./cmd/chat-interactive/
```

**功能特点：**
- 处理 `user_ack`（确认）事件
- 处理 `user_select`（选择）事件
- 处理 `user_input`（文本输入）事件
- 使用用户响应恢复对话
- 内置演示模式（`demo` 命令）

**交互事件类型：**
| 类型 | 描述 | 用户操作 |
|------|------|----------|
| `user_ack` | 确认请求 | 输入 y/yes 或 n/no |
| `user_select` | 从选项中选择 | 输入选项编号 |
| `user_input` | 自由文本输入 | 输入文本内容 |

### 5. 多轮对话 (`cmd/chat-multi-turn`)

演示在同一会话中跨多轮对话保持上下文。

```bash
go run ./cmd/chat-multi-turn/
```

**功能特点：**
- 在单个会话中进行 4 轮以上对话
- 展示 Agent 如何记住之前的上下文
- 演示连贯的多轮对话

**对话流程示例：**
1. 用户自我介绍
2. 用户询问全域智能运维平台相关问题
3. 用户测试上下文（询问 Agent 是否记得用户名字）
4. 用户请求对话摘要

### 6. 会话管理器 (`cmd/thread-manager`)

用于管理对话会话的命令行工具。

```bash
# 列出所有会话
go run ./cmd/thread-manager/ list

# 获取会话详情
go run ./cmd/thread-manager/ get <thread-id>

# 列出会话消息
go run ./cmd/thread-manager/ messages <thread-id>

# 删除会话
go run ./cmd/thread-manager/ delete <thread-id>
```

**命令：**
| 命令 | 描述 |
|------|------|
| `list` | 列出所有会话，包含 ID、标题、状态和创建时间 |
| `get <id>` | 获取指定会话的详细信息 |
| `messages <id>` | 列出会话中的所有消息 |
| `delete <id>` | 删除会话 |

### 7. 文件加载请求 (`cmd/chat-from-file`)

从 JSON 文件加载对话请求，适用于复杂请求场景。

```bash
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json
```

**参数：**
- `-file`（必需）：请求 JSON 文件的路径

### 8. 时间参数示例 (`cmd/chat-with-time`)

演示如何构建带有时间范围参数的 `userContext`。

```bash
# 使用默认时间范围（最近 15 分钟）
go run ./cmd/chat-with-time/ -message "最近有什么异常吗？"

# 指定自定义时间范围
go run ./cmd/chat-with-time/ -from 1770274812 -to 1770275712 -message "这段时间有什么问题吗？"
```

**参数：**
- `-from`：开始时间（Unix 时间戳，秒），默认为 15 分钟前
- `-to`：结束时间（Unix 时间戳，秒），默认为当前时间
- `-message`：要发送的消息内容


## API 参考

### 客户端模块

#### Config（配置）

```go
type Config struct {
    Workspace       string  // 工作空间 ID
    Endpoint        string  // API 端点
    Region          string  // 地域（默认：cn-hangzhou）
    AccessKeyID     string  // 阿里云 Access Key ID
    AccessKeySecret string  // 阿里云 Access Key Secret
    EmployeeName    string  // 数字员工名称（默认：default）
}
```

#### LoadConfigFromEnv（从环境变量加载配置）

```go
func LoadConfigFromEnv() (*Config, error)
```

从环境变量加载配置。当 AK/SK 环境变量未设置时，自动 fallback 到阿里云默认凭据链。

#### AgentClient（Agent 客户端）

```go
type AgentClient struct {
    // 私有字段
}

func NewAgentClient(cfg *Config) (*AgentClient, error)
```

使用提供的配置创建新的 Agent 客户端。

#### Chat 方法（对话方法）

```go
// 基础对话 - 使用默认变量
func (c *AgentClient) Chat(ctx context.Context, threadID, message string) <-chan *ChatEvent

// 带自定义变量的对话
func (c *AgentClient) ChatWithVariables(ctx context.Context, threadID, message string, 
    variables map[string]interface{}) <-chan *ChatEvent

// 带选项的对话（超时、回调、简洁模式）
func (c *AgentClient) ChatWithOptions(ctx context.Context, threadID, message string, 
    opts *ChatOptions) <-chan *ChatEvent

// 带超时的对话（便捷方法）
func (c *AgentClient) ChatWithTimeout(ctx context.Context, threadID, message string, 
    timeout time.Duration) <-chan *ChatEvent
```

#### ChatOptions（对话选项）

```go
type ChatOptions struct {
    Timeout    time.Duration          // 请求超时（0 = 无超时）
    Variables  map[string]interface{} // 自定义请求变量
    OnEvent    func(*ChatEvent)       // 事件回调
    SimpleMode bool                   // 简洁模式（仅文本输出）
}
```

#### ChatEvent（对话事件）

```go
type ChatEvent struct {
    Body       *starops.CreateChatResponseBody  // 响应体
    RawJSON    string                       // 原始 JSON 字符串
    StatusCode int32                        // HTTP 状态码
    IsDone     bool                         // 对话是否完成
    Error      error                        // 错误（如有）
}
```

#### CreateThread（创建会话）

```go
func (c *AgentClient) CreateThread(ctx context.Context) (string, error)
```

创建新的对话会话并返回会话 ID。


### 会话管理

#### ThreadInfo（会话信息）

```go
type ThreadInfo struct {
    ThreadID   string `json:"threadId"`    // 唯一会话标识符
    Title      string `json:"title"`       // 会话标题
    Status     string `json:"status"`      // 会话状态
    CreateTime string `json:"createTime"`  // 创建时间戳
    UpdateTime string `json:"updateTime"`  // 最后更新时间戳
}
```

#### ThreadMessage（会话消息）

```go
type ThreadMessage struct {
    Role      string `json:"role"`       // 消息角色（user/assistant/system）
    Content   string `json:"content"`    // 消息内容
    Timestamp string `json:"timestamp"`  // 消息时间戳
}
```

#### 会话管理方法

```go
// 分页列出会话
func (c *AgentClient) ListThreads(ctx context.Context, pageSize int) ([]*ThreadInfo, int64, error)

// 获取会话详情
func (c *AgentClient) GetThread(ctx context.Context, threadID string) (*ThreadInfo, error)

// 删除会话
func (c *AgentClient) DeleteThread(ctx context.Context, threadID string) error

// 获取会话消息
func (c *AgentClient) GetThreadData(ctx context.Context, threadID string, limit int) ([]*ThreadMessage, error)
```

### 交互处理器

#### InteractiveHandler（交互处理器）

```go
type InteractiveHandler struct {
    // 私有字段
}

func NewInteractiveHandler(client *AgentClient, timeout time.Duration) *InteractiveHandler
```

创建具有指定用户响应超时的交互处理器。

#### InteractiveResponse（交互响应）

```go
type InteractiveResponse struct {
    InteractionID string                 `json:"interactionId"`  // 交互标识符
    Type          types.InteractionType  `json:"type"`           // 交互类型
    Response      map[string]interface{} `json:"response"`       // 用户响应数据
}
```

#### 交互方法

```go
// 处理任意交互事件（分发到特定处理器）
func (h *InteractiveHandler) HandleEvent(ctx context.Context, 
    event *types.ItemEvent) (*InteractiveResponse, error)

// 处理用户确认（是/否）
func (h *InteractiveHandler) HandleUserAck(ctx context.Context, 
    payload *types.ItemInteractivePayload) (*InteractiveResponse, error)

// 处理用户选择（从选项中）
func (h *InteractiveHandler) HandleUserSelect(ctx context.Context, 
    payload *types.ItemInteractivePayload) (*InteractiveResponse, error)

// 处理用户文本输入
func (h *InteractiveHandler) HandleUserInput(ctx context.Context, 
    payload *types.ItemInteractivePayload) (*InteractiveResponse, error)

// 使用交互响应恢复对话
func (h *InteractiveHandler) ResumeChat(ctx context.Context, threadID string, 
    response *InteractiveResponse) <-chan *ChatEvent
```

#### 辅助函数

```go
// 检查事件是否为交互事件
func IsInteractiveEvent(event *types.ItemEvent) bool

// 从消息项中提取交互事件
func ExtractInteractiveEvents(item *types.MessageItem) []*types.ItemEvent
```


### 简洁打印器

#### SimplePrinter（简洁打印器）

```go
type SimplePrinter struct {
    // 私有字段
}

func NewSimplePrinter() *SimplePrinter
```

创建仅从事件中提取文本内容的简洁打印器。

#### 方法

```go
// 处理事件并提取文本内容
// 返回从此事件提取的文本（如有）
func (p *SimplePrinter) ProcessEvent(event *ChatEvent) string

// 获取所有累积的文本
func (p *SimplePrinter) GetFinalText() string

// 重置缓冲区以开始新对话
func (p *SimplePrinter) Reset()
```

**过滤行为：**
- ✅ 包含：助手文本内容
- ❌ 排除：工具调用、思考事件、Agent 调用、非助手消息

### 事件打印器

#### EventPrinter（事件打印器）

```go
type EventPrinter struct {
    PrintRawBody   bool  // 打印原始 JSON 体
    PrintParsed    bool  // 打印解析后的事件详情
    PrintSeparator bool  // 在事件之间打印分隔符
}

func NewEventPrinter(printRawBody, printParsed bool) *EventPrinter
```

创建具有可配置输出选项的事件打印器。

#### 方法

```go
// 打印单个事件
func (p *EventPrinter) PrintEvent(event *ChatEvent, eventIndex int)

// 格式化 JSON 以便美观打印
func PrettyPrintJSON(jsonStr string) (string, error)
```

### 日志模块

#### LogLevel（日志级别）

```go
type LogLevel int

const (
    LevelDebug LogLevel = iota  // 详细调试信息
    LevelInfo                    // 一般操作信息
    LevelWarn                    // 潜在问题警告
    LevelError                   // 错误信息
)
```

#### Logger（日志器）

```go
type Logger struct {
    // 私有字段
}

func NewLogger(level LogLevel, output io.Writer) *Logger
func NewLoggerFromEnv() *Logger  // 使用 LOG_LEVEL 环境变量
```

#### LogEntry（日志条目）

```go
type LogEntry struct {
    Timestamp string                 `json:"timestamp"`        // ISO 8601 时间戳
    Level     string                 `json:"level"`            // 日志级别
    Message   string                 `json:"message"`          // 日志消息
    Context   map[string]interface{} `json:"context,omitempty"` // 附加上下文
    Error     string                 `json:"error,omitempty"`  // 错误详情
    Stack     string                 `json:"stack,omitempty"`  // 堆栈跟踪（仅错误）
}
```

#### 日志方法

```go
func (l *Logger) Debug(msg string, ctx map[string]interface{})
func (l *Logger) Info(msg string, ctx map[string]interface{})
func (l *Logger) Warn(msg string, ctx map[string]interface{})
func (l *Logger) Error(msg string, err error, ctx map[string]interface{})

// 请求/响应日志便捷方法
func (l *Logger) LogRequest(threadID, message string, variables map[string]interface{})
func (l *Logger) LogResponse(threadID string, statusCode int32, rawJSON string, isDone bool, err error)
```


### 响应管理器

#### ResponseManager（响应管理器）

```go
type ResponseManager struct {
    // 私有字段
}

func NewResponseManager(baseDir string, logger *logger.Logger) *ResponseManager
```

创建用于发现和存储响应结构的响应管理器。

#### StructureInfo（结构信息）

```go
type StructureInfo struct {
    EventType   string                 `json:"eventType"`   // 事件类型
    Description string                 `json:"description"` // 结构描述
    Example     map[string]interface{} `json:"example"`     // 示例数据
    Fields      []FieldInfo            `json:"fields"`      // 字段信息
    UpdatedAt   string                 `json:"updatedAt"`   // 最后更新时间戳
}
```

#### FieldInfo（字段信息）

```go
type FieldInfo struct {
    Name        string `json:"name"`        // 字段名称
    Type        string `json:"type"`        // 字段类型
    Description string `json:"description"` // 字段描述
    Required    bool   `json:"required"`    // 是否必需
}
```

#### 方法

```go
// 处理事件并发现/更新结构
func (m *ResponseManager) ProcessEvent(event *types.MessageItem) error

// 保存结构到文件
func (m *ResponseManager) SaveStructure(eventType string, info *StructureInfo) error

// 加载已存储的结构
func (m *ResponseManager) LoadStructure(eventType string) (*StructureInfo, error)

// 检查结构是否有变化
func (m *ResponseManager) HasChanged(eventType string, newExample map[string]interface{}) bool
```

### 错误处理

#### 错误码

```go
const (
    ErrCodeConfigMissing      ErrorCode = "CONFIG_MISSING"       // 配置缺失
    ErrCodeConfigInvalid      ErrorCode = "CONFIG_INVALID"       // 配置无效
    ErrCodeClientCreate       ErrorCode = "CLIENT_CREATE"        // 客户端创建失败
    ErrCodeThreadCreate       ErrorCode = "THREAD_CREATE"        // 会话创建失败
    ErrCodeThreadNotFound     ErrorCode = "THREAD_NOT_FOUND"     // 会话不存在
    ErrCodeChatFailed         ErrorCode = "CHAT_FAILED"          // 对话操作失败
    ErrCodeTimeout            ErrorCode = "TIMEOUT"              // 操作超时
    ErrCodeCancelled          ErrorCode = "CANCELLED"            // 操作已取消
    ErrCodeNetworkError       ErrorCode = "NETWORK_ERROR"        // 网络错误
    ErrCodeAPIError           ErrorCode = "API_ERROR"            // API 错误
    ErrCodeParseError         ErrorCode = "PARSE_ERROR"          // 解析错误
    ErrCodeInteractiveTimeout ErrorCode = "INTERACTIVE_TIMEOUT"  // 交互超时
)
```

#### SDKError（SDK 错误）

```go
type SDKError struct {
    Code       ErrorCode              `json:"code"`                // 错误码
    Message    string                 `json:"message"`             // 错误消息
    Cause      error                  `json:"-"`                   // 底层错误
    Context    map[string]interface{} `json:"context,omitempty"`   // 附加上下文
    Suggestion string                 `json:"suggestion,omitempty"` // 建议操作
}

func (e *SDKError) Error() string
func (e *SDKError) Unwrap() error
```

#### 错误创建函数

```go
func NewSDKError(code ErrorCode, message string) *SDKError
func NewSDKErrorWithCause(code ErrorCode, message string, cause error) *SDKError

// 便捷函数
func ErrConfigMissing(missingVars []string) *SDKError
func ErrConfigInvalid(field, reason string) *SDKError
func ErrClientCreate(cause error) *SDKError
func ErrThreadCreate(cause error) *SDKError
func ErrThreadNotFound(threadID string) *SDKError
func ErrChatFailed(cause error) *SDKError
func ErrTimeout(duration string) *SDKError
func ErrCancelled() *SDKError
func ErrNetworkError(cause error) *SDKError
func ErrAPIError(code string, message string) *SDKError
func ErrParseError(cause error) *SDKError
func ErrInteractiveTimeout(duration string) *SDKError
```

#### TimeoutError（超时错误）

```go
type TimeoutError struct {
    Duration time.Duration  // 配置的超时时间
    Elapsed  time.Duration  // 实际已用时间
}

func IsTimeout(err error) bool  // 检查错误是否为超时
```


## 类型定义

### 事件类型

#### MessageItem（消息项）

```go
type MessageItem struct {
    ParentCallID string           `json:"parentCallId"` // 父调用 ID（根节点为空）
    CallID       string           `json:"callId"`       // 当前调用 ID
    Role         MessageRole      `json:"role"`         // 消息角色
    Timestamp    string           `json:"timestamp"`    // Unix 时间戳（纳秒）
    Contents     []*ItemContent   `json:"contents"`     // 文本/媒体内容
    Tools        []*ItemTool      `json:"tools"`        // 工具调用
    Agents       []*ItemAgent     `json:"agents"`       // 子 Agent 调用
    Events       []*ItemEvent     `json:"events"`       // 事件
    Artifacts    []map[string]any `json:"artifacts"`    // 产物
}
```

#### 消息角色

| 角色 | 常量 | 描述 |
|------|------|------|
| `user` | `MessageItemRoleUser` | 用户输入 |
| `assistant` | `MessageItemRoleAssistant` | Agent 回复或操作 |
| `system` | `MessageItemRoleSystem` | 系统消息 |

#### 内容类型

| 类型 | 常量 | 描述 |
|------|------|------|
| `text` | `MessageItemContentTypeText` | 纯文本 |
| `spin_text` | `MessageItemContentTypeSpinText` | 滚动文本（工作/思考过程） |
| `image` | `MessageItemContentTypeImage` | 图片 |

#### 事件类型

| 类型 | 常量 | 描述 |
|------|------|------|
| `thread_title_updated` | `EventTypeThreadTitleUpdated` | 会话标题已更新 |
| `error` | `EventTypeError` | 错误事件 |
| `thinking` | `EventTypeThinking` | 思考事件 |
| `interactive` | `EventTypeInteractive` | 交互事件 |
| `interactive_response` | `EventTypeInteractiveResponse` | 交互响应 |
| `task_finished` | `EventTypeTaskFinished` | 任务完成 |
| `cancel` | `EventTypeCancel` | 取消事件 |

#### 项目状态

| 状态 | 常量 | 描述 |
|------|------|------|
| `init` | `ItemStatusInit` | 初始化 |
| `start` | `ItemStatusStart` | 已开始 |
| `progress` | `ItemStatusProgress` | 进行中 |
| `suspended` | `ItemStatusSuspended` | 已暂停（等待用户） |
| `success` | `ItemStatusSuccess` | 成功完成 |
| `fail` | `ItemStatusFail` | 失败完成 |

#### 交互类型

| 类型 | 常量 | 描述 |
|------|------|------|
| `user_ack` | `InteractionTypeUserAck` | 用户确认 |
| `user_select` | `InteractionTypeUserSelect` | 用户选择 |
| `user_input` | `InteractionTypeUserInput` | 用户文本输入 |
| `sls_query` | `InteractionTypeSlsQuery` | SLS 查询 |


### 输入类型

#### UserInputParams（用户输入参数）

```go
type UserInputParams struct {
    RegionID    string `json:"region,omitempty"`
    Workspace   string `json:"workspace,omitempty"`
    Project     string `json:"project,omitempty"`
    LogStore    string `json:"logstore,omitempty"`
    MetricStore string `json:"metricstore,omitempty"`
    Language    string `json:"language,omitempty"`
    TimeZone    string `json:"timeZone,omitempty"`   // 标准时区格式
    TimeStamp   string `json:"timeStamp,omitempty"`  // Unix 时间戳（秒）
    UserContext string `json:"userContext,omitempty"` // []UserContext 的 JSON 字符串
    Config      string `json:"config,omitempty"`
    UserInteractiveResp map[string]interface{} `json:"userInteractive,omitempty"`
}
```

#### UserContext 类型

| 类型 | 常量 | 描述 |
|------|------|------|
| `metadata` | `UserContextTypeMetadata` | 元数据上下文（时间范围） |
| `entity` | `UserContextTypeEntity` | 实体上下文 |
| `sql_generation` | `UserContextTypeSQLGenerated` | SQL 生成上下文 |
| `spl_generation` | `UserContextTypeSPLGenerated` | SPL 生成上下文 |

#### 构建 UserContext 示例

```go
import (
    "encoding/json"
    "github.com/vibeops/samples/golang/types"
)

// 构建时间范围上下文
contexts := []types.UserContext{
    {
        Type: types.UserContextTypeMetadata,
        Data: types.MetadataUserData{
            FromTime: 1770274812,
            ToTime:   1770275712,
        },
    },
}

// 序列化为 JSON 字符串
userContextJSON, _ := json.Marshal(contexts)

// 在变量中使用
variables := map[string]interface{}{
    "workspace":   "your-workspace",
    "region":      "cn-hongkong",
    "language":    "zh",
    "timeZone":    "Asia/Shanghai",
    "timeStamp":   "1770275712",
    "userContext": string(userContextJSON),
}
```

## 请求文件格式

请求文件是定义对话请求的 JSON 文件。参见 `requests/starops/` 目录中的示例。

```json
{
    "region": "cn-hongkong",
    "digitalEmployeeName": "apsara-ops",
    "action": "create",
    "messages": [
        {
            "role": "user",
            "contents": [
                {
                    "type": "text",
                    "value": "cart 实体的延迟是多少？"
                }
            ]
        }
    ],
    "variables": {
        "workspace": "rca-benchmark",
        "region": "cn-hongkong",
        "language": "zh",
        "timeZone": "Asia/Shanghai",
        "userContext": "[{\"type\":\"metadata\",\"data\":{\"fromTime\":1770274812,\"toTime\":1770275712}}]"
    }
}
```

### 请求文件字段

| 字段 | 类型 | 描述 |
|------|------|------|
| `region` | string | API 地域 |
| `digitalEmployeeName` | string | 数字员工名称 |
| `threadId` | string | 会话 ID（可选，用于继续对话） |
| `action` | string | 操作类型（`create`） |
| `messages` | array | 消息对象数组 |
| `variables` | object | 请求变量，包含 userContext |

## SDK 依赖

- `github.com/alibabacloud-go/starops-20260428` - 阿里云 STAROps SDK
- `github.com/alibabacloud-go/darabonba-openapi/v2` - OpenAPI 客户端
- `github.com/alibabacloud-go/tea` - Tea 运行时
- `github.com/aliyun/credentials-go` - 阿里云凭据链 SDK（用于默认凭据链支持）

## 许可证

本项目采用 Apache License 2.0 许可证。
