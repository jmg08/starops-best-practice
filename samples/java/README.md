# VibeOps CMS SDK Samples for Java

阿里云 CMS SDK Java 语言示例程序。

## 快速开始

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入配置

# 2. 编译
mvn clean compile

# 3. 运行
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.Chat"
```

## 环境变量

| 变量 | 必需 | 说明 |
|-----|------|-----|
| VIBEOPS_WORKSPACE | ✅ | 工作空间 ID |
| VIBEOPS_ENDPOINT | ✅ | API 端点 |
| ALIBABA_CLOUD_ACCESS_KEY_ID | ✅ | Access Key ID |
| ALIBABA_CLOUD_ACCESS_KEY_SECRET | ✅ | Access Key Secret |
| VIBEOPS_EMPLOYEE_NAME | ❌ | 数字员工名称 (默认 default) |

## 示例程序

### Chat - 交互式对话

```bash
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.Chat"
```

支持多轮对话，在同一会话中保持上下文。

### ChatFromFile - 从文件加载请求

```bash
# 处理单个文件
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatFromFile" \
    -Dexec.args="-file ../../requests/cms/entity.json"

# 批量处理目录
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatFromFile" \
    -Dexec.args="-dir ../../requests/cms/"
```

### ChatInteractive - 交互事件处理

```bash
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ChatInteractive"
```

处理 Agent 返回的确认、选择、输入等交互事件。

### ThreadManager - 会话管理

```bash
# 列出会话
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ThreadManager" \
    -Dexec.args="list"

# 查看详情
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ThreadManager" \
    -Dexec.args="get <thread-id>"

# 删除会话
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.cms.samples.examples.ThreadManager" \
    -Dexec.args="delete <thread-id>"
```

## 测试

```bash
mvn test
```

## 目录结构

```
samples/java/
├── src/
│   ├── main/java/com/alibaba/cloud/cms/samples/
│   │   ├── client/        # 客户端实现
│   │   ├── types/         # 类型定义
│   │   ├── logger/        # 日志工具
│   │   └── examples/      # 示例程序
│   └── test/              # 测试代码
├── pom.xml                # Maven 配置
└── README.md
```

## 依赖

- Java 11+
- Maven 3.6+
- Alibaba Cloud CMS SDK
