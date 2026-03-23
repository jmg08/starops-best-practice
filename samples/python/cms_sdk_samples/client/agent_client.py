"""
Agent client for CMS SDK
CMS SDK Agent 客户端
"""

import asyncio
import json
import time
from dataclasses import dataclass
from typing import Any, AsyncIterator, Dict, List, Optional

from alibabacloud_cms20240330.client import Client
from alibabacloud_cms20240330 import models as cms_models
from alibabacloud_tea_openapi import models as openapi_models
from alibabacloud_tea_util import models as util_models

from .config import Config
from .errors import SDKException, ErrorCode


@dataclass
class ChatEvent:
    """聊天事件 / Chat event"""
    body: Optional[Dict[str, Any]] = None
    raw_json: str = ""
    status_code: int = 0
    is_done: bool = False
    error: Optional[Exception] = None
    id: Optional[str] = None
    event: Optional[str] = None

    @classmethod
    def done(cls) -> "ChatEvent":
        return cls(is_done=True)

    @classmethod
    def from_error(cls, error: Exception) -> "ChatEvent":
        return cls(error=error)

    @classmethod
    def from_response(cls, body: Dict[str, Any], raw_json: str, status_code: int) -> "ChatEvent":
        is_done = cls._is_done_message(body)
        return cls(
            body=body,
            raw_json=raw_json,
            status_code=status_code,
            is_done=is_done,
            id=body.get("id") if body else None,
            event=body.get("event") if body else None,
        )

    @staticmethod
    def _is_done_message(body: Optional[Dict[str, Any]]) -> bool:
        if not body:
            return False
        # 优先使用 response 级别的 event 字段
        if body.get("event") == "done":
            return True
        # fallback: 遍历 messages
        messages = body.get("messages", [])
        for msg in messages:
            if isinstance(msg, dict) and msg.get("type") == "done":
                return True
        return False

    def has_error(self) -> bool:
        return self.error is not None


@dataclass
class ThreadInfo:
    """会话信息 / Thread information"""
    thread_id: str
    title: str = ""
    status: str = ""
    create_time: str = ""
    update_time: str = ""


@dataclass
class ThreadMessage:
    """会话消息 / Thread message"""
    role: str
    content: str
    timestamp: str = ""


class AgentClient:
    """Agent 客户端 / Agent client"""

    def __init__(self, config: Config):
        self.config = config
        try:
            openapi_config = openapi_models.Config(
                access_key_id=config.access_key_id,
                access_key_secret=config.access_key_secret,
                endpoint=config.endpoint,
                signature_version="v3",
            )
            self._client = Client(openapi_config)
        except Exception as e:
            raise SDKException.client_create(e)

    def create_thread(self) -> str:
        """创建会话 / Create thread"""
        try:
            variables = cms_models.CreateThreadRequestVariables(
                workspace=self.config.workspace
            )
            request = cms_models.CreateThreadRequest(
                title=f"Chat-{int(time.time())}",
                variables=variables,
            )
            response = self._client.create_thread(self.config.employee_name, request)

            if not response.body or not response.body.thread_id:
                raise SDKException(ErrorCode.THREAD_CREATE, "无效响应: 缺少ThreadID")

            return response.body.thread_id
        except SDKException:
            raise
        except Exception as e:
            raise SDKException.thread_create(e)

    async def chat(self, thread_id: str, message: str) -> AsyncIterator[ChatEvent]:
        """开始 SSE 对话 / Start SSE chat"""
        variables = {
            "workspace": self.config.workspace,
            "region": self.config.region,
            "language": "zh",
            "timeZone": "Asia/Shanghai",
            "timeStamp": str(int(time.time())),
        }
        async for event in self.chat_with_variables(thread_id, message, variables):
            yield event


    async def chat_with_variables(
        self, thread_id: str, message: str, variables: Optional[Dict[str, Any]] = None
    ) -> AsyncIterator[ChatEvent]:
        """开始 SSE 对话（支持自定义 variables）/ Start SSE chat with custom variables"""
        try:
            # Build request
            content = cms_models.CreateChatRequestMessagesContents(
                type="text",
                value=message,
            )
            msg = cms_models.CreateChatRequestMessages(
                role="user",
                contents=[content],
            )

            # Ensure required fields
            if variables is None:
                variables = {}
            variables.setdefault("workspace", self.config.workspace)
            variables.setdefault("region", self.config.region)
            variables.setdefault("language", "zh")
            variables.setdefault("timeZone", "Asia/Shanghai")
            variables.setdefault("timeStamp", str(int(time.time())))

            request = cms_models.CreateChatRequest(
                action="create",
                thread_id=thread_id,
                digital_employee_name=self.config.employee_name,
                messages=[msg],
                variables=variables,
            )

            # Use SSE streaming
            runtime = util_models.RuntimeOptions()
            runtime.connect_timeout = 30000
            runtime.read_timeout = 300000
            
            response_iterator = self._client.create_chat_with_sse(request, {}, runtime)

            for response in response_iterator:
                if response.body:
                    body_dict = response.body.to_map()
                    raw_json = json.dumps(body_dict, ensure_ascii=False)
                    event = ChatEvent.from_response(body_dict, raw_json, 200)
                    yield event
                    if event.is_done:
                        return

            yield ChatEvent.done()
        except Exception as e:
            yield ChatEvent.from_error(SDKException.chat_failed(e))

    async def chat_with_timeout(
        self, thread_id: str, message: str, timeout: float
    ) -> AsyncIterator[ChatEvent]:
        """带超时的对话 / Chat with timeout"""
        try:
            async for event in asyncio.wait_for(
                self._collect_events(thread_id, message), timeout=timeout
            ):
                yield event
        except asyncio.TimeoutError:
            yield ChatEvent.from_error(SDKException.timeout(f"{timeout}s"))

    async def _collect_events(self, thread_id: str, message: str) -> AsyncIterator[ChatEvent]:
        async for event in self.chat(thread_id, message):
            yield event


    def list_threads(self, page_size: int = 20) -> tuple[List[ThreadInfo], int]:
        """列出会话 / List threads"""
        try:
            if page_size <= 0:
                page_size = 20
            if page_size > 100:
                page_size = 100

            request = cms_models.ListThreadsRequest(max_results=page_size)
            response = self._client.list_threads(self.config.employee_name, request)

            if not response.body:
                raise SDKException(
                    ErrorCode.PARSE_ERROR, "无效响应: 响应体为空"
                ).with_suggestion("请稍后重试")

            threads = []
            if response.body.threads:
                for t in response.body.threads:
                    threads.append(ThreadInfo(
                        thread_id=t.thread_id or "",
                        title=t.title or "",
                        status=t.status or "",
                        create_time=t.create_time or "",
                        update_time=t.update_time or "",
                    ))

            total = response.body.total or 0
            return threads, total
        except SDKException:
            raise
        except Exception as e:
            raise SDKException(
                ErrorCode.API_ERROR, "获取会话列表失败", e
            ).with_suggestion("请检查网络连接和 API 权限")

    def get_thread(self, thread_id: str) -> ThreadInfo:
        """获取会话详情 / Get thread details"""
        self._validate_thread_id(thread_id)

        try:
            response = self._client.get_thread(self.config.employee_name, thread_id)

            if not response.body:
                raise SDKException(
                    ErrorCode.PARSE_ERROR, "无效响应: 响应体为空"
                ).with_context("threadId", thread_id).with_suggestion("请稍后重试")

            return ThreadInfo(
                thread_id=response.body.thread_id or "",
                title=response.body.title or "",
                status=response.body.status or "",
                create_time=response.body.create_time or "",
                update_time=response.body.update_time or "",
            )
        except SDKException:
            raise
        except Exception as e:
            if self._is_thread_not_found_error(e):
                raise SDKException.thread_not_found(thread_id)
            raise SDKException(
                ErrorCode.API_ERROR, f"获取会话详情失败: {thread_id}", e
            ).with_context("threadId", thread_id).with_suggestion("请检查会话 ID 是否正确")

    def delete_thread(self, thread_id: str) -> None:
        """删除会话 / Delete thread"""
        self._validate_thread_id(thread_id)

        try:
            self._client.delete_thread(self.config.employee_name, thread_id)
        except Exception as e:
            if self._is_thread_not_found_error(e):
                raise SDKException.thread_not_found(thread_id)
            raise SDKException(
                ErrorCode.API_ERROR, f"删除会话失败: {thread_id}", e
            ).with_context("threadId", thread_id).with_suggestion("请检查会话 ID 是否正确")


    def get_thread_data(self, thread_id: str, limit: int = 50) -> List[ThreadMessage]:
        """获取会话消息 / Get thread messages"""
        self._validate_thread_id(thread_id)

        try:
            if limit <= 0:
                limit = 50
            if limit > 100:
                limit = 100

            request = cms_models.GetThreadDataRequest(max_results=limit)
            response = self._client.get_thread_data(
                self.config.employee_name, thread_id, request
            )

            if not response.body:
                raise SDKException(
                    ErrorCode.PARSE_ERROR, "无效响应: 响应体为空"
                ).with_context("threadId", thread_id).with_suggestion("请稍后重试")

            # Strategy: prefer system Result over assistant streaming messages
            # 策略：优先使用 system Result，而不是 assistant 流式消息
            message_map: Dict[str, Dict[str, str]] = {}
            message_order: List[str] = []
            
            # Check if any system Result exists
            # 检查是否存在 system Result
            has_system_result = False
            if response.body.data:
                for data in response.body.data:
                    if data.messages:
                        for msg in data.messages:
                            if msg.role == 'system':
                                artifacts = getattr(msg, 'artifacts', None)
                                if artifacts:
                                    for artifact in artifacts:
                                        if isinstance(artifact, dict) and artifact.get('name') == 'Result':
                                            has_system_result = True
                                            break
                            if has_system_result:
                                break
                    if has_system_result:
                        break

            # Process messages
            if response.body.data:
                for data in response.body.data:
                    if data.messages:
                        for msg in data.messages:
                            role = msg.role or ''
                            timestamp = msg.timestamp or ''
                            
                            # Skip assistant streaming messages if system Result exists
                            if role == 'assistant' and has_system_result:
                                continue
                            
                            # Use different key strategy based on role
                            if role == 'user':
                                key = f"user-{timestamp}"
                            elif role == 'system':
                                key = f"system-{timestamp}"
                            else:
                                call_id = getattr(msg, 'call_id', '') or ''
                                key = f"assistant-{call_id}"
                            
                            content = self._extract_message_content(msg)
                            if not content:
                                continue

                            if key in message_map:
                                message_map[key]['content'] += content
                            else:
                                # For system messages, display as assistant role
                                display_role = 'assistant' if role == 'system' else role
                                message_map[key] = {
                                    'role': display_role,
                                    'content': content,
                                    'timestamp': timestamp,
                                }
                                message_order.append(key)

            return [
                ThreadMessage(
                    role=message_map[key]['role'],
                    content=message_map[key]['content'],
                    timestamp=message_map[key]['timestamp'],
                )
                for key in message_order
            ]
        except SDKException:
            raise
        except Exception as e:
            if self._is_thread_not_found_error(e):
                raise SDKException.thread_not_found(thread_id)
            raise SDKException(
                ErrorCode.API_ERROR, f"获取会话消息失败: {thread_id}", e
            ).with_context("threadId", thread_id).with_suggestion("请检查会话 ID 是否正确")

    def _validate_thread_id(self, thread_id: str) -> None:
        if not thread_id:
            raise SDKException(
                ErrorCode.CONFIG_INVALID, "会话 ID 不能为空"
            ).with_context("threadId", thread_id).with_suggestion("请提供有效的会话 ID")

        if any(c in thread_id for c in " \t\n\r"):
            raise SDKException(
                ErrorCode.CONFIG_INVALID, f"会话 ID 包含非法字符: {thread_id}"
            ).with_context("threadId", thread_id).with_suggestion("会话 ID 不能包含空白字符")

    def _is_thread_not_found_error(self, e: Exception) -> bool:
        if not e:
            return False
        err_str = str(e)
        patterns = ["NotFound", "not found", "NOT_FOUND", "ThreadNotFound", 
                   "InvalidThreadId", "does not exist"]
        return any(p in err_str for p in patterns)

    def _extract_message_content(self, msg: Any) -> str:
        # 1. Try to extract from contents (streaming text chunks)
        # 尝试从 contents 提取（流式文本块）
        if msg and hasattr(msg, "contents") and msg.contents:
            result = []
            for content in msg.contents:
                if isinstance(content, dict):
                    if content.get("type") == "text" and content.get("value"):
                        result.append(content["value"])
            if result:
                return "".join(result)

        # 2. Try to extract from artifacts (final result)
        # 尝试从 artifacts 提取（最终结果）
        if msg and hasattr(msg, "artifacts") and msg.artifacts:
            for artifact in msg.artifacts:
                if isinstance(artifact, dict) and artifact.get("name") == "Result":
                    parts = artifact.get("parts", [])
                    text_parts = []
                    for part in parts:
                        if isinstance(part, dict) and part.get("kind") == "text" and part.get("text"):
                            text_parts.append(part["text"])
                    if text_parts:
                        return "".join(text_parts)

        return ""
