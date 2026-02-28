package com.alibaba.cloud.cms.samples.client;

import com.fasterxml.jackson.databind.JsonNode;

/**
 * 聊天事件
 * Chat event
 */
public class ChatEvent {
    private JsonNode body;
    private String rawJson;
    private int statusCode;
    private boolean done;
    private Exception error;

    public ChatEvent() {}

    public static ChatEvent done() {
        ChatEvent event = new ChatEvent();
        event.done = true;
        return event;
    }

    public static ChatEvent error(Exception error) {
        ChatEvent event = new ChatEvent();
        event.error = error;
        return event;
    }

    public static ChatEvent fromResponse(JsonNode body, String rawJson, int statusCode) {
        ChatEvent event = new ChatEvent();
        event.body = body;
        event.rawJson = rawJson;
        event.statusCode = statusCode;
        event.done = isDoneMessage(body);
        return event;
    }

    private static boolean isDoneMessage(JsonNode body) {
        if (body == null || !body.has("messages")) {
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

    public JsonNode getBody() {
        return body;
    }

    public void setBody(JsonNode body) {
        this.body = body;
    }

    public String getRawJson() {
        return rawJson;
    }

    public void setRawJson(String rawJson) {
        this.rawJson = rawJson;
    }

    public int getStatusCode() {
        return statusCode;
    }

    public void setStatusCode(int statusCode) {
        this.statusCode = statusCode;
    }

    public boolean isDone() {
        return done;
    }

    public void setDone(boolean done) {
        this.done = done;
    }

    public Exception getError() {
        return error;
    }

    public void setError(Exception error) {
        this.error = error;
    }

    public boolean hasError() {
        return error != null;
    }
}
