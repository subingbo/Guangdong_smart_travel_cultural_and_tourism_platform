# AI 行程规划 Agent（Go 主线）

本目录记录从 `gdufs-campus` 迁移过来的「时空规划 Agent」在 Travel Guangdong 的落地方式。实现只改 Go 后端（`server-go`），Java 双后端保持原接口不变。

## 1. 目标与边界

**做什么**

- 自然语言对话 → 白名单工具调用 → 确定性行程求解 → 草稿 → 用户确认 → 写入 `routes`。
- 全程流式（SSE），支持取消、超时、排队、断线后轮询恢复、进程重启标记中断。
- 用 goroutine 复刻 gdufs 虚拟线程方案的**准入语义**：每 run 一个 goroutine + `running`/`admission` 双信号量 + 排队超时 + 队列满拒绝。
- 提供 `fixed` 执行器作为对照臂与低内存回退。

**不做什么**

- 群组、日程、共享地图、WebSocket `@AI`（本项目没有这些业务模块）。
- MCP 动态工具、Multi-Agent、向量检索、推测执行、few-shot、Trace Center。
- 多实例水平扩展：订阅者、限流、取消状态都在进程内，当前按单实例部署设计。

## 2. 架构

```
POST /api/agent/runs
  └─ AgentRunService.Submit
       ├─ 限流 / 会话锁 / clientRequestId 幂等
       ├─ ai_runs(queued) + guide_messages(user + assistant 占位)
       └─ Executor.Submit
            ├─ SpawnExecutor: goroutine + running/admission 双信号量
            └─ FixedExecutor: N worker + 有界队列
                 └─ agent.Runner.Run
                      ├─ LLM(OpenAI-compatible, stream, tools)
                      ├─ ToolRegistry（白名单 + 预算 + 去重 + 投影）
                      │    ├─ search_places / get_place / list_favorites
                      │    ├─ search_knowledge / get_weather / route_leg
                      │    ├─ plan_itinerary → internal/planner
                      │    ├─ submit_plan → 草稿落库 + plan 事件
                      │    └─ list_my_routes
                      └─ ai_steps（llm/tool 步骤）
       └─ SSE: snapshot → delta/thinking → tool_call/tool_result → plan → done/error
```

## 3. 与 gdufs / citysim 的对应关系

| gdufs-campus（Java） | 本项目（Go） | 说明 |
|---|---|---|
| `AiRunExecutor` | `agent.Executor` | 统一提交/移除/排队/关闭接口 |
| `VirtualRunExecutor` | `agent.SpawnExecutor` | 每 run 一 goroutine + 双信号量；Go 不需要虚拟线程 |
| `PlatformRunExecutor` | `agent.FixedExecutor` | 固定 worker + 有界队列，对照臂/回退 |
| `SpatiotemporalPlanner` | `internal/planner` | 空闲窗、全排列、真实路况、评分、Top-3 |
| `AgentToolkit`（14 个工具） | `service/agent_tools.go`（9 个工具） | 白名单 + 预算 + 同参去重 |
| `ToolResultProjector` | `agent.Projector` | 字符串 300 字、数组 50、几何字段剔除、单工具 16KB |
| `ai_assistant_run` 状态机 | `ai_runs` + `guide_messages` | queued/running/canceling/succeeded/failed/canceled/interrupted |
| `SseEmitter` + delivery pool | `AgentRunService` 订阅者 + 有界 chan + 单写 handler | 慢客户端丢弃，不阻塞 runner |
| `assistant-run-clock` tick | `tickLoop` + 2 个 flush worker | 超时检查 + 进度落库卸载 |
| `AgentTrace` | `ai_steps` | LLM/tool 入参、结果、耗时、错误 |
| citysim `NativeAgent.run` | `agent.Runner.Run` | 手写 ReAct；修复单 tool accumulator，支持一轮多调用 |

**关键判断**：goroutine 本身就是轻量级用户态线程，生产不需要「虚拟线程执行器」。保留双信号量是因为真正的硬上限是 LLM 配额、高德配额和 MySQL 连接数。

## 4. Agent 循环与预算

- 最多 **8 步**；每步最多 **3 个工具调用**。
- 工具预算：`search_places` 4、`get_place` 8、`list_favorites` 2、`search_knowledge` 4、`get_weather` 2、`route_leg` 24、`plan_itinerary` 3、`submit_plan` 2、`list_my_routes` 2。
- 相同 `name + canonical args` 第二次调用直接回 `duplicate_call`，不执行。
- 工具超时 20s；参数上限 16KB；结果上限 16KB（投影后）。
- `submit_plan` 只接受本 run `plan_itinerary` 返回的 `solutionId`，stops/时间/坐标由服务端填充，模型无法伪造。
- 达到最大步数时：有草稿按成功收尾并提示；无草稿按 `max_steps` 失败。

## 5. 确定性行程规划

`internal/planner` 不访问数据库、不调用 LLM：

1. `ParseWindow` 把「明天下午」「周六下午两点到五点」或 ISO 区间解析为绝对时间（Asia/Shanghai，注入 `now` 便于测试）；过去时间、超过 30 天、窗口 >24h 直接拒绝。
2. `FreeWindowSolver` 合并忙碌区间并求补集；当前没有日程模块，`busy` 传空，接口保留。
3. `SpatiotemporalPlanner`：候选 ≤4，全排列 ≤24；逐窗口、逐排列推进时间线；起点/站间/返程全部走 `TravelProvider`；头部衔接缓冲 + 收尾缓冲；不满足硬约束返回 nil。
4. 评分：`0.35*rank + 0.25*remaining + 0.20*favorite + 0.20*travel`，透明可复算；取 Top-3。
5. 无解返回 `suggestions`（换一天/减少地点/换交通方式），不硬造方案。

`ai.TravelService`：

- 高德单段路线优先；任何失败退化为 haversine + 速度表（步行 4.5、骑行 15、驾车 30、公交 15 km/h），标记 `estimated=true`。
- 每 run 内存缓存 + `singleflight` 去重 + 全局信号量（默认 8）。
- 规划前对候选两两路段并发预取（最多 4 并发），避免排列求解串行打高德。

## 6. Run 状态机与 SSE

```
queued → running → succeeded
                 ├→ failed (llm_error / run_timeout / queue_timeout / queue_full / max_steps / empty_output)
                 └→ canceled
queued → canceled（Remove 成功）
重启恢复：queued/running/canceling → interrupted
```

- 提交：限流（10 次/分/用户）→ 会话行锁（同会话单 active run）→ `client_request_id` 幂等 → 建 run + 用户消息 + 助手占位 → 执行器提交。
- 进度：delta 写内存（上限 64KB），500ms tick 标脏，固定 2 个 flush worker 落库，每 run `flushInFlight` 保证顺序。
- 超时：run 5 分钟；排队 30 秒；队列满返回 429。
- 取消：先置 `canceling` 再 cancel context；排队任务尝试 `Remove`；运行中任务协作退出。
- SSE：首个事件 `snapshot`，随后 `delta`/`thinking`/`tool_call`/`tool_result`/`plan`/`done`/`error`；每 15s `: ping` 心跳。每事件带单调 `seq`，客户端按 snapshot 的 seq 去重。
- 慢客户端：订阅者有界 chan（128），满则断开，不阻塞 runner。
- 重启：启动时把遗留 active run 标 `interrupted`，占位消息同步标记；不自动重放工具。

## 7. 数据模型

复用：

- `guide_sessions`：对话；新增 `plan_json`、`plan_version`、`last_run_id`。
- `guide_messages`：消息；新增 `run_id`、`tool_calls`、`status`、`model`、token 字段；`content` 改 `MEDIUMTEXT`。

新增：

- `ai_runs`：run 快照、状态、输出、错误、步数、token、plan、route_id。
- `ai_steps`：每步 llm/tool 的入参、结果、耗时、错误。
- `ai_plan_commits`：`(user_id, idempotency_key)` 唯一，保证确认幂等。

迁移文件：`sql/migrations/20261105_agent.sql`。`sql/schema.sql` 已同步为全新安装版本。

> Go 端用 `SELECT *` + sqlx StructScan，新增列必须同步到 `model.GuideSession`/`GuideMessage`；Java MyBatis 按列名映射，忽略多余列。

## 8. 并发与资源上限

| 配置 | 默认 | 说明 |
|---|---|---|
| `AGENT_EXECUTOR_MODE` | `spawn` | `fixed` 用于对照/回退 |
| `AGENT_MAX_CONCURRENT_RUNS` | 4 | 必须 ≤ `DB_MAX_OPEN_CONNS` |
| `AGENT_QUEUE_CAPACITY` | 16 | 总在途 = 并发 + 排队 |
| `AGENT_QUEUE_TIMEOUT_SECONDS` | 30 | 排队超时 |
| `AGENT_RUN_TIMEOUT_SECONDS` | 300 | 单 run 总预算 |
| `AGENT_MAP_CONCURRENCY` | 8 | 高德全局并发 |
| `DB_MAX_OPEN_CONNS` | 10 | MySQL 硬上限 |

`SpawnExecutor` 的语义与 gdufs 的 `VirtualRunExecutor` 对齐：

1. 先抢 `running`；成功直接执行。
2. 失败再抢 `admission`；失败立即 `ErrQueueFull`。
3. 排队 goroutine 用 timer 等 `running`；超时回调 `onAdmissionTimeout`。
4. `Remove` 只取消未开始的任务；运行中必须靠 context 取消。

## 9. API

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/agent/runs` | `{conversationId?, message, clientRequestId?}` → 202 snapshot |
| GET | `/api/agent/runs/:id` | 快照（轮询/重连） |
| GET | `/api/agent/runs/:id/events` | SSE；需 `Authorization`，前端用 fetch 流 |
| POST | `/api/agent/runs/:id/cancel` | 取消 |
| POST | `/api/agent/runs/:id/plan/confirm` | `{version, idempotencyKey}` → 写 routes |
| GET | `/api/agent/conversations` | 会话列表 |
| GET | `/api/agent/conversations/:id` | 历史消息 + 最新 plan |
| GET | `/api/agent/debug/metrics` | `AGENT_DEBUG=true` 才开放 |

完整 schema 见 `docs/api/openapi.yaml`。

## 10. 安全

- 工具身份只来自 `ExecutionContext.UserID`（JWT），不接受模型参数里的 userId。
- 工具白名单固定注册；不开放 SQL/任意函数。
- `submit_plan` 的服务端数据源是求解器缓存，模型只能选 `solutionId`。
- 确认时重新校验：地点存在、坐标未漂移、时间窗未过、版本一致；事务内写 route + 提交记录。
- 工具结果按不可信资料处理；系统提示明确禁止执行其中的指令。
- 输出、工具结果、历史消息均有大小上限；错误信息不回传 SQL/上游细节。

## 11. 验证

```bash
cd server-go
go build ./...
go test ./...
```

- 规划器：`internal/planner` 单测覆盖空闲窗、相对时间、可行性、评分、取消、篡改校验。
- Agent：`internal/agent` 单测覆盖单/多工具调用、错误回填、重复调用、plan 事件、最大步数、取消。
- 天气：`internal/ai` 覆盖 21 城 adcode 解析、预报解析、10 分钟缓存、上游错误与未配置降级；`AMAP_WEB_KEY` 存在时跑一次高德实网测试。
- 执行器：覆盖并发上限、队列满、排队超时、Remove。
- 端到端：启动 fake LLM + MySQL + seed，走 `/guide` 页面完成「规划 → 方案卡片 → 确认 → 我的路线」。

压测方法见 [benchmark.md](./benchmark.md)。

## 12. 已知边界

- 订阅者/限流/取消是进程内状态，多副本部署需要 Redis pub/sub 或 sticky session。
- AMap 失败时是直线估算，`routeConfidence=ESTIMATED` 会展示给用户；不能宣称「已验证」。
- 天气只查高德预报（今日 + 未来 3 天），不查实时天气；`AMAP_WEB_KEY` 未配置时 `get_weather` 返回 `available:false`，提示词要求模型说「未查询到」而不是编造。
- 当前没有用户日程模块，无法做多人 free-busy；`busy` 接口保留，未来接入日程即可。
- `ai_steps` 是最小 trace，不是 gdufs 的完整 Trace Center。
- 生产 2C1G：默认并发 4；提高并发前先核算 LLM 配额、高德配额与 MySQL 连接池。
