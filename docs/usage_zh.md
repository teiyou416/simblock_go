# SimBlock-Go 使用说明（中文）

## 1. 环境准备

- Go 1.22+（或以 `go.mod` 为准）
- `make`
- 可选（用于 Java 对齐实验）：JDK 11 与 Gradle Wrapper

## 2. 构建与运行

```bash
make build
make run
```

也可以直接运行：

```bash
go run ./cmd/simblock
```

默认输出会写入 `output/` 目录。

## 3. 命令行参数

模拟器本体目前还不支持通过命令行直接传入模拟参数。

这些命令是支持的：

```bash
make run
go run ./cmd/simblock
./bin/simblock_go
```

这些命令目前还不支持：

```bash
go run ./cmd/simblock --num-nodes 100
./bin/simblock_go --end-block-height 5
```

模拟参数需要通过 `config/simulator.yaml` 配置。

辅助脚本支持少量命令行参数：

- `./scripts/run_tests.sh --with-align`：先运行 Go 测试，再运行一次 Java/Go 对齐检查
- `./scripts/alignment.sh --runs 10`：运行 10 次 Java/Go 对齐对比

## 4. 配置模拟参数

主配置文件：

- `config/simulator.yaml`

常用字段：

- `simulation.num_nodes`：模拟节点数量
- `simulation.block_interval`：平均出块间隔，单位毫秒
- `simulation.block_size`：区块大小，单位字节
- `simulation.end_time`：普通 Go 模式下的停止时间
- `simulation.end_block_height`：Java 兼容模式下的停止高度
- `simulation.java_compatible`：是否启用 Java SimBlock 兼容行为
- `network.latency_matrix_file`：网络延迟矩阵文件路径

如果要和 Java 版对齐，使用 `java_compatible: true`。如果要跑 Go 版自己的普通模拟，使用 `java_compatible: false`。

## 5. 运行测试

运行单元测试 + 集成测试：

```bash
make test
```

或使用脚本：

```bash
./scripts/run_tests.sh
```

## 6. Java/Go 对齐实验

单次对比：

```bash
./scripts/alignment.sh
```

批量对比：

```bash
./scripts/alignment.sh --runs 10
```

测试 + 对齐一条命令：

```bash
./scripts/run_tests.sh --with-align
```

## 7. 输出文件

主要输出文件：

- `output/output.json`
- `output/static.json`
- `output/metrics.json`

这些文件属于运行生成产物，默认不纳入版本控制。
