# VibeOps STAROps SDK Samples for Go

阿里云 STAROps SDK Go 语言示例程序。

## 快速开始

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入配置

# 2. 运行
go run ./cmd/chat/
```

## 环境变量

| 变量 | 必需 | 说明 |
|-----|------|-----|
| VIBEOPS_ENDPOINT | ✅ | STAROps API 端点，格式: `starops.{region-id}.aliyuncs.com` |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ❌* | Access Key ID（如使用默认凭据链则可选） |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ❌* | Access Key Secret（如使用默认凭据链则可选） |
| VIBEOPS_EMPLOYEE_NAME | ❌ | 数字员工名称 (默认 default) |
| VIBEOPS_MAX_RETRIES | ❌ | SSE 连接最大重试次数（默认 10） |

> *当环境变量 AK/SK 未设置时，SDK 会自动使用阿里云默认凭据链。

## Credential Management

The SDK supports two credential acquisition methods:

### 1. Environment Variables (Highest Priority)

Directly set `ALIBABA_CLOUD_ACCESS_KEY_ID` and `ALIBABA_CLOUD_ACCESS_KEY_SECRET` environment variables:

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID="your-access-key-id"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="your-access-key-secret"
```

### 2. Alibaba Cloud Default Credential Chain (Auto-fallback)

When environment variable AK/SK are not set, `LoadConfigFromEnv()` automatically falls back to the Alibaba Cloud default credential chain (via `credentials-go` SDK).

**Credential chain priority:**
1. Environment variables (`ALIBABA_CLOUD_ACCESS_KEY_ID` / `ALIBABA_CLOUD_ACCESS_KEY_SECRET`)
2. Credentials file (`~/.alibabacloud/credentials`)
3. ECS RAM Role
4. OIDC Role SSO
5. IMDSv2

> **Recommendation:** Use the default credential chain in production environments to avoid hardcoding secrets.

## Retry Configuration

The SDK has built-in SSE connection auto-reconnect capability for handling network interruptions.

**Configuration:**
- Environment variable `VIBEOPS_MAX_RETRIES`: Maximum retry attempts (default: 10)

**Retry strategy:**
- Exponential backoff: 1s, 2s, 4s, 8s, ... up to 30s maximum
- Uses `action="reconnect"` to resume the session on reconnection
- Deduplication via timestamp to avoid processing duplicate events

**Normal completion:**
- Receiving a `stream_done` event indicates the conversation completed normally
- If the connection drops before `stream_done`, the SDK will automatically attempt reconnection

### Testing Retry Logic

Use the `-simulate-error` flag to simulate a network disconnection and verify the reconnection mechanism:

```bash
go run ./cmd/chat/ -simulate-error
```

**Behavior when enabled:**
1. Creates a session and sends a message normally
2. After receiving the first SSE event, actively simulates a network disconnection
3. Client outputs retry logs and performs exponential backoff
4. After automatic reconnection, deduplicates via timestamp and continues receiving subsequent events
5. Eventually receives all messages completely until `stream_done`

## 示例程序

### chat - 交互式对话

```bash
go run ./cmd/chat/
```

支持多轮对话，在同一会话中保持上下文。

### chat-from-file - 从文件加载请求

```bash
# 处理单个文件
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json

# 批量处理目录
go run ./cmd/chat-from-file/ -dir ../../requests/starops/
```

日志自动输出到 `output/` 目录。

### chat-interactive - 交互事件处理

```bash
go run ./cmd/chat-interactive/
```

处理 Agent 返回的确认、选择、输入等交互事件。

### thread-manager - 会话管理

```bash
go run ./cmd/thread-manager/ list              # 列出会话
go run ./cmd/thread-manager/ get <thread-id>   # 查看详情
go run ./cmd/thread-manager/ delete <thread-id> # 删除会话
```

## 请求文件

`requests/starops/` 目录包含各类请求示例：

| 文件 | 场景 |
|-----|------|
| entity.json | 实体查询 |
| sls_query.json | 日志查询 |
| text_to_sql.json | SQL 生成 |
| alert_management.json | 告警管理 |

## 目录结构

```
samples/golang/
├── cmd/
│   ├── chat/              # 交互式对话
│   ├── chat-from-file/    # 从文件加载请求
│   ├── chat-interactive/  # 交互事件处理
│   └── thread-manager/    # 会话管理
├── internal/client/       # 客户端实现
└── types/                 # 类型定义
```
