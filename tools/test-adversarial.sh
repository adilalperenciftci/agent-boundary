#!/bin/sh
set -eu

object=${1:-build/out/rpf-sensor.bpf.o}
sensor=${2:-build/out/rpf-sensor}
verifier=${3:-build/out/rpf}
evidence=build/out/adversarial-events.jsonl
graph=build/out/adversarial-graph.json
artifact=/src/build/out/adversarial-artifact.txt
credential=/src/lab/fixtures/synthetic-credential.txt
renamed=/src/build/out/rpf-renamed-shell
provenance=build/out/adversarial-provenance.json
bundle=build/out/adversarial-bundle
policy=lab/kernel/policy.json

if ! mountpoint -q /sys/kernel/tracing; then
  mount -t tracefs tracefs /sys/kernel/tracing
fi
fixture_cgroup=/sys/fs/cgroup/rpf-adversarial-test-$$
mkdir "$fixture_cgroup"
cleanup() {
  rm -f "$renamed"
  rmdir "$fixture_cgroup" 2>/dev/null || true
}
trap cleanup EXIT INT TERM
cgroup_id=$(stat -c %i "$fixture_cgroup")
boot_id=$(cat /proc/sys/kernel/random/boot_id)
cgroup_path_hash=sha256:$(printf '%s' "$fixture_cgroup" | sha256sum | cut -d ' ' -f 1)
rm -f "$evidence" "$graph" "$artifact" "$provenance"
cp /bin/sh "$renamed"
chmod 0700 "$renamed"
if [ -d "$bundle" ]; then
  rm -f "$bundle/execution-graph.json" "$bundle/evidence-manifest.json" "$bundle/runtime-trace.json"
  rmdir "$bundle"
fi

status=0
timeout --signal=INT 4 "$sensor" --object "$object" --cgroup-id "$cgroup_id" \
  --build-id rpf-adversarial-sensitive --run-id local-adversarial-sensitive --boot-id "$boot_id" \
  --cgroup-path-hash "$cgroup_path_hash" --artifact "$artifact" --output "$evidence" \
  --sensitive-path "$credential" --sensitive-category synthetic_credential &
sensor_pid=$!
sleep 1
/bin/sh -c 'echo $$ > "$1/cgroup.procs"; exec /bin/sh -c "IFS= read -r ignored < \"$2\"; /usr/bin/id; printf rpf-adversarial-artifact > \"$3\""' \
  sh "$fixture_cgroup" "$credential" "$artifact"
/bin/sh -c 'echo $$ > "$1/cgroup.procs"; exec "$2" -c "IFS= read -r ignored < \"$3\""' \
  sh "$fixture_cgroup" "$renamed" "$credential"
wait "$sensor_pid" || status=$?
if [ "$status" -ne 0 ] && [ "$status" -ne 124 ] && [ "$status" -ne 130 ]; then
  echo "adversarial sensor exited unexpectedly: $status" >&2
  exit "$status"
fi

"$verifier" validate-events --events "$evidence"
grep -q '"operation":"file_open_sensitive"' "$evidence"
grep -q '"category":"synthetic_credential"' "$evidence"
grep -q '"path":"/usr/bin/id"' "$evidence"
grep -q '"path":"/src/build/out/rpf-renamed-shell"' "$evidence"
test "$(grep -c '"operation":"file_open_sensitive"' "$evidence")" -eq 2
if grep -q 'not-a-real-secret' "$evidence"; then
  echo "synthetic credential value leaked into evidence" >&2
  exit 1
fi
grep -q '"decode":0' "$evidence"
grep -q '"kernel_correlation":0' "$evidence"
grep -q '"kernel_reserve":0' "$evidence"

"$verifier" graph-events --events "$evidence" --output "$graph"
grep -q '"kind":"file_open_sensitive"' "$graph"
"$verifier" create-local-provenance --artifact "$artifact" --events "$evidence" \
  --repository https://example.test/agent-boundary --revision 1111111111111111111111111111111111111111 \
  --output "$provenance"
"$verifier" assemble --artifact "$artifact" --events "$evidence" --provenance "$provenance" \
  --policy "$policy" --output "$bundle"
decision_status=0
decision_output=$("$verifier" verify-fixture --artifact "$artifact" --events "$evidence" --provenance "$provenance" \
  --policy "$policy" --bundle "$bundle") || decision_status=$?
printf '%s\n' "$decision_output"
if [ "$decision_status" -ne 3 ]; then
  echo "synthetic credential access did not produce REJECT" >&2
  exit 1
fi
printf '%s' "$decision_output" | grep -q 'RPF-SENSITIVE-001'
