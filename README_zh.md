# STAROps SDK 示例

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](samples/java/)
[![Java8](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](samples/java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](samples/typescript/)

[English](README.md)

为阿里云 STAROps（全域智能运维平台服务）数字员工提供的生产就绪示例程序，支持多种编程语言。

---

## 工作原理

STAROps 数字员工的核心交互模式是**自然语言对话**——你用自然语言描述需求，数字员工理解语义后自动完成查询、分析和回答。

```
┌───────────────┐       HTTP POST        ┌──────────────────┐
│   你的应用     │  ──────────────────►   │  STAROps 数字员工 API │
│  (SDK 客户端)  │  ◄── SSE 流式响应 ──  │                  │
└───────────────┘                        └──────────────────┘
```

**交互流程：**

1. **创建会话** — 调用 `createThread` 获取一个 `threadId`，用于多轮对话的上下文关联
2. **发送消息** — 构造包含自然语言问题和上下文变量的请求，通过 `chat` 方法发送给数字员工
3. **流式接收** — 服务端通过 SSE（Server-Sent Events）实时推送结构化的处理结果
4. **多轮续聊** — 使用同一 `threadId` 可以继续追问，数字员工会保持上下文

### 请求结构

每次调用需要构造完整的请求参数。以下是一个典型请求的结构：

```json
{
    "region": "cn-hongkong",
    "digitalEmployeeName": "apsara-ops",
    "threadId": "",
    "action": "create",
    "messages": [
        {
            "role": "user",
            "contents": [{ "type": "text", "value": "查询最近1小时的告警数量" }]
        }
    ],
    "variables": {
        "workspace": "your-workspace-id",
        "region": "cn-hongkong",
        "project": "your-project-id",
        "language": "zh",
        "timeZone": "Asia/Shanghai",
        "timeStamp": "1770710677",
        "userContext": "[{\"type\":\"metadata\",\"data\":{\"fromTime\":1770685976,\"toTime\":1770686876}}]",
        "startTime": 1770685976,
        "endTime": 1770686876
    }
}
```

**顶层字段：**

| 字段 | 说明 |
|------|------|
| `region` | 地域，如 `cn-hongkong`、`cn-hangzhou` |
| `digitalEmployeeName` | 数字员工名称 |
| `threadId` | 会话 ID，空字符串表示创建新会话 |
| `action` | 操作类型，通常为 `create`, 此外还有 `stop`、`reconnect` |
| `messages` | 用户消息，包含 `role` 和 `contents`（文本内容） |

**`variables` 字段：**

| 字段 | 必需 | 说明 |
|------|------|------|
| `workspace` | ✅ | 工作空间 ID |
| `region` | ✅ | 地域 |
| `project` | ✅ | 项目名称 |
| `language` | ✅ | 语言，`zh` 或 `en` |
| `timeZone` | ✅ | 时区，如 `Asia/Shanghai` |
| `timeStamp` | ✅ | 当前 Unix 时间戳（秒） |
| `userContext` | ✅ | 用户上下文 JSON 字符串，至少包含 `metadata`（时间范围），也可携带 `entity`（实体信息）等 |
| `logstore` | ❌ | 日志库名称，日志分析 / SQL 生成场景需要 |
| `metricstore` | ❌ | 指标库名称，指标查询场景需要 |
| `startTime` / `endTime` | ❌ | 查询时间范围（Unix 时间戳，秒） |
| `skill` | ❌ | 技能标识，仅 SQL 生成场景需传入 `sql_generation` |

### 推荐使用方式

实际开发中，建议参考 `requests/starops/` 下的示例文件，**构造一个临时 JSON 文件**，然后通过 `chat-from-file` 示例直接调用：

```bash
# 1. 复制一个最接近你场景的示例文件
cp requests/starops/data_agent.json /tmp/my_request.json

# 2. 编辑填入你的实际参数（workspace、project、问题内容等）

# 3. 使用任意语言的 chat-from-file 发起调用
go run ./cmd/chat-from-file/ -file /tmp/my_request.json
```

这样可以快速验证不同场景的请求参数，无需编写代码。

### 响应结构

服务端通过 SSE 流式推送结构化的 `MessageItem`，每条消息包含以下字段（详见 [`samples/golang/types/`](samples/golang/types/)）：

| 字段 | 说明 |
|------|------|
| `role` | 消息角色：`user`、`assistant`、`system` |
| `contents` | 文本或富媒体内容，类型包括 `text`（纯文本）、`spin_text`（思考过程）、`image`（图片） |
| `tools` | 工具调用详情：工具名称、参数、执行状态（`start` / `success` / `fail`）、返回结果 |
| `agents` | 子 Agent 调用详情：Agent 名称、输入、输出 |
| `events` | 事件通知：`thinking`（思考）、`interactive`（交互确认）、`task_finished`（任务完成）、`error`（错误） |
| `artifacts` | 产物输出，遵循 Google A2A 协议格式 |

消息通过 `parentCallId` / `callId` 构建树形调用链，可以还原完整的 Agent 执行过程。你可以针对不同场景解析和优化展示逻辑。

---

## 功能概览

数字员工支持多种运维场景，你只需用自然语言描述需求：

| 场景 | 示例问题 | 说明 |
|------|---------|------|
| **实体查询** | "cart 服务延迟多少" | 查询 APM 服务、主机等实体的性能指标，需在 `userContext` 中携带实体信息 |
| **根因定位** | "payment 服务报错，帮我排查根因" | 根据错误日志，自动定位根因 |
| **SQL 生成** | "查看有多少种类的 admin_emails" | 自然语言自动转 SQL，需指定 `logstore` 和 `skill: sql_generation` |
| **通用对话** | "统计错误数量" | 开放式运维问答，仅需基础 `variables` |

> 不同场景需要的 `variables` 字段略有差异，参考 `requests/starops/` 下对应的 JSON 文件即可了解具体参数。

---

## 语言支持

| 语言 | 状态 | 目录 | 最低版本 |
|------|------|------|----------|
| **Go** | ✅ 已完成 | [`samples/golang/`](samples/golang/) | Go 1.21+ |
| **Java** | ✅ 已完成 | [`samples/java/`](samples/java/) | Java 11+ |
| **Java 8** | ✅ 已完成 | [`samples/java8/`](samples/java8/) | Java 8+ |
| **Python** | ✅ 已完成 | [`samples/python/`](samples/python/) | Python 3.8+ |
| **TypeScript** | ✅ 已完成 | [`samples/typescript/`](samples/typescript/) | Node 18+ |

> **说明**: Java 8 版本（`samples/java8/`）是 Java 11 项目的语法兼容分支，适用于只能使用 Java 8 的环境。功能完全一致。

---

## 快速开始

### 前置条件

- 具有 STAROps 访问权限的阿里云账号
- Access Key ID 和 Secret

### 环境变量

所有语言示例使用相同的环境变量（通过 `.env` 文件）：

| 变量 | 必需 | 说明 |
|------|------|------|
| `VIBEOPS_ENDPOINT` | ✅ | STAROps API 端点，格式: `cms.{region-id}.aliyuncs.com` |
| `VIBEOPS_REGION` | ❌ | 地域，默认 `cn-hangzhou` |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅ | Access Key ID |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅ | Access Key Secret |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | 数字员工名称，默认 `default` |

### Go

```bash
cd samples/golang
cp .env.example .env  # 编辑填入您的凭证
go run ./cmd/chat/
```

### Java (11+)

```bash
cd samples/java
cp .env.example .env  # 编辑填入您的凭证
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

### Java 8

```bash
cd samples/java8
cp .env.example .env  # 编辑填入您的凭证
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

### Python

```bash
cd samples/python
python -m venv .venv && source .venv/bin/activate
pip install -e .
cp .env.example .env  # 编辑填入您的凭证
python -m starops_sdk_samples.examples.chat
```

### TypeScript

```bash
cd samples/typescript
npm install
cp .env.example .env  # 编辑填入您的凭证
npx tsx src/examples/chat.ts
```

---

## 示例程序

所有语言实现相同的示例集：

| 示例 | 说明 |
|------|------|
| `chat` | 交互式多轮对话 |
| `chat-from-file` | 从 JSON 文件加载请求 |
| `thread-manager` | 会话管理（列出/查看/删除） |

### chat-from-file

支持从 `requests/starops/` 下的共享 JSON 文件加载请求参数：

```bash
# Go
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json

# Java / Java 8
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
    -Dexec.args="-file ../../requests/starops/entity.json"

# Python
python -m starops_sdk_samples.examples.chat_from_file -file ../../requests/starops/entity.json

# TypeScript
npx tsx src/examples/chat-from-file.ts -file ../../requests/starops/entity.json
```

### thread-manager

```bash
# Go
go run ./cmd/thread-manager/ list

# Java / Java 8
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" \
    -Dexec.args="list"

# Python
python -m starops_sdk_samples.examples.thread_manager list

# TypeScript
npx tsx src/examples/thread-manager.ts list
```

---

## 项目结构

```
.
├── README.md
├── README_zh.md
├── .env.example
├── requests/                          # 共享请求 JSON 文件
│   └── starops/
│       ├── entity.json                # 实体查询
│       ├── sls_chat.json              # SLS 日志查询
│       ├── sql_generation.json        # SQL 生成
│       ├── general_chat.json          # 通用对话
│       └── data_agent.json            # 数据 Agent
│
└── samples/
    ├── golang/                        # Go (1.21+)
    ├── java/                          # Java (11+)
    ├── java8/                         # Java (8+) - 语法兼容分支
    ├── python/                        # Python (3.8+)
    └── typescript/                    # TypeScript (Node 18+)
```

---

## Java 与 Java 8 的区别

`java8` 项目是 `java` 项目的纯语法适配版本。主要区别：

| 特性 | Java (11+) | Java 8 |
|------|-----------|--------|
| `var` 关键字 | ✅ | ❌ → 显式类型声明 |
| `String.repeat()` | ✅ | ❌ → `repeatStr()` 辅助方法 |
| `Map.putIfAbsent()` | ✅ | ✅ → `containsKey()` + `put()` |
| dotenv-java | 3.0.0 | 2.3.2 |
| mockito | 5.5.0 | 4.11.0 |
| Lambda / Runnable | Lambda | 匿名 `Runnable` |

---

## 许可证

Apache License 2.0
