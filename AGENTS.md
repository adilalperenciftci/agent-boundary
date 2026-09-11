# Repository guide

Read `docs/architecture.md`, `docs/threat-model.md`, and `docs/security-invariants.md` before changing runtime behavior.

- Keep the trusted runtime path small and deterministic.
- Treat every event field, tool definition, result, and rule file as attacker-controlled until validated.
- Never persist a matched secret value. Tests must use synthetic credentials.
- A finding is detection unless an inline adapter enforces its decision before action.
- Add positive, negative, malformed, and edge-case tests for detection changes.
- Record security-significant design changes under `docs/decisions/`.
- Run `ruff`, `pyright`, unit tests, rule validation, Bandit, and `pip-audit` before commit.
