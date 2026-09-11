# Provenance verification

Strict correlation never merges conflicting claims. The artifact digest must match SLSA and
Runtime Trace subjects. Build ID, provider, run ID/attempt, source repository/revision, graph,
evidence, policy, and sensor configuration commitments must agree exactly where required.
Missing required values fail; they are not wildcards.

The fixture slice parses SLSA Provenance v1 Statements and a project-namespaced build identity
inside `buildDefinition.internalParameters`. Signature, builder authorization, source VSA, and
transparency verification are not implemented yet and therefore are not claimed.
