# goPerf 测试步骤文档

本文档描述如何在本机完成 goPerf 的编译、端到端及基本功能验证。

---

## 1. 测试目的与范围

| 项目 | 说明 |
|------|------|
| 编译验证 | 确认工程可成功编译出可执行文件 |
| 端到端测试 | 使用内置 HTTP 服务桩，完成「压测端 → 服务端 → 统计报告」全流程 |
| 基本功能 | 定时退出（-t）、Ctrl+C 退出、周期报告、SUMMARY 汇总 |

---

## 2. 环境准备

### 2.1 要求

- **Go**：1.16 或以上（`go version` 检查）
- **工作目录**：进入 goPerf 项目根目录，后续命令均在此执行

```bash
cd /path/to/goPerf
```

### 2.2 确认文件存在

```bash
# 应有以下关键文件/目录
ls cmd/main.go cmd/testserver/main.go
ls internal/app internal/client internal/config internal/model
ls e2e.properties client.properties
```

---

## 3. 编译验证

### 3.1 编译主程序

```bash
go build -o goPerf ./cmd/
```

**预期**：无报错，当前目录下生成可执行文件 `goPerf`。

### 3.2 编译测试服务桩（可选）

```bash
go build -o testserver ./cmd/testserver/
```

**预期**：无报错，生成 `testserver`。后续步骤也可直接用 `go run ./cmd/testserver/`，无需此步。

### 3.3 检查帮助行为

```bash
./goPerf -h
```

**预期**：打印 usage，包含 `-f`、`-n`、`-t`、`-rt`、`-c`、`-d` 等参数说明。

---

## 4. 端到端测试（推荐流程）

本流程使用项目自带的 HTTP 服务桩和 `e2e.properties`，在本地完成一次完整压测。

### 4.0 一键自动化（推荐）

在项目根目录执行以下命令即可自动完成「编译 → 启动测试桩 → 等待就绪 → 运行压测 → 停止测试桩」：

```bash
./scripts/run_e2e.sh
```

脚本会：编译 `goPerf` 与 `testserver`，后台启动 testserver（端口 8080），用 curl 轮询直到服务就绪，再运行 `./goPerf -f e2e.properties -n 2 -t 5 -rt 1`，最后无论成功失败都会结束 testserver 进程。脚本退出码与 goPerf 一致（0 表示通过）。

若需手动分步执行，请按下面 4.1、4.2 操作。

### 4.1 第一步：启动测试服务

**新开终端（终端 A）**，在项目根目录执行：

```bash
cd /path/to/goPerf
go run ./cmd/testserver/
```

**预期输出示例**：

```
202x/xx/xx xx:xx:xx testserver listening on http://127.0.0.1:8080
```

保持该终端运行，不要关闭。

**可选**：若 8080 被占用，可指定端口，例如 9090：

```bash
go run ./cmd/testserver/ -port 9090
```

若使用 9090，后续需用自建配置将 `e2e.properties` 中的 `url` 改为 `http://127.0.0.1:9090/hello`，或新建一份配置。

### 4.2 第二步：运行压测

**再开一个终端（终端 B）**，在项目根目录执行：

```bash
cd /path/to/goPerf
./goPerf -f e2e.properties -n 2 -t 5 -rt 1
```

参数含义：

- `-f e2e.properties`：使用端到端测试配置（请求 `http://127.0.0.1:8080/hello`）
- `-n 2`：2 个并发 worker
- `-t 5`：运行 5 秒后自动结束
- `-rt 1`：每 1 秒打印一次周期报告

### 4.3 预期输出说明

**（1）启动后立即看到表头**，例如：

```
│****************************************************************************************│
│******************* This is goPerf v1.0   Author: J-H.Wang *******************│
│****************************************************************************************│
│----------------------------------------------------------------------------------------│
│   runtime  │  tps0     delay0  │   tps     delay  │   min    max   │   total   success  │
│------------┼------------------┼------------------┼----------------┼--------------------┼
```

**（2）运行期间每秒一行周期报告**，例如：

```
│ 00:00:01  │   1234    1      │   1234    1     │   0      5     │   1234    1234    │
│ 00:00:02  │   1200    1      │   1217    1     │   0      8     │   2434    2434    │
...
```

列含义：当前周期 TPS、周期平均延迟、累计 TPS、累计平均延迟、最小/最大延迟(ms)、总请求数、成功数。

**（3）约 5 秒后自动结束，并打印 SUMMARY**，例如：

```
******************************* SUMMARY REPORT*******************************
runtime:   00:00:05, total requet:     xxxxx, successRequest:     xxxxx, falureReuqest:        0
averTps:     xxxxx, averDelay:         x
```

### 4.4 如何判断端到端通过

- 终端 A 的 testserver 无 panic，持续在监听。
- 终端 B 的 goPerf：
  - 有表头、有多行周期报告、最后有 SUMMARY。
  - **successRequest 与 total requet 相等**（或失败数为 0），表示所有请求均被服务桩正常处理。

若 **successRequest 明显小于 total**，或大量失败：请确认终端 A 的 testserver 已启动、端口为 8080，且无防火墙拦截。

---

## 5. 其他验证场景（可选）

### 5.1 定时退出（-t）

已在上节体现：`-t 5` 应在约 5 秒后自动退出并打 SUMMARY。可将 `-t` 改为 3 或 10 再跑一次确认行为一致。

### 5.2 Ctrl+C 优雅退出

1. 终端 A 保持 testserver 运行。
2. 终端 B 执行（不设 `-t`，即一直跑）：

   ```bash
   ./goPerf -f e2e.properties -n 2 -rt 1
   ```

3. 运行几秒后，在终端 B 按 **Ctrl+C**。

**预期**：进程退出前打印 SUMMARY，无 panic；终端 A 的 testserver 仍正常。

### 5.3 增加并发与报告周期

```bash
./goPerf -f e2e.properties -n 4 -t 10 -rt 2
```

**预期**：10 秒内约每 2 秒一行报告，总请求量较 `-n 2` 明显增多。

### 5.4 使用根路径 /

若希望压测 `http://127.0.0.1:8080/`，可复制一份配置并改 url：

```bash
cp e2e.properties e2e-root.properties
# 编辑 e2e-root.properties，将 url 改为: url=http://127.0.0.1:8080/
./goPerf -f e2e-root.properties -n 2 -t 3 -rt 1
```

行为应与 `/hello` 类似，仅响应体为 `ok`。

---

## 6. 常见问题

| 现象 | 可能原因 | 处理 |
|------|----------|------|
| `config file xxx does not exist` | 未在项目根目录执行，或 `-f` 路径错误 | 使用 `-f e2e.properties` 且确保该文件存在 |
| `connection refused` / 大量 500 | testserver 未启动或端口不对 | 先启动 `go run ./cmd/testserver/`，确认 8080 可用 |
| 编译报错 `package main` 与 `package app` 冲突 | 误把 app 代码放回 cmd/ | 保持 app 在 `internal/app/`，cmd 下仅保留 main.go 与 testserver |
| SUMMARY 中 success 为 0 | 服务未就绪或 URL/端口错误 | 确认 e2e.properties 中 url 与 testserver 监听地址、端口一致 |

---

## 7. 检查清单（发布前自测）

- [ ] `go build -o goPerf ./cmd/` 成功
- [ ] `./goPerf -h` 正常打印帮助
- [ ] 一键脚本通过：`./scripts/run_e2e.sh` 退出码 0
- [ ] 或按第 4 节手动完成端到端：testserver + `./goPerf -f e2e.properties -n 2 -t 5 -rt 1`
- [ ] 周期报告与 SUMMARY 正常，且 successRequest == total request
- [ ] Ctrl+C 能优雅退出并打印 SUMMARY
- [ ] （可选）不同 `-n`、`-t`、`-rt` 组合行为符合预期

完成以上步骤即表示本地测试通过，可进行提交或发布。
