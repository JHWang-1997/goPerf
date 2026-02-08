#!/usr/bin/env bash
# 一键端到端测试：编译 → 启动测试桩 → 运行压测 → 清理退出
# 用法: ./scripts/run_e2e.sh  或  bash scripts/run_e2e.sh

set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[e2e] project root: $ROOT"
echo "[e2e] building goPerf and testserver..."
go build -o goPerf ./cmd/
go build -o testserver ./cmd/testserver/

PORT=8080
echo "[e2e] starting testserver on port $PORT..."
./testserver -port "$PORT" &
TS_PID=$!

cleanup() {
  echo "[e2e] stopping testserver (PID $TS_PID)..."
  kill "$TS_PID" 2>/dev/null || true
  wait "$TS_PID" 2>/dev/null || true
}
trap cleanup EXIT

echo "[e2e] waiting for testserver to be ready..."
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -s -o /dev/null "http://127.0.0.1:$PORT/hello" 2>/dev/null; then
    echo "[e2e] testserver is ready."
    break
  fi
  if [ "$i" -eq 10 ]; then
    echo "[e2e] ERROR: testserver did not become ready in time."
    exit 1
  fi
  sleep 0.5
done

echo "[e2e] running goPerf (config=e2e.properties, workers=2, duration=5s, report=1s)..."
echo "----------------------------------------"
./goPerf -f e2e.properties -n 2 -t 5 -rt 1
EXIT=$?
echo "----------------------------------------"
echo "[e2e] goPerf exited with code $EXIT"
exit $EXIT
