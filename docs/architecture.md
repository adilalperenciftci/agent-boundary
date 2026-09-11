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

Solid implementation currently includes the fixture correlation path plus a separately tested
cgroup-filtered exec sensor. Converting raw sensor records into canonical build evidence and
Sigstore verification remain target components.

## Components

### Kernel sensor (M2 exec slice implemented)

The first BPF C CO-RE program attaches to `sched_process_exec`, compares
`bpf_get_current_cgroup_id()` with one loader-supplied cgroup ID, and emits bounded identity and
filename records through a BPF ring buffer. Reservation failure increments a per-CPU counter;
userspace sums and emits it during finalization. This slice deliberately does not claim file,
network, namespace, process-start-time, or build-nonce attribution yet.

### Go sensor loader (M2 implemented; canonical collector planned)

`cmd/rpf-sensor` loads the CO-RE object using `cilium/ebpf`, rewrites the target cgroup constant,
attaches the tracepoint, defensively decodes fixed-size records, and emits JSON lines plus the
final loss count. It is a diagnostic boundary, not the canonical evidence collector. Build nonce
registration, composite identities, append-only persistence, and privilege separation remain.

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

The sensor needs BPF/perfmon-style privileges appropriate to the host configuration; the tested
lab uses a privileged disposable container. Future BPF LSM programs may need additional host
configuration. After loading,
capabilities should be reduced; evidence output and policy should be read-only to the build.
Seccomp, Landlock, map freezing, and split loader/collector processes will be evaluated against
actual required syscalls. None of these controls defeats hostile root.
