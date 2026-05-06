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

Default simulation outputs are written to `output/`.

## 3. Command-Line Arguments

The simulator executable currently does not accept simulation parameters from the command line.

These commands are valid:

```bash
make run
go run ./cmd/simblock
./bin/simblock_go
```

These commands are not supported yet:

```bash
go run ./cmd/simblock --num-nodes 100
./bin/simblock_go --end-block-height 5
```

Simulation parameters are configured through `config/simulator.yaml`.

Helper scripts do support a small number of command-line arguments:

- `./scripts/run_tests.sh --with-align`: run Go tests, then run one Java/Go alignment check
- `./scripts/alignment.sh --runs 10`: run Java/Go alignment comparison 10 times

## 4. Configure Simulation

Main config file:

- `config/simulator.yaml`

Typical fields:

- `simulation.num_nodes`: number of simulated nodes
- `simulation.block_interval`: expected block interval in milliseconds
- `simulation.block_size`: block size in bytes
- `simulation.end_time`: stop time for normal Go mode
- `simulation.end_block_height`: stop height for Java-compatible mode
- `simulation.java_compatible`: enable Java SimBlock-compatible behavior
- `network.latency_matrix_file`: latency matrix file path

Use `java_compatible: true` when comparing against the Java version. Use `java_compatible: false` for the normal Go simulation mode.

## 5. Run Tests

Run unit + integrated suite:

```bash
make test
```

Or use the helper script:

```bash
./scripts/run_tests.sh
```

## 6. Java/Go Alignment Check

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

## 7. Output Files

Main output artifacts:

- `output/output.json`
- `output/static.json`
- `output/metrics.json`

These files are generated at runtime and are not tracked by default.
