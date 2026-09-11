# References and design influences

- Model Context Protocol, [2026-07-28 specification announcement](https://blog.modelcontextprotocol.io/posts/2026-07-28/). Influenced stateless adapter assumptions and header-addressable policy boundaries.
- Model Context Protocol, [Authorization specification](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization). Influenced audience binding, token-passthrough non-goals, and separation from OAuth enforcement.
- OpenAI, [Running Codex safely at OpenAI](https://openai.com/index/running-codex-safely/). Influenced agent-native event coverage and OTel interoperability.
- OpenAI, [Designing AI agents to resist prompt injection](https://openai.com/index/designing-agents-to-resist-prompt-injection/). Influenced source-to-sink correlation and rejection of classifier-only prevention.
- OWASP, [Top 10 for Agentic Applications 2026](https://genai.owasp.org/download/52117/) and [LLM06 Excessive Agency](https://genai.owasp.org/llmrisk/llm062025-excessive-agency/). Influenced threat categories and least-authority assumptions.
- MITRE ATT&CK, [T1567 Exfiltration Over Web Service](https://attack.mitre.org/techniques/T1567/) and [T1195 Supply Chain Compromise](https://attack.mitre.org/techniques/T1195/). Used only where observable behavior supports mapping.
- OpenTelemetry, [GenAI semantic attributes](https://opentelemetry.io/docs/specs/semconv/registry/attributes/gen-ai/). Influenced field interoperability and content-retention warnings.
- SLSA, [Specification 1.2](https://slsa.dev/spec/v1.2/) and [artifact verification](https://slsa.dev/spec/v1.2/verifying-artifacts). Influenced the distinction between presence, integrity, authenticity, and verification of provenance.
- Wang et al., [MCPTox](https://ojs.aaai.org/index.php/AAAI/article/download/40895/44856), AAAI 2026. Evidence for metadata-level tool poisoning.
- Song et al., [Beyond the Protocol](https://arxiv.org/abs/2506.02040), 2025. Influenced catalog continuity and rug-pull roadmap.
- Wang et al., [MindGuard](https://arxiv.org/abs/2508.20412), 2025 preprint. Informed decision provenance; model-internal attention dependence is deliberately not adopted.
