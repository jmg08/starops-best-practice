# VibeOps CMS SDK Samples for Go

阿里云 CMS SDK Go 语言示例程序。

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
| VIBEOPS_WORKSPACE | ✅ | 工作空间 ID |
| VIBEOPS_ENDPOINT | ✅ | API 端点 |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ✅ | Access Key ID |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ✅ | Access Key Secret |
| VIBEOPS_EMPLOYEE_NAME | ❌ | 数字员工名称 (默认 default) |

## 示例程序

### chat - 交互式对话

```bash
go run ./cmd/chat/
```

支持多轮对话，在同一会话中保持上下文。

### chat-from-file - 从文件加载请求

```bash
# 处理单个文件
go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json

# 批量处理目录
go run ./cmd/chat-from-file/ -dir ../../requests/cms/
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

`requests/cms/` 目录包含各类请求示例：

| 文件 | 场景 |
|-----|------|
| entity.json | 实体查询 |
| metric_query.json | 指标查询 |
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
