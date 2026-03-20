# CMS SDK 6.0.1 → 6.2.1 升级实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将五种语言的 CMS SDK 从 6.0.1 升级到 6.2.1，并利用新增的 `id`/`event` 字段优化 SSE 事件处理逻辑，同时为 `createThread` 暴露 `attributes` 参数。

**Architecture:** 纯依赖升级 + 现有代码优化。6.2.1 完全向后兼容 6.0.1，无破坏性变更。优化点集中在 `ChatEvent` 数据结构和 `isDoneMessage` 判断逻辑。

**Tech Stack:** Go 1.22+, Java 11 (Maven), Java 8 (Maven), Python 3.8+ (setuptools), TypeScript 5.0+ (npm)

**Spec:** `docs/superpowers/specs/2026-03-19-cms-sdk-upgrade-6.2.1-design.md`

---

### Task 1: TypeScript — 依赖升级 + ChatEvent + isDone + createThread

**Files:**
- Modify: `samples/typescript/package.json:29`
- Modify: `samples/typescript/src/client/agent-client.ts:15-22` (ChatEvent interface)
- Modify: `samples/typescript/src/client/agent-client.ts:40-44` (isDoneMessage)
- Modify: `samples/typescript/src/client/agent-client.ts:72-96` (createThread)
- Modify: `samples/typescript/src/client/simple-printer.ts` (event 字段快速跳过)
- Modify: `samples/typescript/src/client/interactive-handler.ts` (event 字段前置过滤)
- Test: `samples/typescript/tests/chat-event.test.ts` (新建)

- [ ] **Step 1: 升级依赖版本**

修改 `samples/typescript/package.json`:
```json
"@alicloud/cms20240330": "^6.2.1",
```

运行:
```bash
cd samples/typescript && npm install
```

- [ ] **Step 2: 编写 isDoneMessage 的测试**

`isDoneMessage` 是模块私有函数，需要先将其导出以便测试。在 `agent-client.ts` 中将 `function isDoneMessage` 改为 `export function isDoneMessage`。

创建 `samples/typescript/tests/chat-event.test.ts`:
```typescript
import { describe, it, expect } from 'vitest';
import { isDoneMessage } from '../src/client/agent-client.js';

describe('isDoneMessage', () => {
  it('event 字段为 done 时返回 true', () => {
    const body = { event: 'done', messages: [] };
    expect(isDoneMessage(body)).toBe(true);
  });

  it('无 event 字段时 fallback 到 messages[].type', () => {
    const body = { messages: [{ type: 'done' }] };
    expect(isDoneMessage(body)).toBe(true);
  });

  it('非 done 事件返回 false', () => {
    const body = { event: 'text', messages: [{ type: 'text' }] };
    expect(isDoneMessage(body)).toBe(false);
  });

  it('body 为 undefined 时返回 false', () => {
    expect(isDoneMessage(undefined)).toBe(false);
  });
});
```

运行验证失败:
```bash
cd samples/typescript && npm run test
```
Expected: FAIL（isDoneMessage 尚未支持 event 字段优先判断）

- [ ] **Step 3: ChatEvent 新增 id 和 event 字段**

修改 `samples/typescript/src/client/agent-client.ts` 的 `ChatEvent` interface:
```typescript
/** 聊天事件 / Chat event */
export interface ChatEvent {
  id?: string;
  event?: string;
  body?: Record<string, unknown>;
  rawJson: string;
  statusCode: number;
  isDone: boolean;
  error?: Error;
}
```

- [ ] **Step 4: 优化 isDoneMessage 函数**

修改 `samples/typescript/src/client/agent-client.ts` 的 `isDoneMessage`:
```typescript
function isDoneMessage(body?: Record<string, unknown>): boolean {
  // 优先使用 response 级别的 event 字段
  if (body?.event === 'done') return true;
  // fallback: 遍历 messages
  if (!body || !body.messages) return false;
  const messages = body.messages as Array<Record<string, unknown>>;
  return messages.some((msg) => msg.type === 'done');
}
```

- [ ] **Step 5: 构造 ChatEvent 时提取 id/event**

在 `agent-client.ts` 中 `chat()` 方法的 SSE 事件构造处，确保 `id` 和 `event` 从 body 中提取并赋值到 ChatEvent。找到构造 ChatEvent 对象的位置，添加:
```typescript
id: (body as Record<string, unknown>)?.id as string | undefined,
event: (body as Record<string, unknown>)?.event as string | undefined,
```

- [ ] **Step 6: createThread 支持 attributes**

修改 `samples/typescript/src/client/agent-client.ts` 的 `createThread`:
```typescript
/** 创建会话 / Create thread */
async createThread(attributes?: Record<string, string>): Promise<string> {
  try {
    const variables = new $CMS20240330.CreateThreadRequestVariables({
      workspace: this.config.workspace,
    });
    const request = new $CMS20240330.CreateThreadRequest({
      title: `Chat-${Math.floor(Date.now() / 1000)}`,
      variables,
      attributes,
    });
    // ... 其余不变
```

- [ ] **Step 7: SimplePrinter 利用 event 字段快速跳过**

修改 `samples/typescript/src/client/simple-printer.ts` 的 `processEvent` 方法，在方法开头增加快速跳过逻辑：
```typescript
processEvent(event: ChatEvent | null | undefined): string {
  if (!event?.body) return '';
  // 利用 event 字段快速跳过非文本事件
  if (event.event && event.event !== 'text' && event.event !== 'task_finished') return '';
  // ... 其余逻辑不变
```

- [ ] **Step 8: InteractiveHandler 利用 event 字段快速识别**

修改 `samples/typescript/src/client/interactive-handler.ts` 的 `handleEvent` 方法，在类型判断处增加 `ChatEvent.event` 的快速路径：
```typescript
// 如果 ChatEvent 携带 event 字段，可用于快速判断是否为交互事件
// 保留现有的 event.type 判断作为 fallback
```

注意：InteractiveHandler 接收的是 `Record<string, unknown>`（从 body.messages 中提取的单条消息），不是 ChatEvent。此处优化仅在调用 `handleEvent` 之前的分发逻辑中利用 `ChatEvent.event` 做前置过滤，减少不必要的 `handleEvent` 调用。

- [ ] **Step 9: 构建和测试**

```bash
cd samples/typescript && npm run build && npm run test
```
Expected: 构建成功，所有测试通过

- [ ] **Step 10: 提交**

```bash
cd samples/typescript
git add package.json package-lock.json src/client/ tests/
git commit -m "feat(typescript): upgrade CMS SDK to 6.2.1, add id/event fields, support attributes"
```

---

### Task 2: Python — 依赖升级 + ChatEvent + isDone + createThread

**Files:**
- Modify: `samples/python/pyproject.toml`
- Modify: `samples/python/cms_sdk_samples/client/agent_client.py` (ChatEvent + _is_done_message + create_thread)
- Modify: `samples/python/cms_sdk_samples/client/simple_printer.py` (event 字段快速跳过)
- Modify: `samples/python/cms_sdk_samples/client/interactive_handler.py` (event 字段前置过滤)
- Test: `samples/python/tests/test_chat_event.py` (新建)

- [ ] **Step 1: 升级依赖版本**

修改 `samples/python/pyproject.toml`:
```toml
"alibabacloud-cms20240330>=6.2.1",
```

运行:
```bash
cd samples/python && pip install -e ".[dev]"
```

- [ ] **Step 2: 编写测试**

创建 `samples/python/tests/test_chat_event.py`:
```python
"""ChatEvent 测试"""
from cms_sdk_samples.client.agent_client import ChatEvent


class TestChatEvent:
    def test_from_response_with_event_field_done(self):
        """event 字段为 done 时 is_done 应为 True"""
        body = {"event": "done", "messages": []}
        event = ChatEvent.from_response(body, '{"event":"done"}', 200)
        assert event.is_done is True

    def test_from_response_with_messages_done_fallback(self):
        """无 event 字段时 fallback 到 messages[].type"""
        body = {"messages": [{"type": "done"}]}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.is_done is True

    def test_from_response_not_done(self):
        """非 done 事件"""
        body = {"event": "text", "messages": [{"type": "text"}]}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.is_done is False

    def test_from_response_extracts_id_and_event(self):
        """提取 id 和 event 字段"""
        body = {"id": "evt-123", "event": "text", "messages": []}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.id == "evt-123"
        assert event.event == "text"

    def test_from_response_missing_id_event(self):
        """id/event 缺失时为 None"""
        body = {"messages": []}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.id is None
        assert event.event is None
```

运行验证失败:
```bash
cd samples/python && pytest tests/test_chat_event.py -v
```
Expected: FAIL（ChatEvent 尚无 id/event 字段）

- [ ] **Step 3: ChatEvent 新增 id 和 event 字段**

修改 `samples/python/cms_sdk_samples/client/agent_client.py`:
```python
@dataclass
class ChatEvent:
    """聊天事件 / Chat event"""
    body: Optional[Dict[str, Any]] = None
    raw_json: str = ""
    status_code: int = 0
    is_done: bool = False
    error: Optional[Exception] = None
    id: Optional[str] = None
    event: Optional[str] = None

    @classmethod
    def done(cls) -> "ChatEvent":
        return cls(is_done=True)

    @classmethod
    def from_error(cls, error: Exception) -> "ChatEvent":
        return cls(error=error)

    @classmethod
    def from_response(cls, body: Dict[str, Any], raw_json: str, status_code: int) -> "ChatEvent":
        is_done = cls._is_done_message(body)
        return cls(
            body=body,
            raw_json=raw_json,
            status_code=status_code,
            is_done=is_done,
            id=body.get("id") if body else None,
            event=body.get("event") if body else None,
        )
```

- [ ] **Step 4: 优化 _is_done_message**

```python
@staticmethod
def _is_done_message(body: Optional[Dict[str, Any]]) -> bool:
    if not body:
        return False
    # 优先使用 response 级别的 event 字段
    if body.get("event") == "done":
        return True
    # fallback: 遍历 messages
    messages = body.get("messages", [])
    for msg in messages:
        if isinstance(msg, dict) and msg.get("type") == "done":
            return True
    return False
```

- [ ] **Step 5: create_thread 支持 attributes**

修改 `samples/python/cms_sdk_samples/client/agent_client.py`:
```python
def create_thread(self, attributes: Optional[Dict[str, str]] = None) -> str:
    """创建会话 / Create thread"""
    try:
        variables = cms_models.CreateThreadRequestVariables(
            workspace=self.config.workspace
        )
        request = cms_models.CreateThreadRequest(
            title=f"Chat-{int(time.time())}",
            variables=variables,
            attributes=attributes,
        )
        # ... 其余不变
```

- [ ] **Step 6: SimplePrinter 利用 event 字段快速跳过**

修改 `samples/python/cms_sdk_samples/client/simple_printer.py` 的 `process_event` 方法，在方法开头增加快速跳过逻辑：
```python
def process_event(self, event: Optional[Any]) -> str:
    """处理事件，提取文本内容 / Process event and extract text content"""
    if not event:
        return ""

    # 利用 event 字段快速跳过非文本事件
    event_type = getattr(event, 'event', None)
    if event_type and event_type not in ('text', 'task_finished'):
        return ""

    body = getattr(event, 'body', None)
    if not body:
        return ""
    # ... 其余逻辑不变
```

- [ ] **Step 7: InteractiveHandler 利用 event 字段前置过滤**

在调用 `InteractiveHandler.handle_event()` 之前的分发逻辑中，利用 `ChatEvent.event` 字段做前置过滤。InteractiveHandler 本身接收的是从 `body.messages` 中提取的单条消息字典，此处优化在调用方减少不必要的 `handle_event` 调用。

- [ ] **Step 8: 运行测试**

```bash
cd samples/python && pytest tests/ -v
```
Expected: 所有测试通过

- [ ] **Step 9: 提交**

```bash
cd samples/python
git add pyproject.toml cms_sdk_samples/client/ tests/test_chat_event.py
git commit -m "feat(python): upgrade CMS SDK to 6.2.1, add id/event fields, support attributes"
```

---

### Task 3: Go — 依赖升级 + ChatEvent + isDone + CreateThread

**Files:**
- Modify: `samples/golang/go.mod:6`
- Modify: `samples/golang/internal/client/client.go:126-133` (ChatEvent struct)
- Modify: `samples/golang/internal/client/client.go:363-373` (isDoneMessage)
- Modify: `samples/golang/internal/client/client.go:105-124` (CreateThread)
- Modify: `samples/golang/internal/client/simple_printer.go` (Event 字段快速跳过)
- Modify: `samples/golang/internal/client/interactive.go` (Event 字段前置过滤)
- Test: `samples/golang/internal/client/chat_event_test.go` (新建)

- [ ] **Step 1: 升级依赖版本**

修改 `samples/golang/go.mod`:
```
github.com/alibabacloud-go/cms-20240330/v6 v6.2.1
```

运行:
```bash
cd samples/golang && go mod tidy
```

- [ ] **Step 2: 编写测试**

创建或追加到 `samples/golang/internal/client/chat_event_test.go`:
```go
package client

import (
	"testing"

	cms "github.com/alibabacloud-go/cms-20240330/v6/client"
	"github.com/alibabacloud-go/tea/dara"
)

func TestIsDoneMessage_EventField(t *testing.T) {
	body := &cms.CreateChatResponseBody{}
	body.SetEvent("done")
	if !isDoneMessage(body) {
		t.Error("event=done 时应返回 true")
	}
}

func TestIsDoneMessage_MessagesFallback(t *testing.T) {
	msg := &cms.CreateChatResponseBodyMessages{}
	msg.SetType("done")
	body := &cms.CreateChatResponseBody{
		Messages: []*cms.CreateChatResponseBodyMessages{msg},
	}
	if !isDoneMessage(body) {
		t.Error("messages[].type=done 时应返回 true")
	}
}

func TestIsDoneMessage_NotDone(t *testing.T) {
	body := &cms.CreateChatResponseBody{}
	body.SetEvent("text")
	if isDoneMessage(body) {
		t.Error("event=text 时应返回 false")
	}
}

func TestChatEvent_IdAndEvent(t *testing.T) {
	body := &cms.CreateChatResponseBody{}
	body.SetEvent("text")
	// 验证 ChatEvent 结构体包含 Id 和 Event 字段
	event := &ChatEvent{
		Body:   body,
		Id:     dara.StringValue(body.Id),
		Event:  dara.StringValue(body.Event),
	}
	if event.Event != "text" {
		t.Errorf("期望 event=text, 实际=%s", event.Event)
	}
}
```

运行验证失败:
```bash
cd samples/golang && go test ./internal/client/ -run TestIsDoneMessage -v
```
Expected: FAIL（ChatEvent 尚无 Id/Event 字段）

- [ ] **Step 3: ChatEvent 新增 Id 和 Event 字段**

修改 `samples/golang/internal/client/client.go`:
```go
// ChatEvent 聊天事件
type ChatEvent struct {
	Body       *cms.CreateChatResponseBody
	RawJSON    string
	StatusCode int32
	IsDone     bool
	Error      error
	Id         string
	Event      string
}
```

- [ ] **Step 4: 优化 isDoneMessage（注意指针类型）**

修改 `samples/golang/internal/client/client.go`:
```go
func isDoneMessage(body *cms.CreateChatResponseBody) bool {
	if body == nil {
		return false
	}
	// 优先使用 response 级别的 Event 字段
	if body.Event != nil && *body.Event == "done" {
		return true
	}
	// fallback: 遍历 Messages
	for _, msg := range body.Messages {
		if msg.Type != nil && *msg.Type == "done" {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: 构造 ChatEvent 时提取 Id/Event**

在 `client.go` 中构造 `ChatEvent` 的位置，添加 Id 和 Event 的提取:
```go
event := &ChatEvent{
	Body:       resp.Body,
	RawJSON:    rawJSON,
	StatusCode: statusCode,
	IsDone:     isDoneMessage(resp.Body),
	Id:         dara.StringValue(resp.Body.Id),
	Event:      dara.StringValue(resp.Body.Event),
}
```

- [ ] **Step 6: CreateThread 支持 attributes**

注意：`CreateThread` 实际定义在 `client.go`（非需求文档中标注的 `thread.go`），以下修改针对 `client.go`。

修改 `samples/golang/internal/client/client.go`:
```go
// CreateThread 创建会话
func (c *AgentClient) CreateThread(ctx context.Context, attributes ...map[string]string) (string, error) {
	req := &cms.CreateThreadRequest{}
	req.SetTitle(fmt.Sprintf("Chat-%d", time.Now().Unix()))

	variables := &cms.CreateThreadRequestVariables{}
	variables.SetWorkspace(c.config.Workspace)
	req.SetVariables(variables)

	if len(attributes) > 0 && attributes[0] != nil {
		req.SetAttributes(attributes[0])
	}

	resp, err := c.client.CreateThread(dara.String(c.config.EmployeeName), req)
	// ... 其余不变
```

注意：使用 variadic 参数保持向后兼容，现有调用 `CreateThread(ctx)` 无需修改。

- [ ] **Step 7: SimplePrinter 利用 Event 字段快速跳过**

修改 `samples/golang/internal/client/simple_printer.go` 的 `ProcessEvent` 方法，在方法开头增加快速跳过逻辑：
```go
func (p *SimplePrinter) ProcessEvent(event *ChatEvent) string {
	if event == nil || event.Body == nil {
		return ""
	}
	// 利用 Event 字段快速跳过非文本事件
	if event.Event != "" && event.Event != "text" && event.Event != "task_finished" {
		return ""
	}
	// ... 其余逻辑不变
```

- [ ] **Step 8: InteractiveHandler 利用 Event 字段前置过滤**

在调用 `InteractiveHandler.HandleEvent()` 之前的分发逻辑中，利用 `ChatEvent.Event` 字段做前置过滤。InteractiveHandler 本身接收的是 `*types.ItemEvent`（从 body.Messages 中提取的单条消息），此处优化在调用方减少不必要的 `HandleEvent` 调用。

- [ ] **Step 9: 构建和测试**

```bash
cd samples/golang && make lint && make test
```
Expected: 构建成功，所有测试通过

- [ ] **Step 10: 提交**

```bash
cd samples/golang
git add go.mod go.sum internal/client/
git commit -m "feat(golang): upgrade CMS SDK to 6.2.1, add Id/Event fields, support attributes"
```

---

### Task 4: Java — 依赖升级 + ChatEvent + isDone + createThread

**Files:**
- Modify: `samples/java/pom.xml:29`
- Modify: `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/ChatEvent.java`
- Modify: `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/AgentClient.java:66-87` (createThread)
- Modify: `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/SimplePrinter.java` (event 字段快速跳过)
- Modify: `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/InteractiveHandler.java` (event 字段前置过滤)
- Test: `samples/java/src/test/java/com/alibaba/cloud/cms/samples/client/ChatEventTest.java` (新建)

- [ ] **Step 1: 升级依赖版本**

修改 `samples/java/pom.xml`:
```xml
<version>6.2.1</version>
```

运行:
```bash
cd samples/java && mvn compile
```

- [ ] **Step 2: 编写测试**

创建 `samples/java/src/test/java/com/alibaba/cloud/cms/samples/client/ChatEventTest.java`:
```java
package com.alibaba.cloud.cms.samples.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.JsonNode;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class ChatEventTest {
    private final ObjectMapper mapper = new ObjectMapper();

    @Test
    void fromResponse_eventFieldDone_isDoneTrue() throws Exception {
        JsonNode body = mapper.readTree("{\"event\":\"done\",\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertTrue(event.isDone());
    }

    @Test
    void fromResponse_messagesFallback_isDoneTrue() throws Exception {
        JsonNode body = mapper.readTree("{\"messages\":[{\"type\":\"done\"}]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertTrue(event.isDone());
    }

    @Test
    void fromResponse_notDone() throws Exception {
        JsonNode body = mapper.readTree("{\"event\":\"text\",\"messages\":[{\"type\":\"text\"}]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertFalse(event.isDone());
    }

    @Test
    void fromResponse_extractsIdAndEvent() throws Exception {
        JsonNode body = mapper.readTree("{\"id\":\"evt-123\",\"event\":\"text\",\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertEquals("evt-123", event.getId());
        assertEquals("text", event.getEvent());
    }

    @Test
    void fromResponse_missingIdEvent_returnsNull() throws Exception {
        JsonNode body = mapper.readTree("{\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertNull(event.getId());
        assertNull(event.getEvent());
    }
}
```

运行验证失败:
```bash
cd samples/java && mvn test -Dtest=ChatEventTest
```
Expected: FAIL（ChatEvent 尚无 id/event 字段和 getter）

- [ ] **Step 3: ChatEvent 新增 id/event 字段、getter/setter、优化 isDoneMessage**

修改 `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/ChatEvent.java`:

新增字段:
```java
private String id;
private String event;
```

修改 `fromResponse`:
```java
public static ChatEvent fromResponse(JsonNode body, String rawJson, int statusCode) {
    ChatEvent event = new ChatEvent();
    event.body = body;
    event.rawJson = rawJson;
    event.statusCode = statusCode;
    event.id = body != null ? body.path("id").asText(null) : null;
    event.event = body != null ? body.path("event").asText(null) : null;
    event.done = isDoneMessage(body);
    return event;
}
```

修改 `isDoneMessage`:
```java
private static boolean isDoneMessage(JsonNode body) {
    if (body == null) {
        return false;
    }
    // 优先使用 response 级别的 event 字段
    JsonNode eventNode = body.get("event");
    if (eventNode != null && "done".equals(eventNode.asText())) {
        return true;
    }
    // fallback: 遍历 messages
    if (!body.has("messages")) {
        return false;
    }
    JsonNode messages = body.get("messages");
    if (messages.isArray()) {
        for (JsonNode msg : messages) {
            if (msg.has("type") && "done".equals(msg.get("type").asText())) {
                return true;
            }
        }
    }
    return false;
}
```

新增 getter/setter:
```java
public String getId() { return id; }
public void setId(String id) { this.id = id; }
public String getEvent() { return event; }
public void setEvent(String event) { this.event = event; }
```

- [ ] **Step 4: createThread 支持 attributes**

修改 `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/AgentClient.java`:
```java
public String createThread() throws SDKException {
    return createThread(null);
}

public String createThread(Map<String, String> attributes) throws SDKException {
    try {
        CreateThreadRequest request = new CreateThreadRequest();
        request.setTitle("Chat-" + Instant.now().getEpochSecond());

        CreateThreadRequest.CreateThreadRequestVariables variables = new CreateThreadRequest.CreateThreadRequestVariables();
        variables.setWorkspace(config.getWorkspace());
        request.setVariables(variables);

        if (attributes != null && !attributes.isEmpty()) {
            request.setAttributes(attributes);
        }

        CreateThreadResponse response = client.createThread(config.getEmployeeName(), request);
        // ... 其余不变
```

- [ ] **Step 5: SimplePrinter 利用 event 字段快速跳过**

修改 `samples/java/src/main/java/com/alibaba/cloud/cms/samples/client/SimplePrinter.java` 的 `processEvent` 方法，在方法开头增加快速跳过逻辑：
```java
public String processEvent(ChatEvent event) {
    if (event == null || event.getBody() == null) {
        return "";
    }
    // 利用 event 字段快速跳过非文本事件
    String eventType = event.getEvent();
    if (eventType != null && !"text".equals(eventType) && !"task_finished".equals(eventType)) {
        return "";
    }
    // ... 其余逻辑不变
```

- [ ] **Step 6: InteractiveHandler 利用 event 字段前置过滤**

在调用 `InteractiveHandler.handleEvent()` 之前的分发逻辑中，利用 `ChatEvent.getEvent()` 做前置过滤，减少不必要的 `handleEvent` 调用。

- [ ] **Step 7: 运行测试**

```bash
cd samples/java && mvn test
```
Expected: 所有测试通过

- [ ] **Step 8: 提交**

```bash
cd samples/java
git add pom.xml src/
git commit -m "feat(java): upgrade CMS SDK to 6.2.1, add id/event fields, support attributes"
```

---

### Task 5: Java8 — 依赖升级 + ChatEvent + isDone + createThread

**Files:**
- Modify: `samples/java8/pom.xml:29`
- Modify: `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/ChatEvent.java`
- Modify: `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/AgentClient.java:69-90` (createThread)
- Modify: `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/SimplePrinter.java` (event 字段快速跳过)
- Modify: `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/InteractiveHandler.java` (event 字段前置过滤)
- Test: `samples/java8/src/test/java/com/alibaba/cloud/cms/samples/client/ChatEventTest.java` (新建)

- [ ] **Step 1: 升级依赖版本**

修改 `samples/java8/pom.xml`:
```xml
<version>6.2.1</version>
```

运行:
```bash
cd samples/java8 && mvn compile
```

- [ ] **Step 2: 编写测试**

创建 `samples/java8/src/test/java/com/alibaba/cloud/cms/samples/client/ChatEventTest.java`:
```java
package com.alibaba.cloud.cms.samples.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.JsonNode;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class ChatEventTest {
    private final ObjectMapper mapper = new ObjectMapper();

    @Test
    void fromResponse_eventFieldDone_isDoneTrue() throws Exception {
        JsonNode body = mapper.readTree("{\"event\":\"done\",\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertTrue(event.isDone());
    }

    @Test
    void fromResponse_messagesFallback_isDoneTrue() throws Exception {
        JsonNode body = mapper.readTree("{\"messages\":[{\"type\":\"done\"}]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertTrue(event.isDone());
    }

    @Test
    void fromResponse_notDone() throws Exception {
        JsonNode body = mapper.readTree("{\"event\":\"text\",\"messages\":[{\"type\":\"text\"}]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertFalse(event.isDone());
    }

    @Test
    void fromResponse_extractsIdAndEvent() throws Exception {
        JsonNode body = mapper.readTree("{\"id\":\"evt-123\",\"event\":\"text\",\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertEquals("evt-123", event.getId());
        assertEquals("text", event.getEvent());
    }

    @Test
    void fromResponse_missingIdEvent_returnsNull() throws Exception {
        JsonNode body = mapper.readTree("{\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertNull(event.getId());
        assertNull(event.getEvent());
    }
}
```

运行验证失败:
```bash
cd samples/java8 && mvn test -Dtest=ChatEventTest
```

- [ ] **Step 3: ChatEvent 新增 id/event 字段、getter/setter、优化 isDoneMessage**

修改 `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/ChatEvent.java`:

新增字段:
```java
private String id;
private String event;
```

修改 `fromResponse`:
```java
public static ChatEvent fromResponse(JsonNode body, String rawJson, int statusCode) {
    ChatEvent event = new ChatEvent();
    event.body = body;
    event.rawJson = rawJson;
    event.statusCode = statusCode;
    event.id = body != null ? body.path("id").asText(null) : null;
    event.event = body != null ? body.path("event").asText(null) : null;
    event.done = isDoneMessage(body);
    return event;
}
```

修改 `isDoneMessage`:
```java
private static boolean isDoneMessage(JsonNode body) {
    if (body == null) {
        return false;
    }
    // 优先使用 response 级别的 event 字段
    JsonNode eventNode = body.get("event");
    if (eventNode != null && "done".equals(eventNode.asText())) {
        return true;
    }
    // fallback: 遍历 messages
    if (!body.has("messages")) {
        return false;
    }
    JsonNode messages = body.get("messages");
    if (messages.isArray()) {
        for (JsonNode msg : messages) {
            if (msg.has("type") && "done".equals(msg.get("type").asText())) {
                return true;
            }
        }
    }
    return false;
}
```

新增 getter/setter:
```java
public String getId() { return id; }
public void setId(String id) { this.id = id; }
public String getEvent() { return event; }
public void setEvent(String event) { this.event = event; }
```

- [ ] **Step 4: createThread 支持 attributes**

修改 `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/AgentClient.java`:
```java
public String createThread() throws SDKException {
    return createThread(null);
}

public String createThread(Map<String, String> attributes) throws SDKException {
    try {
        CreateThreadRequest request = new CreateThreadRequest();
        request.setTitle("Chat-" + Instant.now().getEpochSecond());

        CreateThreadRequest.CreateThreadRequestVariables variables = new CreateThreadRequest.CreateThreadRequestVariables();
        variables.setWorkspace(config.getWorkspace());
        request.setVariables(variables);

        if (attributes != null && !attributes.isEmpty()) {
            request.setAttributes(attributes);
        }

        CreateThreadResponse response = client.createThread(config.getEmployeeName(), request);
        // ... 其余不变
```

注意 Java 8 语法：不要使用 `Map.of()`，使用 `new HashMap<String, String>()` 等兼容写法。

- [ ] **Step 5: SimplePrinter 利用 event 字段快速跳过**

修改 `samples/java8/src/main/java/com/alibaba/cloud/cms/samples/client/SimplePrinter.java` 的 `processEvent` 方法，在方法开头增加快速跳过逻辑：
```java
public String processEvent(ChatEvent event) {
    if (event == null || event.getBody() == null) {
        return "";
    }
    // 利用 event 字段快速跳过非文本事件
    String eventType = event.getEvent();
    if (eventType != null && !"text".equals(eventType) && !"task_finished".equals(eventType)) {
        return "";
    }
    // ... 其余逻辑不变
```

- [ ] **Step 6: InteractiveHandler 利用 event 字段前置过滤**

在调用 `InteractiveHandler.handleEvent()` 之前的分发逻辑中，利用 `ChatEvent.getEvent()` 做前置过滤，减少不必要的 `handleEvent` 调用。

- [ ] **Step 7: 运行测试**

```bash
cd samples/java8 && mvn test
```
Expected: 所有测试通过

- [ ] **Step 8: 提交**

```bash
cd samples/java8
git add pom.xml src/
git commit -m "feat(java8): upgrade CMS SDK to 6.2.1, add id/event fields, support attributes"
```

---

### Task 6: 全量验证 + 更新文档

**Files:**
- Modify: `CLAUDE.md` (更新 SDK 版本说明)

- [ ] **Step 1: 全量构建验证**

```bash
cd samples/typescript && npm run build && npm run test
cd ../python && pytest tests/ -v
cd ../golang && make lint && make test
cd ../java && mvn test
cd ../java8 && mvn test
```
Expected: 五种语言全部构建成功、测试通过

- [ ] **Step 2: 更新 CLAUDE.md**

将 `CLAUDE.md` 中的 SDK 版本说明从 6.0.1 更新为 6.2.1。

- [ ] **Step 3: 最终提交**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for CMS SDK 6.2.1"
```
