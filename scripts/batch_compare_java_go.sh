#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

RUNS="${1:-10}"
JAVA_HOME_DEFAULT="/opt/homebrew/opt/openjdk@11"
JAVA_HOME="${JAVA_HOME:-$JAVA_HOME_DEFAULT}"
export JAVA_HOME
export PATH="$JAVA_HOME/bin:$PATH"

OUT_DIR="scripts/results"
mkdir -p "$OUT_DIR"
TS="$(date +%Y%m%d_%H%M%S)"
CSV="$OUT_DIR/java_go_compare_${TS}.csv"
SUMMARY_JSON="$OUT_DIR/java_go_compare_${TS}_summary.json"

cat > "$CSV" <<'CSV_HEADER'
run,go_add_block,java_add_block,go_add_link,java_add_link,go_add_node,java_add_node,go_flow_block,java_flow_block,go_total_events,java_total_events,go_end_ts,java_end_ts,flow_rel_err,total_len_rel_err,end_ts_rel_err,first_diff_index
CSV_HEADER

for i in $(seq 1 "$RUNS"); do
  echo "[run $i/$RUNS] run Go"
  go run ./cmd/simblock >/tmp/go_align_run_${i}.log 2>&1

  echo "[run $i/$RUNS] run Java"
  (
    cd simblock
    ./gradlew :simulator:run >/tmp/java_align_run_${i}.log 2>&1
  )

  python3 - "$i" "$CSV" <<'PY'
import collections
import json
import sys

run = int(sys.argv[1])
csv_path = sys.argv[2]

j = json.load(open('simblock/simulator/src/dist/output/output.json'))
g = json.load(open('output/output.json'))

jc = collections.Counter(e['kind'] for e in j)
gc = collections.Counter(e['kind'] for e in g)

def rel_err(go, java):
    if java == 0:
        return 0.0 if go == 0 else 1.0
    return abs(go - java) / java

def end_ts(events):
    for e in reversed(events):
        if e.get('kind') == 'simulation-end':
            return int(e['content']['timestamp'])
    return -1

java_ts = end_ts(j)
go_ts = end_ts(g)
flow_rel = rel_err(gc['flow-block'], jc['flow-block'])
len_rel = rel_err(len(g), len(j))
end_rel = rel_err(go_ts, java_ts) if java_ts > 0 else 1.0

first_diff = -1
for idx, (je, ge) in enumerate(zip(j, g)):
    if je != ge:
        first_diff = idx
        break
if first_diff == -1 and len(j) != len(g):
    first_diff = min(len(j), len(g))

row = [
    run,
    gc['add-block'], jc['add-block'],
    gc['add-link'], jc['add-link'],
    gc['add-node'], jc['add-node'],
    gc['flow-block'], jc['flow-block'],
    len(g), len(j),
    go_ts, java_ts,
    f"{flow_rel:.9f}",
    f"{len_rel:.9f}",
    f"{end_rel:.9f}",
    first_diff,
]
with open(csv_path, 'a', encoding='utf-8') as f:
    f.write(','.join(map(str, row)) + '\n')

print(f"run={run} flow {gc['flow-block']}/{jc['flow-block']} rel={flow_rel:.6f}; total {len(g)}/{len(j)} rel={len_rel:.6f}; end_ts {go_ts}/{java_ts} rel={end_rel:.6f}; first_diff={first_diff}")
PY

done

python3 - "$CSV" "$SUMMARY_JSON" <<'PY'
import csv
import json
import statistics
import sys

csv_path = sys.argv[1]
summary_path = sys.argv[2]
rows = list(csv.DictReader(open(csv_path, encoding='utf-8')))

flow_rel = [float(r['flow_rel_err']) for r in rows]
len_rel = [float(r['total_len_rel_err']) for r in rows]
end_rel = [float(r['end_ts_rel_err']) for r in rows]
first_diff = [int(r['first_diff_index']) for r in rows]

summary = {
    'runs': len(rows),
    'flow_rel_err': {
        'mean': statistics.mean(flow_rel),
        'min': min(flow_rel),
        'max': max(flow_rel),
    },
    'total_len_rel_err': {
        'mean': statistics.mean(len_rel),
        'min': min(len_rel),
        'max': max(len_rel),
    },
    'end_ts_rel_err': {
        'mean': statistics.mean(end_rel),
        'min': min(end_rel),
        'max': max(end_rel),
    },
    'first_diff_index': {
        'mean': statistics.mean(first_diff),
        'min': min(first_diff),
        'max': max(first_diff),
    },
}

with open(summary_path, 'w', encoding='utf-8') as f:
    json.dump(summary, f, ensure_ascii=False, indent=2)

print('\n=== summary ===')
print(json.dumps(summary, ensure_ascii=False, indent=2))
print(f"csv={csv_path}")
print(f"summary_json={summary_path}")
PY
