#!/usr/bin/env python3
"""
从文件加载请求示例
Load requests from file example

Usage:
    python -m cms_sdk_samples.examples.chat_from_file -file ../../requests/cms/entity.json
    python -m cms_sdk_samples.examples.chat_from_file -dir ../../requests/cms/
"""

import argparse
import asyncio
import json
import os
import sys
import time
from datetime import datetime
from pathlib import Path

from ..client import AgentClient, Config, SDKException, SimplePrinter, EventPrinter


def parse_args():
    parser = argparse.ArgumentParser(description="从文件加载请求")
    parser.add_argument("-file", dest="file_path", help="请求 JSON 文件路径")
    parser.add_argument("-dir", dest="dir_path", help="请求文件目录")
    parser.add_argument("-simple", action="store_true", help="简洁模式")
    parser.add_argument("-output", default="../../requests/output", help="输出目录")
    return parser.parse_args()


async def process_file(client: AgentClient, file_path: str, output_dir: str, simple_mode: bool):
    try:
        # Load request file
        with open(file_path, "r", encoding="utf-8") as f:
            req_file = json.load(f)

        # Extract message
        message = extract_message(req_file)
        if not message:
            print("⚠️ 文件中没有消息内容")
            return

        file_name = Path(file_path).name
        print(f"📄 文件: {file_name}")
        print(f"💬 消息: {truncate(message, 60)}")

        # Create thread
        thread_id = client.create_thread()

        # Create output file
        output_file = create_output_file(file_path, output_dir)

        # Write request info
        write_output(output_file, f"# Request: {file_name}")
        write_output(output_file, f"# Time: {datetime.now().isoformat()}")
        write_output(output_file, f"# ThreadID: {thread_id}")
        write_output(output_file, f"# Message: {message}\n")

        print("-" * 60)

        # Extract variables
        variables = req_file.get("variables", {})

        # Send request
        start_time = time.time()
        simple_printer = SimplePrinter() if simple_mode else None
        event_printer = None if simple_mode else EventPrinter(print_raw_body=False, print_parsed=True)
        event_index = 0

        async for event in client.chat_with_variables(thread_id, message, variables):
            event_index += 1

            if event.has_error():
                print(f"❌ 错误: {event.error}")
                write_output(output_file, f"[ERROR] {event.error}")
                break

            # Write raw event
            if event.raw_json:
                write_output(output_file, f"[EVENT {event_index}]\n{event.raw_json}\n")

            # Output
            if simple_mode:
                text = simple_printer.process_event(event)
                if text:
                    print(text, end="", flush=True)
            else:
                event_printer.print_event(event, event_index)

        elapsed = time.time() - start_time
        print()

        # Write final result
        if simple_mode and simple_printer:
            final_text = simple_printer.get_final_text()
            write_output(output_file, f"\n# Final Result:\n{final_text}")
            print(f"📄 最终文本:\n{final_text}")

        write_output(output_file, f"\n# Duration: {elapsed:.2f}s")
        print(f"⏱️  耗时: {elapsed:.2f}s")

        if output_file:
            output_file.close()
            print(f"📁 输出: {output_file.name}")

    except Exception as e:
        print(f"❌ 处理文件失败: {e}")


async def process_directory(client: AgentClient, dir_path: str, output_dir: str, simple_mode: bool):
    files = list(Path(dir_path).glob("*.json"))
    if not files:
        print(f"⚠️ 目录中没有 JSON 文件: {dir_path}")
        return

    print(f"📁 找到 {len(files)} 个请求文件\n")

    for i, file_path in enumerate(files):
        print(f"━━━ [{i + 1}/{len(files)}] {file_path.name} ━━━")
        await process_file(client, str(file_path), output_dir, simple_mode)
        print()

    print(f"✅ 处理完成，共 {len(files)} 个文件")


def extract_message(req_file: dict) -> str:
    messages = req_file.get("messages", [])
    if messages and messages[0].get("contents"):
        contents = messages[0]["contents"]
        if contents and contents[0].get("value"):
            return contents[0]["value"]
    return ""


def create_output_file(input_file: str, output_dir: str):
    try:
        os.makedirs(output_dir, exist_ok=True)
        base_name = Path(input_file).stem
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        output_path = os.path.join(output_dir, f"{base_name}-{timestamp}.log")
        return open(output_path, "w", encoding="utf-8")
    except Exception:
        return None


def write_output(f, content: str):
    if f:
        f.write(content + "\n")


def truncate(s: str, max_len: int) -> str:
    if len(s) <= max_len:
        return s
    return s[:max_len] + "..."


async def main_async():
    args = parse_args()

    if not args.file_path and not args.dir_path:
        print("用法:")
        print("  -file <path>   处理单个文件")
        print("  -dir <path>    处理目录下所有 JSON 文件")
        sys.exit(1)

    print("🚀 VibeOps Chat - 从文件加载请求 (Python)")
    print("=" * 60)

    try:
        # Load configuration
        cfg = Config.load_from_env()

        # Create client
        client = AgentClient(cfg)

        # Ensure output directory exists
        os.makedirs(args.output, exist_ok=True)

        # Process files
        if args.dir_path:
            await process_directory(client, args.dir_path, args.output, args.simple)
        else:
            await process_file(client, args.file_path, args.output, args.simple)

    except SDKException as e:
        print(f"❌ 配置加载失败: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"❌ 错误: {e}")
        sys.exit(1)


def main():
    asyncio.run(main_async())


if __name__ == "__main__":
    main()
