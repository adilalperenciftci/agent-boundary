# Agent Boundary

Agent Boundary is a local decision-point security engine for AI-agent and MCP tool calls. It turns bounded, versioned tool-call events into deterministic `allow`, `review`, or `deny` decisions and records redacted evidence in a hash-chained ledger.

The problem is not only malicious text. Agentic failures occur when untrusted influence reaches a privileged sink: a credential is placed in an outbound call, an unchanged-looking tool gains authority, or individually ordinary calls compose into an unsafe action. General application logs rarely retain the trust, destination, approval, and catalog context needed to explain that boundary crossing.

The current vertical slice detects secret-like values in tool arguments and structurally unapproved HTTP destinations. It rejects malformed/ambiguous events, redacts secret material before persistence, and verifies ledger modification or reordering.

## Security model

The engine is useful inline only when an authenticated adapter calls it before tool execution and enforces the result. In post-execution pipelines it is a detection and forensic component. An `allow` means no configured rule required intervention; it is not proof that an action is safe.

Hash chaining detects modification within a ledger segment. It does not prevent deletion, valid-prefix rollback, or forgery by a host administrator. Producer trust labels are claims unless an adapter binds them to an authenticated identity.

See [threat model](docs/threat-model.md), [security invariants](docs/security-invariants.md), and [architecture](docs/architecture.md).

## Install

Python 3.12 or newer is required.

```console
python -m venv .venv
.venv\Scripts\python -m pip install -e .
```

For development, use `uv sync --extra dev --locked` with the committed lock file.

## Reproducible local demo

All fixtures use reserved `.test` domains and synthetic credentials. No network request is made.

```console
agent-boundary evaluate --input fixtures/benign/approved-call.json --policy policy/example.json --ledger out/events.jsonl
agent-boundary evaluate --input fixtures/suspicious/secret-egress.json --policy policy/example.json --ledger out/events.jsonl
agent-boundary verify-ledger --ledger out/events.jsonl
```

Exit codes are `0` allow, `2` review, `3` deny, and `4` validation or operational failure. The suspicious fixture returns deny, so shells may display a non-zero status by design.

## Architecture

An adapter supplies a bounded JSON event. Strict validation and URI normalization occur before deterministic rules. Evidence is redacted, canonicalized, and appended to a single-writer JSONL ledger. The standard-library runtime has no third-party dependencies.

OpenTelemetry-compatible names are an export concern; raw arguments are not exported by default. The local ledger remains authoritative because collectors can sample or transform telemetry.

## Deliberate exclusions

- No prompt-injection text classifier
- No transparent MCP/OAuth proxy yet
- No tool sandbox or credential broker
- No claim of tamper-proof storage
- No external scanning or third-party targets

## Test and assurance

```console
uv run ruff check .
uv run pyright
uv run python -m unittest discover -s tests -v
uv run python tools/validate_rules.py
uv run bandit -q -r src
uv run pip-audit
```

The fixtures establish only repository test behavior. No production detection-rate or performance claims are published. See [testing strategy](docs/testing-strategy.md).

## Limitations and roadmap

The ledger currently requires one writer, destination checks do not observe DNS resolution or redirects, format detectors can miss or misclassify secrets, and provenance is adapter-supplied. Next work is tool-catalog digest continuity bound to decisions, then session-level read-to-egress correlation. Multi-writer durability and external ledger checkpoints precede daemon deployment.
