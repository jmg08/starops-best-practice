package com.alibaba.cloud.cms.samples.examples;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.concurrent.BlockingQueue;

import com.alibaba.cloud.cms.samples.client.AgentClient;
import com.alibaba.cloud.cms.samples.client.ChatEvent;
import com.alibaba.cloud.cms.samples.client.Config;
import com.alibaba.cloud.cms.samples.client.SDKException;
import com.alibaba.cloud.cms.samples.client.SimplePrinter;

/**
 * 交互式对话示例
 * Interactive chat example
 *
 * Usage: mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.Chat"
 */
public class Chat {
    public static void main(String[] args) {
        System.out.println("🚀 VibeOps Chat (Java 8)");
        System.out.println(repeatStr("=", 60));

        try {
            Config cfg = Config.loadFromEnv();
            System.out.printf("📋 Employee: %s%n%n", cfg.getEmployeeName());

            AgentClient client = new AgentClient(cfg);

            System.out.println("📝 创建会话...");
            String threadId = client.createThread();
            System.out.printf("✅ ThreadID: %s%n%n", threadId);

            SimplePrinter printer = new SimplePrinter();

            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
            while (true) {
                System.out.print("👤 请输入 (quit 退出): ");
                String input = reader.readLine();

                if (input == null) {
                    System.out.println("\n👋 再见!");
                    break;
                }

                input = input.trim();
                if (input.isEmpty()) {
                    continue;
                }
                if (input.equals("quit") || input.equals("exit")) {
                    System.out.println("👋 再见!");
                    break;
                }

                System.out.println(repeatStr("-", 60));

                printer.reset();
                BlockingQueue<ChatEvent> events = client.chat(threadId, input);

                while (true) {
                    ChatEvent event = events.take();

                    if (event.hasError()) {
                        System.out.printf("❌ 错误: %s%n", event.getError().getMessage());
                        break;
                    }

                    String text = printer.processEvent(event);
                    if (!text.isEmpty()) {
                        System.out.print(text);
                    }

                    if (event.isDone()) {
                        break;
                    }
                }

                System.out.println();
                System.out.println(repeatStr("=", 60));
                System.out.println();
            }

            client.shutdown();
        } catch (SDKException e) {
            System.out.printf("❌ 配置加载失败: %s%n", e.getMessage());
            System.out.println("\n请设置环境变量:");
            System.out.println("  VIBEOPS_ENDPOINT");
            System.out.println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET");
            System.exit(1);
        } catch (Exception e) {
            System.out.printf("❌ 错误: %s%n", e.getMessage());
            System.exit(1);
        }
    }

    private static String repeatStr(String s, int count) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < count; i++) {
            sb.append(s);
        }
        return sb.toString();
    }
}
