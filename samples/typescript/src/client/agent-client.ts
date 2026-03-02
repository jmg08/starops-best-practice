/**
 * Agent client for CMS SDK
 * CMS SDK Agent 客户端
 */

import CMS20240330Module, * as $CMS20240330 from '@alicloud/cms20240330';
import * as $OpenApi from '@alicloud/openapi-client';
import * as $dara from '@darabonba/typescript';
import { Config } from './config.js';
import { SDKException, ErrorCode } from './errors.js';

// Handle ESM/CJS interop
const CMS20240330Client = (CMS20240330Module as any).default || CMS20240330Module;

/** 聊天事件 / Chat event */
export interface ChatEvent {
  body?: Record<string, unknown>;
  rawJson: string;
  statusCode: number;
  isDone: boolean;
  error?: Error;
}

/** 会话信息 / Thread information */
export interface ThreadInfo {
  threadId: string;
  title: string;
  status: string;
  createTime: string;
  updateTime: string;
}

/** 会话消息 / Thread message */
export interface ThreadMessage {
  role: string;
  content: string;
  timestamp: string;
}

function isDoneMessage(body?: Record<string, unknown>): boolean {
  if (!body || !body.messages) return false;
  const messages = body.messages as Array<Record<string, unknown>>;
  return messages.some((msg) => msg.type === 'done');
}

/** Agent 客户端 / Agent client */
export class AgentClient {
  private client: any;
  private config: Config;

  constructor(config: Config) {
    this.config = config;

    try {
      const openApiConfig = new $OpenApi.Config({
        accessKeyId: config.accessKeyId,
        accessKeySecret: config.accessKeySecret,
        endpoint: config.endpoint,
        signatureVersion: 'v3',
      });
      this.client = new CMS20240330Client(openApiConfig);
    } catch (e) {
      throw SDKException.clientCreate(e as Error);
    }
  }

  getConfig(): Config {
    return this.config;
  }

  /** 创建会话 / Create thread */
  async createThread(): Promise<string> {
    try {
      const variables = new $CMS20240330.CreateThreadRequestVariables({
        workspace: this.config.workspace,
      });
      const request = new $CMS20240330.CreateThreadRequest({
        title: `Chat-${Math.floor(Date.now() / 1000)}`,
        variables,
      });

      const response = await this.client.createThread(
        this.config.employeeName,
        request
      );

      if (!response.body?.threadId) {
        throw new SDKException(ErrorCode.THREAD_CREATE, '无效响应: 缺少ThreadID');
      }

      return response.body.threadId;
    } catch (e) {
      if (e instanceof SDKException) throw e;
      throw SDKException.threadCreate(e as Error);
    }
  }

  /** 开始 SSE 对话 / Start SSE chat */
  async *chat(threadId: string, message: string): AsyncIterable<ChatEvent> {
    const variables: Record<string, unknown> = {
      workspace: this.config.workspace,
      region: this.config.region,
      language: 'zh',
      timeZone: 'Asia/Shanghai',
      timeStamp: String(Math.floor(Date.now() / 1000)),
    };
    yield* this.chatWithVariables(threadId, message, variables);
  }

  /** 开始 SSE 对话（支持自定义 variables）/ Start SSE chat with custom variables */
  async *chatWithVariables(
    threadId: string,
    message: string,
    variables?: Record<string, unknown>
  ): AsyncIterable<ChatEvent> {
    try {
      const content = new $CMS20240330.CreateChatRequestMessagesContents({
        type: 'text',
        value: message,
      });
      const msg = new $CMS20240330.CreateChatRequestMessages({
        role: 'user',
        contents: [content],
      });

      // Ensure required fields
      const vars = variables || {};
      vars.workspace = vars.workspace || this.config.workspace;
      vars.region = vars.region || this.config.region;
      vars.language = vars.language || 'zh';
      vars.timeZone = vars.timeZone || 'Asia/Shanghai';
      vars.timeStamp = vars.timeStamp || String(Math.floor(Date.now() / 1000));

      const request = new $CMS20240330.CreateChatRequest({
        action: 'create',
        threadId,
        digitalEmployeeName: this.config.employeeName,
        messages: [msg],
        variables: vars,
      });

      const runtime = new $dara.RuntimeOptions({
        readTimeout: 120000,  // 2 minutes
        connectTimeout: 30000, // 30 seconds
      });
      const responseIterator = await this.client.createChatWithSSE(request, {}, runtime);

      for await (const response of responseIterator) {
        if (response.body) {
          const bodyObj = response.body as unknown as Record<string, unknown>;
          const rawJson = JSON.stringify(bodyObj);
          const event: ChatEvent = {
            body: bodyObj,
            rawJson,
            statusCode: 200,
            isDone: isDoneMessage(bodyObj),
          };
          yield event;
          if (event.isDone) return;
        }
      }

      yield { rawJson: '', statusCode: 200, isDone: true };
    } catch (e) {
      const err = e as Error;
      const sdkErr = SDKException.chatFailed(err);
      // Preserve original error details
      if (err.message) {
        sdkErr.withContext('originalMessage', err.message);
      }
      if ((err as any).code) {
        sdkErr.withContext('errorCode', (err as any).code);
      }
      if ((err as any).data) {
        sdkErr.withContext('errorData', JSON.stringify((err as any).data));
      }
      yield { rawJson: '', statusCode: 0, isDone: false, error: sdkErr };
    }
  }

  /** 带超时的对话 / Chat with timeout */
  async *chatWithTimeout(
    threadId: string,
    message: string,
    timeout: number
  ): AsyncIterable<ChatEvent> {
    const deadline = Date.now() + timeout;

    for await (const event of this.chat(threadId, message)) {
      if (Date.now() > deadline) {
        yield { rawJson: '', statusCode: 0, isDone: false, error: SDKException.timeout(`${timeout}ms`) };
        return;
      }
      yield event;
      if (event.isDone || event.error) return;
    }
  }

  /** 列出会话 / List threads */
  async listThreads(pageSize = 20): Promise<{ threads: ThreadInfo[]; total: number }> {
    try {
      if (pageSize <= 0) pageSize = 20;
      if (pageSize > 100) pageSize = 100;

      const request = new $CMS20240330.ListThreadsRequest({
        maxResults: pageSize,
      });

      const response = await this.client.listThreads(
        this.config.employeeName,
        request
      );

      if (!response.body) {
        throw new SDKException(ErrorCode.PARSE_ERROR, '无效响应: 响应体为空')
          .withSuggestion('请稍后重试');
      }

      const threads: ThreadInfo[] = (response.body.threads || []).map((t: any) => ({
        threadId: t.threadId || '',
        title: t.title || '',
        status: t.status || '',
        createTime: t.createTime || '',
        updateTime: t.updateTime || '',
      }));

      return { threads, total: response.body.total || 0 };
    } catch (e) {
      if (e instanceof SDKException) throw e;
      throw new SDKException(ErrorCode.API_ERROR, '获取会话列表失败', e as Error)
        .withSuggestion('请检查网络连接和 API 权限');
    }
  }

  /** 获取会话详情 / Get thread details */
  async getThread(threadId: string): Promise<ThreadInfo> {
    this.validateThreadId(threadId);

    try {
      const response = await this.client.getThread(
        this.config.employeeName,
        threadId
      );

      if (!response.body) {
        throw new SDKException(ErrorCode.PARSE_ERROR, '无效响应: 响应体为空')
          .withContext('threadId', threadId)
          .withSuggestion('请稍后重试');
      }

      return {
        threadId: response.body.threadId || '',
        title: response.body.title || '',
        status: response.body.status || '',
        createTime: response.body.createTime || '',
        updateTime: response.body.updateTime || '',
      };
    } catch (e) {
      if (e instanceof SDKException) throw e;
      if (this.isThreadNotFoundError(e as Error)) {
        throw SDKException.threadNotFound(threadId);
      }
      throw new SDKException(ErrorCode.API_ERROR, `获取会话详情失败: ${threadId}`, e as Error)
        .withContext('threadId', threadId)
        .withSuggestion('请检查会话 ID 是否正确');
    }
  }

  /** 删除会话 / Delete thread */
  async deleteThread(threadId: string): Promise<void> {
    this.validateThreadId(threadId);

    try {
      await this.client.deleteThread(this.config.employeeName, threadId);
    } catch (e) {
      if (this.isThreadNotFoundError(e as Error)) {
        throw SDKException.threadNotFound(threadId);
      }
      throw new SDKException(ErrorCode.API_ERROR, `删除会话失败: ${threadId}`, e as Error)
        .withContext('threadId', threadId)
        .withSuggestion('请检查会话 ID 是否正确');
    }
  }

  /** 获取会话消息 / Get thread messages */
  async getThreadData(threadId: string, limit = 50): Promise<ThreadMessage[]> {
    this.validateThreadId(threadId);

    try {
      if (limit <= 0) limit = 50;
      if (limit > 100) limit = 100;

      const request = new $CMS20240330.GetThreadDataRequest({
        maxResults: limit,
      });

      const response = await this.client.getThreadData(
        this.config.employeeName,
        threadId,
        request
      );

      if (!response.body) {
        throw new SDKException(ErrorCode.PARSE_ERROR, '无效响应: 响应体为空')
          .withContext('threadId', threadId)
          .withSuggestion('请稍后重试');
      }

      // Strategy: prefer system Result over assistant streaming messages
      // 策略：优先使用 system Result，而不是 assistant 流式消息
      // - user messages: use timestamp as key
      // - system messages with Result: use timestamp as key (display as assistant)
      // - assistant messages: only include if no system Result exists
      
      const messageMap = new Map<string, { role: string; content: string; timestamp: string }>();
      const messageOrder: string[] = [];
      
      // Check if any system Result exists
      // 检查是否存在 system Result
      let hasSystemResult = false;
      for (const data of response.body.data || []) {
        for (const msg of data.messages || []) {
          if (msg.role === 'system') {
            const artifacts = (msg as any).artifacts as Array<Record<string, unknown>> | undefined;
            if (artifacts?.some(a => a.name === 'Result')) {
              hasSystemResult = true;
              break;
            }
          }
        }
        if (hasSystemResult) break;
      }

      // Process messages
      // 处理消息
      for (const data of response.body.data || []) {
        for (const msg of data.messages || []) {
          const role = msg.role || '';
          const timestamp = msg.timestamp || '';
          
          // Skip assistant streaming messages if system Result exists
          // 如果存在 system Result，跳过 assistant 流式消息
          if (role === 'assistant' && hasSystemResult) {
            continue;
          }
          
          // Use different key strategy based on role
          // 根据角色使用不同的键策略
          let key: string;
          if (role === 'user') {
            key = `user-${timestamp}`;
          } else if (role === 'system') {
            key = `system-${timestamp}`;
          } else {
            const callId = (msg as any).callId || '';
            key = `assistant-${callId}`;
          }
          
          const content = this.extractMessageContent(msg);
          if (!content) continue;

          if (messageMap.has(key)) {
            // Append content to existing message
            const existing = messageMap.get(key)!;
            existing.content += content;
          } else {
            // For system messages, display as assistant role
            // 对于 system 消息，显示为 assistant 角色
            const displayRole = role === 'system' ? 'assistant' : role;
            messageMap.set(key, { role: displayRole, content, timestamp });
            messageOrder.push(key);
          }
        }
      }

      // Return messages in order
      return messageOrder.map(key => messageMap.get(key)!);
    } catch (e) {
      if (e instanceof SDKException) throw e;
      if (this.isThreadNotFoundError(e as Error)) {
        throw SDKException.threadNotFound(threadId);
      }
      throw new SDKException(ErrorCode.API_ERROR, `获取会话消息失败: ${threadId}`, e as Error)
        .withContext('threadId', threadId)
        .withSuggestion('请检查会话 ID 是否正确');
    }
  }

  private validateThreadId(threadId: string): void {
    if (!threadId) {
      throw new SDKException(ErrorCode.CONFIG_INVALID, '会话 ID 不能为空')
        .withContext('threadId', threadId)
        .withSuggestion('请提供有效的会话 ID');
    }
    if (/[\s\t\n\r]/.test(threadId)) {
      throw new SDKException(ErrorCode.CONFIG_INVALID, `会话 ID 包含非法字符: ${threadId}`)
        .withContext('threadId', threadId)
        .withSuggestion('会话 ID 不能包含空白字符');
    }
  }

  private isThreadNotFoundError(e: Error): boolean {
    if (!e) return false;
    const errStr = e.message || '';
    const patterns = ['NotFound', 'not found', 'NOT_FOUND', 'ThreadNotFound', 'InvalidThreadId', 'does not exist'];
    return patterns.some((p) => errStr.includes(p));
  }

  private extractMessageContent(msg: Record<string, unknown>): string {
    // 1. Try to extract from contents (streaming text chunks)
    // 尝试从 contents 提取（流式文本块）
    const contents = msg.contents as Array<Record<string, unknown>> | undefined;
    if (contents) {
      const textContent = contents
        .filter((c) => c.type === 'text' && c.value)
        .map((c) => c.value as string)
        .join('');
      if (textContent) return textContent;
    }

    // 2. Try to extract from artifacts (final result)
    // 尝试从 artifacts 提取（最终结果）
    const artifacts = msg.artifacts as Array<Record<string, unknown>> | undefined;
    if (artifacts) {
      for (const artifact of artifacts) {
        // Look for Result artifact with text parts
        // 查找包含文本的 Result artifact
        if (artifact.name === 'Result') {
          const parts = artifact.parts as Array<Record<string, unknown>> | undefined;
          if (parts) {
            const textParts = parts
              .filter((p) => p.kind === 'text' && p.text)
              .map((p) => p.text as string);
            if (textParts.length > 0) return textParts.join('');
          }
        }
      }
    }

    return '';
  }
}
