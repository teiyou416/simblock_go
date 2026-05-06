# SimBlock-Go

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

## Test

```bash
make test
```

More details:

- `docs/testing.md`
- `docs/usage_en.md` (English)
- `docs/usage_zh.md` (中文)
