/**
 * Interactive handler for CMS SDK
 * CMS SDK 交互处理器
 */

import * as readline from 'readline';
import { EventType, InteractionType } from '../types/index.js';
import { AgentClient, ChatEvent } from './agent-client.js';
import { SDKException, ErrorCode } from './errors.js';

/** 交互响应 / Interactive response */
export interface InteractiveResponse {
  interactionId: string;
  type: InteractionType;
  response: Record<string, unknown>;
}

/** 交互事件处理器 / Interactive event handler */
export class InteractiveHandler {
  private client: AgentClient;
  private timeout?: number;
  private rl?: readline.Interface;

  constructor(client: AgentClient, timeout?: number) {
    this.client = client;
    this.timeout = timeout;
  }

  /** 处理交互事件 / Handle interactive event */
  async handleEvent(event: Record<string, unknown>): Promise<InteractiveResponse> {
    if (!event) {
      throw new SDKException(ErrorCode.PARSE_ERROR, '事件为空');
    }

    const eventType = event.type as string;
    if (eventType !== EventType.INTERACTIVE) {
      throw new SDKException(ErrorCode.PARSE_ERROR, `不支持的事件类型: ${eventType}`);
    }

    const payload = event.payload as Record<string, unknown>;
    if (!payload) {
      throw new SDKException(ErrorCode.PARSE_ERROR, '交互负载为空');
    }

    const interactiveType = payload.type as string;

    switch (interactiveType) {
      case InteractionType.USER_ACK:
        return this.handleUserAck(payload);
      case InteractionType.USER_SELECT:
        return this.handleUserSelect(payload);
      case InteractionType.USER_INPUT:
        return this.handleUserInput(payload);
      default:
        throw new SDKException(ErrorCode.PARSE_ERROR, `不支持的交互类型: ${interactiveType}`);
    }
  }

  /** 处理用户确认 / Handle user acknowledgment */
  async handleUserAck(payload: Record<string, unknown>): Promise<InteractiveResponse> {
    const interactionId = this.getInteractionId(payload);
    const title = this.getMetaField(payload, 'title');
    const description = this.getMetaField(payload, 'description');

    console.log('\n🔔 确认请求');
    if (title) console.log(`   标题: ${title}`);
    if (description) console.log(`   描述: ${description}`);

    const input = await this.prompt('   请输入 [y/yes] 确认，[n/no] 取消: ');
    const confirmed = ['', 'y', 'yes', '是'].includes(input.trim().toLowerCase());

    return {
      interactionId,
      type: InteractionType.USER_ACK,
      response: { confirmed },
    };
  }

  /** 处理用户选择 / Handle user selection */
  async handleUserSelect(payload: Record<string, unknown>): Promise<InteractiveResponse> {
    const interactionId = this.getInteractionId(payload);
    const title = this.getMetaField(payload, 'title');

    console.log('\n📋 请选择');
    if (title) console.log(`   标题: ${title}`);

    const options = this.getOptions(payload);
    if (options.length === 0) {
      throw new SDKException(ErrorCode.PARSE_ERROR, '没有可选项');
    }

    console.log('   选项:');
    options.forEach((opt, i) => {
      const label = this.getOptionLabel(opt, i);
      console.log(`   [${i + 1}] ${label}`);
    });

    const input = await this.prompt(`   请输入选项编号 (1-${options.length}): `);
    const selectedIndex = parseInt(input.trim(), 10);

    if (isNaN(selectedIndex) || selectedIndex < 1 || selectedIndex > options.length) {
      throw new SDKException(
        ErrorCode.PARSE_ERROR,
        `无效的选择: ${input}，请输入 1-${options.length} 之间的数字`
      );
    }

    return {
      interactionId,
      type: InteractionType.USER_SELECT,
      response: {
        selectedIndex: selectedIndex - 1,
        selectedValue: options[selectedIndex - 1],
      },
    };
  }

  /** 处理用户输入 / Handle user input */
  async handleUserInput(payload: Record<string, unknown>): Promise<InteractiveResponse> {
    const interactionId = this.getInteractionId(payload);
    const title = this.getMetaField(payload, 'title');
    const description = this.getMetaField(payload, 'description');
    const placeholder = this.getMetaField(payload, 'placeholder');

    console.log('\n✏️  请输入');
    if (title) console.log(`   标题: ${title}`);
    if (description) console.log(`   描述: ${description}`);
    if (placeholder) console.log(`   提示: ${placeholder}`);

    const input = await this.prompt('   请输入内容: ');

    return {
      interactionId,
      type: InteractionType.USER_INPUT,
      response: { value: input.trim() },
    };
  }

  /** 使用交互响应恢复对话 / Resume chat with interactive response */
  async *resumeChat(
    threadId: string,
    response: InteractiveResponse
  ): AsyncIterable<ChatEvent> {
    if (!this.client) {
      yield {
        rawJson: '',
        statusCode: 0,
        isDone: false,
        error: new SDKException(ErrorCode.CLIENT_CREATE, '客户端未初始化'),
      };
      return;
    }

    if (!response) {
      yield {
        rawJson: '',
        statusCode: 0,
        isDone: false,
        error: new SDKException(ErrorCode.PARSE_ERROR, '交互响应为空'),
      };
      return;
    }

    const variables: Record<string, unknown> = {
      workspace: this.client.getConfig().workspace,
      region: this.client.getConfig().region,
      language: 'zh',
      timeZone: 'Asia/Shanghai',
      timeStamp: String(Math.floor(Date.now() / 1000)),
      interactionId: response.interactionId,
      interactionType: response.type,
      interactionResult: response.response,
    };

    const message = `[交互响应] ${JSON.stringify(response)}`;

    yield* this.client.chatWithVariables(threadId, message, variables);
  }

  private async prompt(question: string): Promise<string> {
    return new Promise((resolve) => {
      this.rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
      });

      this.rl.question(question, (answer) => {
        this.rl?.close();
        resolve(answer);
      });
    });
  }

  private getInteractionId(payload: Record<string, unknown>): string {
    const meta = payload.meta as Record<string, unknown> | undefined;
    if (meta?.id) return meta.id as string;
    if (meta?.interactionId) return meta.interactionId as string;
    return `interaction_${Date.now()}`;
  }

  private getMetaField(payload: Record<string, unknown>, field: string): string | undefined {
    const meta = payload.meta as Record<string, unknown> | undefined;
    return meta?.[field] as string | undefined;
  }

  private getOptions(payload: Record<string, unknown>): Array<Record<string, unknown>> {
    // Try data field first
    const data = payload.data as Array<Record<string, unknown>> | undefined;
    if (data && data.length > 0) return data;

    // Try meta.options
    const meta = payload.meta as Record<string, unknown> | undefined;
    const options = meta?.options as Array<Record<string, unknown>> | undefined;
    return options || [];
  }

  private getOptionLabel(option: Record<string, unknown>, index: number): string {
    for (const field of ['label', 'name', 'title', 'value']) {
      const value = option[field];
      if (typeof value === 'string' && value) return value;
    }
    return `选项 ${index + 1}`;
  }

  /** 检查事件是否为交互事件 / Check if event is interactive */
  static isInteractiveEvent(event: Record<string, unknown>): boolean {
    return event?.type === EventType.INTERACTIVE;
  }

  /** 从消息中提取交互事件 / Extract interactive events from message */
  static extractInteractiveEvents(
    message: Record<string, unknown>
  ): Array<Record<string, unknown>> {
    const events = message?.events as Array<Record<string, unknown>> | undefined;
    if (!events) return [];
    return events.filter((e) => InteractiveHandler.isInteractiveEvent(e));
  }
}
