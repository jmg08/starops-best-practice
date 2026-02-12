package com.alibaba.cloud.cms.samples.client;

import com.alibaba.cloud.cms.samples.types.EventType;
import com.alibaba.cloud.cms.samples.types.InteractionType;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.*;
import java.time.Duration;
import java.time.Instant;
import java.util.*;
import java.util.concurrent.*;

/**
 * 交互事件处理器
 * Interactive event handler
 */
public class InteractiveHandler {
    private final AgentClient client;
    private final Duration timeout;
    private BufferedReader reader;
    private PrintWriter writer;
    private final ObjectMapper objectMapper;

    public InteractiveHandler(AgentClient client, Duration timeout) {
        this.client = client;
        this.timeout = timeout;
        this.reader = new BufferedReader(new InputStreamReader(System.in));
        this.writer = new PrintWriter(System.out, true);
        this.objectMapper = new ObjectMapper();
    }

    public void setIO(Reader reader, Writer writer) {
        this.reader = new BufferedReader(reader);
        this.writer = new PrintWriter(writer, true);
    }

    /**
     * 处理交互事件
     * Handle interactive event
     */
    public InteractiveResponse handleEvent(JsonNode event) throws SDKException {
        if (event == null) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "事件为空");
        }

        String eventType = event.has("type") ? event.get("type").asText() : null;
        if (!EventType.INTERACTIVE.getValue().equals(eventType)) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "不支持的事件类型: " + eventType);
        }

        JsonNode payload = event.get("payload");
        if (payload == null) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "交互负载为空");
        }

        String interactiveType = payload.has("type") ? payload.get("type").asText() : null;
        InteractionType type = InteractionType.fromValue(interactiveType);

        if (type == null) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "不支持的交互类型: " + interactiveType);
        }

        switch (type) {
            case USER_ACK:
                return handleUserAck(payload);
            case USER_SELECT:
                return handleUserSelect(payload);
            case USER_INPUT:
                return handleUserInput(payload);
            default:
                throw new SDKException(ErrorCode.PARSE_ERROR, "不支持的交互类型: " + interactiveType);
        }
    }

    /**
     * 处理用户确认
     * Handle user acknowledgment
     */
    public InteractiveResponse handleUserAck(JsonNode payload) throws SDKException {
        String interactionId = getInteractionId(payload);
        String title = getMetaField(payload, "title");
        String description = getMetaField(payload, "description");

        writer.println("\n🔔 确认请求");
        if (title != null && !title.isEmpty()) {
            writer.println("   标题: " + title);
        }
        if (description != null && !description.isEmpty()) {
            writer.println("   描述: " + description);
        }
        writer.print("   请输入 [y/yes] 确认，[n/no] 取消: ");
        writer.flush();

        String input = readInputWithTimeout();
        input = input.trim().toLowerCase();
        boolean confirmed = input.isEmpty() || input.equals("y") || input.equals("yes") || input.equals("是");

        Map<String, Object> response = new HashMap<>();
        response.put("confirmed", confirmed);

        return new InteractiveResponse(interactionId, InteractionType.USER_ACK, response);
    }

    /**
     * 处理用户选择
     * Handle user selection
     */
    public InteractiveResponse handleUserSelect(JsonNode payload) throws SDKException {
        String interactionId = getInteractionId(payload);
        String title = getMetaField(payload, "title");

        writer.println("\n📋 请选择");
        if (title != null && !title.isEmpty()) {
            writer.println("   标题: " + title);
        }

        List<Map<String, Object>> options = getOptions(payload);
        if (options.isEmpty()) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "没有可选项");
        }

        writer.println("   选项:");
        for (int i = 0; i < options.size(); i++) {
            String label = getOptionLabel(options.get(i), i);
            writer.println("   [" + (i + 1) + "] " + label);
        }
        writer.print("   请输入选项编号 (1-" + options.size() + "): ");
        writer.flush();

        String input = readInputWithTimeout();
        input = input.trim();

        int selectedIndex;
        try {
            selectedIndex = Integer.parseInt(input);
        } catch (NumberFormatException e) {
            throw new SDKException(ErrorCode.PARSE_ERROR, 
                    "无效的选择: " + input + "，请输入 1-" + options.size() + " 之间的数字");
        }

        if (selectedIndex < 1 || selectedIndex > options.size()) {
            throw new SDKException(ErrorCode.PARSE_ERROR,
                    "无效的选择: " + input + "，请输入 1-" + options.size() + " 之间的数字");
        }

        Map<String, Object> response = new HashMap<>();
        response.put("selectedIndex", selectedIndex - 1);
        response.put("selectedValue", options.get(selectedIndex - 1));

        return new InteractiveResponse(interactionId, InteractionType.USER_SELECT, response);
    }

    /**
     * 处理用户输入
     * Handle user input
     */
    public InteractiveResponse handleUserInput(JsonNode payload) throws SDKException {
        String interactionId = getInteractionId(payload);
        String title = getMetaField(payload, "title");
        String description = getMetaField(payload, "description");
        String placeholder = getMetaField(payload, "placeholder");

        writer.println("\n✏️  请输入");
        if (title != null && !title.isEmpty()) {
            writer.println("   标题: " + title);
        }
        if (description != null && !description.isEmpty()) {
            writer.println("   描述: " + description);
        }
        if (placeholder != null && !placeholder.isEmpty()) {
            writer.println("   提示: " + placeholder);
        }
        writer.print("   请输入内容: ");
        writer.flush();

        String input = readInputWithTimeout();
        input = input.trim();

        Map<String, Object> response = new HashMap<>();
        response.put("value", input);

        return new InteractiveResponse(interactionId, InteractionType.USER_INPUT, response);
    }

    /**
     * 使用交互响应恢复对话
     * Resume chat with interactive response
     */
    public BlockingQueue<ChatEvent> resumeChat(String threadId, InteractiveResponse response) {
        if (client == null) {
            BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();
            try {
                events.put(ChatEvent.error(new SDKException(ErrorCode.CLIENT_CREATE, "客户端未初始化")));
            } catch (InterruptedException ignored) {}
            return events;
        }

        if (response == null) {
            BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();
            try {
                events.put(ChatEvent.error(new SDKException(ErrorCode.PARSE_ERROR, "交互响应为空")));
            } catch (InterruptedException ignored) {}
            return events;
        }

        Map<String, Object> variables = new HashMap<>();
        variables.put("workspace", client.getConfig().getWorkspace());
        variables.put("region", client.getConfig().getRegion());
        variables.put("language", "zh");
        variables.put("timeZone", "Asia/Shanghai");
        variables.put("timeStamp", String.valueOf(Instant.now().getEpochSecond()));
        variables.put("interactionId", response.getInteractionId());
        variables.put("interactionType", response.getType().getValue());
        variables.put("interactionResult", response.getResponse());

        String message;
        try {
            message = "[交互响应] " + objectMapper.writeValueAsString(response);
        } catch (Exception e) {
            message = "[交互响应] " + response.toString();
        }

        return client.chatWithVariables(threadId, message, variables);
    }

    private String readInputWithTimeout() throws SDKException {
        if (timeout == null || timeout.isZero()) {
            try {
                return reader.readLine();
            } catch (IOException e) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "读取输入失败", e);
            }
        }

        ExecutorService executor = Executors.newSingleThreadExecutor();
        Future<String> future = executor.submit(() -> reader.readLine());

        try {
            return future.get(timeout.toMillis(), TimeUnit.MILLISECONDS);
        } catch (TimeoutException e) {
            future.cancel(true);
            throw SDKException.interactiveTimeout(timeout.toString());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw SDKException.cancelled();
        } catch (ExecutionException e) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "读取输入失败", e.getCause());
        } finally {
            executor.shutdownNow();
        }
    }

    private String getInteractionId(JsonNode payload) {
        JsonNode meta = payload.get("meta");
        if (meta != null) {
            if (meta.has("id")) return meta.get("id").asText();
            if (meta.has("interactionId")) return meta.get("interactionId").asText();
        }
        return "interaction_" + System.nanoTime();
    }

    private String getMetaField(JsonNode payload, String field) {
        JsonNode meta = payload.get("meta");
        if (meta != null && meta.has(field)) {
            return meta.get(field).asText();
        }
        return null;
    }

    @SuppressWarnings("unchecked")
    private List<Map<String, Object>> getOptions(JsonNode payload) {
        List<Map<String, Object>> options = new ArrayList<>();

        // Try data field first
        if (payload.has("data") && payload.get("data").isArray()) {
            for (JsonNode item : payload.get("data")) {
                try {
                    options.add(objectMapper.convertValue(item, Map.class));
                } catch (Exception ignored) {}
            }
        }

        // Try meta.options
        if (options.isEmpty()) {
            JsonNode meta = payload.get("meta");
            if (meta != null && meta.has("options") && meta.get("options").isArray()) {
                for (JsonNode item : meta.get("options")) {
                    try {
                        options.add(objectMapper.convertValue(item, Map.class));
                    } catch (Exception ignored) {}
                }
            }
        }

        return options;
    }

    private String getOptionLabel(Map<String, Object> option, int index) {
        String[] fields = {"label", "name", "title", "value"};
        for (String field : fields) {
            Object value = option.get(field);
            if (value instanceof String && !((String) value).isEmpty()) {
                return (String) value;
            }
        }
        return "选项 " + (index + 1);
    }

    /**
     * 检查事件是否为交互事件
     * Check if event is interactive
     */
    public static boolean isInteractiveEvent(JsonNode event) {
        if (event == null) return false;
        String type = event.has("type") ? event.get("type").asText() : null;
        return EventType.INTERACTIVE.getValue().equals(type);
    }

    /**
     * 从消息中提取交互事件
     * Extract interactive events from message
     */
    public static List<JsonNode> extractInteractiveEvents(JsonNode message) {
        List<JsonNode> events = new ArrayList<>();
        if (message == null || !message.has("events")) return events;

        JsonNode eventsNode = message.get("events");
        if (eventsNode.isArray()) {
            for (JsonNode event : eventsNode) {
                if (isInteractiveEvent(event)) {
                    events.add(event);
                }
            }
        }
        return events;
    }
}
