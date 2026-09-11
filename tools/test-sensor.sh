#!/bin/sh
set -eu

object=${1:-build/out/rpf-sensor.bpf.o}
binary=${2:-build/out/rpf-sensor}
output=${3:-build/out/sensor-test.jsonl}
validator=${4:-build/out/rpf}
if ! mountpoint -q /sys/kernel/tracing; then
  mount -t tracefs tracefs /sys/kernel/tracing
fi
cgroup_id=$(stat -c %i /sys/fs/cgroup)
boot_id=$(cat /proc/sys/kernel/random/boot_id)
cgroup_path_hash=sha256:$(sha256sum /proc/self/cgroup | cut -d ' ' -f 1)
rm -f "$output"

status=0
timeout --signal=INT 4 "$binary" --object "$object" --cgroup-id "$cgroup_id" \
  --build-id rpf-sensor-smoke --run-id local-container-smoke --boot-id "$boot_id" \
  --cgroup-path-hash "$cgroup_path_hash" --output "$output" &
sensor_pid=$!
sleep 1
/usr/bin/id >/dev/null
/bin/echo rpf-synthetic-exec >/dev/null
wait "$sensor_pid" || status=$?

if [ "$status" -ne 0 ] && [ "$status" -ne 124 ] && [ "$status" -ne 130 ]; then
  echo "sensor exited unexpectedly: $status" >&2
  exit "$status"
fi
grep -q '"operation":"sensor_started"' "$output"
grep -q '"path":"/usr/bin/id"' "$output"
grep -q '"path":"/bin/echo"' "$output"
grep -q '"kernel_reserve":0' "$output"
grep -q '"operation":"sensor_finalized"' "$output"
"$validator" validate-events --events "$output"
before=$(sha256sum "$output")
if "$binary" --object "$object" --cgroup-id "$cgroup_id" \
  --build-id rpf-sensor-smoke --run-id local-container-smoke --boot-id "$boot_id" \
  --cgroup-path-hash "$cgroup_path_hash" --output "$output" >/dev/null 2>&1; then
  echo "sensor overwrote an existing evidence stream" >&2
  exit 1
fi
test "$before" = "$(sha256sum "$output")"
cat "$output"
