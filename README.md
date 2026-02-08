# goPerf

> 轻量级 HTTP/SSE 性能压测工具，单二进制、无外部依赖，支持周期报告与汇总统计。

## 特性

- **多协议**：HTTP 通用请求 + SSE 流式（首 token 延迟、token/s 等）
- **可配置**：`.properties` 定义请求，命令行控制并发与运行时长
- **指标**：TPS、平均/最小/最大延迟、成功数，支持周期打印与最终汇总
- **优雅退出**：`-t` 定时结束或 Ctrl+C，统一收尾并输出 SUMMARY

## 前置要求

- Go 1.16+

## 安装

```bash
git clone https://github.com/your-org/goPerf.git
cd goPerf
go build -o goPerf ./cmd/
```

## 快速开始

```bash
# 使用默认配置 client.properties，1 并发，Ctrl+C 结束
./goPerf

# 4 并发，运行 60 秒，每 5 秒打印一次报告
./goPerf -n 4 -t 60 -rt 5

# 指定配置文件
./goPerf -f my.properties -n 8
```

## 端到端测试

### 一键自动化（推荐）

在项目根目录执行，会自动编译、启动测试桩、跑压测并清理：

```bash
./scripts/run_e2e.sh
```

### 手动两步

**终端 1** — 启动测试服务（默认 `http://127.0.0.1:8080`）：

```bash
go run ./cmd/testserver/
# 指定端口: go run ./cmd/testserver/ -port 9090
```

**终端 2** — 编译并压测（使用 `e2e.properties`，2 并发、跑 5 秒、每 1 秒打报告）：

```bash
go build -o goPerf ./cmd/
./goPerf -f e2e.properties -n 2 -t 5 -rt 1
```

预期：周期输出 TPS/延迟，结束时打印 SUMMARY（total、success、averTps、averDelay 等）。

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-f` | `client.properties` | 请求配置文件路径（.properties） |
| `-n` | `1` | 并发 worker 数 |
| `-t` | `0` | 运行时长（秒），`0` 表示持续运行直至 Ctrl+C |
| `-rt` | `1` | 报告打印周期（秒） |
| `-c` | 当前 CPU 核数 | 使用的 CPU 核数（GOMAXPROCS） |
| `-d` | `false` | 调试模式 |

## 请求配置（.properties）

在配置文件中定义被测请求，例如：

```properties
method=GET
url=https://127.0.0.1:8448/hello
protocol=http
connTimeout=10
readTimeout=60
writeTimeout=60
keepalive=true
body={"key":"value"}
# 请求头：header.名称=值
header.Accept=application/json
```

| 配置项 | 说明 |
|--------|------|
| `url` | 请求 URL（需以 http/https 开头） |
| `method` | HTTP 方法：GET / POST / PUT / DELETE |
| `protocol` | `http` 或 `sse` |
| `body` | 请求体（如 POST 报文） |
| `connTimeout` / `readTimeout` / `writeTimeout` | 超时时间（秒） |
| `keepalive` | 是否使用长连接 |
| `header.<name>` | 自定义请求头 |

## 项目结构

```
goPerf/
├── cmd/
│   ├── main.go       # goPerf 入口
│   └── testserver/   # 测试用 HTTP 服务桩（端到端）
├── scripts/
│   └── run_e2e.sh    # 一键端到端测试脚本
├── e2e.properties    # 端到端测试配置
├── internal/
│   ├── app/          # 压测编排与指标汇总
│   ├── client/       # HTTP/SSE 客户端
│   ├── config/       # 配置加载
│   └── model/        # 请求/统计/校验模型
├── pkg/utils/        # 工具函数
├── client.properties # 示例配置
└── docs/             # 设计文档
```

## 文档

- [功能设计文档](docs/FEATURE_DESIGN.md) — 架构、数据流、功能规格与扩展建议
- [测试步骤文档](docs/TEST_STEPS.md) — 编译、端到端及基本功能验证的详细步骤

## License

见项目根目录 LICENSE 文件（如有）。
