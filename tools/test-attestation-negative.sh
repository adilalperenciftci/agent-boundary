#!/bin/sh
set -eu

validator=${1:-build/out/rpf}
artifact=${2:-build/out/sensor-test-artifact.txt}
events=${3:-build/out/sensor-test.jsonl}
provenance=${4:-build/out/sensor-test-provenance.json}
policy=${5:-lab/kernel/policy.json}
malformed=build/out/provenance.malformed.json
wrong_digest=build/out/provenance.wrong-digest.json
output=build/out/negative-attestation-bundle

cleanup() {
  rm -f "$malformed" "$wrong_digest"
  rm -rf "$output"
}
trap cleanup EXIT INT TERM
cleanup

printf '{' >"$malformed"
status=0
"$validator" assemble --artifact "$artifact" --events "$events" --provenance "$malformed" \
  --policy "$policy" --output "$output" >/dev/null 2>&1 || status=$?
test "$status" -eq 4
test ! -e "$output"

sed '0,/"sha256":"[0-9a-f]*"/s//"sha256":"0000000000000000000000000000000000000000000000000000000000000000"/' \
  "$provenance" >"$wrong_digest"
status=0
"$validator" assemble --artifact "$artifact" --events "$events" --provenance "$wrong_digest" \
  --policy "$policy" --output "$output" >/dev/null 2>&1 || status=$?
test "$status" -eq 4
test ! -e "$output"

printf '%s\n' '{"malformed_provenance":"REJECTED","wrong_artifact_digest":"REJECTED"}'
