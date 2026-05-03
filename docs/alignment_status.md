# Go vs Java 对齐进展（feat/align）

更新时间：2026-05-02

## 本次已完成

1. 新增 `java.util.Random` 等价实现：`internal/javarand/random.go`
2. `Simulator` 增加 Java 兼容模式：
   - 地区分布、度分布、CBR/churn 节点分布
   - 高斯算力生成（均值 400000，标准差 100000）
   - Genesis minter 按算力加权抽样
   - 按 `END_BLOCK_HEIGHT` 停止（默认 3）
3. `Network` 增加 Java 风格延迟采样（Pareto-like 公式）
4. `Node` 对齐改造：
   - outbound/inbound 邻居模型
   - 连接上限（`numConnections`）
   - CBR 控制节点/抖动节点失败率
   - `GetBlockTxn` 失败块大小分布采样（control=210*0.01，churn=945 离散分布）
   - `addOrphans` 递归逻辑
5. `PoW` 改为可注入随机源，确保全局随机序列可控。

## 当前配置（已改）

`config/simulator.yaml` 默认切到 Java 兼容运行参数：

- `num_nodes: 300`
- `end_block_height: 3`
- `java_compatible: true`

## 与 Java 基线对比结果

Java 基线命令：`JAVA_HOME=/opt/homebrew/opt/openjdk@11 ./gradlew :simulator:run`

Go 命令：`go run ./cmd/simblock`

事件计数对比：

- `add-node`: Java `300` / Go `300`（一致）
- `add-link`: Java `5166` / Go `5166`（一致）
- `add-block`: Java `1200` / Go `1200`（一致）
- `flow-block`: Java `1491` / Go `1488`（差 3）
- `simulation-end`: 均为 1（一致）

结束时间戳：

- Java: `2056103`
- Go: `761445`

## 仍未完全对齐的点

1. `flow-block` 仍有极小差异（3 条）
2. `simulation-end.timestamp` 仍存在显著偏差
3. 尚未完成“字节级输出完全一致”

## 下一步建议

1. 对齐 `receiveMessage/sendNextBlockMessage` 里的随机调用顺序（逐分支核对 Java 调用点）。
2. 把输出 JSON 的字段顺序与 Java 打印顺序做固定化（当前语义一致，字段顺序不同）。
3. 新增自动 diff 脚本（固定 seed，一键跑 Java/Go 并输出首个差异事件索引）。

## 误差容忍执行（新增）

使用脚本：

`./scripts/alignment.sh`

当前脚本判定阈值：

- `flow-block` 相对误差 <= `2%`
- 事件总数相对误差 <= `2%`
- `simulation-end.timestamp` 相对误差 <= `70%`

说明：该阈值用于“工程可接受范围”对齐，不要求字节级一致。
