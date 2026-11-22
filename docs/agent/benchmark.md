# Agent 执行器压测（Go 版三臂对照）

目标不是证明「goroutine 比虚拟线程快」，而是回答一个可复现的问题：

> 在同一台机器、同一份数据、同一个 fake LLM 下，只切换 Agent 执行器的并发/排队模型，p50/p95/p99、吞吐、排队等待、拒绝率和 goroutine/堆内存如何变化？

Go 的 goroutine 已经是轻量级用户态线程，生产不需要虚拟线程。本压测保留 `fixed` 执行器作为对照臂，验证的是**准入控制语义**（并发上限、排队容量、排队超时、拒绝、取消），不是线程实现。

## 1. 三个臂

| 臂 | 配置 | 等价物 |
|---|---|---|
| A `fixed4` | `AGENT_EXECUTOR_MODE=fixed` `AGENT_MAX_CONCURRENT_RUNS=4` | gdufs 改造前的平台线程池 |
| B `fixed32` | `AGENT_EXECUTOR_MODE=fixed` `AGENT_MAX_CONCURRENT_RUNS=32` | 只加线程 |
| C `spawn32` | `AGENT_EXECUTOR_MODE=spawn` `AGENT_MAX_CONCURRENT_RUNS=32` | gdufs 虚拟线程臂 |

生产默认是 `spawn` + 4 并发 + 16 排队；32 并发只用于压测，不要直接照搬。

## 2. 固定变量

- 同一台机器、同一份二进制、同一 MySQL 快照；每组前恢复 DB。
- `AGENT_MAP_CONCURRENCY=8`、`DB_MAX_OPEN_CONNS=32`（压测专用；生产仍是 10）。
- `AMAP_WEB_KEY` 置空 → 规划器走直线估算，结果确定。
- `AGENT_DEBUG=true` → 每秒采样 `/api/agent/debug/metrics`。
- fake LLM 固定 `4 × 50ms` 文本块；工具链固定 `search_places → plan_itinerary → submit_plan`。
- 每组 `concurrency × runs` 个 run，统一使用「明天下午 / DRIVE / 3 个景点」的 prompt。
- 3 次重复；A/B/C 交替执行，抵消机器漂移。

## 3. 运行

### 3.1 启动 fake LLM

```bash
cd server-go
go run ./cmd/agent-bench -mode fake-llm -addr :9099
```

fake LLM 是 OpenAI 兼容的 SSE 服务：第一轮返回 `search_places` 工具调用，拿到 placeId 后调用 `plan_itinerary`，再调用 `submit_plan`，最后返回文本。它不访问外网。

### 3.2 启动后端（每个臂重启一次）

```bash
# A 臂
AGENT_EXECUTOR_MODE=fixed AGENT_MAX_CONCURRENT_RUNS=4 AGENT_QUEUE_CAPACITY=16 \
AGENT_DEBUG=true AI_BASE_URL=http://127.0.0.1:9099/v1 AI_STREAM=true \
DB_MAX_OPEN_CONNS=32 AMAP_WEB_KEY= \
go run .
```

B 臂改 `AGENT_MAX_CONCURRENT_RUNS=32`；C 臂改 `AGENT_EXECUTOR_MODE=spawn`。其余环境变量保持不变。

### 3.3 驱动

```bash
go run ./cmd/agent-bench -mode bench \
  -arm fixed4 -concurrency 4 -runs 3 \
  -base-url http://127.0.0.1:8081 \
  -out results-fixed4.jsonl
```

并发级别建议 4 / 8 / 16 / 32；每个级别 3 次。脚本会自动注册/登录 `bench_<arm>_<i>` 用户。

### 3.4 汇总

```bash
go run ./cmd/agent-bench -mode analyze -in results-fixed4.jsonl
```

输出：

```
arm        conc   runs    ok%     p50ms     p95ms     p99ms   queueP50      req/s
fixed4        4     12  100.0       812      1103      1280         12       3.21
spawn32      32     96  100.0      2335      3489      4102        420      10.50
```

`queueP50` 是 `startedAt - createdAt`，即排队等待；A 臂在 32 并发下大量快速失败时，p50 不能当作性能。

## 4. 采集指标

客户端（JSONL）：

- `latencyMs`：提交 → 终态。
- `queueMs`：`startedAt - createdAt`。
- `status`：succeeded / failed / canceled / request_failed / poll_failed / client_timeout。
- `steps`、`hasPlan`、`errorCode`。
- 吞吐：run 数 / 墙钟时间（analyze 计算）。

服务端（`/api/agent/debug/metrics`）：

- `goroutines`、`heapAlloc`、`numGC`。
- `activeRuns`、`executorRunning`、`executorQueued`。

## 5. 预期与判读

- A 臂（4 并发）在低并发下延迟最低；32 并发时会大量 `queue_timeout`/`queue_full`，成功率下降。
- B/C 臂把并发提到 32 后，队列等待上升，但成功率应保持高位；C 与 B 的差异主要在 goroutine 调度开销和内存，不在「线程池耗尽」。
- 总在途上限 = `AGENT_MAX_CONCURRENT_RUNS + AGENT_QUEUE_CAPACITY`；64 并发时只完成 48 是**有意背压**，不是执行器失效。
- 真正先打满的通常是 MySQL 连接池和 LLM 配额，而不是 Go 调度器；报告必须把 `DB_MAX_OPEN_CONNS` 和 fake LLM 延迟写清楚。

## 6. 结果模板

```markdown
## 环境
- 机器：2C1G / <OS>；Go <version>
- MySQL：<version>，max_open=32；Redis：未使用
- fake LLM：4 chunk × 50ms；AMAP_WEB_KEY 为空
- 样本：每臂 4/8/16/32 并发 × 3 次

## 结果
| 臂 | 并发 | 成功率 | p50 | p95 | p99 | 排队 p50 | 吞吐 |
|---|---|---|---|---|---|---|---|
| fixed4 | 4 | | | | | | |
| fixed32 | 32 | | | | | | |
| spawn32 | 32 | | | | | | |

## 结论
- <只写数据支持的结论；不要写「虚拟线程更快」这类无法归因的话>
- <记录失败样本的 errorCode 分布>
- <记录 goroutine/堆峰值，以及是否触发 MemoryMax>
```

## 7. 与 gdufs 压测的差异

- gdufs 用 `jcmd`/JFR 观察虚拟线程与 pinning；Go 用 `runtime.NumGoroutine`/`MemStats`。
- gdufs 的三臂需要同一 JDK 21；Go 的三臂只需要同一二进制 + 不同环境变量重启。
- gdufs 的 C 臂是 `VirtualRunExecutor`；Go 的 C 臂是 `SpawnExecutor`，语义相同（每任务一个 goroutine + 双信号量）。
- 不要把本压测的 32 并发结果当作生产建议；生产 2C1G 默认 4 并发，先看 LLM 配额。
