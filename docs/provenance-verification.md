# Provenance verification

Strict correlation never merges conflicting claims. The artifact digest must match SLSA and
Runtime Trace subjects. Build ID, provider, run ID/attempt, source repository/revision, graph,
evidence, policy, and sensor configuration commitments must agree exactly where required.
Missing required values fail; they are not wildcards.

The verifier parses SLSA Provenance v1 Statements and a project-namespaced build identity inside
`buildDefinition.internalParameters`. Its local profile requires non-empty build type, external
parameters, a matching resolved source dependency, builder ID, invocation ID, and start/finish
timestamps. External source, resolved revision, extension identity, runtime build/run identity,
and subject digest must agree. Regression tests reject conflicting revisions and provenance
replayed from another build.

`create-local-provenance` exists only for the disposable lab and reports assurance
`unsigned-local-fixture`. Signature, builder authorization, source VSA, and transparency
verification are not implemented yet and therefore are not claimed.
