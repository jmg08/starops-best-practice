# STAROps SDK Samples

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](samples/java/)
[![Java8](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](samples/java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](samples/typescript/)

[中文文档](README_zh.md)

Production-ready sample programs for interacting with Alibaba Cloud STAROps digital employees across multiple programming languages.

---

## How It Works

The core interaction model of STAROps digital employees is **natural language conversation** — you describe your needs in plain language, and the digital employee understands, queries, analyzes, and responds automatically.

```
┌───────────────┐       HTTP POST        ┌──────────────────┐
│  Your App     │  ──────────────────►   │  STAROps Digital      │
│  (SDK Client) │  ◄── SSE Streaming ──  │  Employee API     │
└───────────────┘                        └──────────────────┘
```

**Interaction flow:**

1. **Create a thread** — Call `createThread` to get a `threadId` for multi-turn context
2. **Send a message** — Construct a request with the natural language question and context variables, then send it via the `chat` method
3. **Stream the response** — The server pushes structured results in real-time via SSE (Server-Sent Events)
4. **Continue the conversation** — Reuse the same `threadId` to ask follow-up questions with full context

### Request Structure

Each call requires a fully constructed request. Here is a typical request structure:

```json
{
    "region": "cn-hongkong",
    "digitalEmployeeName": "apsara-ops",
    "threadId": "",
    "action": "create",
    "messages": [
        {
            "role": "user",
            "contents": [{ "type": "text", "value": "How many alerts in the last hour?" }]
        }
    ],
    "variables": {
        "workspace": "your-workspace-id",
        "region": "cn-hongkong",
        "project": "your-project-id",
        "language": "en",
        "timeZone": "Asia/Shanghai",
        "timeStamp": "1770710677",
        "userContext": "[{\"type\":\"metadata\",\"data\":{\"fromTime\":1770685976,\"toTime\":1770686876}}]",
        "startTime": 1770685976,
        "endTime": 1770686876
    }
}
```

**Top-level fields:**

| Field | Description |
|-------|-------------|
| `region` | Region, e.g. `cn-hongkong`, `cn-hangzhou` |
| `digitalEmployeeName` | Digital employee name |
| `threadId` | Thread ID; empty string creates a new thread |
| `action` | Operation type, typically `create` |
| `messages` | User messages with `role` and `contents` (text content) |

**`variables` fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `workspace` | ✅ | Workspace ID |
| `region` | ✅ | Region |
| `project` | ✅ | Project name |
| `language` | ✅ | Language, `zh` or `en` |
| `timeZone` | ✅ | Timezone, e.g. `Asia/Shanghai` |
| `timeStamp` | ✅ | Current Unix timestamp (seconds) |
| `userContext` | ✅ | User context JSON string; must include at least `metadata` (time range), and optionally `entity` (entity info), etc. |
| `logstore` | ❌ | Log store name, required for log analysis / SQL generation |
| `metricstore` | ❌ | Metric store name, required for metric queries |
| `startTime` / `endTime` | ❌ | Query time range (Unix timestamp, seconds) |
| `skill` | ❌ | Skill identifier, only required for SQL generation (`sql_generation`) |

### Recommended Usage

In practice, we recommend copying a sample JSON file from `requests/starops/`, **creating a temporary JSON file** with your parameters, and running it via `chat-from-file`:

```bash
# 1. Copy the sample closest to your scenario
cp requests/starops/data_agent.json /tmp/my_request.json

# 2. Edit with your actual parameters (workspace, project, question, etc.)

# 3. Run with any language's chat-from-file
go run ./cmd/chat-from-file/ -file /tmp/my_request.json
```

This lets you quickly validate request parameters for different scenarios without writing code.

### Response Structure

The server streams structured `MessageItem` objects via SSE. Each message contains the following fields (see [`samples/golang/types/`](samples/golang/types/) for full definitions):

| Field | Description |
|-------|-------------|
| `role` | Message role: `user`, `assistant`, `system` |
| `contents` | Text or rich media content; types include `text` (plain text), `spin_text` (thinking process), `image` |
| `tools` | Tool call details: tool name, arguments, execution status (`start` / `success` / `fail`), results |
| `agents` | Sub-agent call details: agent name, inputs, outputs |
| `events` | Event notifications: `thinking`, `interactive` (user confirmation), `task_finished`, `error` |
| `artifacts` | Output artifacts, following the Google A2A protocol format |

Messages are linked via `parentCallId` / `callId` to form a tree-shaped call chain, enabling full reconstruction of the agent execution process. You can parse and optimize the display logic for different scenarios.

---

## Features

The digital employee supports a variety of operations scenarios. Describe your needs in natural language:

| Scenario | Example Question | Description |
|----------|-----------------|-------------|
| **Entity Query** | "What's the latency of the cart service?" | Query performance metrics for APM services, hosts, etc. Requires entity info in `userContext` |
| **Metric Query** | "ECS instances with CPU usage above 80% in the last hour" | Query Cloud Monitor metrics; requires `metricstore` |
| **Log Analysis** | "Show 5xx error logs from the last 15 minutes" | Intelligent SLS log query and analysis; requires `logstore` |
| **SQL Generation** | "How many types of admin_emails are there?" | Auto-converts natural language to SQL; requires `logstore` and `skill: sql_generation` |
| **Data Insights** | "How many alerts in the last hour?" | Complex data analysis via the Data Agent |
| **General Chat** | "Count the number of errors" | Open-ended operations Q&A, only basic `variables` needed |

> Different scenarios require slightly different `variables` fields. Refer to the corresponding JSON files under `requests/starops/` for specific parameters.

---

## Language Support

| Language | Status | Directory | Min Version |
|----------|--------|-----------|-------------|
| **Go** | ✅ Complete | [`samples/golang/`](samples/golang/) | Go 1.21+ |
| **Java** | ✅ Complete | [`samples/java/`](samples/java/) | Java 11+ |
| **Java 8** | ✅ Complete | [`samples/java8/`](samples/java8/) | Java 8+ |
| **Python** | ✅ Complete | [`samples/python/`](samples/python/) | Python 3.8+ |
| **TypeScript** | ✅ Complete | [`samples/typescript/`](samples/typescript/) | Node 18+ |

> **Note**: The Java 8 variant (`samples/java8/`) is a syntax-compatible fork of the Java 11 project for environments restricted to Java 8. Functionality is identical.

---

## Quick Start

### Prerequisites

- Alibaba Cloud account with STAROps access
- Access Key ID and Secret

### Environment Variables

All language samples use the same environment variables (via `.env` file):

| Variable | Required | Description |
|----------|----------|-------------|
| `VIBEOPS_ENDPOINT` | ✅ | STAROps API endpoint, format: `starops.cn-beijing.aliyuncs.com` |
| `VIBEOPS_REGION` | ❌ | Region, default `cn-hangzhou` |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅ | Access Key ID |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅ | Access Key Secret |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | Employee name, default `default` |

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
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

### Java 8

```bash
cd samples/java8
cp .env.example .env  # edit with your credentials
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

### Python

```bash
cd samples/python
python -m venv .venv && source .venv/bin/activate
pip install -e .
cp .env.example .env  # edit with your credentials
python -m starops_sdk_samples.examples.chat
```

### TypeScript

```bash
cd samples/typescript
npm install
cp .env.example .env  # edit with your credentials
npx tsx src/examples/chat.ts
```

---

## Sample Programs

All languages implement the same set of examples:

| Sample | Description |
|--------|-------------|
| `chat` | Interactive multi-turn chat |
| `chat-from-file` | Load request from JSON file |
| `thread-manager` | List/get/delete conversation threads |

### chat-from-file

Supports loading request parameters from shared JSON files in `requests/starops/`:

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

## Project Structure

```
.
├── README.md
├── README_zh.md
├── .env.example
├── requests/                          # Shared request JSON files
│   └── starops/
│       ├── entity.json                # Entity query
│       ├── sls_chat.json              # SLS log query (sql_generation)
│       ├── sql_generation.json        # SQL generation
│       ├── general_chat.json          # General chat
│       └── data_agent.json            # Data agent
│
└── samples/
    ├── golang/                        # Go (1.21+)
    ├── java/                          # Java (11+)
    ├── java8/                         # Java (8+) - syntax-compatible fork
    ├── python/                        # Python (3.8+)
    └── typescript/                    # TypeScript (Node 18+)
```

---

## Java vs Java 8

The `java8` project is a syntax-only adaptation of the `java` project. Key differences:

| Feature | Java (11+) | Java 8 |
|---------|-----------|--------|
| `var` keyword | ✅ | ❌ → explicit types |
| `String.repeat()` | ✅ | ❌ → `repeatStr()` helper |
| `Map.putIfAbsent()` | ✅ | ✅ → `containsKey()` + `put()` |
| dotenv-java | 3.0.0 | 2.3.2 |
| mockito | 5.5.0 | 4.11.0 |
| Lambda / Runnable | Lambda | Anonymous `Runnable` |

---

## License

Apache License 2.0
