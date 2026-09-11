# MBE-001: synthetic sensitive-file read and child execution

## Classification and scope

This is malware-behavior emulation, not deployable malware. It has no scanning, propagation,
persistence, Internet communication, credential collection, destructive action, or host-targeting
logic. It runs in a disposable cgroup, reads only
`lab/fixtures/synthetic-credential.txt`, and writes only generated files under `build/out`.

## Intended observations

The target-cgroup `/bin/sh` must successfully open the configured synthetic credential path,
launch `/usr/bin/id`, and write the synthetic artifact. Telemetry must retain category
`synthetic_credential`, process identity and ancestry, but never the fixture value. A second run
inside the same monitored interval executes a copied shell named
`build/out/rpf-renamed-shell` and repeats the read to test representation robustness.

Expected policy result: `REJECT` with `RPF-SENSITIVE-001`. The renamed executable may add
`RPF-PROCESS-001` but must not suppress the sensitive-access finding.

## Reproduction and actual result

Run `./tools/test-adversarial.sh` in the checked-in kernel lab image. On 2026-09-11, Linux 6.8
WSL2 produced a valid nine-event stream, three graph nodes, seven graph edges, two
`file_open_sensitive` events, zero implemented loss counters, and:

```json
{"decision":"REJECT","completeness":"complete","reasons":[{"code":"RPF-SENSITIVE-001","effect":"REJECT","sequence":3},{"code":"RPF-PROCESS-001","effect":"REVIEW","sequence":6},{"code":"RPF-SENSITIVE-001","effect":"REJECT","sequence":7}]}
```

The exact message fields and process IDs vary; the script asserts stable reason codes and
security invariants. It also searches the evidence bytes and fails if the synthetic value occurs.

## Evasion result and blind spots

Executable renaming did not bypass category detection because kernel filtering keys on the exact
sensitive path, not executable name. Current blind spots include aliases/symlinks, relative paths,
`openat2`, inherited descriptors, mmap access, path replacement, and reads that begin before the
process has an observed exec identity. The test does not claim that reading a file is malicious;
policy makes this specific synthetic category forbidden for the build.
