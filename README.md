# CMS SDK Samples | CMS SDK 示例

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-Planned-orange?style=flat&logo=openjdk)](samples/java/)
[![Python](https://img.shields.io/badge/Python-Planned-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-Planned-3178C6?style=flat&logo=typescript)](samples/typescript/)

Production-ready sample programs for interacting with Alibaba Cloud CMS (Cloud Monitor Service) digital employees across multiple programming languages.

为阿里云 CMS（云监控服务）数字员工提供的生产就绪示例程序，支持多种编程语言。

---

## Table of Contents | 目录

- [Overview | 概述](#overview--概述)
- [Language Support | 语言支持](#language-support--语言支持)
- [Quick Start | 快速开始](#quick-start--快速开始)
- [Project Structure | 项目结构](#project-structure--项目结构)
- [Features | 功能特性](#features--功能特性)
- [Documentation | 文档](#documentation--文档)
- [Contributing | 贡献指南](#contributing--贡献指南)
- [License | 许可证](#license--许可证)

---

## Overview | 概述

### English

The CMS SDK Samples project provides comprehensive, production-ready examples for interacting with Alibaba Cloud CMS digital employees. These samples demonstrate best practices for:

- **Multi-turn Conversations**: Maintain context across multiple conversation rounds
- **Interactive Event Handling**: Handle user confirmations, selections, and inputs
- **Thread Management**: Create, list, retrieve, and delete conversation threads
- **Timeout Control**: Configure and handle request timeouts
- **Structured Logging**: Debug and monitor SDK behavior with JSON logs
- **Response Structure Management**: Automatically discover and document response formats

### 中文

CMS SDK 示例项目为阿里云 CMS 数字员工提供全面的、生产就绪的交互示例。这些示例展示了以下最佳实践：

- **多轮对话**：在多轮对话中保持上下文
- **交互事件处理**：处理用户确认、选择和输入
- **会话管理**：创建、列出、获取和删除对话会话
- **超时控制**：配置和处理请求超时
- **结构化日志**：使用 JSON 日志调试和监控 SDK 行为
- **响应结构管理**：自动发现和记录响应格式

---

## Language Support | 语言支持

We are committed to providing SDK samples in multiple programming languages to serve developers across different technology stacks.

我们致力于提供多种编程语言的 SDK 示例，以服务于不同技术栈的开发者。

### Current Status | 当前状态

| Language | Status | Directory | Documentation |
|----------|--------|-----------|---------------|
| **Go** | ✅ Complete | [`samples/golang/`](samples/golang/) | [README](samples/golang/README.md) \| [中文](samples/golang/README_zh.md) |
| **Java** | 🔜 Planned (Q2 2025) | `samples/java/` | Coming Soon |
| **Python** | 🔜 Planned (Q3 2025) | `samples/python/` | Coming Soon |
| **TypeScript** | 🔜 Planned (Q4 2025) | `samples/typescript/` | Coming Soon |

| 语言 | 状态 | 目录 | 文档 |
|------|------|------|------|
| **Go** | ✅ 已完成 | [`samples/golang/`](samples/golang/) | [English](samples/golang/README.md) \| [中文](samples/golang/README_zh.md) |
| **Java** | 🔜 计划中 (2025 Q2) | `samples/java/` | 即将推出 |
| **Python** | 🔜 计划中 (2025 Q3) | `samples/python/` | 即将推出 |
| **TypeScript** | 🔜 计划中 (2025 Q4) | `samples/typescript/` | 即将推出 |

### Language Roadmap | 语言路线图

For detailed information about our multi-language support plan, including:
- Timeline and milestones
- Common interface specifications
- Implementation guidelines
- Quality standards

Please see the [Language Support Roadmap](docs/LANGUAGE_ROADMAP.md).

有关多语言支持计划的详细信息，包括：
- 时间线和里程碑
- 通用接口规范
- 实现指南
- 质量标准

请参阅[语言支持路线图](docs/LANGUAGE_ROADMAP.md)。

---

## Quick Start | 快速开始

### Prerequisites | 前置条件

- Alibaba Cloud account with CMS access | 具有 CMS 访问权限的阿里云账号
- Access Key ID and Secret | Access Key ID 和 Secret
- Workspace ID | 工作空间 ID

### Go (Available Now | 现已可用)

```bash
# Clone the repository | 克隆仓库
git clone <repository-url>
cd samples/golang

# Set environment variables | 设置环境变量
export VIBEOPS_WORKSPACE="your-workspace"
export VIBEOPS_ENDPOINT="cms.cn-hongkong.aliyuncs.com"
export ALIBABA_CLOUD_ACCESS_KEY_ID="your-access-key-id"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="your-access-key-secret"

# Run the interactive chat example | 运行交互式对话示例
go run ./cmd/chat/
```

For more Go examples, see the [Go README](samples/golang/README.md).

更多 Go 示例，请参阅 [Go README](samples/golang/README.md)。

### Java (Coming Q2 2025 | 2025 Q2 推出)

```bash
cd samples/java
mvn compile exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.Chat"
```

### Python (Coming Q3 2025 | 2025 Q3 推出)

```bash
cd samples/python
pip install -e .
python -m cms_sdk_samples.examples.chat
```

### TypeScript (Coming Q4 2025 | 2025 Q4 推出)

```bash
cd samples/typescript
npm install
npm run chat
```

---

## Project Structure | 项目结构

```
.
├── README.md                          # This file | 本文件
├── .env.example                       # Environment variables template | 环境变量模板
├── requests/                          # Shared request JSON files | 共享请求 JSON 文件
│   └── cms/                           # CMS-specific requests | CMS 特定请求
│       ├── entity.json                # Entity query example | 实体查询示例
│       ├── sls_query.json             # SLS query example | SLS 查询示例
│       └── text_to_sql.json           # Text-to-SQL example | Text-to-SQL 示例
│
└── samples/                           # Language-specific implementations | 语言特定实现
    ├── golang/                        # ✅ Go implementation (complete) | Go 实现（已完成）
    │   ├── cmd/                       # Sample programs | 示例程序
    │   ├── internal/                  # Internal packages | 内部包
    │   ├── types/                     # Type definitions | 类型定义
    │   ├── docs/                      # Documentation | 文档
    │   │   └── LANGUAGE_ROADMAP.md    # Multi-language roadmap | 多语言路线图
    │   ├── README.md                  # English documentation | 英文文档
    │   └── README_zh.md               # Chinese documentation | 中文文档
    │
    ├── java/                          # 🔜 Java implementation (planned) | Java 实现（计划中）
    ├── python/                        # 🔜 Python implementation (planned) | Python 实现（计划中）
    └── typescript/                    # 🔜 TypeScript implementation (planned) | TypeScript 实现（计划中）
```

---

## Features | 功能特性

All language implementations will include the following sample programs:

所有语言实现将包含以下示例程序：

| Sample | Description | 示例 | 描述 |
|--------|-------------|------|------|
| `chat` | Basic interactive chat | `chat` | 基础交互式对话 |
| `chat-simple` | Simple output mode (text only) | `chat-simple` | 简洁输出模式（仅文本） |
| `chat-timeout` | Timeout control | `chat-timeout` | 超时控制 |
| `chat-interactive` | Interactive event handling | `chat-interactive` | 交互事件处理 |
| `chat-multi-turn` | Multi-turn conversation | `chat-multi-turn` | 多轮对话 |
| `thread-manager` | Thread management CLI | `thread-manager` | 会话管理命令行工具 |
| `chat-from-file` | Load requests from JSON | `chat-from-file` | 从 JSON 文件加载请求 |
| `chat-with-time` | Time parameter example | `chat-with-time` | 时间参数示例 |

---

## Documentation | 文档

### Language-Specific Documentation | 语言特定文档

| Language | English | 中文 |
|----------|---------|------|
| Go | [README.md](samples/golang/README.md) | [README_zh.md](samples/golang/README_zh.md) |
| Java | Coming Soon | 即将推出 |
| Python | Coming Soon | 即将推出 |
| TypeScript | Coming Soon | 即将推出 |

### Additional Resources | 其他资源

- [Language Support Roadmap | 语言支持路线图](docs/LANGUAGE_ROADMAP.md)
- [Alibaba Cloud CMS Documentation | 阿里云 CMS 文档](https://help.aliyun.com/product/28572.html)

---

## Contributing | 贡献指南

We welcome contributions to extend language support and improve existing implementations!

我们欢迎贡献以扩展语言支持和改进现有实现！

### How to Contribute | 如何贡献

1. **Fork the repository** | 复刻仓库
2. **Create a feature branch** | 创建功能分支
3. **Follow the [interface specification](docs/LANGUAGE_ROADMAP.md#common-interface-specification--通用接口规范)** | 遵循[接口规范](docs/LANGUAGE_ROADMAP.md#common-interface-specification--通用接口规范)
4. **Write tests (≥80% coverage)** | 编写测试（≥80% 覆盖率）
5. **Update documentation in both languages** | 更新双语文档
6. **Submit a pull request** | 提交拉取请求

### Review Checklist | 审查清单

- [ ] Follows common interface specification | 遵循通用接口规范
- [ ] All sample programs implemented | 所有示例程序已实现
- [ ] Unit tests with ≥80% coverage | 单元测试覆盖率 ≥80%
- [ ] Documentation in both languages | 双语文档
- [ ] No breaking changes to existing APIs | 无破坏性 API 变更

---

## License | 许可证

This project is licensed under the Apache License 2.0.

本项目采用 Apache License 2.0 许可证。

---

*Last Updated: 2024-12 | 最后更新：2024年12月*
