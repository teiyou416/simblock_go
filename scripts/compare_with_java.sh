#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

JAVA_HOME_DEFAULT="/opt/homebrew/opt/openjdk@11"
JAVA_HOME="${JAVA_HOME:-$JAVA_HOME_DEFAULT}"
export JAVA_HOME
export PATH="$JAVA_HOME/bin:$PATH"

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
    if a!=b:
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
