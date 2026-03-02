# CMS SDK Samples | CMS SDK 示例

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](samples/java/)
[![Java8](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](samples/java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](samples/typescript/)

Production-ready sample programs for interacting with Alibaba Cloud CMS (Cloud Monitor Service) digital employees across multiple programming languages.

为阿里云 CMS（云监控服务）数字员工提供的生产就绪示例程序，支持多种编程语言。

---

## Language Support | 语言支持

| Language | Status | Directory | Min Version |
|----------|--------|-----------|-------------|
| **Go** | ✅ Complete | [`samples/golang/`](samples/golang/) | Go 1.21+ |
| **Java** | ✅ Complete | [`samples/java/`](samples/java/) | Java 11+ |
| **Java 8** | ✅ Complete | [`samples/java8/`](samples/java8/) | Java 8+ |
| **Python** | ✅ Complete | [`samples/python/`](samples/python/) | Python 3.8+ |
| **TypeScript** | ✅ Complete | [`samples/typescript/`](samples/typescript/) | Node 18+ |

| 语言 | 状态 | 目录 | 最低版本 |
|------|------|------|----------|
| **Go** | ✅ 已完成 | [`samples/golang/`](samples/golang/) | Go 1.21+ |
| **Java** | ✅ 已完成 | [`samples/java/`](samples/java/) | Java 11+ |
| **Java 8** | ✅ 已完成 | [`samples/java8/`](samples/java8/) | Java 8+ |
| **Python** | ✅ 已完成 | [`samples/python/`](samples/python/) | Python 3.8+ |
| **TypeScript** | ✅ 已完成 | [`samples/typescript/`](samples/typescript/) | Node 18+ |

> **Note**: The Java 8 variant (`samples/java8/`) is a syntax-compatible fork of the Java 11 project for environments restricted to Java 8. Functionality is identical.
>
> **说明**: Java 8 版本（`samples/java8/`）是 Java 11 项目的语法兼容分支，适用于只能使用 Java 8 的环境。功能完全一致。

---

## Quick Start | 快速开始

### Prerequisites | 前置条件

- Alibaba Cloud account with CMS access | 具有 CMS 访问权限的阿里云账号
- Access Key ID and Secret | Access Key ID 和 Secret

### Environment Variables | 环境变量

All language samples use the same environment variables (via `.env` file):

所有语言示例使用相同的环境变量（通过 `.env` 文件）：

| Variable | Required | Description |
|----------|----------|-------------|
| `VIBEOPS_WORKSPACE` | ✅ | Workspace ID（工作空间 ID） |
| `VIBEOPS_ENDPOINT` | ✅ | API endpoint（API 端点） |
| `VIBEOPS_REGION` | ❌ | Region, default `cn-hangzhou`（地域） |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅ | Access Key ID |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅ | Access Key Secret |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | Employee name, default `default`（数字员工名称） |

### Go

```bash
cd samples/golang
cp .env.example .env  # edit with your credentials
go run ./cmd/chat/
```

### Java (11+)

```bash
cd samples/java
cp .env.example .env  # edit with your credentials
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.Chat"
```

### Java 8

```bash
cd samples/java8
cp .env.example .env  # edit with your credentials
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.Chat"
```

### Python

```bash
cd samples/python
python -m venv .venv && source .venv/bin/activate
pip install -e .
cp .env.example .env  # edit with your credentials
python -m cms_sdk_samples.examples.chat
```

### TypeScript

```bash
cd samples/typescript
npm install
cp .env.example .env  # edit with your credentials
npx tsx src/examples/chat.ts
```

---

## Sample Programs | 示例程序

All languages implement the same set of examples:

所有语言实现相同的示例集：

| Sample | Description | 描述 |
|--------|-------------|------|
| `chat` | Interactive multi-turn chat | 交互式多轮对话 |
| `chat-from-file` | Load request from JSON file | 从 JSON 文件加载请求 |
| `thread-manager` | List/get/delete conversation threads | 会话管理（列出/查看/删除） |

### chat-from-file

Supports loading request parameters from shared JSON files in `requests/cms/`:

支持从 `requests/cms/` 下的共享 JSON 文件加载请求参数：

```bash
# Go
go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json

# Java / Java 8
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatFromFile" \
    -Dexec.args="-file ../../requests/cms/entity.json"

# Python
python -m cms_sdk_samples.examples.chat_from_file -file ../../requests/cms/entity.json

# TypeScript
npx tsx src/examples/chat-from-file.ts -file ../../requests/cms/entity.json
```

### thread-manager

```bash
# Go
go run ./cmd/thread-manager/ list

# Java / Java 8
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ThreadManager" \
    -Dexec.args="list"

# Python
python -m cms_sdk_samples.examples.thread_manager list

# TypeScript
npx tsx src/examples/thread-manager.ts list
```

---

## Project Structure | 项目结构

```
.
├── README.md
├── .env.example
├── requests/                          # Shared request JSON files | 共享请求 JSON 文件
│   └── cms/
│       ├── entity.json                # Entity query | 实体查询
│       ├── sls_chat.json              # SLS log query (sql_generation) | SLS 日志查询
│       ├── sql_generation.json        # SQL generation | SQL 生成
│       ├── general_chat.json          # General chat | 通用对话
│       ├── metric_query.json          # Metric query | 指标查询
│       └── data_agent.json            # Data agent | 数据 Agent
│
└── samples/
    ├── golang/                        # Go (1.21+)
    ├── java/                          # Java (11+)
    ├── java8/                         # Java (8+) - syntax-compatible fork
    ├── python/                        # Python (3.8+)
    └── typescript/                    # TypeScript (Node 18+)
```

---

## Java vs Java 8 | Java 与 Java 8 的区别

The `java8` project is a syntax-only adaptation of the `java` project. Key differences:

`java8` 项目是 `java` 项目的纯语法适配版本。主要区别：

| Feature | Java (11+) | Java 8 |
|---------|-----------|--------|
| `var` keyword | ✅ | ❌ → explicit types |
| `String.repeat()` | ✅ | ❌ → `repeatStr()` helper |
| `Map.putIfAbsent()` | ✅ | ✅ → `containsKey()` + `put()` |
| dotenv-java | 3.0.0 | 2.3.2 |
| mockito | 5.5.0 | 4.11.0 |
| Lambda / Runnable | Lambda | Anonymous `Runnable` |

> **Note on Java SDK**: The Alibaba Cloud CMS Java SDK (`cms20240330:6.0.1`) does not have a `createChatWithSSE()` method. The `createChat()` API returns an SSE stream, but the `tea-openapi` runtime attempts to parse it as JSON, causing an exception. The SDK samples automatically extract SSE data from the exception chain and parse each event — no manual handling is needed. Python, Go, and TypeScript all support SSE streaming natively via `createChatWithSSE()`.
>
> **Java SDK 说明**: 阿里云 CMS Java SDK（`cms20240330:6.0.1`）没有 `createChatWithSSE()` 方法。`createChat()` API 返回 SSE 流，但 `tea-openapi` 运行时会尝试将其解析为 JSON 导致异常。SDK 示例会自动从异常链中提取 SSE 数据并逐条解析事件，无需手动处理。Python、Go、TypeScript 均通过 `createChatWithSSE()` 原生支持 SSE 流式输出。

---

## Documentation | 文档

- [Go README](samples/golang/README.md) | [Go 中文文档](samples/golang/README_zh.md)
- [Java README](samples/java/README.md) | [Java 中文文档](samples/java/README_zh.md)
- [Language Roadmap | 语言路线图](docs/LANGUAGE_ROADMAP.md)

---

## License | 许可证

Apache License 2.0
