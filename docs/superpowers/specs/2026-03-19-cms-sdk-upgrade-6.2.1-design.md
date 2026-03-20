# CMS SDK 6.0.1 → 6.2.1 升级需求文档

## 概述

将 vibeops-samples 项目中五种语言（Go、Java、Java8、Python、TypeScript）的阿里云 CMS SDK 依赖从 6.0.1 升级到 6.2.1，并利用新版本新增的字段优化现有 SSE 事件处理逻辑。

## 背景

### 版本线

6.0.0 → 6.0.1 → 6.1.0 → 6.2.0 → 6.2.1

### 6.2.1 相对于 6.0.1 的变更分析

**无破坏性变更**，现有 API 签名完全向后兼容。

#### 新增 API（本次不涉及）

- 记忆管理（Memory Store）：`addMemories`、`deleteMemories`、`deleteMemory`、`createMemoryStore`、`deleteMemoryStore`、`getMemories`、`getMemory`、`getMemoryHistory`、`getMemoryStore`、`listMemoryStores`、`searchMemories`、`updateMemory`、`updateMemoryStore`
- 技能管理（Digital Employee Skill）：`createDigitalEmployeeSkill`、`deleteDigitalEmployeeSkill`、`getDigitalEmployeeSkill`、`listDigitalEmployeeSkills`、`listDigitalEmployeeSkillVersions`、`updateDigitalEmployeeSkill`

#### 现有模型扩展（本次优化重点）

1. `CreateChatResponse` 新增可选字段：
   - `id`：事件唯一标识符
   - `event`：事件类型标识
2. `CreateThreadRequest` 新增可选字段：
   - `attributes`：自定义属性字典（`map[string]string`）

## 需求范围

### 第一部分：依赖版本升级

修改五种语言的依赖声明文件，将 CMS SDK 版本从 6.0.1 升级到 6.2.1。

| 语言 | 文件 | 变更 |
|------|------|------|
| Go | `samples/golang/go.mod` | `cms-20240330/v6 v6.0.1` → `v6.2.1` |
| Java | `samples/java/pom.xml` | `<version>6.0.1</version>` → `<version>6.2.1</version>` |
| Java8 | `samples/java8/pom.xml` | `<version>6.0.1</version>` → `<version>6.2.1</version>` |
| Python | `samples/python/pyproject.toml` | `>=6.0.1` → `>=6.2.1` |
| TypeScript | `samples/typescript/package.json` | `^6.0.1` → `^6.2.1` |

升级后需执行各语言的依赖更新命令并确保构建通过：
- Go：`go mod tidy`（同时更新 `go.sum`）
- Java/Java8：`mvn compile`
- Python：`pip install -e .`
- TypeScript：`npm install`（同时更新 `package-lock.json`）

注意：Go 的 `go.sum` 和 TypeScript 的 `package-lock.json` 是自动生成的锁文件，需一并提交。

### 第二部分：ChatEvent 新增 id/event 字段

在各语言的 `ChatEvent`（或等价数据结构）中新增 `id` 和 `event` 字段，从 `CreateChatResponse` 中提取。

#### 当前状态

各语言的 `ChatEvent` 结构：

| 语言 | 文件 | 当前字段 |
|------|------|---------|
| Go | `internal/client/types.go` 或 `client.go` | `Body`、`RawJSON`、`StatusCode`、`IsDone`、`Error` |
| TypeScript | `src/client/agent-client.ts` | `body`、`rawJson`、`statusCode`、`isDone`、`error` |
| Python | `cms_sdk_samples/client/agent_client.py` | `body`、`raw_json`、`status_code`、`is_done`、`error` |
| Java | `client/ChatEvent.java` | `body`（JsonNode）、`rawJson`、`statusCode`、`isDone`、`error` |
| Java8 | `client/ChatEvent.java` | `body`（JsonNode）、`rawJson`、`statusCode`、`isDone`、`error` |

#### 目标状态

每种语言的 `ChatEvent` 新增：
- `id`（string，可选）：从 response 的 `id` 字段提取
- `event`（string，可选）：从 response 的 `event` 字段提取

各语言的提取方式：
- Go：从 `CreateChatResponseBody` 的 `Id *string` 和 `Event *string` 字段提取（Go SDK 使用指针类型，需 nil 检查后解引用）
- TypeScript：从 response body 对象的 `id` 和 `event` 属性提取
- Python：从 response body 字典的 `body.get("id")` 和 `body.get("event")` 提取
- Java/Java8：从 `JsonNode` 的 `jsonNode.path("id").asText(null)` 和 `jsonNode.path("event").asText(null)` 提取（Java 使用 `callApi` 返回原始 JSON，不经过 `CreateChatResponse` 对象）。`ChatEvent.fromResponse(JsonNode, ...)` 方法需同步修改以提取这两个字段

### 第三部分：优化事件类型判断逻辑

#### 当前状态

所有语言通过遍历 `body.messages[]` 数组检查 `type` 字段来判断事件类型：

```
// 伪代码 - 当前逻辑
function isDone(body):
    for msg in body.messages:
        if msg.type == "done": return true
    return false
```

#### 目标状态

优先使用 response 级别的 `event` 字段判断事件类型，保留 `body.messages[].type` 作为 fallback。

各语言的实现要点：

**Go**：`isDoneMessage(body *cms.CreateChatResponseBody)` 不需要改签名，因为升级后 `CreateChatResponseBody` 自动包含 `Event *string` 字段。在函数内部增加优先判断（注意 Go SDK 使用指针类型）：
```go
if body.Event != nil && *body.Event == "done" {
    return true
}
// fallback: 遍历 body.Messages
```

**TypeScript / Python**：`isDoneMessage` / `_is_done_message` 不需要改签名，直接从 `body` 对象中读取 `event` 字段即可（`body.event` / `body.get("event")`）。

**Java / Java8**：`isDoneMessage` 逻辑封装在 `ChatEvent.java` 的 `isDoneMessage(JsonNode body)` 静态方法中（不在 `AgentClient.java` 中），从 `body.path("event").asText(null)` 读取。

#### 涉及文件

| 语言 | 需修改的文件 | 修改内容 |
|------|------------|---------|
| Go | `internal/client/client.go` | `isDoneMessage()` 函数增加 `body.Event` 指针字段优先判断 |
| TypeScript | `src/client/agent-client.ts` | `isDoneMessage()` 函数增加 `body.event` 字段优先判断 |
| Python | `cms_sdk_samples/client/agent_client.py` | `_is_done_message()` 静态方法增加 `body.get("event")` 优先判断，不改签名 |
| Java | `client/ChatEvent.java` | `isDoneMessage()` 增加 `body.path("event")` 优先判断 |
| Java8 | `client/ChatEvent.java` | `isDoneMessage()` 增加 `body.path("event")` 优先判断 |

#### SimplePrinter 和 InteractiveHandler 优化

在事件分发时，利用 `ChatEvent.event` 字段进行快速类型判断：
- `SimplePrinter`：用 `event` 字段快速跳过非文本事件
- `InteractiveHandler`：用 `event` 字段快速识别交互事件

保留现有的 `body.messages[].type` 解析作为 fallback，确保向后兼容。

涉及文件（五种语言各自的 SimplePrinter 和 InteractiveHandler）：

| 语言 | SimplePrinter | InteractiveHandler |
|------|--------------|-------------------|
| Go | `internal/client/simple_printer.go` | `internal/client/interactive.go` |
| TypeScript | `src/client/simple-printer.ts` | `src/client/interactive-handler.ts` |
| Python | `cms_sdk_samples/client/simple_printer.py` | `cms_sdk_samples/client/interactive_handler.py` |
| Java | `client/SimplePrinter.java` | `client/InteractiveHandler.java` |
| Java8 | `client/SimplePrinter.java` | `client/InteractiveHandler.java` |

### 第四部分：createThread 支持 attributes 参数

#### 当前状态

各语言的 `AgentClient.createThread()` 方法只接受 `workspace` 参数（部分语言还有 `title`），不支持传入自定义属性。

#### 目标状态

在 `createThread()` 方法中新增可选的 `attributes` 参数（`map[string]string` 类型），传递给 `CreateThreadRequest.attributes`。

| 语言 | 文件 | 修改内容 |
|------|------|---------|
| Go | `internal/client/thread.go` | `CreateThread()` 新增 `attributes` 可选参数 |
| TypeScript | `src/client/agent-client.ts` | `createThread()` 新增 `attributes` 可选参数 |
| Python | `cms_sdk_samples/client/agent_client.py` | `create_thread()` 新增 `attributes` 可选参数 |
| Java | `client/AgentClient.java` | `createThread()` 新增 `attributes` 可选参数 |
| Java8 | `client/AgentClient.java` | `createThread()` 新增 `attributes` 可选参数 |

不修改现有 examples 中的调用方式，保持默认不传 attributes 的行为不变。

## 验收标准

1. 五种语言的 CMS SDK 依赖版本均为 6.2.1
2. 各语言构建成功（`make build`、`mvn compile`、`npm run build`、`pip install -e .`）
3. 各语言测试通过（`make test`、`mvn test`、`npm run test`、`pytest`）
4. `ChatEvent` 包含 `id` 和 `event` 字段，字段为空时安全处理（Go 的 nil 指针、Java 的 null、Python 的 None、TypeScript 的 undefined）
5. `isDoneMessage` 优先使用 `event` 字段判断，fallback 到 `body.messages[].type`；两条路径均需测试覆盖
6. `createThread` 支持可选的 `attributes` 参数，不传时行为与升级前一致
7. 现有 examples 行为不变，无回归

## 不在范围内

- 不新增记忆管理或技能管理的示例代码
- 不修改 examples 层的用户接口
- 不升级其他依赖（tea、openapi-client 等）
