# Event schema

The accepted event is `tool_call.requested` version `1.0`. Unknown fields are rejected to avoid producer/consumer disagreement.

| Field | Type | Meaning |
| --- | --- | --- |
| `schema_version` | string, `1.0` | Schema contract |
| `event_id` | UUID string | Producer-assigned event identity |
| `event_type` | `tool_call.requested` | Pre-action decision point |
| `occurred_at` | timezone-aware RFC 3339 string | Producer event time |
| `producer.id` | identifier | Adapter identity claim |
| `session.id` | identifier | Correlation key |
| `source.trust` | enum | `trusted`, `untrusted`, `mixed`, or `unknown` |
| `tool.name` | identifier | Selected tool |
| `arguments` | object | Proposed arguments, subject to limits and redaction |
| `destination` | absolute HTTP(S) URI or null | Proposed network sink |

Limits are 256 KiB encoded input, depth 16, 2,000 traversed values, 32 KiB per string, and 2,048 characters for a destination. JSON `NaN`, infinities, duplicate keys, URI userinfo, and destination fragments that cannot be parsed are rejected.

Normalized event digests are calculated after secret redaction. They support replay comparison without committing a recoverable secret to disk; they are not a digest of the raw source event.

OpenTelemetry mapping, when an exporter is added: `tool.name` maps to the GenAI tool name convention; arguments and results remain disabled by default because the semantic conventions mark them as potentially sensitive.
