# SimBlock-Go Implementation Checklist

This checklist is designed for day-to-day execution. Follow stages in order and check items only when the acceptance criteria are met.

## Current Baseline
- Event scheduler and unit tests exist: `engine/timer.go`, `engine/timer_test.go`
- Minimal domain skeleton exists: `block/`, `node/`, `network/`
- Minimal end-to-end loop test exists: `engine/simulation_loop_test.go`
- Build/run baseline is not fully wired yet (entry/config mismatch)

## Stage 0: Buildable and Runnable Baseline
- [ ] Fix `config/simulator.yaml` structure/indentation.
- [ ] Unify entrypoint to `cmd/simblock/main.go`.
- [ ] Fix `Makefile` `MAIN_FILE` to the real entrypoint.
- [ ] Ensure `go test ./...`, `go build ./...`, `make build` all pass.

Acceptance:
- Program starts and successfully loads config via `go run`.

## Stage 1: Freeze Engine Contracts
- [ ] Add tests for same-timestamp FIFO ordering.
- [ ] Add tests for reschedule/overwrite behavior of same task instance.
- [ ] Add tests for cancel/remove behavior.
- [ ] Add tests for absolute-time scheduling behavior.
- [ ] Keep core simulation single-threaded (no goroutines in core path).

Acceptance:
- Scheduler behavior is locked by tests and deterministic under fixed seed.

## Stage 2: Configuration Parity with Java
- [ ] Map Java simulation/network parameters to Go config structs.
- [ ] Support region list, region distribution, degree distribution.
- [ ] Support upload/download bandwidth and latency matrix loading.
- [ ] Add config validation (dimension mismatch, invalid values, missing fields).

Acceptance:
- Go config can express major Java experiment settings.

## Stage 3: Domain Model for Extensibility
- [ ] Expand `Block` for consensus extensibility (parent/height/time/consensus data).
- [ ] Expand `Node` state: mining power, neighbors, orphan set, download set, send queue/state.
- [ ] Define decoupled interfaces: `consensus`, `routing`.

Acceptance:
- Node model can host full PoW and message-protocol state transitions.

## Stage 4: Network and Routing
- [ ] Implement region model (`network/region.go`).
- [ ] Implement latency jitter + bandwidth bottleneck network behavior.
- [ ] Implement routing interface and `BitcoinCoreTable` equivalent.
- [ ] Implement node join/bootstrap and degree-constrained linking.

Acceptance:
- Given fixed seed, topology and propagation timing are reproducible.

## Stage 5: PoW and Mining Tasks
- [ ] Implement consensus interface and PoW implementation.
- [ ] Implement `MiningTask` with exponential interval sampling.
- [ ] Implement block validation (difficulty + total difficulty).
- [ ] On better-chain arrival: cancel old mining task and remint.

Acceptance:
- Competing forks resolve according to chain-selection rules.

## Stage 6: Full Message Protocol Tasks
- [ ] Implement `InvMessageTask`.
- [ ] Implement `RecMessageTask`.
- [ ] Implement `BlockMessageTask`.
- [ ] Implement `CmpctBlockMessageTask`.
- [ ] Implement `GetBlockTxnMessageTask`.
- [ ] Implement send queue + `sendNextBlockMessage` semantics.

Acceptance:
- Both legacy block propagation and compact-block flow work end-to-end.

## Stage 7: Simulator Integration Loop
- [ ] Implement `engine/simulator.go` (nodes, target interval, propagation observation).
- [ ] Implement initialization flow (region/degree/mining power assignment).
- [ ] Implement genesis minter selection by mining power.
- [ ] Implement main loop and stop conditions (`END_BLOCK_HEIGHT` or equivalent).

Acceptance:
- Simulation runs from genesis to termination condition without manual intervention.

## Stage 8: Output and Visualizer Compatibility
- [ ] Output `output.json`.
- [ ] Output `static.json`.
- [ ] Output `graph/*.txt`.
- [ ] Output `blockList.txt`.
- [ ] Align event kinds/fields with Java (`add-node`, `add-block`, `flow-block`, `simulation-end`).
- [ ] Add metrics: propagation delay, orphan rate, average orphan count.

Acceptance:
- Output can be consumed by existing visualizer/analysis tooling.

## Stage 9: Equivalence and Performance
- [ ] Fix random seed behavior for deterministic replay.
- [ ] Run Java and Go with same settings and compare outputs/metrics.
- [ ] Add automated diff scripts for key artifacts.
- [ ] Profile and optimize only after parity (`sync.Pool` optional).

Acceptance:
- Functional parity is demonstrated with repeatable evidence.

## Daily Development Loop
1. Pick 1-2 small tasks from a single stage.
2. Read corresponding Java source before implementation.
3. Write/adjust tests first, then implement.
4. Run `go test ./...` and one small integration run.
5. Record progress by checking boxes in this file and updating notes.

## Suggested Next 3 Tasks
1. Fix baseline run/build path and YAML config.
2. Create consensus/routing interface skeletons with test doubles.
3. Implement first real `MiningTask + BlockMessageTask` path to replace temporary test tasks.
