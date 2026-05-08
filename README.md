# SimBlock-Go

[English](./README.md) | [中文](./README_zh.md) | [日本語](./README_ja.md)

Go version of SimBlock.

## Quick Start

```bash
make build
make run
```

You can also override `config/simulator.yaml` directly from the command line:

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible false
```

Use `--config` to point at a different YAML file if needed:

```bash
go run ./cmd/simblock --config ./config/simulator.yaml --latency-matrix-file ./data/latency.txt
```

To enable Java SimBlock compatible mode:

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible true --end-block-height 1000
```

## Test

```bash
make test
```

## Documentation

For more details, please refer to:

- [Testing Guide](./docs/testing.md)
- [Usage Guide (English)](./docs/usage_en.md)
- [Usage Guide (中文)](./docs/usage_zh.md)
- [Usage Guide (日本語)](./docs/usage_ja.md)
