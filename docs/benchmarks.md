# Benchmarks

Benchmark claims are limited to commands and results recorded here. The microbenchmarks use a
five-event canonical fixture and therefore measure parser/correlation machinery, not kernel sensor
overhead or realistic build scale.

## Reproduction

```sh
go test ./internal/rpf -run '^$' -bench 'Benchmark(ParseEventStream|Assemble|Verify)$' \
  -benchmem -count=5
```

Record the Go version, CPU model, operating environment, fixture event/byte count, and every raw
sample. Compare distributions from the same machine; do not compare a WSL2 sample directly with a
native Linux runner. A benchmark regression gate is not enabled until variance and representative
event-stream sizes are characterized.

## Current measurement status

| Metric | Status | Reason |
| --- | --- | --- |
| canonical event parse latency/allocations | benchmark implemented | five-event fixture only |
| bundle assembly latency/allocations | benchmark implemented | cryptographic signing excluded |
| policy verification latency/allocations | benchmark implemented | local fixture profile only |
| ring-buffer event throughput/loss | not measured | requires a controlled kernel load generator |
| sensor CPU and memory overhead | not measured | no paired monitored/unmonitored harness yet |
| build-time overhead | not measured | no statistically repeated representative builds yet |
| Cosign/keyless latency | not measured | current lab uses offline keys and no transparency log |
| controlled false positives/negatives | scenario assertions only | corpus is too small for a rate claim |

Measured values must not be promoted to README performance claims without the corresponding raw
results and environment metadata. Event-loss counters in functional tests establish zero observed
loss for those runs only; they are not throughput measurements.

## Recorded sanity sample

On 2026-09-11, one non-comparative sanity sample was run with Go 1.27.1 in the pinned lab
container on Linux `6.18.33.2-microsoft-standard-WSL2`, amd64, Intel i5-11400H. The fixture was
five events / 4,462 bytes. This single sample confirms the harness executes; it is not a stable
performance baseline.

| Operation | ns/op | bytes allocated/op | allocations/op |
| --- | ---: | ---: | ---: |
| parse event stream | 1,231,373 | 329,761 | 4,884 |
| assemble bundle | 1,805,237 | 389,625 | 5,869 |
| verify bundle | 3,521,905 | 807,724 | 12,103 |

The exact command was the reproduction command with `-count=1`. A future baseline must retain the
five raw samples, characterize variance, and use larger representative streams before setting a
regression threshold.
