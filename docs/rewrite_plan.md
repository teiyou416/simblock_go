# SimBlock-Go 重写计划书 (Rewrite Plan)

## 1. 项目概述
本计划旨在将 Java 版本的区块链模拟器 **SimBlock** 重写为 **Go** 版本。
**核心目标**：保持功能等效性（Functional Parity）、提升执行效率、利用 Go 的组合特性优化 Java 的深层继承架构。

---

## 2. 核心技术规范 (Technical Guidelines)
- **并发控制**：严禁在模拟器核心逻辑中使用 `goroutine`。必须保持单线程离散事件模拟，确保实验可复现。
- **组合优于继承**：将 Java 的 `Abstract...` 类转换为 Go 的 `interface` 和 `struct` 嵌套。
- **确定性随机性**：使用统一的随机数种子（Seed），对齐 Java 的 `java.util.Random` 行为。
- **工程化**：使用 `Makefile` 管理构建，使用 `viper` 管理配置，使用 `logrus` 记录日志。

---

## 3. 重写进度里程碑 (Milestones)

### 阶段 1：工程基建与核心引擎 [已完成 80%]
- [x] 项目目录结构初始化 (`cmd`, `core`, `engine`, `network`, `node`, `tasks`)
- [x] 配置解析模块 (`config/config.go` + `viper`)
- [x] 核心任务接口 (`core/task.go`)
- [x] 全局时间调度器 (`engine/timer.go`) - 基于 `container/heap` 的最小堆
- [ ] 调度器 Sequence ID 稳定性校验

### 阶段 2：物理网络与节点模型 [已完成]
- [x] **Region 建模**：实现 `network/region.go`。
- [x] **网络延迟矩阵**：重写 `Network.java` 逻辑，支持加载 `latency.txt`。
- [x] **Node 骨架**：实现 `node/node.go`，定义基础属性（NodeID, Region, HashPower）。
- [x] **带宽计算**：实现 `GetBandwidth` 和数据传输耗时计算逻辑。

### 阶段 3：区块链协议抽象 [已完成]
- [x] **Block 结构**：实现 `core/block.go`。支持父块索引、高度和 `ConsensusData any` 字段。
- [x] **共识接口定义**：实现 `node/consensus/consensus.go`，定义 `IsReceivedBlockValid` 等方法。
- [x] **路由表接口定义**：实现 `node/routing/routing.go`，解耦 `AbstractRoutingTable`。
- [x] **PoW 算法实现**：重写 `ProofOfWork.java`，实现难度调整与验证。

### 阶段 4：消息协议与任务实装 (Tasks) [已完成]
- [x] **挖掘任务**：实现 `tasks/mining_task.go`。
- [x] **标准区块传输**：实现 `tasks/block_message_task.go`。
- [x] **致密区块 (Compact Block)**：实现 `tasks/cmpct_block_task.go` 及其对应的 `GetBlockTxn` 逻辑。
- [x] **广播协议**：实现 `InvMessage` 和 `RecMessage` 的逻辑流转。

### 阶段 5：主循环与集成模拟 [已完成]
- [x] **Simulator 封装**：实现 `engine/simulator.go` 管理全局节点列表。
- [x] **Main 入口实装**：在 `cmd/simblock/main.go` 中串联配置、网络、节点初始化。
- [x] **主循环启动**：跑通从第一个创世块生成到第一个区块传播的全流程。

### 阶段 6：数据输出与可视化对齐 [已完成]
- [x] **JSON 日志导出**：格式化输出模拟结果，确保兼容 `simblock-visualizer`。
- [x] **统计指标采集**：实现传播延迟、孤块率等数据的实时收集。

### 阶段 7：等效性验证与调优
- [ ] **随机数序列对齐**：验证 Go 与 Java 随机序列的一致性。
- [ ] **字节级输出比对**：通过固定 Seed，对比 `diff` 两个版本的输出日志。
- [ ] **内存池优化**：引入 `sync.Pool` 优化海量 Task 产生的 GC 压力。

---

## 4. 给 Agent 的执行指令 (Instructions for Agent)
1. **增量开发**：每次只专注于一个 Milestone 中的 1-2 个任务。
2. **源码参考**：在实装每个模块前，请务必读取 `simulator/src/main/java/simblock/` 下对应的 Java 源文件。
3. **保持同步**：每完成一个子任务，更新本 `rewrite_plan.md` 中的复选框状态。
4. **代码风格**：优先产出 idiomatic Go 代码，若 Java 逻辑过于冗余，请主动提出优化建议。
