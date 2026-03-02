# CMS SDK Samples

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](samples/java/)
[![Java8](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](samples/java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](samples/typescript/)

[中文文档](README_zh.md)

Production-ready sample programs for interacting with Alibaba Cloud CMS (Cloud Monitor Service) digital employees across multiple programming languages.

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

- Alibaba Cloud account with CMS access
- Access Key ID and Secret

### Environment Variables

All language samples use the same environment variables (via `.env` file):

| Variable | Required | Description |
|----------|----------|-------------|
| `VIBEOPS_WORKSPACE` | ✅ | Workspace ID |
| `VIBEOPS_ENDPOINT` | ✅ | API endpoint |
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

## Sample Programs

All languages implement the same set of examples:

| Sample | Description |
|--------|-------------|
| `chat` | Interactive multi-turn chat |
| `chat-from-file` | Load request from JSON file |
| `thread-manager` | List/get/delete conversation threads |

### chat-from-file

Supports loading request parameters from shared JSON files in `requests/cms/`:

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

## Project Structure

```
.
├── README.md
├── README_zh.md
├── .env.example
├── requests/                          # Shared request JSON files
│   └── cms/
│       ├── entity.json                # Entity query
│       ├── sls_chat.json              # SLS log query (sql_generation)
│       ├── sql_generation.json        # SQL generation
│       ├── general_chat.json          # General chat
│       ├── metric_query.json          # Metric query
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
