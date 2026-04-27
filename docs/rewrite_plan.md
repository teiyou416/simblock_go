# SimBlock Go Rewrite Plan

## Scope

### MVP (first phase)
- PoW consensus
- Network topology generation
- `inv -> rec -> block` message flow
- Fork handling and orphan tracking
- Basic propagation statistics

### Out of MVP
- PoS
- Compact Block Relay (CBR)
- Full Java output format compatibility

## Milestones

### M0: Baseline Definition (0.5 day)
Goal: Freeze rewrite scope to avoid requirement drift.

Tasks:
1. Confirm MVP and non-goals.
2. Define core terms: `tip`, `event`, `sim time`.
3. Record acceptance criteria per milestone.

Definition of done:
1. This document is complete and agreed.
2. Every later milestone has a testable outcome.

### M1: Package Skeleton (1 day)
Goal: Lock module boundaries before implementation.

Planned packages:
- `core` (`Task`, `SimTime`)
- `engine` (`Timer`)
- `simulator` (global state and main loop)
- `node`
- `block`
- `consensus/pow`
- `routing/bitcoincore`
- `task/message`
- `network`
- `metrics`
- `config`

Definition of done:
1. All packages compile.
2. `go test ./...` passes.
3. No circular dependencies.

### M2: Block Model Upgrade (1 day)
Goal: Move from height-only chain selection to PoW-ready chain data.

Tasks:
1. Add `ProofOfWorkBlock` with:
   - `difficulty`
   - `totalDifficulty`
   - `nextDifficulty`
2. Keep ancestor lookup and chain relationship checks.

Definition of done:
1. Genesis block can derive next difficulty.
2. Unit tests cover total difficulty accumulation and ancestor traversal.

### M3: PoW Consensus Logic (1 day)
Goal: Implement Java-equivalent fork choice semantics for PoW.

Tasks:
1. Define consensus interface:
   - `MintingTask(...)`
   - `IsReceivedBlockValid(...)`
   - `GenesisBlock(...)`
2. Implement PoW mint interval sampling (exponential).
3. Use strict `totalDifficulty` comparison for chain switch.

Definition of done:
1. Fork choice is no longer height/ID based.
2. Tests cover equal-height and different-total-difficulty scenarios.

### M4: Node State Machine (2 days)
Goal: Expand node behavior from simple receive/mint into protocol state machine.

Tasks:
1. Add node state:
   - `tip`
   - `orphans`
   - `mintingTask`
   - `neighbors`
   - `downloadingBlocks`
   - `messageQueue`
2. On better chain:
   - cancel previous minting task
   - switch tip
   - schedule new minting
3. Wire:
   - `receiveBlock`
   - `minting`
   - `sendInv`
   - `receiveMessage`

Definition of done:
1. Node completes `receive -> re-mint -> advertise` loop.
2. Tests verify old mint task cancellation on reorg.

### M5: Topology and Network Model (1 day)
Goal: Make propagation graph-aware.

Tasks:
1. Implement BitcoinCore-like routing table (inbound/outbound).
2. Use latency and bandwidth model in message delay calculations.

Definition of done:
1. Nodes only propagate to neighbors.
2. Tests verify hop count impacts end-to-end delay.

### M6: Message Task Pipeline (2 days)
Goal: Complete the minimal protocol pipeline.

Tasks:
1. Implement message tasks:
   - `InvMessageTask`
   - `RecMessageTask`
   - `BlockMessageTask`
2. Add sender-side queueing (bandwidth serialization).
3. Transmission delay model:
   - `latency + size / bandwidth + processing`

Definition of done:
1. Observable event chain: `inv -> rec -> block -> receiveBlock`.
2. Tests verify duplicate `inv` does not trigger duplicate download.

### M7: Simulator Main Loop (1 day)
Goal: Run full simulation from `main`.

Tasks:
1. Build nodes from config:
   - region distribution
   - degree distribution
   - mining power distribution
2. Select genesis minter.
3. Start timer loop and stop on:
   - `end_block_height`, or
   - `end_time`

Definition of done:
1. `go run .` executes one full simulation.
2. Works without test-only scaffolding.

### M8: Metrics and Outputs (1 day)
Goal: Produce usable outputs before format parity.

Tasks:
1. Track block propagation arrivals.
2. Track orphan/stale counts.
3. Print summary:
   - main chain height
   - orphan count
   - average propagation time

Definition of done:
1. CLI summary is stable and reproducible with fixed seed.
2. Tests cover propagation metric updates.

### M9: Regression and Stress (1 day)
Goal: Stabilize rewrite for iteration.

Tasks:
1. Add integration tests (4-node and 20-node fixed-seed scenarios).
2. Run `go test -race ./...`.

Definition of done:
1. Integration tests pass consistently.
2. No race conditions.

## Post-MVP (Parity Phase)

### M10+: Optional Parity Work
1. Implement CBR:
   - `CmpctBlockMessageTask`
   - `GetBlockTxnMessageTask`
2. Align output artifacts with Java:
   - `output.json`
   - `static.json`
   - `graph/*.txt`
   - `blockList.txt`
3. Add PoS sample mode.

## Immediate Next Steps
1. Finish `M0` confirmation.
2. Start `M1` package skeleton.
3. Implement `M2` and make `totalDifficulty` fork choice a blocking test gate.
