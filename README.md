# gooo-foundation-rotation

Rotatable Foundation authorization for protected Gooo self-improvement.

Version `v0.1.0` closes the verifier/protocol scope. It defines an exact
authorization tuple, single-use receipts, sequential rotation, replay and
scope/head mismatch rejection, revocation, and explicit issuer/attestation
unknowns. Approval meaning and state transitions are owned by the released
`.gooo` semantic graph in [`semantic/foundation-authorization.gooo`](semantic/foundation-authorization.gooo).

The Go implementation is a parsing, cryptographic-verification, and
serialization backend. A separate consumer reads the same public fixtures
through an independent path. No command in this repository issues an
authorization or self-approves a candidate. A candidate/approver identity
collision is `REFUTED`.

The live issuer and `meta-ontology-go` PR #619 integration remain `UNKNOWN`
until an external human, protected GitHub environment, OIDC issuer, or key
authority provides evidence. The repository does not claim to unblock #619.

## Usage

```text
go run ./cmd/gooo-foundation-rotation verify --case fixtures/cases/normal-rotation.json
go run ./cmd/gooo-foundation-consumer consume --case fixtures/cases/normal-rotation.json
```

These commands are examples only. Local tests are intentionally not part of
the authority model; GitHub Actions with Go 1.27 is the verification
authority. The workflow records exact integers for the semantic denominator,
cases, proof choices, indicators, receipts, rotations, replays, revocations,
inventory, outputs, bytes, wall time, peak RSS, and test observation.

## Evidence inventory

- `semantic/foundation-authorization.gooo`: released semantic graph.
- `fixtures/authorization-request.json`: machine-readable request derived from
  the public #619 failure tuple.
- `fixtures/unsigned-candidate-receipt.json`: unsigned candidate; external
  attestation is absent, so verification is `UNKNOWN`.
- `fixtures/signed-externally-attested-receipt.json`: deterministic public
  signature fixture; no private key is stored.
- `fixtures/rotation-chain.json`: sequential receipt chain.
- `fixtures/revocation-set.json`: public revocation inputs.
- `fixtures/cases/`: normal, unknown, and refuted conformance corpus.
- `reports/human-report.md`: protocol boundary and evidence interpretation.

## Public record and authority notes

The implementation records the read-only investigation of:

- `meta-ontology-go` PR #619: open; its public Guardian run failed with
  `CI-FOUNDATION-AUTHORIZATION-001` because the live candidate tuple was not
  exact.
- PR #606 and #609: closed dev/main synchronization candidates.
- PR #610: merged FOUNDATION authorization route with a single-use override.

No file, branch, workflow, PR, or release in `meta-ontology-go` is modified by
this repository.

## License

MIT.
