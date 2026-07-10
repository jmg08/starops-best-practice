# STAROps SDK Samples for Python 🐍

Python client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming with
automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Python 3.8+**
- An Alibaba Cloud account with STAROps access
- Credentials configured with Alibaba Cloud CLI, or AK/SK environment variables

## Installation

```bash
cd python
python -m venv .venv && source .venv/bin/activate
pip install -e .
cp .env.example .env   # edit .env with STAROps settings, not credentials
```

## Quick Start

```bash
python -m starops_sdk_samples.examples.chat
```

## Running the Samples

### chat — interactive chat

```bash
python -m starops_sdk_samples.examples.chat
```

Multi-turn interactive chat with context preserved within a thread.

### chat_from_file — run requests from JSON

```bash
# Single request (detailed event output by default)
python -m starops_sdk_samples.examples.chat_from_file -file ../sample-requests/entity.json

# Batch-process a directory
python -m starops_sdk_samples.examples.chat_from_file -dir ../sample-requests/

# Simple mode: text-only output
python -m starops_sdk_samples.examples.chat_from_file -file ../sample-requests/entity.json -simple
```

By default the detailed `EventPrinter` shows role, content, tool calls, agent calls, and durations.
Use `-simple` to switch to text-only output.

### thread_manager — manage threads

```bash
python -m starops_sdk_samples.examples.thread_manager list                # list threads
python -m starops_sdk_samples.examples.thread_manager get <thread-id>     # thread details
python -m starops_sdk_samples.examples.thread_manager delete <thread-id>  # delete a thread
```

## SSE Retry & Reconnection

The client streams responses over SSE and recovers transparently from interruptions.

```
create ──► stream events ──► stream_done ✅ (normal completion)
              │
              ├─ connection dropped ─┐
              ├─ idle timeout ───────┤──► backoff ──► reconnect (action="reconnect") ──► dedupe ──► resume
              └─ SSE error ──────────┘
```

- **Normal completion** is marked by a `stream_done` event; a stream ending before it triggers a reconnect.
- **Exponential backoff**: `1s, 2s, 4s, 8s, 16s, 30s` (capped at 30s), up to `STAROPS_MAX_RETRIES` attempts.
- **Reconnect** sends `action="reconnect"` to resume the session.
- **Deduplication**: after reconnecting, messages are filtered by timestamp so none are delivered twice.

> [!NOTE]
> After exceeding the maximum retries, the client raises an error instead of hanging.

### Testing the retry logic

```bash
python -m starops_sdk_samples.examples.chat -simulate-error
python -m starops_sdk_samples.examples.chat_from_file -file ../sample-requests/entity.json -simulate-error
```

With `-simulate-error`, the client simulates a network disconnection, backs off, reconnects,
deduplicates by timestamp, and finishes at `stream_done`.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STAROPS_ENDPOINT` | ✅ | — | STAROps API endpoint, e.g. `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | ✅ | — | Workspace ID |
| `STAROPS_REGION` | ❌ | `cn-hangzhou` | Region (should match the endpoint) |
| `STAROPS_EMPLOYEE_NAME` | ❌ | `apsara-ops` | Digital employee name |
| `STAROPS_MAX_RETRIES` | ❌ | `10` | Max SSE reconnect attempts |
| `STAROPS_IDLE_TIMEOUT` | ❌ | `60` | Idle timeout (seconds); reconnect if no message arrives within this window |

Credential configuration:

1. **Recommended**: configure credentials with Alibaba Cloud CLI.
2. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview).
3. If you do not want to install the CLI, configure AK/SK with environment variables:

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

## Command-Line Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `-simulate-error` | `chat`, `chat_from_file` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `chat_from_file` | Load a single request from a JSON file |
| `-dir <path>` | `chat_from_file` | Batch-process every JSON request in a directory |
| `-simple` | `chat_from_file` | Text-only output |

## Testing

```bash
pip install -e ".[dev]"
pytest
```

## Project Structure

```
python/
├── starops_sdk_samples/
│   ├── client/        # Core client: chat, threads, retry, printers
│   ├── types/         # Type definitions
│   ├── logger/        # Logging utilities
│   └── examples/      # chat, chat_from_file, thread_manager
├── tests/             # Test suite
├── pyproject.toml
└── README.md
```
