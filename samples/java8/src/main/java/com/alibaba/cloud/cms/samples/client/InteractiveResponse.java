package com.alibaba.cloud.cms.samples.client;

import com.alibaba.cloud.cms.samples.types.InteractionType;
import java.util.Map;

/**
 * 交互响应
 * Interactive response
 */
public class InteractiveResponse {
    private final String interactionId;
    private final InteractionType type;
    private final Map<String, Object> response;

    public InteractiveResponse(String interactionId, InteractionType type, Map<String, Object> response) {
        this.interactionId = interactionId;
        this.type = type;
        this.response = response;
    }

    public String getInteractionId() {
        return interactionId;
    }

    public InteractionType getType() {
        return type;
    }

    public Map<String, Object> getResponse() {
        return response;
    }

    @Override
    public String toString() {
        return "InteractiveResponse{" +
                "interactionId='" + interactionId + '\'' +
                ", type=" + type +
                ", response=" + response +
                '}';
    }
}
