# Security review

This document records demonstrated results separately from hypotheses. “Observed” means runtime
evidence exists; “detected” means a rule emitted a finding; “rejected” means the final policy
decision was `REJECT`.

## Kernel/security reviewer

Interim review found that kernel-reported parent keys could produce graph edges whose parent node
was absent from the observation interval without an explicit uncertainty marker. Graph schema
v0.2 fixes this by labelling each parent `observed`, `unobserved`, or `none`; unit and privileged
graph assertions cover the distinction. This improves claim precision but does not recover missing
ancestry or defend against a privileged hostile host.

## Controlled malware-behavior experiments

MBE-001 emulates a synthetic credential-file read, child execution, artifact staging, an
executable-renaming variation, and a fixed localhost-only callback. It is not malware. Both reads
and the numeric callback attempt were observed and detected, the secret value was absent from
evidence, process attribution was retained, and policy rejected the bundle with
`RPF-SENSITIVE-001` and `RPF-EGRESS-001`. See
[MBE-001](lab/MBE-001-sensitive-read-and-child.md).

## Intentionally vulnerable fixtures

EXP-001 provides a single-request localhost service whose explicitly vulnerable mode trusts
attacker-controlled adapter-role metadata. The paired patched mode derives authorization from the
synthetic session role. Both are repository-owned, disposable, and use only fixed synthetic data.

## Exploitability-confirmed findings

EXP-001 exists and was dynamically exploitable: the fixed proof crossed the synthetic reader/admin
boundary and obtained a harmless marker. See
[EXP-001](lab/EXP-001-adapter-role-confusion.md). Telemetry observed the proof but did not identify
the authorization semantic; policy rejection came from undeclared egress.

## Exploitability-rejected hypotheses

None yet; no exploit hypothesis has completed the required dynamic validation cycle.

## Authorization-boundary experiments

EXP-001 vulnerable mode granted a synthetic admin marker based on adapter metadata. A separate
patched build denied identical input. Each run has an independent cgroup, identity, evidence chain,
graph, artifact, provenance, and runtime trace. The responsible proof process and connection were
attributed correctly. The detector cannot distinguish grant from denial; this is a confirmed blind
spot, not a negative exploitability result.

## Detection-evasion experiments

Renaming `/bin/sh` to the generated local path `build/out/rpf-renamed-shell` did not bypass the
sensitive-path finding. Telemetry observed the renamed executable, detection emitted the same
`RPF-SENSITIVE-001`, and policy remained `REJECT`. This tests one representation change only and
does not establish general evasion resistance.

The same numeric-connect signal is exercised negatively and positively: undeclared ports 18080
and 18081 produce `RPF-EGRESS-001`, while policy-declared port 18082 remains finding-free in the
benign baseline. This validates exact endpoint semantics only, not domain, proxy, or IPv6 handling.

The kernel baseline now feeds permanent stream-integrity regressions. Interior deletion and
byte modification are rejected by sequence/hash validation. Tail truncation is structurally a
valid hash-chain prefix, so the lifecycle invariant—not the chain—forces `unknown` completeness
and strict `REJECT`. A signed external checkpoint remains necessary to detect rollback to another
complete historical stream.

The signing path rejects modified or malformed attestation bytes and rejects a genuine bundle
under an unrelated generated public key. Separately, the verifier CLI rejects malformed SLSA
provenance and a provenance subject with a wrong artifact digest without creating an output
bundle. This demonstrates local fixture rejection, not keyless identity or transparency-log
verification; the current lab intentionally uses an offline key and skips tlog verification.
The baseline now invokes one fail-closed offline entry point that verifies both signed blobs before
semantic correlation; signature failure prevents the policy verifier from running. This closes the
local orchestration gap but does not supply keyless workload identity or transparency freshness.

Runtime replay is exercised with two independently captured, complete kernel event streams. A
bundle from build B presented with build A's artifact, events, and provenance produces explicit
identity, manifest, and runtime-trace integrity reasons and `REJECT`; build B provenance presented
with build A events fails assembly. This establishes cross-build mismatch handling for the local
identity profile, not freshness against a trusted external clock or transparency checkpoint.

## Exploit-to-telemetry correlation

For MBE-001, the non-exploit chain is: fixture shell input → successful `openat` → kernel
entry/exit correlation → categorized event → process/graph edge → `RPF-SENSITIVE-001` → `REJECT`.

For EXP-001: fixed proof input → vulnerable role selection → synthetic marker grant → proof
process/connect observation → graph edge → undeclared-egress finding → `REJECT`. The vulnerability
was exploitable and observable, but the authorization violation itself was not detected.

## Patch validation

EXP-001 patched mode was rerun as an independent monitored build with the exact proof and denied
the marker while returning a valid response and still producing its benign test artifact. Kernel
telemetry remained present. Benign authorization functionality is unit-tested for the trusted
admin role through the common authorization function. Parser resource limits and artifact
substitution checks remain defensive regressions, not this target's remediation.

## Claim matrix

| Experiment | Vulnerability exists | Exploit demonstrated | Telemetry observed | Detection identified | Policy rejected | Remediation validated |
| --- | --- | --- | --- | --- | --- | --- |
| MBE-001 baseline | not applicable | not applicable | yes | yes | yes | not applicable |
| MBE-001 renamed shell | not applicable | not applicable | yes | yes | yes | not applicable |
| MBE-001 localhost callback | not applicable | not applicable | yes | yes | yes | not applicable |
| EXP-001 vulnerable role confusion | yes | yes | yes | no (authorization semantic) | yes (egress reason) | not applicable |
| EXP-001 patched role handling | no for original flaw | original proof rejected | yes | no (authorization semantic) | yes (egress reason) | yes |
