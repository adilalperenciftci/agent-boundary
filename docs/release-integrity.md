# Release integrity

No release artifacts are published yet. A future release workflow must build from a protected tag on a hosted ephemeral runner, generate a CycloneDX or SPDX SBOM, and attach identity-bound provenance. Consumers must verify repository identity, workflow identity, source revision, and artifact digest rather than checking only that an attestation exists.

The current CI builds distributions as a packaging check but does not claim a SLSA level. Publishing, keyless signing, and provenance generation require a separate workflow with narrowly scoped `id-token: write` and `contents: read` permissions. Pull-request workflows must not receive release credentials.

Recommended branch controls are required reviews for runtime, policy, and workflow changes; successful CI; no force pushes on the release branch; private vulnerability reporting; and reviewed dependency updates. Host configuration cannot be enforced from this repository and must be verified separately.
