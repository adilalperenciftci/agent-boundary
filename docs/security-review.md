# Security review

This document records demonstrated results separately from hypotheses. “Observed” means runtime
evidence exists; “detected” means a rule emitted a finding; “rejected” means the final policy
decision was `REJECT`.

## Controlled malware-behavior experiments

MBE-001 emulates a synthetic credential-file read, child execution, artifact staging, an
executable-renaming variation, and a fixed localhost-only callback. It is not malware. Both reads
and the numeric callback attempt were observed and detected, the secret value was absent from
evidence, process attribution was retained, and policy rejected the bundle with
`RPF-SENSITIVE-001` and `RPF-EGRESS-001`. See
[MBE-001](lab/MBE-001-sensitive-read-and-child.md).

## Intentionally vulnerable fixtures

No intentionally vulnerable service has been implemented or executed yet. No vulnerability or
exploitability claim is made in this section.

## Exploitability-confirmed findings

None yet. MBE-001 validates behavior detection, not exploitation of a vulnerability.

## Exploitability-rejected hypotheses

None yet; no exploit hypothesis has completed the required dynamic validation cycle.

## Authorization-boundary experiments

Not yet executed. No unauthorized-access detection claim is made.

## Detection-evasion experiments

Renaming `/bin/sh` to the generated local path `build/out/rpf-renamed-shell` did not bypass the
sensitive-path finding. Telemetry observed the renamed executable, detection emitted the same
`RPF-SENSITIVE-001`, and policy remained `REJECT`. This tests one representation change only and
does not establish general evasion resistance.

## Exploit-to-telemetry correlation

No exploit chain has been confirmed. For MBE-001, the non-exploit chain is: fixture shell input →
successful `openat` → kernel entry/exit correlation → categorized event → process/graph edge →
`RPF-SENSITIVE-001` → `REJECT`.

## Patch validation

No vulnerable target patch cycle has been completed yet. Parser resource limits and artifact
substitution checks are regression-tested defensive controls, not remediations of a demonstrated
lab vulnerability.

## Claim matrix

| Experiment | Vulnerability exists | Exploit demonstrated | Telemetry observed | Detection identified | Policy rejected | Remediation validated |
| --- | --- | --- | --- | --- | --- | --- |
| MBE-001 baseline | not applicable | not applicable | yes | yes | yes | not applicable |
| MBE-001 renamed shell | not applicable | not applicable | yes | yes | yes | not applicable |
| MBE-001 localhost callback | not applicable | not applicable | yes | yes | yes | not applicable |
