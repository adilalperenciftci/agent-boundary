#!/bin/sh
set -eu

object=${1:-build/out/rpf-sensor.bpf.o}
binary=${2:-build/out/rpf-sensor}
output=${3:-build/out/sensor-test.jsonl}
if ! mountpoint -q /sys/kernel/tracing; then
  mount -t tracefs tracefs /sys/kernel/tracing
fi
cgroup_id=$(stat -c %i /sys/fs/cgroup)

status=0
timeout --signal=INT 4 "$binary" --object "$object" --cgroup-id "$cgroup_id" > "$output" &
sensor_pid=$!
sleep 1
/usr/bin/id >/dev/null
/bin/echo rpf-synthetic-exec >/dev/null
wait "$sensor_pid" || status=$?

if [ "$status" -ne 0 ] && [ "$status" -ne 124 ] && [ "$status" -ne 130 ]; then
  echo "sensor exited unexpectedly: $status" >&2
  exit "$status"
fi
grep -q '"filename":"/usr/bin/id"' "$output"
grep -q '"filename":"/bin/echo"' "$output"
grep -q '"sensor_finalized":true' "$output"
cat "$output"
