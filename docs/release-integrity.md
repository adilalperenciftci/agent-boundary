# Release integrity

No release artifacts are published yet. A future release workflow must build from a protected tag on a hosted ephemeral runner, generate a CycloneDX or SPDX SBOM, and attach identity-bound provenance. Consumers must verify repository identity, workflow identity, source revision, and artifact digest rather than checking only that an attestation exists.

The current CI builds distributions as a packaging check but does not claim a SLSA level. Publishing, keyless signing, and provenance generation require a separate workflow with narrowly scoped `id-token: write` and `contents: read` permissions. Pull-request workflows must not receive release credentials.

CI now runs Go race tests/vet and Python tests/static analysis with read-only default permissions.
Separate SHA-pinned CodeQL and OpenSSF Scorecard workflows grant `security-events: write` only to
their analysis jobs; only Scorecard receives `id-token: write` for authenticated result
publication. These workflows improve repository checks but do not constitute a release process,
signed release, or SLSA provenance.

`tools/generate-sbom.sh` scans a read-only source mount with Syft v1.51.1 pinned by OCI manifest
digest, excludes generated/VCS/virtual-environment trees, and emits SPDX 2.3 JSON plus CycloneDX
1.7 JSON under `build/sbom`. A stdlib validator checks required top-level format fields, and a
read-only CI job uploads both as short-lived workflow artifacts. These are source-tree SBOMs, not
release-binary SBOMs; they are not signed and do not imply a published release.
The source component version is the full Git commit on a clean tree and gains a `-dirty` suffix
when local tracked or untracked changes are present, avoiding a false clean-revision claim.

Recommended branch controls are required reviews for runtime, policy, and workflow changes; successful CI; no force pushes on the release branch; private vulnerability reporting; and reviewed dependency updates. Host configuration cannot be enforced from this repository and must be verified separately.
