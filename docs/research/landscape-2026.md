# Agent and MCP runtime security landscape, 2026

## Scope

This review covers controls at the boundary where an agent selects and invokes a tool. It prioritizes protocol specifications, vendor security guidance, standards, and original research available on 11 September 2026. Claims from experimental papers are treated as evidence of attack feasibility, not as production guarantees.

## What the ecosystem already provides

MCP standardizes discovery and invocation, but deliberately leaves consequential security decisions to hosts and implementations. The 2026-07-28 protocol is stateless: requests carry protocol, identity, and capability metadata; method and tool names can be exposed in headers for routing and authorization; authorization adds issuer validation and moves toward Client ID Metadata Documents. These changes improve enforceability at HTTP boundaries but do not establish that a tool description, result, or requested action is trustworthy.[1]

MCP security guidance correctly treats tool descriptions and annotations as untrusted, forbids token passthrough, requires token audience validation, and emphasizes consent and least privilege.[2] These are necessary protocol and authorization controls. They do not answer whether a particular invocation is consistent with the user's objective or whether information flowed from an attacker-controlled source to a sensitive sink.

OpenAI's current guidance combines sandboxing, network policy, approvals, managed configuration, identities, and agent-native telemetry. Codex can export prompts, approvals, tool results, MCP use, and proxy decisions through OpenTelemetry.[3] OpenAI also describes prompt injection as a source-to-sink problem: dangerous behavior generally requires attacker influence plus a consequential action. It cautions that classifying text as malicious is not sufficient and recommends constraining impact even when manipulation succeeds.[4]

OWASP's 2025 LLM risks cover prompt injection, sensitive-information disclosure, supply chain, and excessive agency. The 2026 Agentic Top 10 extends these into goal hijack, tool misuse, identity abuse, agentic supply-chain failures, memory poisoning, insecure inter-agent communication, and untraceability.[5] The material is useful for threat enumeration and control design, but it is not an event schema or executable detection model.

MITRE ATT&CK remains applicable only at observable endpoints. For example, a tool invocation that sends data to a webhook can support T1567.004 only when the event shows the relevant exfiltration behavior; a suspicious string alone does not. Supply Chain Compromise (T1195) is defensible when an artifact or manifest changes unexpectedly and verification fails.[6] Agent-specific taxonomies should not be mechanically converted into ATT&CK mappings.

OpenTelemetry's GenAI semantic conventions now define tool call arguments, results, definitions, and tool types, while explicitly warning that arguments and results may contain sensitive information.[7] This supplies interoperable field names, not retention policy, trust labels, provenance integrity, or security decisions.

SLSA 1.2 distinguishes provenance existence from authentic and unforgeable provenance. Its verification guidance is directly relevant to released artifacts, but runtime tool catalogs often lack equivalent attestations. Sigstore provides identity-bound signing and verification mechanisms; neither automatically models runtime catalog continuity.[8]

## Public attack and defense research

Research since 2025 demonstrates that instructions embedded in tool metadata can alter agent decisions without the poisoned tool itself being called. MCPTox evaluates this class across real tool schemas; broader MCP ecosystem research describes tool poisoning, puppet attacks, rug pulls, and malicious external resources.[9] MindGuard explores model-internal decision dependence for attribution, but its assumptions require access to attention signals that hosted models normally do not expose.[10]

Indirect prompt injection research repeatedly shows the same structural failure: untrusted retrieved content influences an agent that holds ambient authority. Consequences include credential disclosure, unauthorized tool composition, and external actions. Results vary substantially by model, task, and defense; no published detector should be treated as a complete prevention boundary.[11]

The durable lesson is classical. Agents can become confused deputies when identity, intent, and authority are collapsed. Controls should bind each privileged action to an explicit policy decision and preserve evidence about the source, sink, authority, and catalog state that existed at that moment.

## Three candidate theses

| Candidate | Novelty | Defensive value | Technical depth | Maintainability | Daybreak Blue fit |
| --- | --- | --- | --- | --- | --- |
| A. Transparent MCP reverse proxy with argument signatures | Medium | High for HTTP MCP | High transport complexity | Medium; protocol churn is costly | Good |
| B. Static MCP manifest linter and signer | Low-medium | Useful before deployment | Medium | High | Fair; misses runtime composition |
| C. Provenance-aware source-to-sink decision monitor and replay ledger | High | High across transports/frameworks | High in correlation, integrity, and testing | High with a small core | Strong |

Candidate A can enforce network policy, but transport termination and authorization quickly dominate the project. It also cannot infer origin trust unless hosts provide context. Candidate B addresses tool poisoning and rug pulls, but duplicates conventional signing workflows and cannot observe misuse of unchanged tools.

Candidate C is selected. Its narrow thesis is:

> A small, deterministic decision-point engine can make agent/tool activity meaningfully auditable by binding normalized events to explicit trust labels, catalog digests, source-to-sink policy, redacted evidence, and a hash-chained replay ledger.

The contribution is not another prompt-injection classifier. It is a security control plane with explicit evidence requirements. It can block policy violations before invocation when integrated inline; in observe-only deployments it produces detections, not prevention.

## Initial vertical slice

The first slice accepts one versioned `tool_call.requested` event, validates and normalizes it, evaluates deterministic rules, emits an `allow`, `review`, or `deny` decision with redacted evidence, and appends the input and decision to a hash-chained JSONL ledger. Initial rules cover secret-like values in arguments and untrusted network destinations. Positive, negative, malformed-input, redaction, and ledger-tamper tests establish the claims.

Deferred capabilities include natural-language prompt-injection classification, live OAuth termination, model introspection, autonomous remediation, and cross-host identity federation. Catalog continuity and multi-event correlation follow only after the slice is stable.

## Sources

1. Model Context Protocol, "[The 2026-07-28 Specification](https://blog.modelcontextprotocol.io/posts/2026-07-28/)," 28 July 2026.
2. Model Context Protocol, "[Authorization](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization)" and protocol security principles, accessed 11 September 2026.
3. OpenAI, "[Running Codex safely at OpenAI](https://openai.com/index/running-codex-safely/)," 8 May 2026.
4. OpenAI, "[Designing AI agents to resist prompt injection](https://openai.com/index/designing-agents-to-resist-prompt-injection/)," 11 March 2026.
5. OWASP GenAI Security Project, "[LLM06:2025 Excessive Agency](https://genai.owasp.org/llmrisk/llm062025-excessive-agency/)" and "[Top 10 for Agentic Applications 2026](https://genai.owasp.org/download/52117/)," accessed 11 September 2026.
6. MITRE ATT&CK, "[Exfiltration Over Web Service (T1567)](https://attack.mitre.org/techniques/T1567/)" and "[Supply Chain Compromise (T1195)](https://attack.mitre.org/techniques/T1195/)," versions current 11 September 2026.
7. OpenTelemetry, "[Generative AI semantic attributes](https://opentelemetry.io/docs/specs/semconv/registry/attributes/gen-ai/)," accessed 11 September 2026.
8. SLSA, "[Specification 1.2](https://slsa.dev/spec/v1.2/)" and "[Build requirements](https://slsa.dev/spec/v1.2/build-requirements)," accessed 11 September 2026; Sigstore, "[Verifying signatures](https://docs.sigstore.dev/cosign/verifying/verify/)."
9. Wang et al., "[MCPTox](https://ojs.aaai.org/index.php/AAAI/article/download/40895/44856)," AAAI 2026; Song et al., "[Beyond the Protocol](https://arxiv.org/abs/2506.02040)," 2025.
10. Wang et al., "[MindGuard](https://arxiv.org/abs/2508.20412)," 2025 preprint.
11. "[Securing Tool-Using AI Agents Against Injection and Authority Misuse](https://doi.org/10.3390/computation14050098)," Computation 14(5), 2026; "[A Framework for Formalizing LLM Agent Security](https://openreview.net/pdf?id=iQzd6qzIs5)," 2026 preprint.
