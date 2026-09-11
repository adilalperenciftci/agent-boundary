#!/bin/sh
set -eu

object=${1:-build/out/rpf-sensor.bpf.o}
binary=${2:-build/out/rpf-sensor}
output=${3:-build/out/sensor-test.jsonl}
validator=${4:-build/out/rpf}
graph=${5:-build/out/sensor-test-graph.json}
artifact=${6:-/src/build/out/sensor-test-artifact.txt}
if ! mountpoint -q /sys/kernel/tracing; then
  mount -t tracefs tracefs /sys/kernel/tracing
fi
fixture_cgroup=/sys/fs/cgroup/rpf-sensor-test-$$
mkdir "$fixture_cgroup"
cleanup() {
  rmdir "$fixture_cgroup" 2>/dev/null || true
}
trap cleanup EXIT INT TERM
cgroup_id=$(stat -c %i "$fixture_cgroup")
boot_id=$(cat /proc/sys/kernel/random/boot_id)
cgroup_path_hash=sha256:$(printf '%s' "$fixture_cgroup" | sha256sum | cut -d ' ' -f 1)
rm -f "$output"
rm -f "$graph"
rm -f "$artifact"

status=0
timeout --signal=INT 4 "$binary" --object "$object" --cgroup-id "$cgroup_id" \
  --build-id rpf-sensor-smoke --run-id local-container-smoke --boot-id "$boot_id" \
  --cgroup-path-hash "$cgroup_path_hash" --artifact "$artifact" --output "$output" &
sensor_pid=$!
sleep 1
/usr/bin/whoami >/dev/null
/bin/sh -c 'echo $$ > "$1/cgroup.procs"; exec /usr/bin/id' sh "$fixture_cgroup"
/bin/sh -c 'echo $$ > "$1/cgroup.procs"; exec /bin/echo rpf-synthetic-exec' sh "$fixture_cgroup"
/bin/sh -c 'echo $$ > "$1/cgroup.procs"; exec /bin/sh -c "printf rpf-artifact-v1 > \"$2\""' sh "$fixture_cgroup" "$artifact"
wait "$sensor_pid" || status=$?

if [ "$status" -ne 0 ] && [ "$status" -ne 124 ] && [ "$status" -ne 130 ]; then
  echo "sensor exited unexpectedly: $status" >&2
  exit "$status"
fi
grep -q '"operation":"sensor_started"' "$output"
grep -q '"path":"/usr/bin/id"' "$output"
grep -q '"path":"/bin/echo"' "$output"
grep -q '"operation":"file_open_output"' "$output"
grep -q '"operation":"artifact_finalized"' "$output"
artifact_hash=$(sha256sum "$artifact" | cut -d ' ' -f 1)
grep -q '"sha256":"'"$artifact_hash"'"' "$output"
if grep -q '"path":"/usr/bin/whoami"' "$output"; then
  echo "sensor admitted an executable outside the target cgroup" >&2
  exit 1
fi
grep -q '"kernel_reserve":0' "$output"
grep -q '"kernel_correlation":0' "$output"
grep -q '"decode":0' "$output"
grep -q '"operation":"sensor_finalized"' "$output"
"$validator" validate-events --events "$output"
"$validator" graph-events --events "$output" --output "$graph"
grep -q '"executable":"/usr/bin/id"' "$graph"
grep -q '"executable":"/bin/echo"' "$graph"
grep -q '"kind":"observed_exec_parent"' "$graph"
grep -q '"kind":"file_open_output"' "$graph"
grep -q '"kind":"artifact_finalized"' "$graph"
before=$(sha256sum "$output")
if "$binary" --object "$object" --cgroup-id "$cgroup_id" \
  --build-id rpf-sensor-smoke --run-id local-container-smoke --boot-id "$boot_id" \
  --cgroup-path-hash "$cgroup_path_hash" --artifact "$artifact" --output "$output" >/dev/null 2>&1; then
  echo "sensor overwrote an existing evidence stream" >&2
  exit 1
fi
test "$before" = "$(sha256sum "$output")"
cat "$output"
