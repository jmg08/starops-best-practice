#!/usr/bin/env tsx
/**
 * 交互事件处理示例
 * Interactive event handling example
 *
 * Usage: npm run chat-interactive
 */

import * as readline from 'readline';
import {
  loadConfigFromEnv,
  AgentClient,
  SimplePrinter,
  InteractiveHandler,
  SDKException,
  ChatEvent,
} from '../client/index.js';

const PRESET_QUESTIONS = [
  ['1', '查询 SLS 日志（可能触发 logstore 选择）', '查询最近一小时的错误日志'],
  ['2', '执行危险操作（可能触发确认）', '删除所有过期的告警规则'],
  ['3', '查询指标（可能触发指标选择）', '查询 ECS 实例的 CPU 使用率'],
  ['4', '模糊查询（可能触发输入补充）', '查询服务的性能数据'],
  ['5', '多选项查询（可能触发选择）', '帮我分析系统问题'],
];

async function processEventsWithInteraction(
  events: AsyncIterable<ChatEvent>,
  printer: SimplePrinter,
  handler: InteractiveHandler,
  threadId: string
): Promise<void> {
  for await (const event of events) {
    if (event.error) {
      console.log(`❌ 错误: ${event.error.message}`);
      continue;
    }

    // Extract text
    const text = printer.processEvent(event);
    if (text) {
      process.stdout.write(text);
    }

    // Check for interactive events
    if (event.body?.messages) {
      const messages = event.body.messages as Array<Record<string, unknown>>;
      for (const msg of messages) {
        const interactiveEvents = InteractiveHandler.extractInteractiveEvents(msg);
        for (const interactiveEvent of interactiveEvents) {
          console.log('\n' + '*'.repeat(60));
          console.log('🔔 检测到交互事件!');
          console.log(`   类型: ${interactiveEvent.type}`);
          console.log('*'.repeat(60));

          try {
            // Handle interactive event
            const response = await handler.handleEvent(interactiveEvent);

            console.log('\n✅ 交互响应:');
            console.log(`   ID: ${response.interactionId}`);
            console.log(`   类型: ${response.type}`);
            console.log(`   响应: ${JSON.stringify(response.response)}`);

            // Resume chat
            console.log('\n📤 恢复对话...');
            const resumeEvents = handler.resumeChat(threadId, response);
            await processEventsWithInteraction(resumeEvents, printer, handler, threadId);
          } catch (e) {
            console.log(`❌ 处理交互事件失败: ${(e as Error).message}`);
          }
        }
      }
    }
  }
}

function printHelp(): void {
  console.log('\n📖 命令帮助:');
  console.log('  help  - 显示帮助');
  console.log('  list  - 显示预设问题列表');
  console.log('  1-5   - 使用预设问题（可能触发交互事件）');
  console.log('  quit  - 退出');
  console.log('  <msg> - 发送自定义消息');
}

function printPresetQuestions(): void {
  console.log('\n📋 预设问题（可能触发交互事件）:');
  console.log('-'.repeat(60));
  for (const [id, desc, question] of PRESET_QUESTIONS) {
    console.log(`  [${id}] ${desc}`);
    console.log(`      问题: ${question}`);
  }
  console.log('-'.repeat(60));
}

async function main() {
  console.log('🚀 VibeOps Chat 交互事件处理示例 (TypeScript)');
  console.log('='.repeat(60));

  try {
    // Load configuration
    const cfg = loadConfigFromEnv();
    console.log('📋 配置:');
    console.log(`  Workspace: ${cfg.workspace}`);
    console.log(`  Employee: ${cfg.employeeName}\n`);

    // Create client
    const client = new AgentClient(cfg);

    // Create thread
    console.log('📝 创建会话...');
    const threadId = await client.createThread();
    console.log(`✅ 会话创建成功, ThreadID: ${threadId}\n`);

    // Create printer and handler
    const printer = new SimplePrinter();
    const handler = new InteractiveHandler(client, 60000);

    // Print help
    printHelp();

    // Interactive loop
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    const prompt = () => {
      rl.question('\n👤 请输入 (help 查看帮助): ', async (input) => {
        input = input.trim();

        if (!input) {
          prompt();
          return;
        }

        switch (input) {
          case 'quit':
          case 'exit':
            console.log('👋 再见!');
            rl.close();
            return;
          case 'help':
            printHelp();
            prompt();
            return;
          case 'list':
            printPresetQuestions();
            prompt();
            return;
        }

        // Check for preset question
        for (const [id, , question] of PRESET_QUESTIONS) {
          if (input === id) {
            input = question;
            console.log(`📝 使用预设问题: ${input}`);
            break;
          }
        }

        console.log('-'.repeat(60));

        // Send message and process response
        printer.reset();
        const events = client.chat(threadId, input);
        await processEventsWithInteraction(events, printer, handler, threadId);

        // Output final text
        const finalText = printer.getFinalText();
        if (finalText) {
          console.log(`\n📄 回复:\n${finalText}`);
        }

        console.log('='.repeat(60));

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
