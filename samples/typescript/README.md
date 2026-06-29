# VibeOps STAROps SDK Samples for TypeScript

Alibaba Cloud STAROps SDK samples for TypeScript.

## Quick Start

```bash
# 1. Configure environment variables
cp .env.example .env
# Edit .env with your configuration

# 2. Install dependencies
npm install

# 3. Run
npm run chat
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| VIBEOPS_ENDPOINT | ✅ | STAROps API endpoint, format: `starops.{region-id}.aliyuncs.com` |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ✅ | Access Key ID |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ✅ | Access Key Secret |
| VIBEOPS_EMPLOYEE_NAME | ❌ | Digital employee name (default: default) |

## Sample Programs

### chat - Interactive Chat

```bash
npm run chat
```

Supports multi-turn conversation with context preservation.

### chat-from-file - Load Requests from File

```bash
# Process single file (default: shows detailed event information)
npm run chat-from-file -- -file ../../requests/starops/entity.json

# Batch process directory
npm run chat-from-file -- -dir ../../requests/starops/

# Simple mode (text output only)
npm run chat-from-file -- -file ../../requests/starops/entity.json -simple
```

By default uses `EventPrinter` to display detailed SSE event information (role, content, tool calls, agent calls, duration, etc.). Use `-simple` to switch to `SimplePrinter` for text-only output.

### chat-interactive - Interactive Event Handling

```bash
npm run chat-interactive
```

Handles confirmation, selection, and input events from the Agent.

### thread-manager - Thread Management

```bash
# List threads
npm run thread-manager -- list

# Get thread details
npm run thread-manager -- get <thread-id>

# Delete thread
npm run thread-manager -- delete <thread-id>
```

## Testing

```bash
npm test
```

## Building

```bash
npm run build
```

## Directory Structure

```
samples/typescript/
├── src/
│   ├── client/        # Client implementation
│   ├── types/         # Type definitions
│   └── examples/      # Sample programs
├── tests/             # Test code
├── package.json       # npm configuration
├── tsconfig.json      # TypeScript configuration
└── README.md
```

## Requirements

- Node.js 18+
- TypeScript 5.0+
- Alibaba Cloud STAROps SDK
