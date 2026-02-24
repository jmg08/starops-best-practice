#!/usr/bin/env tsx
/**
 * 交互式对话示例
 * Interactive chat example
 *
 * Usage: npm run chat
 */

import * as readline from 'readline';
import { loadConfigFromEnv, AgentClient, SimplePrinter, SDKException } from '../client/index.js';

async function main() {
  console.log('🚀 VibeOps Chat (TypeScript)');
  console.log('='.repeat(60));

  try {
    // Load configuration
    const cfg = loadConfigFromEnv();
    console.log(`📋 Workspace: ${cfg.workspace}`);
    console.log(`📋 Employee: ${cfg.employeeName}\n`);

    // Create client
    const client = new AgentClient(cfg);

    // Create thread
    console.log('📝 创建会话...');
    const threadId = await client.createThread();
    console.log(`✅ ThreadID: ${threadId}\n`);

    // Create printer
    const printer = new SimplePrinter();

    // Interactive loop
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    const prompt = () => {
      rl.question('👤 请输入 (quit 退出): ', async (input) => {
        input = input.trim();

        if (!input) {
          prompt();
          return;
        }

        if (input === 'quit' || input === 'exit') {
          console.log('👋 再见!');
          rl.close();
          return;
        }

        console.log('-'.repeat(60));

        // Send message
        printer.reset();
        try {
          for await (const event of client.chat(threadId, input)) {
            if (event.error) {
              console.log(`❌ 错误: ${event.error.message}`);
              if (event.error.cause) {
                console.log(`   原因: ${(event.error.cause as Error).message || event.error.cause}`);
              }
              continue;
            }

            const text = printer.processEvent(event);
            if (text) {
              process.stdout.write(text);
            }
          }
        } catch (chatError) {
          console.log(`❌ 对话异常: ${(chatError as Error).message}`);
          if ((chatError as Error).stack) {
            console.log((chatError as Error).stack);
          }
        }

        console.log();
        console.log('='.repeat(60));
        console.log();

        prompt();
      });
    };

    prompt();
  } catch (e) {
    if (e instanceof SDKException) {
      console.log(`❌ 配置加载失败: ${e}`);
      console.log('\n请设置环境变量:');
      console.log('  VIBEOPS_WORKSPACE, VIBEOPS_ENDPOINT');
      console.log('  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET');
    } else {
      console.log(`❌ 错误: ${(e as Error).message}`);
    }
    process.exit(1);
  }
}

main();
