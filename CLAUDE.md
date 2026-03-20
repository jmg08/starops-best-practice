# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

阿里云 CMS（云监控）数字员工 SDK 多语言示例项目，包含 Go、Java、Java8、Python、TypeScript 五种语言的一致实现。所有语言共享相同的架构模式和 `requests/cms/` 下的请求模板文件。

## 构建、测试、运行命令

### Go（`samples/golang/`）

```bash
cd samples/golang
make build          # 构建所有示例
make test           # 运行所有测试
make test-verbose   # 详细模式测试
make lint           # 代码检查（gofmt + go vet）
make fmt            # 格式化代码
make run            # 运行默认 chat 示例
make build-chat     # 构建单个示例
# 调试运行（禁止使用 go run）：
go build -gcflags="all=-N -l" -o bin/chat ./cmd/chat && dlv exec bin/chat
```

### Java（`samples/java/`）/ Java8（`samples/java8/`）

```bash
cd samples/java     # 或 samples/java8
mvn compile         # 编译
mvn test            # 运行测试
mvn test -Dtest=ConfigTest                    # 运行单个测试类
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.Chat"  # 运行示例
```

### Python（`samples/python/`）

```bash
cd samples/python
pip install -e ".[dev]"   # 安装（含开发依赖）
pytest                    # 运行所有测试
pytest tests/test_config.py          # 运行单个测试文件
pytest tests/test_config.py -k test_name  # 运行单个测试
python -m cms_sdk_samples.examples.chat   # 运行示例
```

### TypeScript（`samples/typescript/`）

```bash
cd samples/typescript
npm install         # 安装依赖
npm run build       # 编译（tsc）
npm run test        # 运行测试（vitest --run）
npm run chat        # 运行 chat 示例（tsx）
```

## 架构

所有语言实现遵循统一的分层结构：

- **client/**：核心 SDK 封装层
  - `AgentClient` — 封装阿里云 CMS SDK，提供 `createThread()`、`chat()`、`listThreads()`、`getThread()`、`deleteThread()`
  - `Config` — 从 `.env` 加载环境变量并校验
  - `SimplePrinter` — SSE 事件流处理与输出格式化
  - `InteractiveHandler` — 处理交互事件（确认、选择、输入）
  - `SDKException` / `errors` — 带错误码的自定义异常
- **types/**：事件类型、消息角色、内容类型等枚举和数据结构
- **examples/**（或 Go 的 `cmd/`）：可直接运行的示例程序
  - `chat` — 交互式多轮对话
  - `chat-from-file` — 从 `requests/cms/` 加载 JSON 请求批量处理
  - `thread-manager` — 会话管理（列表/查询/删除）
- **logger/**：结构化日志工具

## 环境变量

复制 `.env.example` 为 `.env`，必填项：
- `VIBEOPS_WORKSPACE` — 工作空间 ID
- `VIBEOPS_ENDPOINT` — API 端点地址
- `ALIBABA_CLOUD_ACCESS_KEY_ID` — 阿里云 AK
- `ALIBABA_CLOUD_ACCESS_KEY_SECRET` — 阿里云 SK

可选项：`VIBEOPS_REGION`（默认 cn-hangzhou）、`VIBEOPS_EMPLOYEE_NAME`（默认 default）、`LOG_LEVEL`

## 开发约定

- 文档和注释使用中文，代码标识符使用英文
- 五种语言的功能和架构保持一致，修改一种语言时考虑其他语言是否需要同步
- Go 调试运行使用 `dlv`，禁止 `go run`（`go test`、`go vet` 等不受此限制）
- 示例程序应可直接运行，配置应简洁清晰
- 不同 cmd/examples 文件之间避免功能重复
- 中间日志输出到文件（`output/` 目录），便于查看和对比
- 不要使用 `../../` 相对路径引用文件

## 核心依赖

所有语言统一使用阿里云 CMS SDK `cms20240330` v6.2.1。

## 请求模板

`requests/cms/` 下包含共享的 JSON 请求模板，用于不同场景测试：
- `general_chat.json` — 通用查询
- `entity.json` — APM 实体查询（`userContext.type: entity`）
- `metric_query.json` — 云产品指标查询（`userContext.type: metric_query`）
- `sql_generation.json` — 自然语言转 SQL（`skill: sql_generation`）
- `data_agent.json` — 数据智能体（`skill: data-agent-pro`）
- `sls_chat.json` — SLS 日志查询
