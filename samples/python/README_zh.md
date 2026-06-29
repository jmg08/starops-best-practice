# VibeOps STAROps SDK Python 示例

阿里云 STAROps SDK Python 语言示例程序。

## 快速开始

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入配置

# 2. 安装依赖
pip install -e .

# 3. 运行
python -m starops_sdk_samples.examples.chat
```

## 环境变量

| 变量 | 必需 | 说明 |
|-----|------|-----|
| VIBEOPS_ENDPOINT | ✅ | STAROps API 端点，格式: `starops.{region-id}.aliyuncs.com` |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ✅ | Access Key ID |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ✅ | Access Key Secret |
| VIBEOPS_EMPLOYEE_NAME | ❌ | 数字员工名称 (默认 default) |

## 示例程序

### chat - 交互式对话

```bash
python -m starops_sdk_samples.examples.chat
```

支持多轮对话，在同一会话中保持上下文。

### chat_from_file - 从文件加载请求

```bash
# 处理单个文件（默认显示每个事件的详细信息）
python -m starops_sdk_samples.examples.chat_from_file -file ../../requests/starops/entity.json

# 批量处理目录
python -m starops_sdk_samples.examples.chat_from_file -dir ../../requests/starops/

# 简洁模式（仅输出最终文本）
python -m starops_sdk_samples.examples.chat_from_file -file ../../requests/starops/entity.json -simple
```

默认使用 `EventPrinter` 打印每个 SSE 事件的详细信息（角色、内容、工具调用、Agent 调用、耗时等），`-simple` 模式使用 `SimplePrinter` 仅输出最终文本。

### chat_interactive - 交互事件处理

```bash
python -m starops_sdk_samples.examples.chat_interactive
```

处理 Agent 返回的确认、选择、输入等交互事件。

### thread_manager - 会话管理

```bash
# 列出会话
python -m starops_sdk_samples.examples.thread_manager list

# 查看详情
python -m starops_sdk_samples.examples.thread_manager get <thread-id>

# 删除会话
python -m starops_sdk_samples.examples.thread_manager delete <thread-id>
```

## 测试

```bash
pip install -e ".[dev]"
pytest
```

## 目录结构

```
samples/python/
├── starops_sdk_samples/
│   ├── client/        # 客户端实现
│   ├── types/         # 类型定义
│   ├── logger/        # 日志工具
│   └── examples/      # 示例程序
├── tests/             # 测试代码
├── pyproject.toml     # 项目配置
└── README.md
```

## 依赖要求

- Python 3.8+
- Alibaba Cloud STAROps SDK
