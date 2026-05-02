# SimBlock-Go 已完成内容说明（截至 2026-05-02）

本文档用于总结当前分支中已经完成的重写工作，帮助快速了解项目现状与下一步重点。

## 1. 总体进度（对应 `docs/rewrite_plan.md`）

- 阶段 1：工程基建与核心引擎（大部分完成，文档中仅剩 Sequence ID 稳定性校验未勾选）
- 阶段 2：物理网络与节点模型（已完成）
- 阶段 3：区块链协议抽象（已完成）
- 阶段 4：消息协议与任务实装（已完成）
- 阶段 5~7：尚未完成

## 2. 已完成的核心能力

### 2.1 运行与工程基线

- 入口可运行：`go run ./cmd/simblock` 可启动
- 配置可加载：`config/simulator.yaml` + `viper`
- 基础构建可用：`go test ./...` 可全量通过

关键文件：

- `cmd/simblock/main.go`
- `internal/app/run.go`
- `config/config.go`
- `config/simulator.yaml`

### 2.2 核心调度引擎（离散事件）

- 实现了基于最小堆的事件调度器
- 支持：
  - 相对时间调度 `PutTask`
  - 绝对时间调度 `PutTaskAt`
  - 任务移除 `RemoveTask`
  - 主循环执行 `RunUntilEmpty`
- 同时间戳按 `sequence` 保持稳定顺序

关键文件：

- `engine/timer.go`
- `engine/timer_test.go`

### 2.3 网络模型（阶段2）

- Region 建模完成（默认地区列表）
- 支持延迟矩阵文件加载（`latency.txt`）
- 支持带宽瓶颈计算与传输时延计算

关键文件：

- `network/region.go`
- `network/latency_matrix.go`
- `network/network.go`
- `data/latency.txt`
- `network/*_test.go`

### 2.4 协议层抽象（阶段3）

- `core.Block` 已抽象化，支持：
  - 父块索引/父指针
  - 高度、时间、出块者
  - `ConsensusData any` 扩展字段
  - 同链判断与按高度回溯
- 定义了共识接口 `node/consensus/consensus.go`
- 定义了路由抽象接口 `node/routing/routing.go`（含 `BaseTable`）
- 实现 PoW：
  - 挖矿间隔抽样（指数分布）
  - 验块规则（难度与总难度）
  - 难度调整窗口逻辑

关键文件：

- `core/block.go`
- `node/consensus/consensus.go`
- `node/routing/routing.go`
- `node/consensus/pow.go`
- 对应测试：`core/block_test.go`、`node/routing/routing_test.go`、`node/consensus/pow_test.go`

### 2.5 消息协议与任务（阶段4）

- 新增 `tasks` 包并实现：
  - `MiningTask`
  - `InvMessageTask`
  - `RecMessageTask`
  - `BlockMessageTask`
  - `CmpctBlockMessageTask`
  - `GetBlockTxnMessageTask`
- 节点侧已接入消息流转：
  - `SendInv`
  - `ReceiveMessage`
  - `SendNextBlockMessage`
  - 下载中集合与消息队列
  - compact block 失败回退到 `GetBlockTxn` 再发标准区块

关键文件：

- `tasks/message_task.go`
- `tasks/mining_task.go`
- `tasks/inv_message_task.go`
- `tasks/rec_message_task.go`
- `tasks/block_message_task.go`
- `tasks/cmpct_block_task.go`
- `tasks/get_block_txn_message_task.go`
- `node/node.go`
- `tasks/tasks_test.go`

## 3. 当前测试与可运行状态

- 全量测试：`go test ./...` 通过
- 启动验证：`go run ./cmd/simblock` 通过

说明：

- 当前已经具备“事件驱动 + 节点消息流转 + PoW 抽象”的可执行基础。
- 但尚未进入完整模拟主循环与结果输出阶段（见下一节）。

## 4. 仍未完成的部分（高优先级）

### 4.1 阶段5：主循环与集成模拟

- 实现 `engine/simulator.go`（统一管理全局节点与运行过程）
- 在主入口串联：
  - 配置读取
  - 网络初始化
  - 节点构建与连边
  - 创世块生成与首轮任务入队
- 跑通“从创世块到首个区块传播”的完整流程

### 4.2 阶段6：输出与可视化对齐

- 输出 `output.json` / `static.json` 等
- 采集传播延迟、孤块率等指标

### 4.3 阶段7：等效性验证

- 固定 seed 与 Java 版做结果比对
- 输出 diff 与性能调优

## 5. 建议的下一步实施顺序

1. 先完成阶段5（这是当前“能不能成为完整模拟器”的关键）
2. 再做阶段6（否则无法验证结果质量）
3. 最后做阶段7（等效性与性能收敛）

