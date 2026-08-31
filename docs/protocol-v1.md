# Foundation authorization protocol v1

## Exact tuple

Every request and receipt carries one exact tuple:

```text
(base_sha, head_sha, pull_request, protected_scope,
 actor_identity, issuer_identity, nonce, issued_at, expires_at,
 generation, prior_receipt_digest, decision)
```

The request target must equal the last receipt tuple. A rotation changes the
nonce, generation, and prior digest while preserving the candidate, scope,
actor, issuer, and decision fields. The first generation has an empty prior
digest; each following generation must point to the digest of the immediately
previous receipt.

## Decisions

The released `.gooo` graph has 12 cells: four each for FOUNDATION, COHERENCE,
and REGRESSION. The axes are independent:

- FOUNDATION is an external trust anchor. Missing issuer or attestation is
  `UNKNOWN`; a trusted public signature is `CLOSED`; a known invalid signature
  or issuer mismatch is `REFUTED`.
- COHERENCE is exact prior-chain and policy continuity. Missing chain evidence
  is `UNKNOWN`; generation gaps, prior mismatches, scope/head mismatches, and
  expiry are `REFUTED`; a valid sequential chain is `CLOSED`.
- REGRESSION is the replay/refutation corpus. A clean corpus path is `CLOSED`;
  replay, a revoked key, or candidate self-authorization is `REFUTED`.

The graph combines axis states only by the explicit precedence
`REFUTED > UNKNOWN > CLOSED`. It never sums scores.

## Authority boundary

The verifier does not issue receipts. It consumes unsigned candidates and
externally attested receipts. An independent consumer separately consumes
verification results and emits an independent receipt with self-approval
disabled. The v0.1.0 release can close verifier/protocol scope only. Live
issuer and #619 integration remain `UNKNOWN` unless an external human,
protected environment, OIDC issuer, DSSE/Sigstore attestation, or equivalent
key authority is observed.
