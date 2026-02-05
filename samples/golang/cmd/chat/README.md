# VibeOps Chat 示例程序

这是一个简单的对话示例程序，演示如何使用阿里云 CMS SDK 与 VibeOps 数字员工进行对话，并打印详细的交互信息。

## 功能特点

- 使用环境变量配置，安全方便
- SSE 流式对话支持
- 打印原始 Response Body（JSON 格式）
- 基于 `types/events.go` 解析并格式化打印事件详情
- 支持多种事件类型：思考、工具调用、Agent 调用、交互事件等
- 正确处理 EOF，优雅退出

## 环境变量

| 变量名 | 必需 | 说明 | 示例 |
|--------|------|------|------|
| `VIBEOPS_WORKSPACE` | ✅ | 工作空间 ID | `rca-benchmark` |
| `VIBEOPS_ENDPOINT` | ✅ | API 端点地址 | `cms.cn-hongkong.aliyuncs.com` |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅ | 阿里云 Access Key ID | |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅ | 阿里云 Access Key Secret | |
| `VIBEOPS_REGION` | ✅ | 区域（**重要：必须与 Endpoint 对应**） | `cn-hongkong` |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | 数字员工名称（默认: default） | `apsara-ops` |

> **注意**: `VIBEOPS_REGION` 必须正确设置，否则会出现 `region_id not found in region_to_endpoint` 错误。

## 使用方法

### 1. 设置环境变量

```bash
export VIBEOPS_WORKSPACE="xxx"
export VIBEOPS_ENDPOINT="cms.cn-hongkong.aliyuncs.com"
export VIBEOPS_REGION="cn-hongkong"
export ALIBABA_CLOUD_ACCESS_KEY_ID="your-access-key-id"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="your-access-key-secret"
export VIBEOPS_EMPLOYEE_NAME="apsara-ops" // 默认数字员工
```

### 2. 运行程序

```bash
# 在 samples/golang 目录下
cd samples/golang

# 下载依赖
go mod tidy

# 运行程序
go run ./cmd/chat/
```

### 3. 开始对话

程序启动后会创建一个会话，然后进入交互式对话模式：

```
🚀 VibeOps Chat 示例程序
============================================================
📋 配置信息:
  - Workspace: rca-benchmark
  - Endpoint: cms.cn-hongkong.aliyuncs.com
  - Region: cn-hongkong
  - Employee: apsara-ops

📝 创建会话...
✅ 会话创建成功, ThreadID: thread-xxx

👤 请输入消息 (输入 'quit' 退出): 当前工作空间下，有哪些服务
```

## 输出示例

每个 SSE 事件会打印：
1. **原始 Body**: 格式化的 JSON 内容
2. **解析详情**: 基于 `types/events.go` 结构解析的详细信息

```
============================== 事件 #1 ==============================

📦 原始 Body:
{
  "messages": [
    {
      "callId": "thread-xxx",
      "contents": [
        {
          "append": false,
          "lastChunk": true,
          "type": "spin_text",
          "value": "让我分析一下..."
        }
      ],
      "role": "assistant",
      "timestamp": "1770278209345158970",
      "version": "v0.1.0"
    }
  ],
  "requestId": "xxx",
  "traceId": "xxx"
}

📋 解析详情:
  📌 角色: assistant
  🔗 CallID: thread-xxx
  📝 内容:
    [0] 类型: spin_text
        值: 让我分析一下...
        最后块: true
```

### 事件类型

程序支持解析以下事件类型：

| 事件类型 | 说明 |
|----------|------|
| `spin_text` | 思考中提示文本 |
| `text` | 文本回复（支持流式追加） |
| `thinking` | 思考内容 |
| `thread_title_updated` | 会话标题更新 |
| `task_finished` | 任务完成 |
| `interactive` | 交互式事件 |
| 工具调用 | DataAgent、generate_diagnosis_report 等 |

## 请求结构

程序构建的 CreateChat 请求包含以下关键字段：

```go
req := &cms.CreateChatRequest{}
req.SetAction("create")
req.SetThreadId(threadID)
req.SetDigitalEmployeeName(employeeName)
req.SetMessages(messages)
req.SetVariables(map[string]interface{}{
    "workspace": workspace,
    "region":    region,      // 关键！
    "language":  "zh",
    "timeZone":  "Asia/Shanghai",
    "timeStamp": timestamp,
})
```

## 代码结构

```
samples/golang/
├── cmd/
│   └── chat/
│       ├── main.go          # 主程序
│       └── README.md        # 本文件
├── types/
│   └── events.go            # 事件类型定义
├── go.mod
└── go.sum
```

## SDK 依赖

- `github.com/alibabacloud-go/cms-20240330/v6` - 阿里云 CMS SDK

## 常见问题

### 1. region_id not found in region_to_endpoint

确保 `VIBEOPS_REGION` 环境变量设置正确，且与 `VIBEOPS_ENDPOINT` 对应。例如：
- Endpoint: `cms.cn-hongkong.aliyuncs.com` → Region: `cn-hongkong`
- Endpoint: `cms.cn-hangzhou.aliyuncs.com` → Region: `cn-hangzhou`

### 2. DigitalEmployeeNotExist

指定的数字员工不存在，请检查 `VIBEOPS_EMPLOYEE_NAME` 是否正确。

### 3. EOF 错误导致程序退出

这是正常行为。当输入流结束时（如管道输入 `echo "xxx" | go run ./cmd/chat/`），程序会优雅退出。
