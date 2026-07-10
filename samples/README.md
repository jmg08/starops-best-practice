# STAROps SDK Samples

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](java/)
[![Java%208](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](typescript/)

Production-ready multi-language samples for building SSE streaming chat clients with Alibaba Cloud **STAROps** digital employees.

This repository provides equivalent sample clients in **Go**, **Java**, **Java 8**, **Python**, and **TypeScript**.
Each implementation demonstrates the same core workflow: load credentials, create or reuse a conversation thread,
stream responses over SSE, handle interactive events, and recover from transient stream failures.

[中文文档](README_zh.md)

## Features

- **Multi-language coverage**: Go, Java, Java 8, Python, and TypeScript implementations.
- **SSE streaming**: Stream structured STAROps responses in real time.
- **Automatic reconnection**: Resume after dropped connections, SSE errors, or idle timeouts.
- **Exponential backoff**: Retry with bounded backoff to avoid aggressive reconnect loops.
- **Message deduplication**: Skip already received messages after reconnecting.
- **Credential configuration**: Prefer Alibaba Cloud CLI-based credential configuration; environment variables can be used when CLI installation is not desired.
- **Shared request fixtures**: Reuse the same JSON request files across languages.
- **Consistent examples**: Run interactive chat, file-based chat, and thread management in every language.

> [!IMPORTANT]
> We recommend configuring credentials with Alibaba Cloud CLI. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview). If you do not want to install it, configure AK/SK with environment variables.

## Repository layout

```text
.
├── README.md / README_zh.md
├── sample-requests/              # Shared STAROps request examples
├── golang/                        # Go 1.22+ sample client
├── java/                          # Java 11+ sample client
├── java8/                         # Java 8-compatible sample client
├── python/                        # Python 3.8+ sample client
└── typescript/                    # TypeScript / Node.js sample client
```

## Quick start

### 1. Choose a language

| Language | Directory | Runtime | Main commands |
| --- | --- | --- | --- |
| Go | [`golang/`](golang/) | Go 1.22+ | `go run ./cmd/chat` |
| Java | [`java/`](java/) | Java 11+ / Maven | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| Java 8 | [`java8/`](java8/) | Java 8+ / Maven | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| Python | [`python/`](python/) | Python 3.8+ | `python -m starops_sdk_samples.examples.chat` |
| TypeScript | [`typescript/`](typescript/) | Node.js 18+ | `npm run chat` |

### 2. Configure environment variables

Each language has its own `.env.example`. Copy it to `.env` before running a sample:

```bash
cp .env.example .env
```

Required `.env` values:

| Variable | Description |
| --- | --- |
| `STAROPS_ENDPOINT` | STAROps endpoint, for example `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | STAROps workspace ID |
| `STAROPS_EMPLOYEE_NAME` | Digital employee name |

Credential configuration:

1. **Recommended**: configure credentials with Alibaba Cloud CLI.
2. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview).
3. If you do not want to install the CLI, configure AK/SK with environment variables:

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

Optional retry settings:

| Variable | Default | Description |
| --- | --- | --- |
| `STAROPS_REGION` | Derived from endpoint when possible | Alibaba Cloud region |
| `STAROPS_MAX_RETRIES` | `10` | Maximum SSE reconnect attempts |
| `STAROPS_IDLE_TIMEOUT` | `60` | Seconds to wait before treating a stream as idle |

> [!TIP]
> `.env` is only for STAROps sample settings. Configure credentials with Alibaba Cloud CLI or AK/SK environment variables.

### 3. Run a sample

Run an interactive chat:

```bash
# Go
cd golang && go run ./cmd/chat

# Python
cd python && python -m starops_sdk_samples.examples.chat

# TypeScript
cd typescript && npm install && npm run chat
```

Run a file-based STAROps request:

```bash
# Go
cd golang && go run ./cmd/chat-from-file -file ../sample-requests/data_agent.json

# Java
cd java && mvn exec:java \
  -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
  -Dexec.args="-file ../sample-requests/data_agent.json"
```

## Shared request examples

The [`sample-requests/`](sample-requests/) directory contains request fixtures that can be reused by every language sample.

| File | Scenario |
| --- | --- |
| `general_chat.json` | General conversation |
| `entity.json` | Entity-based query |
| `metric_query.json` | Metric query |
| `sql_generation.json` | Natural language to SQL |
| `sls_chat.json` | SLS-focused chat |
| `data_agent.json` | Data-agent analysis |
| `user_ack_interactive.json` | User acknowledgement interaction |
| `user_input_interactive.json` | User input interaction |

## Example programs

Each language provides the same three entry points:

| Program | Purpose |
| --- | --- |
| `chat` | Start an interactive multi-turn conversation |
| `chat-from-file` | Load one JSON request or process a directory of requests |

## Retry and reconnection model

The samples implement the same stream recovery behavior across languages:

```text
send request
  └─ stream SSE events
       ├─ stream_done                  -> finish
       ├─ connection dropped           -> retry
       ├─ SSE error                    -> retry
       └─ idle timeout                 -> retry
              └─ reconnect with backoff
                    └─ deduplicate received messages
```

Key behaviors:

- A normal response finishes when a `stream_done` event is received.
- If a stream ends before `stream_done`, the client reconnects automatically.
- Retry delays use exponential backoff and are capped to avoid excessive retry pressure.
- Reconnected streams deduplicate messages that were already delivered.

You can exercise the retry path with `-simulate-error`:

```bash
cd golang
go run ./cmd/chat-from-file -file ../sample-requests/data_agent.json -simulate-error true
```

## VS Code debugging

The repository includes `.vscode/launch.json` with launch configurations for Go, Python, Java, Java 8, and TypeScript.

For Java and Java 8 debugging, compile and copy Maven dependencies first:

```bash
mvn -q -f java/pom.xml -DskipTests dependency:copy-dependencies
mvn -q -f java8/pom.xml -DskipTests dependency:copy-dependencies
```

The launch configurations use explicit classpaths:

```text
java/target/classes
java/target/dependency/*
java8/target/classes
java8/target/dependency/*
```

## Development checks

Run tests for a specific language:

```bash
# Go
cd golang && go test ./...

# Java
cd java && mvn test

# Java 8
cd java8 && mvn test

# Python
cd python && python -m pytest

# TypeScript
cd typescript && npm test
```

## Security notes

- Keep real AK/SK values out of Git.
- Do not commit `.env`, IDE workspace files, or local debug launch files containing credentials.
- Use `.env.example` only for placeholders and documentation.
- Prefer RAM roles, OIDC, or the Alibaba Cloud credential chain for non-local environments.
