# SimBlock-Go 实施清单（中文版）

这份清单用于日常推进。按阶段顺序执行，只有在满足验收标准后再勾选。

## 当前基线
- 已有事件调度器与单元测试：`engine/timer.go`、`engine/timer_test.go`
- 已有最小领域骨架：`block/`、`node/`、`network/`
- 已有最小闭环测试：`engine/simulation_loop_test.go`
- 构建/运行基线尚未完全打通（入口与配置仍有不一致）

## 阶段 0：先打通可构建可运行基线
- [ ] 修复 `config/simulator.yaml` 的结构与缩进。
- [ ] 将入口统一到 `cmd/simblock/main.go`。
- [ ] 修正 `Makefile` 中 `MAIN_FILE` 的路径。
- [ ] 确保 `go test ./...`、`go build ./...`、`make build` 全部通过。

验收标准：
- 使用 `go run` 启动后程序可正常读取配置并运行。

## 阶段 1：冻结引擎行为契约
- [ ] 增加同时间戳 FIFO 执行顺序测试。
- [ ] 增加同一任务实例重排/覆盖行为测试。
- [ ] 增加任务取消/删除行为测试。
- [ ] 增加绝对时间调度行为测试。
- [ ] 保持核心模拟单线程（核心路径不引入 goroutine）。

验收标准：
- 调度器行为被测试完整锁定，在固定 seed 下可重复。

## 阶段 2：配置系统对齐 Java 参数
- [ ] 将 Java 的 simulation/network 参数映射到 Go 配置结构。
- [ ] 支持 region 列表、region 分布、degree 分布。
- [ ] 支持上下行带宽与延迟矩阵加载。
- [ ] 增加配置校验（维度不匹配、非法值、缺失字段）。

验收标准：
- Go 配置能表达 Java 版主要实验参数。

## 阶段 3：领域模型扩展到可插拔
- [ ] 扩展 `Block` 结构，支持共识扩展（parent/height/time/consensus data）。
- [ ] 扩展 `Node` 状态：算力、邻居、孤块集、下载集、发送队列/状态。
- [ ] 定义解耦接口：`consensus`、`routing`。

验收标准：
- 节点模型可以承载完整 PoW 与消息协议状态迁移。

## 阶段 4：网络与路由实现
- [ ] 实现 region 模型（`network/region.go`）。
- [ ] 实现延迟抖动 + 带宽瓶颈网络行为。
- [ ] 实现路由接口与 `BitcoinCoreTable` 等价逻辑。
- [ ] 实现节点入网/建网与连接度约束。

验收标准：
- 固定 seed 下，拓扑与传播时延可复现。

## 阶段 5：PoW 与挖矿任务
- [ ] 实现共识接口及 PoW 逻辑。
- [ ] 实现 `MiningTask`（指数分布抽样出块间隔）。
- [ ] 实现区块合法性判断（difficulty + total difficulty）。
- [ ] 实现“收到更优链时取消旧挖矿并重新挖矿”。

验收标准：
- 竞争分叉出现时，能按规则收敛到正确主链。

## 阶段 6：消息协议任务全量实现
- [ ] 实现 `InvMessageTask`。
- [ ] 实现 `RecMessageTask`。
- [ ] 实现 `BlockMessageTask`。
- [ ] 实现 `CmpctBlockMessageTask`。
- [ ] 实现 `GetBlockTxnMessageTask`。
- [ ] 实现发送队列与 `sendNextBlockMessage` 语义。

验收标准：
- 传统传播流程与 compact-block 流程都可端到端跑通。

## 阶段 7：Simulator 主循环集成
- [ ] 实现 `engine/simulator.go`（节点列表、目标间隔、传播观测）。
- [ ] 实现初始化流程（region/degree/算力分配）。
- [ ] 实现按算力权重选择创世出块节点。
- [ ] 实现主循环和停止条件（如 `END_BLOCK_HEIGHT`）。

验收标准：
- 模拟可从创世块自动运行到终止条件。

## 阶段 8：输出与可视化对齐
- [ ] 输出 `output.json`。
- [ ] 输出 `static.json`。
- [ ] 输出 `graph/*.txt`。
- [ ] 输出 `blockList.txt`。
- [ ] 对齐 Java 事件类型与字段（`add-node`、`add-block`、`flow-block`、`simulation-end`）。
- [ ] 增加指标：传播延迟、孤块率、平均孤块数。

验收标准：
- 输出可直接接入现有 visualizer/分析工具。

## 阶段 9：等效性验证与性能收尾
- [ ] 固化随机 seed 行为，确保可重放。
- [ ] 用同配置运行 Java 与 Go 并对比结果/指标。
- [ ] 增加关键产物自动 diff 脚本。
- [ ] 在等效性达成后再做性能优化（可选 `sync.Pool`）。

验收标准：
- 功能等效有可重复、可量化证据。

## 每日开发循环（建议固定执行）
1. 从同一阶段选择 1-2 个小任务。
2. 开发前先读对应 Java 源码。
3. 先写/补测试，再写实现。
4. 跑 `go test ./...` 并做一次小规模集成运行。
5. 在本清单勾选进度并补充简要备注。

## 下一步建议（优先 3 项）
1. 修复基线构建/运行路径与 YAML 配置。
2. 先搭好 `consensus/routing` 接口骨架和测试桩。
3. 实现第一版真实 `MiningTask + BlockMessageTask`，替换临时测试任务。
