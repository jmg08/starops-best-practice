package com.alibaba.cloud.cms.samples.examples;

import com.alibaba.cloud.cms.samples.client.*;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.time.Duration;
import java.util.List;
import java.util.concurrent.BlockingQueue;

/**
 * 交互事件处理示例
 * Interactive event handling example
 *
 * Usage: mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatInteractive"
 */
public class ChatInteractive {
    private static final ObjectMapper objectMapper = new ObjectMapper();

    private static final String[][] PRESET_QUESTIONS = {
            {"1", "查询 SLS 日志（可能触发 logstore 选择）", "查询最近一小时的错误日志"},
            {"2", "执行危险操作（可能触发确认）", "删除所有过期的告警规则"},
            {"3", "查询指标（可能触发指标选择）", "查询 ECS 实例的 CPU 使用率"},
            {"4", "模糊查询（可能触发输入补充）", "查询服务的性能数据"},
            {"5", "多选项查询（可能触发选择）", "帮我分析系统问题"}
    };

    public static void main(String[] args) {
        System.out.println("🚀 VibeOps Chat 交互事件处理示例 (Java)");
        System.out.println("=".repeat(60));

        try {
            // Load configuration
            Config cfg = Config.loadFromEnv();
            printConfig(cfg);

            // Create client
            AgentClient client = new AgentClient(cfg);

            // Create thread
            System.out.println("📝 创建会话...");
            String threadId = client.createThread();
            System.out.printf("✅ 会话创建成功, ThreadID: %s%n%n", threadId);

            // Create printer and handler
            SimplePrinter printer = new SimplePrinter();
            InteractiveHandler handler = new InteractiveHandler(client, Duration.ofSeconds(60));

            // Print help
            printHelp();

            // Interactive loop
            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
            while (true) {
                System.out.print("\n👤 请输入 (help 查看帮助): ");
                String input = reader.readLine();

                if (input == null) {
                    System.out.println("\n👋 再见!");
                    break;
                }

                input = input.trim();
                if (input.isEmpty()) continue;

                switch (input) {
                    case "quit":
                    case "exit":
                        System.out.println("👋 再见!");
                        client.shutdown();
                        return;
                    case "help":
                        printHelp();
                        continue;
                    case "list":
                        printPresetQuestions();
                        continue;
                }

                // Check for preset question
                for (String[] q : PRESET_QUESTIONS) {
                    if (q[0].equals(input)) {
                        input = q[2];
                        System.out.printf("📝 使用预设问题: %s%n", input);
                        break;
                    }
                }

                System.out.println("-".repeat(60));

                // Send message and process response
                printer.reset();
                BlockingQueue<ChatEvent> events = client.chat(threadId, input);
                processEventsWithInteraction(events, printer, handler, threadId);

                // Output final text
                String finalText = printer.getFinalText();
                if (!finalText.isEmpty()) {
                    System.out.printf("%n📄 回复:%n%s%n", finalText);
                }

                System.out.println("=".repeat(60));
            }

            client.shutdown();
        } catch (SDKException e) {
            System.out.printf("❌ 配置加载失败: %s%n", e.getMessage());
            printEnvHelp();
            System.exit(1);
        } catch (Exception e) {
            System.out.printf("❌ 错误: %s%n", e.getMessage());
            System.exit(1);
        }
    }

    private static void processEventsWithInteraction(
            BlockingQueue<ChatEvent> events,
            SimplePrinter printer,
            InteractiveHandler handler,
            String threadId) throws Exception {

        while (true) {
            ChatEvent event = events.take();

            if (event.hasError()) {
                System.out.printf("❌ 错误: %s%n", event.getError().getMessage());
                break;
            }

            // Extract text
            String text = printer.processEvent(event);
            if (!text.isEmpty()) {
                System.out.print(text);
            }

            // Check for interactive events
            if (event.getBody() != null && event.getBody().has("messages")) {
                for (JsonNode msg : event.getBody().get("messages")) {
                    List<JsonNode> interactiveEvents = InteractiveHandler.extractInteractiveEvents(msg);
                    for (JsonNode interactiveEvent : interactiveEvents) {
                        System.out.println("\n" + "*".repeat(60));
                        System.out.println("🔔 检测到交互事件!");
                        System.out.printf("   类型: %s%n", interactiveEvent.get("type").asText());
                        System.out.println("*".repeat(60));

                        // Handle interactive event
                        InteractiveResponse response = handler.handleEvent(interactiveEvent);

                        System.out.println("\n✅ 交互响应:");
                        System.out.printf("   ID: %s%n", response.getInteractionId());
                        System.out.printf("   类型: %s%n", response.getType());
                        System.out.printf("   响应: %s%n", response.getResponse());

                        // Resume chat
                        System.out.println("\n📤 恢复对话...");
                        BlockingQueue<ChatEvent> resumeEvents = handler.resumeChat(threadId, response);
                        processEventsWithInteraction(resumeEvents, printer, handler, threadId);
                    }
                }
            }

            if (event.isDone()) {
                break;
            }
        }
    }

    private static void printHelp() {
        System.out.println("\n📖 命令帮助:");
        System.out.println("  help  - 显示帮助");
        System.out.println("  list  - 显示预设问题列表");
        System.out.println("  1-5   - 使用预设问题（可能触发交互事件）");
        System.out.println("  quit  - 退出");
        System.out.println("  <msg> - 发送自定义消息");
    }

    private static void printPresetQuestions() {
        System.out.println("\n📋 预设问题（可能触发交互事件）:");
        System.out.println("-".repeat(60));
        for (String[] q : PRESET_QUESTIONS) {
            System.out.printf("  [%s] %s%n", q[0], q[1]);
            System.out.printf("      问题: %s%n", q[2]);
        }
        System.out.println("-".repeat(60));
    }

    private static void printConfig(Config cfg) {
        System.out.println("📋 配置:");
        System.out.printf("  Workspace: %s%n", cfg.getWorkspace());
        System.out.printf("  Employee: %s%n%n", cfg.getEmployeeName());
    }

    private static void printEnvHelp() {
        System.out.println("\n请设置环境变量:");
        System.out.println("  VIBEOPS_WORKSPACE, VIBEOPS_ENDPOINT");
        System.out.println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET");
    }
}
