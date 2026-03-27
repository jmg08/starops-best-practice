# VibeOps CMS SDK Samples for Python

Alibaba Cloud CMS SDK samples for Python.

## Quick Start

```bash
# 1. Configure environment variables
cp .env.example .env
# Edit .env with your configuration

# 2. Install dependencies
pip install -e .

# 3. Run
python -m cms_sdk_samples.examples.chat
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| VIBEOPS_ENDPOINT | ✅ | CMS API endpoint, format: `cms.{region-id}.aliyuncs.com` |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ✅ | Access Key ID |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ✅ | Access Key Secret |
| VIBEOPS_EMPLOYEE_NAME | ❌ | Digital employee name (default: default) |

## Sample Programs

### chat - Interactive Chat

```bash
python -m cms_sdk_samples.examples.chat
```

Supports multi-turn conversation with context preservation.

### chat_from_file - Load Requests from File

```bash
# Process single file (default: shows detailed event information)
python -m cms_sdk_samples.examples.chat_from_file -file ../../requests/cms/entity.json

# Batch process directory
python -m cms_sdk_samples.examples.chat_from_file -dir ../../requests/cms/

# Simple mode (text output only)
python -m cms_sdk_samples.examples.chat_from_file -file ../../requests/cms/entity.json -simple
```

By default uses `EventPrinter` to display detailed SSE event information (role, content, tool calls, agent calls, duration, etc.). Use `-simple` to switch to `SimplePrinter` for text-only output.

### chat_interactive - Interactive Event Handling

```bash
python -m cms_sdk_samples.examples.chat_interactive
```

Handles confirmation, selection, and input events from the Agent.

### thread_manager - Thread Management

```bash
# List threads
python -m cms_sdk_samples.examples.thread_manager list

# Get thread details
python -m cms_sdk_samples.examples.thread_manager get <thread-id>

# Delete thread
python -m cms_sdk_samples.examples.thread_manager delete <thread-id>
```

## Testing

```bash
pip install -e ".[dev]"
pytest
```

## Directory Structure

```
samples/python/
├── cms_sdk_samples/
│   ├── client/        # Client implementation
│   ├── types/         # Type definitions
│   ├── logger/        # Logging utilities
│   └── examples/      # Sample programs
├── tests/             # Test code
├── pyproject.toml     # Project configuration
└── README.md
```

## Requirements

- Python 3.8+
- Alibaba Cloud CMS SDK
