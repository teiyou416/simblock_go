# SimBlock-Go Usage (English)

## 1. Prerequisites

- Go 1.22+ (or the version required by `go.mod`)
- `make`
- Optional for Java parity checks: JDK 11 and Gradle wrapper support

## 2. Build and Run

```bash
make build
make run
```

You can also run directly:

```bash
go run ./cmd/simblock
```

Command-line flags override the same values found in [config/simulator.yaml](../config/simulator.yaml). For example:

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible false
```

To load a different YAML file, pass `--config`:

```bash
go run ./cmd/simblock --config ./config/simulator.yaml --latency-matrix-file ./data/latency.txt
```

Default simulation outputs are written to `output/`.

## 3. Configure Simulation

Main config file:

- `config/simulator.yaml`

Typical fields:

- `simulation.num_nodes`
- `simulation.end_time`
- `simulation.end_block_height`
- `simulation.block_interval`
- `simulation.java_compatible`
- `network.latency_matrix_file`

## 4. Run Tests

Run unit + integrated suite:

```bash
make test
```

Or use the helper script:

```bash
./scripts/run_tests.sh
```

## 5. Java/Go Alignment Check

Single-run comparison:

```bash
./scripts/alignment.sh
```

Batch comparison:

```bash
./scripts/alignment.sh --runs 10
```

Run tests plus alignment in one command:

```bash
./scripts/run_tests.sh --with-align
```

## 6. Output Files

Main output artifacts:

- `output/output.json`
- `output/static.json`
- `output/metrics.json`

These files are generated at runtime and are not tracked by default.
