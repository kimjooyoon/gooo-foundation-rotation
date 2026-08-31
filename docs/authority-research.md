# Authority research and adoption record

The v0.1.0 implementation separates a candidate producer from an authorization
consumer. It accepts no candidate-generated authorization as an external
anchor. If the candidate actor and issuer identities are the same, the
Regression axis is `REFUTED`.

## Public records investigated

The investigation was read-only and did not modify `meta-ontology-go`.

| record | observed fact | protocol consequence |
| --- | --- | --- |
| [PR #619](https://github.com/kimjooyoon/meta-ontology-go/pull/619) | open; base `ac3a56b933d9a9b934fe26709485dc2f36edd916`, head `90f5fdf2198a7da5cde405f9f675cc62975b9226` | exact request fixture |
| [PR #606](https://github.com/kimjooyoon/meta-ontology-go/pull/606) | closed dev/main synchronization candidate | historical context only |
| [PR #609](https://github.com/kimjooyoon/meta-ontology-go/pull/609) | closed re-run dev/main synchronization candidate | historical context only |
| [PR #610](https://github.com/kimjooyoon/meta-ontology-go/pull/610) | merged FOUNDATION authorization route; public body says single-use and replay `REFUTED` | rotation requirement |
| [Guardian run 33433725445](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/33433725445) | failed `CI guardian shadow` | exact failure fixture |

The public failed job ended with:

```text
CI-FOUNDATION-AUTHORIZATION-001: Guardian dispatch live candidate tuple is not exact
```

This repository does not infer that #619 is approved or unblocked.

## External authority options

### GitHub Actions OIDC — adopted as a future live issuer profile

GitHub documents `iss`, `aud`, `sub`, `iat`, and `exp` claims and recommends
using audience and subject conditions when a cloud role trusts a workflow. The
workflow must request `id-token: write`; a bare actor name is not equivalent to
an issuer assertion. See the [OIDC reference](https://docs.github.com/en/actions/reference/security/oidc)
and [OIDC concepts](https://docs.github.com/en/actions/concepts/security/openid-connect).

The tuple therefore carries separate `actor_identity` and `issuer_identity`,
and the verifier requires a trusted issuer binding. No live token is present
in this repository, so live issuance is `UNKNOWN`.

### Protected environments — adopted as a future human/environment profile

GitHub documents required reviewers, environment protection rules, and the
option to prevent self-reviews in protected environments. That is the needed
separation between a candidate workflow and an external approval authority;
the [deployment and environments documentation](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments)
is the primary source. No repository environment approval is used by the
deterministic fixture, so it is not treated as live evidence.

### DSSE/in-toto — adopted as a serialization integration boundary

The [in-toto envelope specification](https://github.com/in-toto/attestation/blob/main/spec/v1/envelope.md)
requires an array of signatures, an authenticated payload type, and a
base64-encoded payload for the standard JSON statement. v0.1.0 keeps the
authorization tuple and transition semantics in `.gooo`; its Go backend
verifies an Ed25519 public fixture and records the external-attestation slot.
Full DSSE/Sigstore transport is not claimed as live evidence.

### Sigstore — rejected for v0.1.0 live closure, retained as an integration option

Sigstore documents keyless signing through Fulcio short-lived certificates,
Rekor transparency logging, and OIDC identities; its CI quickstart describes
GitHub Actions integration. See [Sigstore signing overview](https://docs.sigstore.dev/quickstart/quickstart-ci/)
and [Cosign blob verification](https://github.com/sigstore/cosign/blob/main/README.md).
Because no external Sigstore identity, certificate, or transparency-log entry
is available in this run, v0.1.0 records that profile as `UNKNOWN` rather than
manufacturing a signature. A deterministic public signature fixture is used
only for verifier conformance.

## State boundary

`CLOSED` means the verifier/protocol contract and public fixture corpus are
conformant. `UNKNOWN` means an external authority or integration evidence is
missing. `REFUTED` means known mismatch, replay, revoked key, invalid
signature, or candidate self-authorization. The semantic graph owns this
meaning and applies `REFUTED > UNKNOWN > CLOSED`; there is no aggregate score.
