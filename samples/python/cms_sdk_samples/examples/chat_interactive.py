#!/usr/bin/env python3
"""
交互事件处理示例
Interactive event handling example

Usage: python -m cms_sdk_samples.examples.chat_interactive
"""

import asyncio
import sys

from ..client import (
    AgentClient,
    Config,
    SDKException,
    SimplePrinter,
    InteractiveHandler,
)


PRESET_QUESTIONS = [
    ("1", "查询 SLS 日志（可能触发 logstore 选择）", "查询最近一小时的错误日志"),
    ("2", "执行危险操作（可能触发确认）", "删除所有过期的告警规则"),
    ("3", "查询指标（可能触发指标选择）", "查询 ECS 实例的 CPU 使用率"),
    ("4", "模糊查询（可能触发输入补充）", "查询服务的性能数据"),
    ("5", "多选项查询（可能触发选择）", "帮我分析系统问题"),
]


async def process_events_with_interaction(
    events,
    printer: SimplePrinter,
    handler: InteractiveHandler,
    thread_id: str,
):
    async for event in events:
        if event.has_error():
            print(f"❌ 错误: {event.error}")
            continue

        # Extract text
        text = printer.process_event(event)
        if text:
            print(text, end="", flush=True)

        # Check for interactive events
        if event.body and "messages" in event.body:
            for msg in event.body["messages"]:
                interactive_events = InteractiveHandler.extract_interactive_events(msg)
                for interactive_event in interactive_events:
                    print("\n" + "*" * 60)
                    print("🔔 检测到交互事件!")
                    print(f"   类型: {interactive_event.get('type')}")
                    print("*" * 60)

                    # Handle interactive event
                    try:
                        response = handler.handle_event(interactive_event)

                        print("\n✅ 交互响应:")
                        print(f"   ID: {response.interaction_id}")
                        print(f"   类型: {response.type}")
                        print(f"   响应: {response.response}")

                        # Resume chat
                        print("\n📤 恢复对话...")
                        resume_events = handler.resume_chat(thread_id, response)
                        await process_events_with_interaction(
                            resume_events, printer, handler, thread_id
                        )
                    except SDKException as e:
                        print(f"❌ 处理交互事件失败: {e}")


def print_help():
    print("\n📖 命令帮助:")
    print("  help  - 显示帮助")
    print("  list  - 显示预设问题列表")
    print("  1-5   - 使用预设问题（可能触发交互事件）")
    print("  quit  - 退出")
    print("  <msg> - 发送自定义消息")


def print_preset_questions():
    print("\n📋 预设问题（可能触发交互事件）:")
    print("-" * 60)
    for q_id, desc, question in PRESET_QUESTIONS:
        print(f"  [{q_id}] {desc}")
        print(f"      问题: {question}")
    print("-" * 60)


def print_config(cfg: Config):
    print("📋 配置:")
    print(f"  Workspace: {cfg.workspace}")
    print(f"  Employee: {cfg.employee_name}\n")


def print_env_help():
    print("\n请设置环境变量:")
    print("  VIBEOPS_WORKSPACE, VIBEOPS_ENDPOINT")
    print("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")


async def main_async():
    print("🚀 VibeOps Chat 交互事件处理示例 (Python)")
    print("=" * 60)

    try:
        # Load configuration
        cfg = Config.load_from_env()
        print_config(cfg)

        # Create client
        client = AgentClient(cfg)

        # Create thread
        print("📝 创建会话...")
        thread_id = client.create_thread()
        print(f"✅ 会话创建成功, ThreadID: {thread_id}\n")

        # Create printer and handler
        printer = SimplePrinter()
        handler = InteractiveHandler(client, timeout=60.0)

        # Print help
        print_help()

        # Interactive loop
        while True:
            try:
                user_input = input("\n👤 请输入 (help 查看帮助): ")
            except EOFError:
                print("\n👋 再见!")
                break

            user_input = user_input.strip()
            if not user_input:
                continue

            if user_input in ("quit", "exit"):
                print("👋 再见!")
                break
            elif user_input == "help":
                print_help()
                continue
            elif user_input == "list":
                print_preset_questions()
                continue

            # Check for preset question
            for q_id, _, question in PRESET_QUESTIONS:
                if user_input == q_id:
                    user_input = question
                    print(f"📝 使用预设问题: {user_input}")
                    break

            print("-" * 60)

            # Send message and process response
            printer.reset()
            events = client.chat(thread_id, user_input)
            await process_events_with_interaction(events, printer, handler, thread_id)

            # Output final text
            final_text = printer.get_final_text()
            if final_text:
                print(f"\n📄 回复:\n{final_text}")

            print("=" * 60)

    except SDKException as e:
        print(f"❌ 配置加载失败: {e}")
        print_env_help()
        sys.exit(1)
    except Exception as e:
        print(f"❌ 错误: {e}")
        sys.exit(1)


def main():
    asyncio.run(main_async())


if __name__ == "__main__":
    main()
