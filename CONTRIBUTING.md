# Contributing

Changes should preserve the invariants in `docs/security-invariants.md` and update the threat model when a trust boundary changes.

Detection changes require stable metadata under `rules/`, positive and negative fixtures, edge cases, expected output, and limitations. Fixtures must be synthetic and must not call external systems. Do not add mappings unsupported by observable evidence.

Before submitting a change, run the commands under README's Test and assurance section. Keep commits coherent; architecture, event contracts, evaluator changes, fixtures, and CI hardening should remain reviewable.
