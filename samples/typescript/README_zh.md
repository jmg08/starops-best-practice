# VibeOps CMS SDK TypeScript 示例

阿里云 CMS SDK TypeScript 语言示例程序。

## 快速开始

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入配置

# 2. 安装依赖
npm install

# 3. 运行
npm run chat
```

## 环境变量

| 变量 | 必需 | 说明 |
|-----|------|-----|
| VIBEOPS_ENDPOINT | ✅ | StarOps API 端点，格式: `starops.{region-id}.aliyuncs.com` |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ✅ | Access Key ID |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ✅ | Access Key Secret |
| VIBEOPS_EMPLOYEE_NAME | ❌ | 数字员工名称 (默认 default) |

## 示例程序

### chat - 交互式对话

```bash
npm run chat
```

支持多轮对话，在同一会话中保持上下文。

### chat-from-file - 从文件加载请求

```bash
# 处理单个文件（默认显示每个事件的详细信息）
npm run chat-from-file -- -file ../../requests/cms/entity.json

# 批量处理目录
npm run chat-from-file -- -dir ../../requests/cms/

# 简洁模式（仅输出最终文本）
npm run chat-from-file -- -file ../../requests/cms/entity.json -simple
```

默认使用 `EventPrinter` 打印每个 SSE 事件的详细信息（角色、内容、工具调用、Agent 调用、耗时等），`-simple` 模式使用 `SimplePrinter` 仅输出最终文本。

### chat-interactive - 交互事件处理

```bash
npm run chat-interactive
```

处理 Agent 返回的确认、选择、输入等交互事件。

### thread-manager - 会话管理

```bash
# 列出会话
npm run thread-manager -- list

# 查看详情
npm run thread-manager -- get <thread-id>

# 删除会话
npm run thread-manager -- delete <thread-id>
```

## 测试

```bash
npm test
```

## 构建

```bash
npm run build
```

## 目录结构

```
samples/typescript/
├── src/
│   ├── client/        # 客户端实现
│   ├── types/         # 类型定义
│   └── examples/      # 示例程序
├── tests/             # 测试代码
├── package.json       # npm 配置
├── tsconfig.json      # TypeScript 配置
└── README.md
```

## 依赖要求

- Node.js 18+
- TypeScript 5.0+
- Alibaba Cloud CMS SDK
