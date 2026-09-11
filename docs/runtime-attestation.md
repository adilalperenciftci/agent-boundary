# Runtime attestation

The project composes in-toto Runtime Trace v0.1 rather than defining a competing statement
format. Artifact digests remain in the standard Statement subject. Standard predicate fields
identify monitor and monitored run. The namespaced `runtime-provenance/v0.1` extension commits
to the evidence manifest, execution graph, SLSA provenance, policy, source, build/run identity,
and completeness.

The current fixture slice emits an unsigned Statement. It is not accepted as authenticated
production evidence. A later strict verifier will require a Sigstore bundle and authorized
signer/OIDC issuer. Detailed event bytes remain external and must be supplied by digest for
forensic replay.
