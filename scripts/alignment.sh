#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

RUNS=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --runs)
      if [[ $# -lt 2 ]]; then
        echo "error: --runs requires a number" >&2
        exit 2
      fi
      RUNS="$2"
      shift 2
      ;;
    *)
      echo "error: unknown argument: $1" >&2
      echo "usage: ./scripts/alignment.sh [--runs N]" >&2
      exit 2
      ;;
  esac
done

if ! [[ "$RUNS" =~ ^[0-9]+$ ]] || [[ "$RUNS" -lt 1 ]]; then
  echo "error: runs must be a positive integer" >&2
  exit 2
fi

JAVA_HOME_DEFAULT="/opt/homebrew/opt/openjdk@11"
JAVA_HOME="${JAVA_HOME:-$JAVA_HOME_DEFAULT}"
export JAVA_HOME
export PATH="$JAVA_HOME/bin:$PATH"

if [[ "$RUNS" -eq 1 ]]; then
  echo "[1/4] Run Go simulator"
  go run ./cmd/simblock >/tmp/go_align_run.log 2>&1

  echo "[2/4] Run Java simulator"
  (
    cd simblock
    ./gradlew :simulator:run >/tmp/java_full_run.log 2>&1
  )

  echo "[3/4] Count event kinds"
  printf "Go   : "
  jq -r '.[].kind' output/output.json | sort | uniq -c | tr '\n' '; '; echo
  printf "Java : "
  jq -r '.[].kind' simblock/simulator/src/dist/output/output.json | sort | uniq -c | tr '\n' '; '; echo

  echo "[4/4] Tolerance report"
  python3 - <<'PY'
import collections
import json
j = json.load(open('simblock/simulator/src/dist/output/output.json'))
g = json.load(open('output/output.json'))

max_flow_rel = 0.02
max_total_len_rel = 0.02
max_end_ts_rel = 0.70

java_counts = collections.Counter(e["kind"] for e in j)
go_counts = collections.Counter(e["kind"] for e in g)

def rel_err(go, java):
    if java == 0:
        return 0.0 if go == 0 else 1.0
    return abs(go - java) / java

flow_rel = rel_err(go_counts["flow-block"], java_counts["flow-block"])
len_rel = rel_err(len(g), len(j))

def end_ts(events):
    for e in reversed(events):
        if e.get("kind") == "simulation-end":
            return int(e["content"]["timestamp"])
    return -1

java_ts = end_ts(j)
go_ts = end_ts(g)
end_ts_rel = rel_err(go_ts, java_ts) if java_ts > 0 else 1.0

print("metrics:")
print(f"  flow-block  : java={java_counts['flow-block']} go={go_counts['flow-block']} rel_err={flow_rel:.6f}")
print(f"  event_len   : java={len(j)} go={len(g)} rel_err={len_rel:.6f}")
print(f"  simulation_end_ts: java={java_ts} go={go_ts} rel_err={end_ts_rel:.6f}")

for i, (a, b) in enumerate(zip(j, g)):
    if a != b:
        print("first_diff_index=", i)
        print("java=", a)
        print("go  =", b)
        break
else:
    if len(j) == len(g):
        print("No semantic diff found.")
    else:
        print("Prefix equal; lengths differ.")

ok = True
if flow_rel > max_flow_rel:
    ok = False
if len_rel > max_total_len_rel:
    ok = False
if end_ts_rel > max_end_ts_rel:
    ok = False

if ok:
    print("result=PASS (within acceptable tolerance)")
else:
    print("result=FAIL (out of tolerance)")
    raise SystemExit(1)
PY

  echo "done"
  exit 0
fi

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
