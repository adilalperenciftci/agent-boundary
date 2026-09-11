# Behavioral policy

Policy is versioned JSON and contains data, not executable code. The first slice supports
allowed executable paths, numeric network destinations, forbidden sensitive-path categories,
expected CI provider, and the `REVIEW`/`REJECT` treatment of incomplete evidence.

Findings have stable reason codes and deterministic precedence. `REJECT` dominates `REVIEW`,
which dominates `ALLOW`. Completeness is not a policy preference: no policy can convert lost
or unknown evidence into `complete` or produce `ALLOW` from it.
