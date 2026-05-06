# SimBlock-Go

[English](./README.md) | [中文](./README_zh.md)

SimBlock 的 Go 版本。

## 快速开始

```bash
make build
make run
```

你也可以直接从命令行覆盖 `config/simulator.yaml`：

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible false
```

使用 `--config` 指向不同的 YAML 文件：

```bash
go run ./cmd/simblock --config ./config/simulator.yaml --latency-matrix-file ./data/latency.txt
```

## 测试

```bash
make test
```

## 文档

更多详细信息，请参考：

- [测试指南](./docs/testing.md)
- [使用指南（English）](./docs/usage_en.md)
- [使用指南（中文）](./docs/usage_zh.md)
