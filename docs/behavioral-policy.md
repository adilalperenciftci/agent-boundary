# Behavioral policy

Policy is versioned JSON and contains data, not executable code. The first slice supports
allowed executable paths, numeric network destinations, forbidden sensitive-path categories,
expected CI provider, exact allowed provenance builder IDs, exact allowed source repository IDs,
and the `REVIEW`/`REJECT` treatment of incomplete evidence. Builder and repository allowlists are
required; an empty/missing allowlist makes policy parsing fail rather than acting as a wildcard.

Findings have stable reason codes and deterministic precedence. `REJECT` dominates `REVIEW`,
which dominates `ALLOW`. Completeness is not a policy preference: no policy can convert lost
or unknown evidence into `complete` or produce `ALLOW` from it.
