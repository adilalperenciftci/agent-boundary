# Testing strategy

Every evaluator requires a benign baseline, suspicious fixture, non-detection case, malformed-input cases, and expected findings. Fixtures are deterministic and use `.test` domains and synthetic credentials.

The initial suite verifies allow/deny behavior, exact-host matching, redaction, duplicate-key rejection, depth limits, undeclared nested-field rejection, ledger verification, and refusal to extend a modified ledger.

Unit and vertical-slice tests run without network access or third-party services. Rule metadata validation ensures fixture references exist and implemented IDs match documented IDs. CLI integration tests verify exit status and durable output. Property-based parsing tests are deferred until a dependency is justified; parser regression fixtures should be added for every discovered ambiguity.

No false-positive or false-negative rate is reported from these curated fixtures. Counts from deterministic fixtures describe coverage, not field performance.

## Privileged sensor tests

`tools/test-sensor.sh` runs only in a disposable privileged Linux environment. The sensor stays
outside a temporary fixture cgroup. Fixed `/usr/bin/id` and `/bin/echo` processes move into that
cgroup before exec and must be present, while a `/usr/bin/whoami` control executed in the sensor's
cgroup must be absent. The test also requires `sensor_finalized=true` and `ringbuf_drops=0`, so a
lossy run cannot pass as clean. This proves exact-ID filtering in the tested namespace layout,
not exclusion under every cgroup namespace/delegation arrangement or inclusion of nested cgroups.

The smoke test passes the resulting stream through `rpf validate-events`, then attempts to reuse
the same evidence path. The second collector must fail and the file digest must remain unchanged.
It also executes a target-cgroup shell that opens an absolute synthetic artifact for writing,
requires the final digest to match the file, reconstructs both output-open and artifact edges,
and requires every implemented loss counter to be zero.
The same test creates full-structure unsigned local SLSA provenance, assembles Runtime Trace and
manifest commitments, requires an `ALLOW`, modifies the artifact bytes, and requires verifier
exit status 3 with `RPF-ARTIFACT-001`. This demonstrates correlation and tamper rejection, not
provenance authenticity or signature verification.

`tools/test-signing.sh` generates a synthetic ephemeral Cosign key, signs exact provenance and
Runtime Trace bytes into standardized bundles, verifies both against the public key, and requires
a one-byte statement modification to fail. The private key is removed on exit. The local config
has no Rekor/TSA services, so this test deliberately does not satisfy production transparency or
trusted-time requirements.

`tools/test-adversarial.sh` runs MBE-001 with a repository-owned fake credential. It requires two
category findings (ordinary and renamed shell), observed shell-to-child ancestry, no fixture value
in evidence, all loss counters zero, and final `REJECT`. This is behavior emulation, not malware
execution or a field detection-rate benchmark.
`tools/kernel-lab.sh` builds the checked-in pinned-base lab image and runs BPF compilation, Go
race tests, vet, both binaries, and the privileged smoke test. Debian packages installed into
that image are not yet snapshot-pinned, so the image build is repeatable but not byte-reproducible.

The wire decoder has unprivileged exact-size and bounded-string unit tests. Linux CI must build
and vet all Go packages; privileged kernel coverage is a separate gate because ordinary hosted
CI does not provide equivalent eBPF semantics.
