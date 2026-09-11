# Testing strategy

Every evaluator requires a benign baseline, suspicious fixture, non-detection case, malformed-input cases, and expected findings. Fixtures are deterministic and use `.test` domains and synthetic credentials.

The initial suite verifies allow/deny behavior, exact-host matching, redaction, duplicate-key rejection, depth limits, undeclared nested-field rejection, ledger verification, and refusal to extend a modified ledger.

Unit and vertical-slice tests run without network access or third-party services. Rule metadata validation ensures fixture references exist and implemented IDs match documented IDs. CLI integration tests verify exit status and durable output. Property-based parsing tests are deferred until a dependency is justified; parser regression fixtures should be added for every discovered ambiguity.

No false-positive or false-negative rate is reported from these curated fixtures. Counts from deterministic fixtures describe coverage, not field performance.
