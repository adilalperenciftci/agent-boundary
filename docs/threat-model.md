# Threat model

## Security objective

Prevent or detect policy-inconsistent agent tool calls at an integration boundary and retain enough minimized evidence to reproduce the decision. The engine must not turn attacker-controlled content into executable policy or leak secrets through its own telemetry.

## Assets

- Credentials and sensitive values present in tool arguments or results
- User intent and approval state
- Tool authority, identities, scopes, and destination policy
- Tool catalog definitions and approved digests
- Decision rules and configuration
- Event and finding integrity
- Availability of inline decisions

## Actors

- Authorized user issuing an objective
- Agent host and adapter producing events
- MCP server or other tool provider
- Policy administrator
- Security analyst replaying evidence
- Attacker controlling user input, retrieved content, tool metadata/results, a dependency, or an MCP server
- Local attacker able to alter configuration or ledger files

## Entry points and attacker-controlled inputs

JSON event bytes, every tool-schema field, tool arguments and results, URIs, timestamps, trace identifiers, asserted provenance, rule files, environment configuration, fixture files, and dependency artifacts are inputs. Text returned by a source is data even when it resembles instructions.

## Trust boundaries

1. External content to agent context: untrusted data can influence model decisions.
2. Agent/model to privileged tool: probabilistic output crosses into deterministic authority.
3. Tool provider to host: schema and result metadata are provider-controlled.
4. Adapter to engine: event claims require adapter authentication to have provenance value.
5. Engine to ledger/exporter: sensitive evidence leaves the decision process.
6. Repository/dependency source to build: code and rules become trusted executable artifacts.

## Privileged operations and data flows

Tool calls may read files, query private services, mutate state, execute commands, or transmit data. Credentials should remain in adapter/tool-specific credential brokers and should not be copied into events. If a secret appears in arguments, the detector processes it transiently and persists only redacted evidence.

Network destinations are untrusted unless matched by policy after structural normalization. Localhost is not universally trusted: it may expose privileged local services. Redirects and DNS resolution occur outside the initial engine and require adapter enforcement.

## Threat scenarios

| Scenario | Observable evidence | Initial control | Classification |
| --- | --- | --- | --- |
| Indirect injection causes credential transmission | untrusted source provenance, secret indicator, egress sink | deny/review at requested call | Prevention inline; detection otherwise |
| Poisoned tool description changes selection | catalog content/digest changes | catalog continuity check (next slice) | Validation and detection |
| Rug pull after approval | approved digest differs at call time | bind decision to catalog digest | Prevention when enforced |
| Confused deputy invokes privileged write | actor/objective/approval lacks required capability | capability policy (planned) | Prevention |
| Benign calls compose into exfiltration | read-sensitive then external-write in session | sequence correlation (planned) | Detection/review |
| Malformed event bypasses inspection | parser ambiguity or resource exhaustion | strict schema and limits | Prevention |
| Credential leaks through finding/log | raw match retained | evidence redaction invariant | Prevention |
| Ledger is modified | broken record hash chain | verifier | Detection |
| Ledger is deleted or rolled back | no local evidence remains | external signed checkpoint (not initial) | Not controlled locally |
| Producer forges provenance labels | self-asserted trusted source | authenticated adapter binding (deployment duty) | Not solved by schema |

## Failure modes

- Parser disagreement: canonicalization accepts a value differently from the adapter. Mitigation: strict JSON types, no duplicate keys, no NaN/Infinity, documented URI normalization.
- Rule ambiguity: overlapping findings yield inconsistent decisions. Mitigation: fixed effect precedence and deterministic ordering.
- Excessive retention: arguments become a shadow secret store. Mitigation: redact before append and default-deny raw-content export.
- Availability attack: large/deep inputs exhaust resources. Mitigation: pre-parse byte limit plus bounded traversal.
- Ledger write failure: a call proceeds without evidence. Inline contract fails closed.
- False confidence: an allow result is interpreted as proof of safety. Documentation defines allow as “no configured rule required intervention,” not benignness.

## Assumptions

- Inline adapters call the engine before action and enforce its result.
- Policy files and the engine installation are protected by host access controls.
- Event clocks may be wrong; ingestion order is authoritative for replay.
- Detection cannot reliably infer user intent from arbitrary prose.
- Destination enforcement after DNS resolution and redirects belongs to the network adapter.

## Non-goals

- Proving an LLM is aligned or free of prompt injection
- Detecting every secret format or malicious instruction
- Sandboxing tool implementations
- Replacing OAuth audience/scope enforcement, EDR, DLP, or network controls
- Providing tamper-proof storage against a host administrator
- Scanning or testing third-party MCP servers
- Automatically attributing malicious intent

## Control taxonomy

- Prevention: strict rejection, inline deny, adapter-enforced destination/capability policy, redaction.
- Detection: findings and chain-integrity failures.
- Telemetry: normalized events, decisions, trace context, minimized evidence.
- Validation: schema checks, catalog digest checks, deterministic replay.
- Response: analyst-facing recommendation and caller decision; no autonomous remediation initially.
