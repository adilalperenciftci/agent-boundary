# Runtime Provenance Firewall

Runtime Provenance Firewall is a security research project about one question: can evidence of
what a software build did at runtime be cryptographically bound to the artifact and provenance
that claim to describe that build?

SLSA provenance identifies source, builder, declared inputs, and artifact digests. It does not
normally state which child processes executed, whether an install script opened a credential
location, or which network destinations were contacted. Linux runtime tools observe much of
that behavior, but observation alone does not prove that a trace, artifact, provenance
statement, CI run, and policy belong to the same execution.

This project studies that correlation and verification gap. It is not another general syscall
logger and does not claim to replace SLSA, in-toto, Sigstore, Tetragon, Tracee, Falco, or
cicd-sensor.

## Current state

The repository contains a tested, platform-independent first vertical slice:

```text
synthetic canonical runtime events
        -> verified event chain and loss state
        -> deterministic execution graph
        -> artifact digest commitment
        -> in-toto Runtime Trace v0.1 correlation extension
        -> SLSA v1 identity/digest equality checks
        -> deterministic policy
        -> ALLOW / REVIEW / REJECT
```

This slice uses synthetic events and has assurance level `fixture`. It does **not** yet collect
kernel telemetry or verify Sigstore signatures. `verify-fixture` is named to prevent unsigned
fixture verification from being confused with the later strict signed verifier.

Implemented invariants include:

- artifact bytes must match provenance, evidence, and Runtime Trace subjects;
- build and CI run identity must agree across runtime evidence and SLSA provenance;
- event order and hash-chain integrity must verify;
- execution graph and evidence manifest are recomputed, not trusted;
- non-zero or unknown event loss cannot produce `ALLOW`;
- missing artifact-finalization/process attribution fails closed;
- policy findings carry stable machine-readable reason codes.

## Reproduce the current slice

Go 1.27 or newer is required for the Go research core. Python 3.12 and `uv` remain temporarily
required for the original decision-engine tests while that code is migrated or retired.

```console
go test ./...
go vet ./...
uv run --extra dev --locked ruff check .
uv run --extra dev --locked pyright
uv run --locked python -m unittest discover -s tests -v
```

The disk round-trip demonstration is:

```console
go test ./internal/rpf -run TestBundleDiskRoundTrip -v
```

The test constructs synthetic build evidence, writes the graph/manifest/Runtime Trace bundle,
reloads it through strict parsers, recomputes every binding, and requires `ALLOW`. Adjacent
tests alter the artifact and manifest, inject event loss, and emulate forbidden sensitive-file
access and localhost egress; those paths must reject.

## Intended architecture

The selected target is a narrow BPF CO-RE sensor, a Go collector/graph builder, in-toto
Runtime Trace plus SLSA provenance, standard Sigstore bundles, and a portable fail-closed
verifier. The sensor will filter by cgroup and emit only semantically useful execution,
sensitive-file, output-file, network, privilege, lifecycle, and loss events.

The build cgroup is treated as adversarial. The monitor must start outside it. A root-equivalent
host attacker that can disable kernel telemetry and reach signing authority is explicitly not
solved.

See [architecture](docs/architecture.md), [threat model](docs/threat-model.md),
[trust model](docs/trust-model.md), and the
[2026 landscape review](docs/research/landscape-2026.md).

## Research result so far

The original broad thesis was narrowed after source-level review. `cicd-sensor` already
provides CI-focused eBPF telemetry, ancestry-aware detection, loss counters, and Runtime Trace
predicate output. The remaining hypothesis is whether a loss-aware verification profile can
prove all required artifact/evidence/provenance/CI identity equalities without silently
reconciling missing data. See the [gap analysis](docs/research/runtime-attestation-gap.md).

## Limits

- No kernel sensor has been implemented or validated yet.
- No signature or transparency-log verification is implemented yet.
- Runtime Trace v0.1 is experimental and monitor event fields are not standardized.
- Async eBPF cannot prove atomic file-content identity at access time.
- Process/file observations establish documented edges, not semantic causation.
- No performance, detection-rate, SLSA level, or production-readiness claim is made.

All adversarial work is restricted to synthetic fixtures, localhost, repository-controlled
containers/VMs, and explicitly authorized systems. See [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
