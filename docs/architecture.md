# Architecture

## Thesis

The system correlates runtime evidence with existing provenance standards. The core result is
not a claim that events were collected; it is a verifier result over explicit digest and
identity equalities.

```text
source revision
      |
CI identity + build nonce
      |
isolated build cgroup <--- host-owned CO-RE sensor
      |                           |
artifact bytes              bounded events + loss counters
      |                           |
artifact SHA-256        canonical stream -> execution graph
      |                           |
SLSA provenance <------ evidence manifest ------> Runtime Trace v0.1
              \             |                    /
               \------ signed Sigstore bundle --/
                              |
                strict correlation + policy verifier
                              |
                    ALLOW / REVIEW / REJECT
```

Solid implementation currently begins at canonical synthetic events and ends at an unsigned
fixture verdict. Kernel collection and Sigstore verification are target components, not
current capabilities.

## Components

### Kernel sensor (planned)

Small BPF C CO-RE programs attach to stable tracepoints, cgroup hooks, and BPF LSM hooks where
their semantics justify deployment cost. Kernel filtering limits events to a registered build
cgroup and policy-selected paths/output roots. Ring-buffer reservation failure increments a
per-build or conservatively attributable loss counter.

### Go collector (planned)

The collector loads CO-RE objects using `cilium/ebpf`, registers a build nonce/cgroup scope,
normalizes fixed kernel records, writes a single-writer append-only stream, and finalizes loss
counters. Loading privilege is separated from parsing/graphing where practical.

### Evidence and graph core (implemented for fixtures)

`internal/rpf` strictly parses canonical JSONL, rejects duplicate/unknown fields, verifies
event order and hash chaining, constructs typed observation edges, and commits graph and
stream digests to an evidence manifest. Process identity combines boot ID, PID namespace,
TGID, and start time; PID alone is never a node key.

### Attestation composer (implemented for fixtures)

The composer emits an in-toto Statement v1 with Runtime Trace v0.1. Standard fields identify
the monitor and run; one namespaced extension commits to build/run/source identity, evidence
manifest, graph, SLSA provenance, policy, and completeness. Detailed events remain external
content-addressed evidence.

### Signature verifier (planned)

Maintained Sigstore libraries will verify bundles, trust roots, certificate identity/issuer,
transparency evidence, and signed statement bytes. The project will not implement cryptographic
primitives or hard-code a Rekor shard.

### Correlation and policy verifier (fixture slice implemented)

The verifier recomputes artifact, stream, graph, manifest, provenance, and policy commitments;
checks cross-document build/run/source identities; evaluates deterministic behavior policy;
and applies precedence `REJECT > REVIEW > ALLOW`. Incomplete evidence is a verification state
that policy cannot upgrade to `ALLOW`.

## Data ownership and ordering

The sensor supplies kernel observations, not policy decisions. The collector owns event
sequence and persistence. Graph construction is pure and replayable. The attestation composer
does not mutate evidence. The signer authenticates immutable statement bytes. The verifier
receives untrusted bytes and recomputes all links under local policy.

## Technology choices

Go plus `cilium/ebpf` was selected over Aya and a pure third-party adapter after comparison in
the [gap analysis](research/runtime-attestation-gap.md). The choice favors mature Go in-toto,
Sigstore, and eBPF ecosystems while keeping BPF code small. Protobuf is deferred until kernel
record semantics have been exercised; versioned canonical JSON prevents premature schema
stability claims in the first slice.

## Failure behavior

Malformed data, duplicate keys, unsupported versions, digest mismatch, identity conflict, and
missing artifact attribution fail closed. Event loss produces `incomplete`; strict policy maps
that to `REVIEW` or `REJECT`. Operational parser errors use a distinct exit status from policy
decisions.

## Privilege and host trust

The future sensor needs BPF-related capabilities and possibly BPF LSM support. After loading,
capabilities should be reduced; evidence output and policy should be read-only to the build.
Seccomp, Landlock, map freezing, and split loader/collector processes will be evaluated against
actual required syscalls. None of these controls defeats hostile root.
