# goPerf 功能设计文档

## 1. 项目概述

**goPerf** 是一款基于 Go 的 HTTP/SSE 性能压测工具，用于对目标服务发起并发请求、采集延迟与 TPS 等指标，并输出周期报告与最终汇总。

### 1.1 目标用户

- 开发/测试人员：对 HTTP 或 SSE 接口做压力与稳定性验证
- 运维/SRE：评估服务容量与限流策略

### 1.2 核心价值

- **轻量**：单二进制、无外部运行时依赖
- **协议支持**：HTTP 通用请求 + SSE 流式场景（首 token 延迟、后续 token 速率等）
- **可配置**：`.properties` 描述请求，命令行控制并发与时长

---

## 2. 架构概览

### 2.1 目录与模块

```
goPerf/
├── cmd/                    # 程序入口
│   └── main.go             # 解析参数、构建 context、启动 app
├── internal/
│   ├── app/                # 压测编排与指标汇总
│   │   ├── app.go          # Start()：启动 worker + 统计协程
│   │   └── metrics.go      # HandleReport、周期/最终报告
│   ├── client/             # HTTP 客户端
│   │   └── client.go       # DoSend：发请求、填 Report、写 channel
│   ├── config/             # 配置加载
│   │   └── loader.go       # LoadRequest：从 .properties 加载 Request
│   └── model/              # 领域模型与统计
│       ├── request.go      # 请求定义与校验
│       ├── perf_conf.go    # 压测配置（并发、时长、报告周期等）
│       ├── statistic.go    # Report、Stat 接口、HttpStat/SseStat
│       └── verify.go       # 成功判定（状态码 / 事件数）
├── pkg/utils/              # 通用工具
│   ├── fileUtils.go        # 文件存在、ParseProperties
│   ├── stringUtis.go       # ParseInt
│   └── timeUtils.go        # 时长格式化等
├── client.properties       # 示例请求配置
├── go.mod
└── docs/
    └── FEATURE_DESIGN.md   # 本文档
```

### 2.2 数据流

```
main
  │
  ├─ parseArgs() → Request + PerfConf
  ├─ context (WithTimeout 或 WithCancel) + SIGINT 监听
  └─ app.Start(ctx, request, conf)
        │
        ├─ 1 个统计协程：HandleReport
        │     ├─ 从 statCh 收 Report → Stat.Add(report, verify)
        │     ├─ 周期 ticker → ReportPeriod + ResetPeriod
        │     └─ ctx.Done() → ReportLast，退出
        │
        └─ N 个压测协程：workerLoop
              └─ 循环 DoSend(request, statCh) 直到 ctx 取消
                    │
                    └─ client.DoSend
                          ├─ 发 HTTP 请求（支持 Keep-Alive）
                          ├─ 按 protocol 处理：HTTP 读 body / SSE 读流并计首 token、后续 token
                          └─ 写入 Report → statCh
```

### 2.3 生命周期与退出

- **Ctrl+C (SIGINT)**：`cancel()` → 所有 worker 与 HandleReport 收到 `ctx.Done()`，正常退出并打最终报告。
- **运行时长 `-t N`**：`context.WithTimeout(..., N*time.Second)`，到期自动取消，行为同上。
- 主流程 `WaitGroup.Wait()` 等待统计协程与全部 worker 结束后进程退出。

---

## 3. 功能规格

### 3.1 已实现功能

| 功能 | 说明 |
|------|------|
| 请求配置 | 通过 `.properties` 配置 URL、method、body、超时、header、protocol(http/sse)、keepalive |
| 并发压测 | 指定并发数 N，启动 N 个 goroutine 循环发请求 |
| HTTP 统计 | 周期/累计：TPS、平均/最小/最大延迟(ms)、总请求数、成功数 |
| SSE 统计 | 首 token 延迟(平均/最大)、后续 token 延迟与 token/s、总时延等 |
| 成功判定 | 按响应状态码（默认 200）或按 SSE 事件数（VerifyByEvent） |
| 周期报告 | 按 `-rt` 秒周期打印当前周期 TPS/延迟与累计信息 |
| 最终报告 | 退出时打印 SUMMARY（总请求、成功/失败、平均 TPS、平均延迟） |
| 运行时长 | `-t N`：运行 N 秒后自动结束 |
| 优雅退出 | Ctrl+C 或 timeout 后统一 cancel，所有协程退出后再打最终报告 |

### 3.2 命令行参数（重构后）

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-f` | string | `client.properties` | 请求配置文件路径（.properties） |
| `-n` | int | 1 | 并发 worker 数 |
| `-t` | uint | 0 | 运行时长（秒），0 表示持续运行直到 Ctrl+C |
| `-rt` | int | 1 | 报告周期（秒） |
| `-c` | int | runtime.NumCPU() | 使用的 CPU 核数（GOMAXPROCS） |
| `-d` | bool | false | 调试模式（预留） |

**示例：**

```bash
# 使用默认 client.properties，4 并发，每 5 秒打一次报告，运行 60 秒
./goPerf -n 4 -rt 5 -t 60

# 指定配置文件，8 并发，仅 Ctrl+C 结束
./goPerf -f my.properties -n 8
```

### 3.3 请求配置（.properties）

| 配置项 | 说明 | 示例 |
|--------|------|------|
| url | 请求 URL（需 http/https） | `url=https://127.0.0.1:8448/hello` |
| method | HTTP 方法 | `method=GET` / `method=POST` |
| body | 请求体（POST 等） | `body={"key":"value"}` |
| protocol | 协议类型 | `protocol=http` / `protocol=sse` |
| connTimeout | 连接超时(秒) | `connTimeout=10` |
| readTimeout | 读超时(秒) | `readTimeout=60` |
| writeTimeout | 写超时(秒) | `writeTimeout=60` |
| keepalive | 是否长连接 | `keepalive=true` |
| header.\* | 请求头 | `header.Accept=application/json` |

---

## 4. 扩展与后续设计建议

### 4.1 配置与校验

- **VerifyByCode / VerifyByEvent**：当前在代码中通过 `PerfConf` 传入，可考虑在 .properties 中增加 `verifyByCode=200`、`verifyByEvent=10`，由 config 加载进 PerfConf。
- **YAML/JSON 配置**：可选支持除 .properties 外的配置格式，由 `config` 包统一抽象为 `Request + PerfConf`。

### 4.2 输出与可观测性

- **结构化输出**：支持 JSON 行输出（每周期一行），便于与监控系统或脚本对接。
- **Prometheus metrics**：可选在本地暴露 `/metrics`，暴露请求数、延迟分位数等，便于与现有监控栈集成。

### 4.3 压测策略

- **QPS 限速**：在 worker 侧按目标 QPS 做 throttle（如 token bucket），而不仅是“并发数”。
- **阶梯压测**：按时间阶梯增加并发数或 QPS，由编排层（如 internal/app）控制 worker 数或限速参数变化。

### 4.4 协议与客户端

- **HTTP/2**：显式支持 HTTP/2 或通过 `h2` 标识选择 transport。
- **gRPC**：若需压测 gRPC，可新增 `internal/client/grpc`，实现同一 `DoSend(request, statCh)` 风格接口，由 protocol 选择实现。

### 4.5 稳定性与可维护性

- **超时与重试**：请求级超时已由 ReadTimeout 等控制；可选的有限次重试仅对偶发网络错误，并区分“可重试”与“业务错误”。
- **优雅关闭 statCh**：当前不在 worker 退出时关闭 channel，避免多处 close。若需“收尾阶段再消费完 statCh 中剩余 Report”，可在 app 层在 ctx 取消后做一次有限时间 drain，再 ReportLast。

---

## 5. 重构变更摘要

与重构前相比，主要变更如下：

1. **生命周期**：`-t` 通过 `context.WithTimeout` 真正参与取消逻辑；SIGINT 与 timeout 统一走同一 `cancel()`。
2. **WaitGroup**：统计协程与每个 worker 均 `Add(1)` 并在退出时 `Done()`，主流程 `Wait()` 等待全部结束，避免负计数或提前退出。
3. **包结构**：`cmd/` 仅保留 `package main` 的 main.go；`app` 与 metrics 迁至 `internal/app`，符合单目录单包与 internal 规范。
4. **配置与参数**：请求配置由 `internal/config` 的 `LoadRequest` 从 .properties 加载；命令行改为 flag（`-f`、`-n`、`-t`、`-rt`、`-c`、`-d`），默认 `-f client.properties`、`-n 1`。
5. **客户端健壮性**：DoSend 中对 `resp == nil`、`resp.Body == nil` 做防护，避免 panic；使用 `time.Since`、`strings.EqualFold` 等规范写法。
6. **代码清理**：删除 app 中残留的无效 `Request` 标识符；printTitle 改为 switch；报告表头拼写修正（deley → delay）。

---

## 6. 版本与维护

- **文档版本**：与当前重构后代码一致（含 go.mod、internal/app、internal/config、flag 用法）。
- **后续迭代**：新功能建议在本文“扩展与后续设计建议”中先做条目与接口级描述，再落实现码，并同步更新本文档。
