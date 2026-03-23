#!/usr/bin/env python3
"""
交互式对话示例
Interactive chat example

Usage: python -m cms_sdk_samples.examples.chat
"""

import asyncio
import sys

from ..client import AgentClient, Config, SDKException, SimplePrinter


async def main_async():
    print("🚀 VibeOps Chat (Python)")
    print("=" * 60)

    try:
        # Load configuration
        cfg = Config.load_from_env()
        print(f"📋 Employee: {cfg.employee_name}\n")

        # Create client
        client = AgentClient(cfg)

        # Create thread
        print("📝 创建会话...")
        thread_id = client.create_thread()
        print(f"✅ ThreadID: {thread_id}\n")

        # Create printer
        printer = SimplePrinter()

        # Interactive loop
        while True:
            try:
                user_input = input("👤 请输入 (quit 退出): ")
            except EOFError:
                print("\n👋 再见!")
                break

            user_input = user_input.strip()
            if not user_input:
                continue
            if user_input in ("quit", "exit"):
                print("👋 再见!")
                break

            print("-" * 60)

            # Send message
            printer.reset()
            async for event in client.chat(thread_id, user_input):
                if event.has_error():
                    print(f"❌ 错误: {event.error}")
                    continue

                text = printer.process_event(event)
                if text:
                    print(text, end="", flush=True)

            print()
            print("=" * 60)
            print()

    except SDKException as e:
        print(f"❌ 配置加载失败: {e}")
        print("\n请设置环境变量:")
        print("  VIBEOPS_ENDPOINT")
        print("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")
        sys.exit(1)
    except Exception as e:
        print(f"❌ 错误: {e}")
        sys.exit(1)


def main():
    asyncio.run(main_async())


if __name__ == "__main__":
    main()
