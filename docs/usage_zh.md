# SimBlock-Go 使用说明（中文）

## 1. 环境准备

- Go 1.22+（或以 `go.mod` 为准）
- `make`

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

## 3. 命令行参数

模拟器本体支持用命令行参数覆盖 `config/simulator.yaml` 中的同名配置。

支持的参数：

- `--config`：YAML 配置文件路径
- `--num-nodes`：覆盖 `simulation.num_nodes`
- `--block-interval`：覆盖 `simulation.block_interval`
- `--block-size`：覆盖 `simulation.block_size`
- `--end-time`：覆盖 `simulation.end_time`
- `--end-block-height`：覆盖 `simulation.end_block_height`
- `--java-compatible`：覆盖 `simulation.java_compatible`
- `--network-profile`：覆盖 `network.profile`
- `--latency-matrix-file`：覆盖 `network.latency_matrix_file`

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
- `network.profile`：内置网络 profile 名称，目前支持 `bitcoin_2019`
- `network.latency_matrix_file`：网络延迟矩阵文件路径
- `network.upload_bandwidth`：每个地区的上传带宽，单位 bit/s
- `network.download_bandwidth`：每个地区的下载带宽，单位 bit/s
- `network.region_distribution`：节点地区分布，所有值建议加和为 `1.0`
- `network.degree_distribution`：出站连接数累计分布，最后一个值应为 `1.0`

普通 Go 模拟建议使用 `java_compatible: false`。只有需要 Java 兼容行为时才使用 `java_compatible: true`。

## 5. 运行测试

运行单元测试 + 集成测试：

```bash
make test
```

或使用脚本：

```bash
./scripts/run_tests.sh
```

## 6. 输出文件

主要输出文件：

- `output/output.json`
- `output/static.json`
- `output/metrics.json`

这些文件属于运行生成产物，默认不纳入版本控制。
