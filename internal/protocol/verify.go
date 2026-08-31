package protocol

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

const (
	signalIssuerUnavailable   = "issuer_unavailable"
	signalExternalAttestation = "external_attestation_missing"
	signalSignatureInvalid    = "signature_invalid"
	signalTrustedAnchor       = "trusted_anchor_satisfied"
	signalGenerationGap       = "generation_gap"
	signalPriorMismatch       = "prior_receipt_mismatch"
	signalScopeHeadMismatch   = "scope_or_head_mismatch"
	signalSequentialRotation  = "sequential_rotation_satisfied"
	signalReplay              = "replay_detected"
	signalRevokedKey          = "revoked_key"
	signalSelfAuthorization   = "self_authorized_candidate"
	signalCorpus              = "corpus_satisfied"
)

func VerifyCase(graph SemanticGraph, trust TrustStore, c ConformanceCase) (VerificationResult, error) {
	if c.Schema != CaseSchema {
		return VerificationResult{}, fmt.Errorf("case %s has invalid schema", c.CaseID)
	}
	if c.Request.Schema != RequestSchema || c.Request.RequestID == "" {
		return VerificationResult{}, fmt.Errorf("case %s has invalid request", c.CaseID)
	}
	if c.Chain.Schema != ChainSchema || c.Chain.RequestID != c.Request.RequestID {
		return VerificationResult{}, fmt.Errorf("case %s has invalid rotation chain", c.CaseID)
	}
	if c.Revocations.Schema != RevocationSchema {
		return VerificationResult{}, fmt.Errorf("case %s has invalid revocation set", c.CaseID)
	}
	if err := validateTuple(c.Request.Tuple); err != nil {
		return VerificationResult{}, fmt.Errorf("case %s request tuple: %w", c.CaseID, err)
	}

	result := VerificationResult{
		Schema: ResultSchema, CaseID: c.CaseID, RequestID: c.Request.RequestID,
		DecisionRule: "semantic-graph-precedence:" + joinStates(graph.Precedence),
		AxisResults:  []AxisResult{}, Unknowns: []UnknownDetail{}, Refutations: []string{},
		ReceiptDigests: []string{},
	}
	foundationSignals := []string{}
	coherenceSignals := []string{}
	regressionSignals := []string{}

	revokedKeys := map[string]bool{}
	for _, key := range c.Revocations.Keys {
		revokedKeys[key.KeyID] = true
	}
	revokedReceipts := stringSet(c.Revocations.ReceiptDigests)
	revokedNonces := stringSet(c.Revocations.Nonces)
	trusted := map[string]TrustedIssuer{}
	for _, issuer := range trust.Issuers {
		trusted[issuer.KeyID] = issuer
	}

	if len(c.Chain.Receipts) == 0 {
		result.Unknowns = append(result.Unknowns, unknown("COHERENCE", "REQUIRE_ROTATION_CHAIN", "rotation chain is absent", "MISSING_EVIDENCE", "COLLECT_PRIOR_RECEIPT_CHAIN", []string{"rotation-chain"}))
	}
	seenDigests := map[string]bool{}
	seenNonces := map[string]bool{}
	for index, receipt := range c.Chain.Receipts {
		digest, err := Digest(receipt)
		if err != nil {
			return VerificationResult{}, err
		}
		result.ReceiptDigests = append(result.ReceiptDigests, digest)
		if seenDigests[digest] || revokedReceipts[digest] || seenNonces[receipt.Tuple.Nonce] || revokedNonces[receipt.Tuple.Nonce] {
			regressionSignals = append(regressionSignals, signalReplay)
			result.ReplayCount++
			result.Refutations = append(result.Refutations, "replay:"+receipt.ReceiptID)
		}
		seenDigests[digest] = true
		seenNonces[receipt.Tuple.Nonce] = true
		if revokedKeys[receipt.IssuerKeyID] {
			regressionSignals = append(regressionSignals, signalRevokedKey)
			result.RevocationCount++
			result.Refutations = append(result.Refutations, "revoked-key:"+receipt.IssuerKeyID)
		}
		if receipt.Tuple.ActorIdentity != "" && receipt.Tuple.ActorIdentity == receipt.Tuple.IssuerIdentity {
			regressionSignals = append(regressionSignals, signalSelfAuthorization)
			result.Refutations = append(result.Refutations, "self-authorization:"+receipt.ReceiptID)
		}
		if err := validateTuple(receipt.Tuple); err != nil {
			coherenceSignals = append(coherenceSignals, signalScopeHeadMismatch)
			result.Refutations = append(result.Refutations, "tuple-invalid:"+receipt.ReceiptID)
		}
		if index > 0 {
			previous := c.Chain.Receipts[index-1]
			previousDigest := result.ReceiptDigests[index-1]
			if receipt.Tuple.Generation != previous.Tuple.Generation+1 {
				coherenceSignals = append(coherenceSignals, signalGenerationGap)
				result.Refutations = append(result.Refutations, "generation-gap:"+receipt.ReceiptID)
			}
			if receipt.Tuple.PriorReceiptDigest != previousDigest {
				coherenceSignals = append(coherenceSignals, signalPriorMismatch)
				result.Refutations = append(result.Refutations, "prior-mismatch:"+receipt.ReceiptID)
			}
			if !sameStableTuple(previous.Tuple, receipt.Tuple) {
				coherenceSignals = append(coherenceSignals, signalScopeHeadMismatch)
				result.Refutations = append(result.Refutations, "scope-head-mismatch:"+receipt.ReceiptID)
			}
			result.RotationCount++
		}
		if !sameTuple(c.Request.Tuple, receipt.Tuple) && index == len(c.Chain.Receipts)-1 {
			if c.Request.Tuple.BaseSHA != receipt.Tuple.BaseSHA || c.Request.Tuple.HeadSHA != receipt.Tuple.HeadSHA || c.Request.Tuple.ProtectedScope != receipt.Tuple.ProtectedScope {
				coherenceSignals = append(coherenceSignals, signalScopeHeadMismatch)
			} else {
				coherenceSignals = append(coherenceSignals, signalPriorMismatch)
			}
			result.Refutations = append(result.Refutations, "request-tuple-mismatch:"+receipt.ReceiptID)
		}

		if receipt.Tuple.IssuerIdentity == "" || receipt.IssuerKeyID == "" {
			foundationSignals = append(foundationSignals, signalIssuerUnavailable)
			result.Unknowns = append(result.Unknowns, unknown("FOUNDATION", "REQUIRE_EXTERNAL_TRUST_ANCHOR", "issuer unavailable", "ISSUER_UNAVAILABLE", "OBTAIN_EXTERNAL_ISSUER_ATTESTATION", []string{"external-issuer"}))
		} else if receipt.ExternalAttestation == nil || receipt.Signature == "" {
			foundationSignals = append(foundationSignals, signalExternalAttestation)
			result.Unknowns = append(result.Unknowns, unknown("FOUNDATION", "REQUIRE_EXTERNAL_TRUST_ANCHOR", "external attestation is missing", "ATTESTATION_MISSING", "COLLECT_EXTERNAL_ATTESTATION", []string{"external-attestation"}))
		} else if revokedKeys[receipt.IssuerKeyID] {
			foundationSignals = append(foundationSignals, signalSignatureInvalid)
		} else {
			issuer, ok := trusted[receipt.IssuerKeyID]
			if !ok {
				foundationSignals = append(foundationSignals, signalIssuerUnavailable)
				result.Unknowns = append(result.Unknowns, unknown("FOUNDATION", "RESOLVE_EXTERNAL_ISSUER_KEY", "issuer key is unavailable to the verifier", "ISSUER_KEY_UNAVAILABLE", "OBTAIN_TRUSTED_ISSUER_KEY", []string{"trust-store"}))
			} else {
				payloadDigest, digestErr := Digest(receipt.Signable())
				if receipt.PublicKey != "" && receipt.PublicKey != issuer.PublicKey {
					foundationSignals = append(foundationSignals, signalSignatureInvalid)
					result.Refutations = append(result.Refutations, "public-key-mismatch:"+receipt.ReceiptID)
				} else if digestErr != nil || issuer.IssuerIdentity != receipt.Tuple.IssuerIdentity || receipt.ExternalAttestation.Issuer != receipt.Tuple.IssuerIdentity || receipt.ExternalAttestation.PayloadDigest != payloadDigest {
					foundationSignals = append(foundationSignals, signalSignatureInvalid)
					result.Refutations = append(result.Refutations, "issuer-mismatch:"+receipt.ReceiptID)
				} else if !verifySignature(issuer, receipt) {
					foundationSignals = append(foundationSignals, signalSignatureInvalid)
					result.Refutations = append(result.Refutations, "signature-invalid:"+receipt.ReceiptID)
				} else {
					foundationSignals = append(foundationSignals, signalTrustedAnchor)
					result.IssuerEvidenceObserved = true
					result.ExternalAttestation = true
				}
			}
		}
		if c.NowUnix > receipt.Tuple.ExpiresAt {
			coherenceSignals = append(coherenceSignals, signalScopeHeadMismatch)
			result.Refutations = append(result.Refutations, "receipt-expired:"+receipt.ReceiptID)
		}
	}
	if len(regressionSignals) == 0 {
		regressionSignals = append(regressionSignals, signalCorpus)
	}
	if len(coherenceSignals) == 0 && len(c.Chain.Receipts) > 0 {
		coherenceSignals = append(coherenceSignals, signalSequentialRotation)
	}

	foundationState, err := stateForSignals(graph, Foundation, foundationSignals)
	if err != nil {
		return VerificationResult{}, err
	}
	coherenceState, err := stateForSignals(graph, Coherence, coherenceSignals)
	if err != nil {
		return VerificationResult{}, err
	}
	regressionState, err := stateForSignals(graph, Regression, regressionSignals)
	if err != nil {
		return VerificationResult{}, err
	}
	result.AxisResults = append(result.AxisResults,
		AxisResult{Axis: Foundation, State: foundationState, Signals: uniqueSorted(foundationSignals), Decision: "FOUNDATION_EXTERNAL_TRUST_ANCHOR"},
		AxisResult{Axis: Coherence, State: coherenceState, Signals: uniqueSorted(coherenceSignals), Decision: "COHERENCE_PRIOR_RECEIPT_CHAIN"},
		AxisResult{Axis: Regression, State: regressionState, Signals: uniqueSorted(regressionSignals), Decision: "REGRESSION_REPLAY_REFUTATION_CORPUS"},
	)
	result.Decision = graph.Combine(foundationState, coherenceState, regressionState)
	return result, nil
}

func stateForSignals(graph SemanticGraph, axis Axis, signals []string) (State, error) {
	if len(signals) == 0 {
		return Unknown, nil
	}
	states := make([]State, 0, len(signals))
	for _, signal := range signals {
		state, err := graph.StateFor(axis, signal)
		if err != nil {
			return "", err
		}
		states = append(states, state)
	}
	return graph.Combine(states...), nil
}

func validateTuple(tuple AuthorizationTuple) error {
	if tuple.BaseSHA == "" || tuple.HeadSHA == "" || tuple.PullRequest < 1 || tuple.ProtectedScope == "" || tuple.Nonce == "" || tuple.IssuedAt < 1 || tuple.ExpiresAt <= tuple.IssuedAt || tuple.Generation < 1 || tuple.Decision == "" {
		return fmt.Errorf("incomplete exact authorization tuple")
	}
	return nil
}

func sameStableTuple(a, b AuthorizationTuple) bool {
	return a.BaseSHA == b.BaseSHA && a.HeadSHA == b.HeadSHA && a.PullRequest == b.PullRequest && a.ProtectedScope == b.ProtectedScope && a.ActorIdentity == b.ActorIdentity && a.IssuerIdentity == b.IssuerIdentity && a.Decision == b.Decision
}

func sameTuple(a, b AuthorizationTuple) bool {
	return sameStableTuple(a, b) && a.Nonce == b.Nonce && a.IssuedAt == b.IssuedAt && a.ExpiresAt == b.ExpiresAt && a.Generation == b.Generation && a.PriorReceiptDigest == b.PriorReceiptDigest
}

func verifySignature(issuer TrustedIssuer, receipt CandidateReceipt) bool {
	publicKey, err := base64.RawStdEncoding.DecodeString(issuer.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return false
	}
	signature, err := base64.RawStdEncoding.DecodeString(receipt.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return false
	}
	payload, err := CanonicalJSON(receipt.Signable())
	if err != nil {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(publicKey), payload, signature)
}

func unknown(stage, step, reason, class, next string, blocked []string) UnknownDetail {
	sort.Strings(blocked)
	return UnknownDetail{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blocked}
}

func stringSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}

func uniqueSorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	unique := result[:0]
	for _, value := range result {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}

func joinStates(states []State) string {
	values := make([]string, len(states))
	for i, state := range states {
		values[i] = string(state)
	}
	return strings.Join(values, ">")
}
