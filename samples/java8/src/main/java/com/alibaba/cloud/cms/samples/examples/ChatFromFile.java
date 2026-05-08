package com.alibaba.cloud.cms.samples.examples;

import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.io.PrintWriter;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.BlockingQueue;

import com.alibaba.cloud.cms.samples.client.AgentClient;
import com.alibaba.cloud.cms.samples.client.ChatEvent;
import com.alibaba.cloud.cms.samples.client.Config;
import com.alibaba.cloud.cms.samples.client.EventPrinter;
import com.alibaba.cloud.cms.samples.client.SDKException;
import com.alibaba.cloud.cms.samples.client.SimplePrinter;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 从文件加载请求示例
 * Load requests from file example
 *
 * Usage:
 *   mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatFromFile" -Dexec.args="-file ../../requests/cms/entity.json"
 *   mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatFromFile" -Dexec.args="-dir ../../requests/cms/"
 */
public class ChatFromFile {
    private static String filePath = null;
    private static String dirPath = null;
    private static boolean simpleMode = false;
    private static String outputDir = "../../requests/output";
    private static final ObjectMapper objectMapper = new ObjectMapper();

    public static void main(String[] args) {
        parseArgs(args);

        if (filePath == null && dirPath == null) {
            printUsage();
            System.exit(1);
        }

        System.out.println("🚀 VibeOps Chat - 从文件加载请求 (Java 8)");
        System.out.println(repeatStr("=", 60));

        try {
            Config cfg = Config.loadFromEnv();
            AgentClient client = new AgentClient(cfg);
            Files.createDirectories(Paths.get(outputDir));

            if (dirPath != null) {
                processDirectory(client, dirPath);
            } else {
                processFile(client, filePath);
            }

            client.shutdown();
        } catch (SDKException e) {
            System.out.printf("❌ 配置加载失败: %s%n", e.getMessage());
            System.exit(1);
        } catch (Exception e) {
            System.out.printf("❌ 错误: %s%n", e.getMessage());
            System.exit(1);
        }
    }

    private static void parseArgs(String[] args) {
        for (int i = 0; i < args.length; i++) {
            switch (args[i]) {
                case "-file":
                    if (i + 1 < args.length) filePath = args[++i];
                    break;
                case "-dir":
                    if (i + 1 < args.length) dirPath = args[++i];
                    break;
                case "-simple":
                    simpleMode = true;
                    break;
                case "-output":
                    if (i + 1 < args.length) outputDir = args[++i];
                    break;
            }
        }
    }

    private static void processDirectory(AgentClient client, String dir) throws IOException {
        File[] files = new File(dir).listFiles((d, name) -> name.endsWith(".json"));
        if (files == null || files.length == 0) {
            System.out.printf("⚠️ 目录中没有 JSON 文件: %s%n", dir);
            return;
        }

        System.out.printf("📁 找到 %d 个请求文件%n%n", files.length);

        for (int i = 0; i < files.length; i++) {
            System.out.printf("━━━ [%d/%d] %s ━━━%n", i + 1, files.length, files[i].getName());
            processFile(client, files[i].getPath());
            System.out.println();
        }

        System.out.printf("✅ 处理完成，共 %d 个文件%n", files.length);
    }

    private static void processFile(AgentClient client, String file) {
        try {
            JsonNode reqFile = objectMapper.readTree(new File(file));

            String message = extractMessage(reqFile);
            if (message == null || message.isEmpty()) {
                System.out.println("⚠️ 文件中没有消息内容");
                return;
            }

            System.out.printf("📄 文件: %s%n", new File(file).getName());
            System.out.printf("💬 消息: %s%n", truncate(message, 60));

            String threadId = client.createThread();

            PrintWriter outputFile = createOutputFile(file);

            writeOutput(outputFile, "# Request: " + new File(file).getName());
            writeOutput(outputFile, "# Time: " + Instant.now().toString());
            writeOutput(outputFile, "# ThreadID: " + threadId);
            writeOutput(outputFile, "# Message: " + message + "\n");

            System.out.println(repeatStr("-", 60));

            Map<String, Object> variables = extractVariables(reqFile);

            long startTime = System.currentTimeMillis();
            BlockingQueue<ChatEvent> events = client.chatWithVariables(threadId, message, variables);

            // 与 Java 11 蓝本一致：根据 simpleMode 选择打印器
            // - simpleMode=true:  SimplePrinter，仅输出最终的 system Result（看起来"一次性刷出"）
            // - simpleMode=false: EventPrinter，逐 SSE 事件输出（看起来"分段流式"）
            // Aligned with Java 11 baseline: select printer based on simpleMode
            SimplePrinter simplePrinter = simpleMode ? new SimplePrinter() : null;
            EventPrinter eventPrinter = simpleMode ? null : new EventPrinter(false, true);
            int eventIndex = 0;

            while (true) {
                ChatEvent event = events.take();
                eventIndex++;

                if (event.hasError()) {
                    Exception err = event.getError();
                    System.out.printf("❌ 错误: %s%n", err.toString());
                    if (err.getCause() != null) {
                        System.out.printf("   原因: %s%n", err.getCause().getMessage());
                    }
                    writeOutput(outputFile, "[ERROR] " + err.toString());
                    break;
                }

                if (event.getRawJson() != null && !event.getRawJson().isEmpty()) {
                    writeOutput(outputFile, "[EVENT " + eventIndex + "]\n" + event.getRawJson() + "\n");
                }

                if (simpleMode) {
                    String text = simplePrinter.processEvent(event);
                    if (!text.isEmpty()) {
                        System.out.print(text);
                    }
                } else {
                    eventPrinter.printEvent(event, eventIndex);
                }

                if (event.isDone()) {
                    break;
                }
            }

            long elapsed = System.currentTimeMillis() - startTime;
            System.out.println();

            if (simpleMode) {
                String finalText = simplePrinter.getFinalText();
                writeOutput(outputFile, "\n# Final Result:\n" + finalText);
                System.out.printf("📄 最终文本:%n%s%n", finalText);
            }

            writeOutput(outputFile, "\n# Duration: " + elapsed + "ms");
            System.out.printf("⏱️  耗时: %dms%n", elapsed);

            if (outputFile != null) {
                outputFile.close();
            }
        } catch (Exception e) {
            System.out.printf("❌ 处理文件失败: %s%n", e.getMessage());
        }
    }

    private static String extractMessage(JsonNode reqFile) {
        if (reqFile.has("messages") && reqFile.get("messages").isArray()) {
            JsonNode messages = reqFile.get("messages");
            if (messages.size() > 0) {
                JsonNode firstMsg = messages.get(0);
                if (firstMsg.has("contents") && firstMsg.get("contents").isArray()) {
                    JsonNode contents = firstMsg.get("contents");
                    if (contents.size() > 0 && contents.get(0).has("value")) {
                        return contents.get(0).get("value").asText();
                    }
                }
            }
        }
        return null;
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> extractVariables(JsonNode reqFile) {
        if (reqFile.has("variables")) {
            try {
                return objectMapper.convertValue(reqFile.get("variables"), Map.class);
            } catch (Exception ignored) {}
        }
        return new HashMap<String, Object>();
    }

    private static PrintWriter createOutputFile(String inputFile) {
        try {
            String baseName = new File(inputFile).getName().replace(".json", "");
            String timestamp = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMdd-HHmmss"));
            String outputPath = outputDir + "/" + baseName + "-" + timestamp + ".log";
            return new PrintWriter(new FileWriter(outputPath));
        } catch (IOException e) {
            return null;
        }
    }

    private static void writeOutput(PrintWriter file, String content) {
        if (file != null) {
            file.println(content);
        }
    }

    private static String truncate(String s, int maxLen) {
        if (s.length() <= maxLen) return s;
        return s.substring(0, maxLen) + "...";
    }

    private static String repeatStr(String s, int count) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < count; i++) {
            sb.append(s);
        }
        return sb.toString();
    }

    private static void printUsage() {
        System.out.println("用法:");
        System.out.println("  -file <path>   处理单个文件");
        System.out.println("  -dir <path>    处理目录下所有 JSON 文件");
        System.out.println();
        System.out.println("示例:");
        System.out.println("  -file ../../requests/cms/entity.json");
        System.out.println("  -dir ../../requests/cms/");
        System.out.println("  -file entity.json -simple");
        System.out.println();
        System.out.println("选项:");
        System.out.println("  -file     请求 JSON 文件路径");
        System.out.println("  -dir      请求文件目录");
        System.out.println("  -simple   简洁模式，只输出最终文本");
        System.out.println("  -output   输出目录 (默认: output)");
    }
}
