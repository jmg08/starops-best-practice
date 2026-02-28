package com.alibaba.cloud.cms.samples.client;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;

import com.aliyun.cms20240330.Client;
import com.aliyun.cms20240330.models.CreateChatRequest;
import com.aliyun.cms20240330.models.CreateChatResponse;
import com.aliyun.cms20240330.models.CreateThreadRequest;
import com.aliyun.cms20240330.models.CreateThreadResponse;
import com.aliyun.cms20240330.models.GetThreadDataRequest;
import com.aliyun.cms20240330.models.GetThreadDataResponse;
import com.aliyun.cms20240330.models.GetThreadDataResponseBody;
import com.aliyun.cms20240330.models.GetThreadResponse;
import com.aliyun.cms20240330.models.ListThreadsRequest;
import com.aliyun.cms20240330.models.ListThreadsResponse;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Agent 客户端
 * Agent client for CMS digital employee interactions
 */
public class AgentClient {
    private final Client client;
    private final Config config;
    private final ObjectMapper objectMapper;
    private final ExecutorService executor;

    public AgentClient(Config config) throws SDKException {
        this.config = config;
        this.objectMapper = new ObjectMapper();
        this.executor = Executors.newCachedThreadPool();

        try {
            com.aliyun.teaopenapi.models.Config openApiConfig = new com.aliyun.teaopenapi.models.Config();
            openApiConfig.setAccessKeyId(config.getAccessKeyId());
            openApiConfig.setAccessKeySecret(config.getAccessKeySecret());
            openApiConfig.setEndpoint(config.getEndpoint());
            openApiConfig.setSignatureVersion("v3");

            this.client = new Client(openApiConfig);
        } catch (Exception e) {
            throw SDKException.clientCreate(e);
        }
    }

    public Config getConfig() {
        return config;
    }

    /**
     * 创建会话
     * Create a new thread
     */
    public String createThread() throws SDKException {
        try {
            CreateThreadRequest request = new CreateThreadRequest();
            request.setTitle("Chat-" + Instant.now().getEpochSecond());

            CreateThreadRequest.CreateThreadRequestVariables variables = new CreateThreadRequest.CreateThreadRequestVariables();
            variables.setWorkspace(config.getWorkspace());
            request.setVariables(variables);

            CreateThreadResponse response = client.createThread(config.getEmployeeName(), request);

            if (response.getBody() == null || response.getBody().getThreadId() == null) {
                throw new SDKException(ErrorCode.THREAD_CREATE, "无效响应: 缺少ThreadID");
            }

            return response.getBody().getThreadId();
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            throw SDKException.threadCreate(e);
        }
    }


    /**
     * 开始 SSE 对话（基础版本）
     * Start SSE chat (basic version)
     */
    public BlockingQueue<ChatEvent> chat(String threadId, String message) {
        Map<String, Object> variables = new HashMap<>();
        variables.put("workspace", config.getWorkspace());
        variables.put("region", config.getRegion());
        variables.put("language", "zh");
        variables.put("timeZone", "Asia/Shanghai");
        variables.put("timeStamp", String.valueOf(Instant.now().getEpochSecond()));
        return chatWithVariables(threadId, message, variables);
    }

    /**
     * 开始对话（支持自定义 variables）
     * Start chat with custom variables
     * 
     * Java SDK 的 createChat() 方法将 bodyType 设为 "json"，但 API 实际返回 SSE 流。
     * tea-openapi 运行时尝试用 readAsJSON 解析 SSE 流会抛出 TeaException。
     * 我们从异常信息中提取 SSE data 行，逐条解析为 JSON 事件。
     *
     * The Java SDK's createChat() sets bodyType to "json", but the API returns an SSE stream.
     * The tea-openapi runtime throws TeaException when readAsJSON fails on SSE data.
     * We extract SSE "data:" lines from the exception message and parse each as a JSON event.
     */
    public BlockingQueue<ChatEvent> chatWithVariables(String threadId, String message, Map<String, Object> variables) {
        BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();
        
        // Create a copy of variables to make it effectively final
        final Map<String, Object> finalVariables = variables != null ? new HashMap<>(variables) : new HashMap<>();
        finalVariables.putIfAbsent("workspace", config.getWorkspace());
        finalVariables.putIfAbsent("region", config.getRegion());
        finalVariables.putIfAbsent("language", "zh");
        finalVariables.putIfAbsent("timeZone", "Asia/Shanghai");
        finalVariables.putIfAbsent("timeStamp", String.valueOf(Instant.now().getEpochSecond()));

        executor.submit(() -> {
            try {
                // Build request
                CreateChatRequest.CreateChatRequestMessagesContents content = new CreateChatRequest.CreateChatRequestMessagesContents();
                content.setType("text");
                content.setValue(message);

                CreateChatRequest.CreateChatRequestMessages msg = new CreateChatRequest.CreateChatRequestMessages();
                msg.setRole("user");
                msg.setContents(Collections.singletonList(content));

                CreateChatRequest request = new CreateChatRequest();
                request.setAction("create");
                request.setThreadId(threadId);
                request.setDigitalEmployeeName(config.getEmployeeName());
                request.setMessages(Collections.singletonList(msg));
                request.setVariables(finalVariables);

                com.aliyun.teautil.models.RuntimeOptions runtime = new com.aliyun.teautil.models.RuntimeOptions();
                runtime.setConnectTimeout(30000);
                runtime.setReadTimeout(300000);

                try {
                    // Try normal request first — will succeed if API returns JSON
                    CreateChatResponse response = client.createChatWithOptions(request, new HashMap<>(), runtime);

                    if (response.getBody() != null) {
                        String rawJson = objectMapper.writeValueAsString(response.getBody());
                        JsonNode jsonNode = objectMapper.readTree(rawJson);
                        ChatEvent event = ChatEvent.fromResponse(jsonNode, rawJson, 200);
                        events.put(event);
                    }
                } catch (Exception ex) {
                    // API returns SSE stream, tea-openapi fails to parse as JSON.
                    // The SSE data is in the exception message (or its cause's message).
                    String sseData = extractSSEFromException(ex);
                    if (sseData != null) {
                        parseSSEEvents(sseData, events);
                    } else {
                        throw ex;
                    }
                }

                events.put(ChatEvent.done());
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                try {
                    events.put(ChatEvent.error(SDKException.cancelled()));
                } catch (InterruptedException ignored) {}
            } catch (Exception e) {
                try {
                    events.put(ChatEvent.error(SDKException.chatFailed(e)));
                } catch (InterruptedException ignored) {}
            }
        });

        return events;
    }

    /**
     * 从 SSE 文本中解析事件
     * Parse SSE events from raw SSE text (extracted from exception message)
     */
    private void parseSSEEvents(String sseText, BlockingQueue<ChatEvent> events) throws InterruptedException {
        String[] lines = sseText.split("\n");
        for (String line : lines) {
            String trimmed = line.trim();
            if (trimmed.startsWith("data:")) {
                String jsonStr = trimmed.substring(5).trim();
                if (jsonStr.isEmpty() || "[DONE]".equals(jsonStr)) {
                    continue;
                }
                try {
                    JsonNode jsonNode = objectMapper.readTree(jsonStr);
                    ChatEvent event = ChatEvent.fromResponse(jsonNode, jsonStr, 200);
                    events.put(event);
                } catch (Exception ignored) {
                    // Skip malformed JSON lines
                }
            }
        }
    }

    /**
     * 从异常链中提取 SSE 数据
     * Extract SSE data from exception chain (TeaException, TeaUnretryableException, etc.)
     */
    private String extractSSEFromException(Exception ex) {
        // Check the exception itself
        String msg = ex.getMessage();
        if (msg != null && msg.contains("data:")) {
            return msg;
        }
        // Check the cause chain
        Throwable cause = ex.getCause();
        while (cause != null) {
            String causeMsg = cause.getMessage();
            if (causeMsg != null && causeMsg.contains("data:")) {
                return causeMsg;
            }
            cause = cause.getCause();
        }
        return null;
    }

    /**
     * 带超时的对话
     * Chat with timeout
     */
    public BlockingQueue<ChatEvent> chatWithTimeout(String threadId, String message, java.time.Duration timeout) {
        BlockingQueue<ChatEvent> events = chat(threadId, message);
        BlockingQueue<ChatEvent> timedEvents = new LinkedBlockingQueue<>();

        executor.submit(() -> {
            long deadline = System.currentTimeMillis() + timeout.toMillis();
            try {
                while (true) {
                    long remaining = deadline - System.currentTimeMillis();
                    if (remaining <= 0) {
                        timedEvents.put(ChatEvent.error(SDKException.timeout(timeout.toString())));
                        return;
                    }

                    ChatEvent event = events.poll(remaining, TimeUnit.MILLISECONDS);
                    if (event == null) {
                        timedEvents.put(ChatEvent.error(SDKException.timeout(timeout.toString())));
                        return;
                    }

                    timedEvents.put(event);
                    if (event.isDone() || event.hasError()) {
                        return;
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                try {
                    timedEvents.put(ChatEvent.error(SDKException.cancelled()));
                } catch (InterruptedException ignored) {}
            }
        });

        return timedEvents;
    }


    /**
     * 列出会话
     * List threads
     */
    public ListThreadsResult listThreads(int pageSize) throws SDKException {
        try {
            if (pageSize <= 0) pageSize = 20;
            if (pageSize > 100) pageSize = 100;

            ListThreadsRequest request = new ListThreadsRequest();
            request.setMaxResults((long) pageSize);

            ListThreadsResponse response = client.listThreads(config.getEmployeeName(), request);

            if (response.getBody() == null) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "无效响应: 响应体为空")
                        .withSuggestion("请稍后重试");
            }

            List<ThreadInfo> threads = new ArrayList<>();
            if (response.getBody().getThreads() != null) {
                for (var t : response.getBody().getThreads()) {
                    threads.add(new ThreadInfo(
                            t.getThreadId(),
                            t.getTitle(),
                            t.getStatus(),
                            t.getCreateTime(),
                            t.getUpdateTime()
                    ));
                }
            }

            long total = response.getBody().getTotal() != null ? response.getBody().getTotal() : 0;
            return new ListThreadsResult(threads, total);
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            throw new SDKException(ErrorCode.API_ERROR, "获取会话列表失败", e)
                    .withSuggestion("请检查网络连接和 API 权限");
        }
    }

    /**
     * 获取会话详情
     * Get thread details
     */
    public ThreadInfo getThread(String threadId) throws SDKException {
        validateThreadId(threadId);

        try {
            GetThreadResponse response = client.getThread(config.getEmployeeName(), threadId);

            if (response.getBody() == null) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "无效响应: 响应体为空")
                        .withContext("threadId", threadId)
                        .withSuggestion("请稍后重试");
            }

            return new ThreadInfo(
                    response.getBody().getThreadId(),
                    response.getBody().getTitle(),
                    response.getBody().getStatus(),
                    response.getBody().getCreateTime(),
                    response.getBody().getUpdateTime()
            );
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            if (isThreadNotFoundError(e)) {
                throw SDKException.threadNotFound(threadId);
            }
            throw new SDKException(ErrorCode.API_ERROR, "获取会话详情失败: " + threadId, e)
                    .withContext("threadId", threadId)
                    .withSuggestion("请检查会话 ID 是否正确");
        }
    }

    /**
     * 删除会话
     * Delete thread
     */
    public void deleteThread(String threadId) throws SDKException {
        validateThreadId(threadId);

        try {
            client.deleteThread(config.getEmployeeName(), threadId);
        } catch (Exception e) {
            if (isThreadNotFoundError(e)) {
                throw SDKException.threadNotFound(threadId);
            }
            throw new SDKException(ErrorCode.API_ERROR, "删除会话失败: " + threadId, e)
                    .withContext("threadId", threadId)
                    .withSuggestion("请检查会话 ID 是否正确");
        }
    }

    /**
     * 获取会话消息
     * Get thread messages
     */
    public List<ThreadMessage> getThreadData(String threadId, int limit) throws SDKException {
        validateThreadId(threadId);

        try {
            if (limit <= 0) limit = 50;
            if (limit > 100) limit = 100;

            GetThreadDataRequest request = new GetThreadDataRequest();
            request.setMaxResults((long) limit);

            GetThreadDataResponse response = client.getThreadData(config.getEmployeeName(), threadId, request);

            if (response.getBody() == null) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "无效响应: 响应体为空")
                        .withContext("threadId", threadId)
                        .withSuggestion("请稍后重试");
            }

            // Strategy: prefer system Result over assistant streaming messages
            // 策略：优先使用 system Result，而不是 assistant 流式消息
            Map<String, ThreadMessage> messageMap = new LinkedHashMap<>();
            
            // Check if any system Result exists
            // 检查是否存在 system Result
            boolean hasSystemResult = false;
            if (response.getBody().getData() != null) {
                outer:
                for (var data : response.getBody().getData()) {
                    if (data.getMessages() != null) {
                        for (var msg : data.getMessages()) {
                            if ("system".equals(msg.getRole()) && msg.getArtifacts() != null) {
                                for (var artifact : msg.getArtifacts()) {
                                    if (artifact != null && "Result".equals(artifact.get("name"))) {
                                        hasSystemResult = true;
                                        break outer;
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // Process messages
            if (response.getBody().getData() != null) {
                for (var data : response.getBody().getData()) {
                    if (data.getMessages() != null) {
                        for (var msg : data.getMessages()) {
                            String role = msg.getRole() != null ? msg.getRole() : "";
                            String timestamp = msg.getTimestamp() != null ? msg.getTimestamp() : "";
                            
                            // Skip assistant streaming messages if system Result exists
                            if ("assistant".equals(role) && hasSystemResult) {
                                continue;
                            }
                            
                            // Use different key strategy based on role
                            String key;
                            if ("user".equals(role)) {
                                key = "user-" + timestamp;
                            } else if ("system".equals(role)) {
                                key = "system-" + timestamp;
                            } else {
                                String callId = msg.getCallId() != null ? msg.getCallId() : "";
                                key = "assistant-" + callId;
                            }
                            
                            String content = extractMessageContent(msg);
                            if (content == null || content.isEmpty()) {
                                continue;
                            }

                            if (messageMap.containsKey(key)) {
                                ThreadMessage existing = messageMap.get(key);
                                messageMap.put(key, new ThreadMessage(
                                        existing.getRole(),
                                        existing.getContent() + content,
                                        existing.getTimestamp()
                                ));
                            } else {
                                // For system messages, display as assistant role
                                String displayRole = "system".equals(role) ? "assistant" : role;
                                messageMap.put(key, new ThreadMessage(displayRole, content, timestamp));
                            }
                        }
                    }
                }
            }

            return new ArrayList<>(messageMap.values());
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            if (isThreadNotFoundError(e)) {
                throw SDKException.threadNotFound(threadId);
            }
            throw new SDKException(ErrorCode.API_ERROR, "获取会话消息失败: " + threadId, e)
                    .withContext("threadId", threadId)
                    .withSuggestion("请检查会话 ID 是否正确");
        }
    }


    private void validateThreadId(String threadId) throws SDKException {
        if (threadId == null || threadId.isEmpty()) {
            throw new SDKException(ErrorCode.CONFIG_INVALID, "会话 ID 不能为空")
                    .withContext("threadId", threadId)
                    .withSuggestion("请提供有效的会话 ID");
        }
        if (threadId.contains(" ") || threadId.contains("\t") || threadId.contains("\n")) {
            throw new SDKException(ErrorCode.CONFIG_INVALID, "会话 ID 包含非法字符: " + threadId)
                    .withContext("threadId", threadId)
                    .withSuggestion("会话 ID 不能包含空白字符");
        }
    }

    private boolean isThreadNotFoundError(Exception e) {
        if (e == null) return false;
        String errStr = e.getMessage();
        if (errStr == null) return false;
        return errStr.contains("NotFound") || errStr.contains("not found") ||
               errStr.contains("NOT_FOUND") || errStr.contains("ThreadNotFound") ||
               errStr.contains("InvalidThreadId") || errStr.contains("does not exist");
    }

    private String extractMessageContent(GetThreadDataResponseBody.GetThreadDataResponseBodyDataMessages msg) {
        if (msg == null) return "";

        // 1. Try to extract from contents (streaming text chunks)
        // 尝试从 contents 提取（流式文本块）
        if (msg.getContents() != null) {
            StringBuilder result = new StringBuilder();
            for (var content : msg.getContents()) {
                if (content != null) {
                    Object type = content.get("type");
                    Object value = content.get("value");
                    if ("text".equals(type) && value != null) {
                        result.append(value.toString());
                    }
                }
            }
            if (result.length() > 0) {
                return result.toString();
            }
        }

        // 2. Try to extract from artifacts (final result)
        // 尝试从 artifacts 提取（最终结果）
        if (msg.getArtifacts() != null) {
            for (var artifact : msg.getArtifacts()) {
                if (artifact != null && "Result".equals(artifact.get("name"))) {
                    Object partsObj = artifact.get("parts");
                    if (partsObj instanceof List) {
                        List<?> parts = (List<?>) partsObj;
                        StringBuilder textParts = new StringBuilder();
                        for (Object partObj : parts) {
                            if (partObj instanceof Map) {
                                Map<?, ?> part = (Map<?, ?>) partObj;
                                if ("text".equals(part.get("kind")) && part.get("text") != null) {
                                    textParts.append(part.get("text").toString());
                                }
                            }
                        }
                        if (textParts.length() > 0) {
                            return textParts.toString();
                        }
                    }
                }
            }
        }

        return "";
    }

    public void shutdown() {
        executor.shutdown();
    }

    /**
     * 列出会话结果
     * List threads result
     */
    public static class ListThreadsResult {
        private final List<ThreadInfo> threads;
        private final long total;

        public ListThreadsResult(List<ThreadInfo> threads, long total) {
            this.threads = threads;
            this.total = total;
        }

        public List<ThreadInfo> getThreads() {
            return threads;
        }

        public long getTotal() {
            return total;
        }
    }
}
