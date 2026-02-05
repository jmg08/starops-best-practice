# VibeOps Chat 示例程序

这是一组对话示例程序，演示如何使用阿里云 CMS SDK 与 VibeOps 数字员工进行对话。

## 目录结构

```
samples/golang/
├── cmd/
│   ├── chat/                  # 交互式对话（基础示例）
│   ├── chat-from-file/        # 从 JSON 文件加载请求
│   └── chat-with-time/        # 带时间参数的示例
├── internal/
│   └── client/                # 公共客户端代码
├── types/
│   ├── events.go              # 事件类型定义（响应解析）
│   └── inputs.go              # 输入参数类型定义（请求构建）
├── go.mod
└── go.sum
```

## 环境变量

| 变量名 | 必需 | 说明 | 示例 |
|--------|------|------|------|
| `VIBEOPS_WORKSPACE` | ✅ | 工作空间 ID | `rca-benchmark` |
| `VIBEOPS_ENDPOINT` | ✅ | API 端点地址 | `cms.cn-hongkong.aliyuncs.com` |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅ | 阿里云 Access Key ID | |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅ | 阿里云 Access Key Secret | |
| `VIBEOPS_REGION` | ✅ | 区域（必须与 Endpoint 对应） | `cn-hongkong` |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | 数字员工名称（默认: default） | `apsara-ops` |

## 示例程序

### 1. 交互式对话 (cmd/chat)

基础的交互式对话示例：

```bash
cd samples/golang
go run ./cmd/chat/
```

### 2. 从文件加载请求 (cmd/chat-from-file)

从 JSON 文件加载请求，适用于复杂的请求场景：

```bash
go run ./cmd/chat-from-file/ -file ../../requests/cms/entity.json
```

### 3. 带时间参数 (cmd/chat-with-time)

演示如何构建包含时间范围的 `userContext`：

```bash
# 使用默认时间范围（最近15分钟）
go run ./cmd/chat-with-time/ -message "最近有什么异常吗"

# 指定时间范围
go run ./cmd/chat-with-time/ -from 1770274812 -to 1770275712 -message "这段时间有问题吗"
```

## 输入参数类型 (types/inputs.go)

### UserInputParams

用户输入参数的完整定义：

```go
type UserInputParams struct {
    RegionID    string `json:"region,omitempty"`
    Workspace   string `json:"workspace,omitempty"`
    Language    string `json:"language,omitempty"`
    TimeZone    string `json:"timeZone,omitempty"`
    TimeStamp   string `json:"timeStamp,omitempty"`
    UserContext string `json:"userContext,omitempty"` // []UserContext 的 JSON 字符串
    Config      string `json:"config,omitempty"`
}
```

### UserContext 上下文类型

| 类型 | 常量 | 说明 |
|------|------|------|
| `metadata` | `UserContextTypeMetadata` | 元数据上下文，包含时间范围 |
| `entity` | `UserContextTypeEntity` | 实体上下文，包含实体信息 |
| `sql_generation` | `UserContextTypeSQLGenerated` | SQL 生成上下文 |
| `spl_generation` | `UserContextTypeSPLGenerated` | SPL 生成上下文 |

### 构建 userContext 示例

```go
import "github.com/vibeops/samples/golang/types"

// 构建时间范围上下文
contexts := []types.UserContext{
    {
        Type: types.UserContextTypeMetadata,
        Data: types.MetadataUserData{
            FromTime: 1770274812,
            ToTime:   1770275712,
        },
    },
}

// 序列化为 JSON 字符串
userContextJSON, _ := json.Marshal(contexts)

// 在 variables 中使用
variables := map[string]interface{}{
    "workspace":   "your-workspace",
    "region":      "cn-hongkong",
    "userContext": string(userContextJSON),
}
```

## 请求文件格式

参考 `requests/cms/entity.json`：

```json
{
    "region": "cn-hongkong",
    "digitalEmployeeName": "apsara-ops",
    "action": "create",
    "messages": [
        {
            "role": "user",
            "contents": [{"type": "text", "value": "cart 实体延迟多少"}]
        }
    ],
    "variables": {
        "workspace": "rca-benchmark",
        "region": "cn-hongkong",
        "userContext": "[{\"type\":\"metadata\",\"data\":{...}}]"
    }
}
```

## SDK 依赖

- `github.com/alibabacloud-go/cms-20240330/v6` - 阿里云 CMS SDK
