# Architecture

## Thesis and scope

Agent Boundary is a local-first decision-point engine for agent/tool interactions. It consumes security-relevant events from an agent host or MCP adapter, normalizes them into a stable envelope, evaluates deterministic policy, and records a replayable evidence trail.

The core does not proxy MCP transport in its first release. Adapters are expected to call it before a privileged tool invocation and honor `deny` or `review` results. When events arrive after execution, the same rule is detection-only.

## Components

```text
agent host / MCP adapter
          |
          v
  bounded JSON input
          |
    schema validator
          |
      normalizer ----> canonical event digest
          |
     policy engine ----> decision + findings
          |                       |
          +-----------> redactor -+
                                  |
                           hash-chained JSONL
                                  |
                           verifier / replay
```

The implementation is one Python package with narrow modules. Python is sufficient because this layer performs bounded parsing and deterministic evaluation, not high-throughput packet processing. A Rust gateway would add build and unsafe-FFI review surface before transport interception is justified.

## Event path

1. The caller submits a UTF-8 JSON object subject to byte, depth, collection-size, and string-size limits.
2. Structural validation rejects unknown event versions, invalid identifiers, non-finite numbers, and ambiguous fields.
3. Normalization emits canonical field ordering and URI forms without changing argument semantics.
4. Rules evaluate only declared event requirements. Missing evidence cannot silently become evidence of safety.
5. The engine combines findings by explicit precedence: `deny` > `review` > `allow`.
6. Evidence is redacted before persistence. Raw credentials are never required for replay.
7. Each ledger record commits to the previous record hash and canonical current record. Verification detects deletion, reordering, or modification within a ledger segment; it does not prevent whole-ledger deletion or rollback without an external checkpoint.

## Trust model

Caller-provided provenance labels are claims, not facts. A deployment adapter is responsible for authenticating the caller and assigning trusted labels. The engine records `producer` and `observed_at` separately from event time. Future signed adapters may raise provenance assurance; the event schema must not imply that self-asserted metadata is verified.

Tool manifests are canonicalized and hashed. An approved catalog digest can be supplied by policy. A mismatch supports a catalog-integrity finding but not attribution to a malicious actor.

## Policy model

Policy is deterministic and data oriented. Rules declare stable IDs, applicable event kinds, required fields, severity, decision effect, and parameters. The initial built-in evaluators are intentionally small. YAML loading is deferred until schema and evaluator semantics are proven; executable Python from rule directories is prohibited.

Secret detection uses named synthetic/known-secret fingerprints and conservative format indicators. Evidence retains the JSON path, detector identifier, and keyed or unkeyed digest prefix as configured, never the matched value. Format-only matches result in review unless policy explicitly raises them to deny.

Egress policy parses destinations structurally. It does not use substring allowlists. Schemes and normalized hosts are matched separately; userinfo, malformed hosts, unexpected ports, and IP literals are explicit policy dimensions.

## Failure behavior

Inline mode fails closed for malformed input, unavailable policy, and ledger write failure. Observe-only replay reports validation errors and continues only when explicitly requested. Limits are applied before expensive traversal. Decision output is written only after durable ledger append unless the caller selects a documented non-durable mode.

## Deployment modes

- Library: lowest latency; host invokes the engine in process.
- CLI/stdio: language-neutral local adapter and deterministic laboratory interface.
- Replay: reads fixtures or ledgers with network disabled and compares findings.

An HTTP daemon is not in the initial scope because it introduces authentication, request smuggling, TLS termination, and availability obligations unrelated to proving the decision model.

## Telemetry interoperability

The normalized schema preserves trace and span identifiers and maps compatible fields to OpenTelemetry GenAI names. Sensitive tool arguments/results are not exported by default. OTel is an export surface, not the authoritative forensic store, because collector transformations and sampling can remove evidence.
