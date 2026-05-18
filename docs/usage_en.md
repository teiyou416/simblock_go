# SimBlock-Go Usage (English)

## 1. Prerequisites

- Go 1.22+ (or the version required by `go.mod`)
- `make`

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
By default `output_mode=core`, only `metrics.txt` is saved (smallest footprint).

## 3. Command-Line Arguments

The simulator executable accepts command-line flags that override the same values from `config/simulator.yaml`.

Supported flags:

- `--config`: YAML config file path
- `--num-nodes`: override `simulation.num_nodes`
- `--block-interval`: override `simulation.block_interval`
- `--block-size`: override `simulation.block_size`
- `--end-time`: override `simulation.end_time`
- `--end-block-height`: override `simulation.end_block_height`
- `--java-compatible`: override `simulation.java_compatible`
- `--output-mode`: override `simulation.output_mode` (`core` or `full`)
- `--network-profile`: override `network.profile`
- `--latency-matrix-file`: override `network.latency_matrix_file`

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
- `simulation.output_mode`: output mode; `core` keeps only key metrics, `full` writes full debug artifacts
- `network.profile`: built-in network profile name, currently `bitcoin_2019`
- `network.latency_matrix_file`: latency matrix file path
- `network.upload_bandwidth`: per-region upload bandwidth in bit/s
- `network.download_bandwidth`: per-region download bandwidth in bit/s
- `network.region_distribution`: per-region node distribution; values should sum to `1.0`
- `network.degree_distribution`: cumulative outbound-link distribution; final value should be `1.0`

Use `java_compatible: false` for the normal Go simulation mode. Keep `java_compatible: true` only when you intentionally want Java-compatible simulation behavior.

## 5. Run Tests

Run unit + integrated suite:

```bash
make test
```

Or use the helper script:

```bash
./scripts/run_tests.sh
```

## 6. Output Files

Main output artifacts:

- `output/<run_id>/metrics.txt` (default, `output_mode=core`)
- `output/<run_id>/output.txt` (only in `output_mode=full`)
- `output/<run_id>/static.txt` (only in `output_mode=full`)
- `output/<run_id>/chain_tree.txt` (only in `output_mode=full`)

These files are generated at runtime and are not tracked by default.
