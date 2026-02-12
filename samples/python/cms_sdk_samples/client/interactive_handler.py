"""
Interactive handler for CMS SDK
CMS SDK 交互处理器
"""

import sys
import time
from dataclasses import dataclass
from typing import Any, AsyncIterator, Dict, List, Optional, TextIO

from ..types import EventType, InteractionType
from .agent_client import AgentClient, ChatEvent
from .errors import SDKException, ErrorCode


@dataclass
class InteractiveResponse:
    """交互响应 / Interactive response"""
    interaction_id: str
    type: InteractionType
    response: Dict[str, Any]


class InteractiveHandler:
    """交互事件处理器 / Interactive event handler"""

    def __init__(
        self,
        client: AgentClient,
        timeout: Optional[float] = None,
        reader: Optional[TextIO] = None,
        writer: Optional[TextIO] = None,
    ):
        self.client = client
        self.timeout = timeout
        self.reader = reader or sys.stdin
        self.writer = writer or sys.stdout

    def handle_event(self, event: Dict[str, Any]) -> InteractiveResponse:
        """处理交互事件 / Handle interactive event"""
        if not event:
            raise SDKException(ErrorCode.PARSE_ERROR, "事件为空")

        event_type = event.get("type")
        if event_type != EventType.INTERACTIVE.value:
            raise SDKException(ErrorCode.PARSE_ERROR, f"不支持的事件类型: {event_type}")

        payload = event.get("payload")
        if not payload:
            raise SDKException(ErrorCode.PARSE_ERROR, "交互负载为空")

        interactive_type = payload.get("type")
        
        if interactive_type == InteractionType.USER_ACK.value:
            return self.handle_user_ack(payload)
        elif interactive_type == InteractionType.USER_SELECT.value:
            return self.handle_user_select(payload)
        elif interactive_type == InteractionType.USER_INPUT.value:
            return self.handle_user_input(payload)
        else:
            raise SDKException(ErrorCode.PARSE_ERROR, f"不支持的交互类型: {interactive_type}")

    def handle_user_ack(self, payload: Dict[str, Any]) -> InteractiveResponse:
        """处理用户确认 / Handle user acknowledgment"""
        interaction_id = self._get_interaction_id(payload)
        title = self._get_meta_field(payload, "title")
        description = self._get_meta_field(payload, "description")

        self._print("\n🔔 确认请求")
        if title:
            self._print(f"   标题: {title}")
        if description:
            self._print(f"   描述: {description}")
        self._print("   请输入 [y/yes] 确认，[n/no] 取消: ", end="")

        user_input = self._read_input().strip().lower()
        confirmed = user_input in ("", "y", "yes", "是")

        return InteractiveResponse(
            interaction_id=interaction_id,
            type=InteractionType.USER_ACK,
            response={"confirmed": confirmed},
        )

    def handle_user_select(self, payload: Dict[str, Any]) -> InteractiveResponse:
        """处理用户选择 / Handle user selection"""
        interaction_id = self._get_interaction_id(payload)
        title = self._get_meta_field(payload, "title")

        self._print("\n📋 请选择")
        if title:
            self._print(f"   标题: {title}")

        options = self._get_options(payload)
        if not options:
            raise SDKException(ErrorCode.PARSE_ERROR, "没有可选项")

        self._print("   选项:")
        for i, opt in enumerate(options):
            label = self._get_option_label(opt, i)
            self._print(f"   [{i + 1}] {label}")
        self._print(f"   请输入选项编号 (1-{len(options)}): ", end="")

        user_input = self._read_input().strip()
        try:
            selected_index = int(user_input)
        except ValueError:
            raise SDKException(
                ErrorCode.PARSE_ERROR,
                f"无效的选择: {user_input}，请输入 1-{len(options)} 之间的数字"
            )

        if selected_index < 1 or selected_index > len(options):
            raise SDKException(
                ErrorCode.PARSE_ERROR,
                f"无效的选择: {user_input}，请输入 1-{len(options)} 之间的数字"
            )

        return InteractiveResponse(
            interaction_id=interaction_id,
            type=InteractionType.USER_SELECT,
            response={
                "selectedIndex": selected_index - 1,
                "selectedValue": options[selected_index - 1],
            },
        )

    def handle_user_input(self, payload: Dict[str, Any]) -> InteractiveResponse:
        """处理用户输入 / Handle user input"""
        interaction_id = self._get_interaction_id(payload)
        title = self._get_meta_field(payload, "title")
        description = self._get_meta_field(payload, "description")
        placeholder = self._get_meta_field(payload, "placeholder")

        self._print("\n✏️  请输入")
        if title:
            self._print(f"   标题: {title}")
        if description:
            self._print(f"   描述: {description}")
        if placeholder:
            self._print(f"   提示: {placeholder}")
        self._print("   请输入内容: ", end="")

        user_input = self._read_input().strip()

        return InteractiveResponse(
            interaction_id=interaction_id,
            type=InteractionType.USER_INPUT,
            response={"value": user_input},
        )

    async def resume_chat(
        self, thread_id: str, response: InteractiveResponse
    ) -> AsyncIterator[ChatEvent]:
        """使用交互响应恢复对话 / Resume chat with interactive response"""
        if not self.client:
            yield ChatEvent.from_error(
                SDKException(ErrorCode.CLIENT_CREATE, "客户端未初始化")
            )
            return

        if not response:
            yield ChatEvent.from_error(
                SDKException(ErrorCode.PARSE_ERROR, "交互响应为空")
            )
            return

        variables = {
            "workspace": self.client.config.workspace,
            "region": self.client.config.region,
            "language": "zh",
            "timeZone": "Asia/Shanghai",
            "timeStamp": str(int(time.time())),
            "interactionId": response.interaction_id,
            "interactionType": response.type.value,
            "interactionResult": response.response,
        }

        import json
        message = f"[交互响应] {json.dumps(response.__dict__, default=str)}"

        async for event in self.client.chat_with_variables(thread_id, message, variables):
            yield event

    def _print(self, text: str, end: str = "\n") -> None:
        self.writer.write(text + end)
        self.writer.flush()

    def _read_input(self) -> str:
        # Note: timeout handling would require threading in sync context
        return self.reader.readline()

    def _get_interaction_id(self, payload: Dict[str, Any]) -> str:
        meta = payload.get("meta", {})
        if meta.get("id"):
            return meta["id"]
        if meta.get("interactionId"):
            return meta["interactionId"]
        return f"interaction_{int(time.time() * 1000000)}"

    def _get_meta_field(self, payload: Dict[str, Any], field: str) -> Optional[str]:
        meta = payload.get("meta", {})
        return meta.get(field)

    def _get_options(self, payload: Dict[str, Any]) -> List[Dict[str, Any]]:
        # Try data field first
        data = payload.get("data", [])
        if data:
            return data

        # Try meta.options
        meta = payload.get("meta", {})
        options = meta.get("options", [])
        return options

    def _get_option_label(self, option: Dict[str, Any], index: int) -> str:
        for field in ["label", "name", "title", "value"]:
            value = option.get(field)
            if isinstance(value, str) and value:
                return value
        return f"选项 {index + 1}"

    @staticmethod
    def is_interactive_event(event: Dict[str, Any]) -> bool:
        """检查事件是否为交互事件 / Check if event is interactive"""
        if not event:
            return False
        return event.get("type") == EventType.INTERACTIVE.value

    @staticmethod
    def extract_interactive_events(message: Dict[str, Any]) -> List[Dict[str, Any]]:
        """从消息中提取交互事件 / Extract interactive events from message"""
        events = []
        if not message:
            return events

        msg_events = message.get("events", [])
        for event in msg_events:
            if InteractiveHandler.is_interactive_event(event):
                events.append(event)
        return events
