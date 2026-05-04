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

命令行参数会覆盖 [config/simulator.yaml](../config/simulator.yaml) 中的同名配置，例如：

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible false
```

如果你想切换到别的 YAML 文件，可以使用 `--config`：

```bash
go run ./cmd/simblock --config ./config/simulator.yaml --latency-matrix-file ./data/latency.txt
```

默认输出会写入 `output/` 目录。

## 3. 配置模拟参数

主配置文件：

- `config/simulator.yaml`

常用字段：

- `simulation.num_nodes`
- `simulation.end_time`
- `simulation.end_block_height`
- `simulation.block_interval`
- `simulation.java_compatible`
- `network.latency_matrix_file`

## 4. 运行测试

运行单元测试 + 集成测试：

```bash
make test
```

或使用脚本：

```bash
./scripts/run_tests.sh
```

## 5. Java/Go 对齐实验

单次对比：

```bash
./scripts/compare_with_java.sh
```

批量对比（默认 10 次）：

```bash
./scripts/batch_compare_java_go.sh 10
```

测试 + 对齐一条命令：

```bash
./scripts/run_tests.sh --with-align
```

## 6. 输出文件

主要输出文件：

- `output/output.json`
- `output/static.json`
- `output/metrics.json`

这些文件属于运行生成产物，默认不纳入版本控制。
