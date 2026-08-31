package protocol

import "encoding/json"

const (
	RequestSchema         = "gooo/foundation-authorization/request/v1"
	ReceiptSchema         = "gooo/foundation-authorization/receipt/v1"
	ResultSchema          = "gooo/foundation-authorization/verification-result/v1"
	ChainSchema           = "gooo/foundation-authorization/rotation-chain/v1"
	RevocationSchema      = "gooo/foundation-authorization/revocation-set/v1"
	CaseSchema            = "gooo/foundation-authorization/case/v1"
	CorpusSchema          = "gooo/foundation-authorization/corpus/v1"
	ConsumerReceiptSchema = "gooo/foundation-authorization/independent-consumer-receipt/v1"
)

type State string

const (
	Closed  State = "CLOSED"
	Unknown State = "UNKNOWN"
	Refuted State = "REFUTED"
)

type Axis string

const (
	Foundation Axis = "FOUNDATION"
	Coherence  Axis = "COHERENCE"
	Regression Axis = "REGRESSION"
)

type AuthorizationDecision string

const (
	Authorize AuthorizationDecision = "AUTHORIZE"
	Deny      AuthorizationDecision = "DENY"
)

type AuthorizationTuple struct {
	BaseSHA            string                `json:"base_sha"`
	HeadSHA            string                `json:"head_sha"`
	PullRequest        int                   `json:"pull_request"`
	ProtectedScope     string                `json:"protected_scope"`
	ActorIdentity      string                `json:"actor_identity"`
	IssuerIdentity     string                `json:"issuer_identity"`
	Nonce              string                `json:"nonce"`
	IssuedAt           int64                 `json:"issued_at"`
	ExpiresAt          int64                 `json:"expires_at"`
	Generation         int                   `json:"generation"`
	PriorReceiptDigest string                `json:"prior_receipt_digest"`
	Decision           AuthorizationDecision `json:"decision"`
}

type EvidenceReference struct {
	Kind      string `json:"kind"`
	Reference string `json:"reference"`
	Digest    string `json:"digest"`
	ReadOnly  bool   `json:"read_only"`
}

type AuthorizationRequest struct {
	Schema             string              `json:"schema"`
	RequestID          string              `json:"request_id"`
	Tuple              AuthorizationTuple  `json:"tuple"`
	RequestedOperation string              `json:"requested_operation"`
	IssuerRequired     bool                `json:"issuer_required"`
	ConsumerIdentity   string              `json:"consumer_identity"`
	EvidenceBasis      []EvidenceReference `json:"evidence_basis"`
}

type ExternalAttestation struct {
	Type            string `json:"type"`
	Issuer          string `json:"issuer"`
	Subject         string `json:"subject"`
	PayloadDigest   string `json:"payload_digest"`
	TransparencyLog string `json:"transparency_log"`
	Environment     string `json:"environment"`
}

type ReceiptSignable struct {
	Schema      string             `json:"schema"`
	ReceiptID   string             `json:"receipt_id"`
	RequestID   string             `json:"request_id"`
	Tuple       AuthorizationTuple `json:"tuple"`
	IssuerKeyID string             `json:"issuer_key_id"`
}

type CandidateReceipt struct {
	Schema              string               `json:"schema"`
	ReceiptID           string               `json:"receipt_id"`
	RequestID           string               `json:"request_id"`
	Tuple               AuthorizationTuple   `json:"tuple"`
	IssuerKeyID         string               `json:"issuer_key_id"`
	PublicKey           string               `json:"public_key,omitempty"`
	Signature           string               `json:"signature,omitempty"`
	ExternalAttestation *ExternalAttestation `json:"external_attestation,omitempty"`
}

type RotationChain struct {
	Schema    string             `json:"schema"`
	RequestID string             `json:"request_id"`
	Receipts  []CandidateReceipt `json:"receipts"`
}

type RevokedKey struct {
	KeyID          string `json:"key_id"`
	IssuerIdentity string `json:"issuer_identity"`
	RevokedAt      int64  `json:"revoked_at"`
	Reason         string `json:"reason"`
}

type RevocationSet struct {
	Schema         string       `json:"schema"`
	Keys           []RevokedKey `json:"keys"`
	ReceiptDigests []string     `json:"receipt_digests"`
	Nonces         []string     `json:"nonces"`
}

type TrustedIssuer struct {
	KeyID          string `json:"key_id"`
	IssuerIdentity string `json:"issuer_identity"`
	PublicKey      string `json:"public_key"`
	Anchor         string `json:"anchor"`
}

type TrustStore struct {
	Schema  string          `json:"schema"`
	Issuers []TrustedIssuer `json:"issuers"`
}

type UnknownDetail struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type AxisResult struct {
	Axis     Axis     `json:"axis"`
	State    State    `json:"state"`
	Signals  []string `json:"signals"`
	Decision string   `json:"decision"`
}

type VerificationResult struct {
	Schema                 string          `json:"schema"`
	CaseID                 string          `json:"case_id"`
	RequestID              string          `json:"request_id"`
	Decision               State           `json:"decision"`
	DecisionRule           string          `json:"decision_rule"`
	AxisResults            []AxisResult    `json:"axis_results"`
	Unknowns               []UnknownDetail `json:"unknowns"`
	Refutations            []string        `json:"refutations"`
	ReceiptDigests         []string        `json:"receipt_digests"`
	RotationCount          int             `json:"rotation_count"`
	ReplayCount            int             `json:"replay_count"`
	RevocationCount        int             `json:"revocation_count"`
	IssuerEvidenceObserved bool            `json:"issuer_evidence_observed"`
	ExternalAttestation    bool            `json:"external_attestation"`
}

type ExpectedResult struct {
	Decision   State `json:"decision"`
	Foundation State `json:"foundation"`
	Coherence  State `json:"coherence"`
	Regression State `json:"regression"`
}

type ConformanceCase struct {
	Schema      string               `json:"schema"`
	CaseID      string               `json:"case_id"`
	ProofChoice Axis                 `json:"proof_choice"`
	Indicator   string               `json:"indicator"`
	NowUnix     int64                `json:"now_unix"`
	Request     AuthorizationRequest `json:"request"`
	Chain       RotationChain        `json:"chain"`
	Revocations RevocationSet        `json:"revocations"`
	Expected    ExpectedResult       `json:"expected"`
}

type Corpus struct {
	Schema string            `json:"schema"`
	Cases  []ConformanceCase `json:"cases"`
}

type CorpusIndexCase struct {
	CaseID          string `json:"case_id"`
	Decision        State  `json:"decision"`
	ProofChoice     Axis   `json:"proof_choice"`
	Indicator       string `json:"indicator"`
	ReceiptCount    int    `json:"receipt_count"`
	RotationCount   int    `json:"rotation_count"`
	ReplayCount     int    `json:"replay_count"`
	RevocationCount int    `json:"revocation_count"`
	OutputFiles     int    `json:"output_files"`
}

type CorpusIndex struct {
	Schema       string            `json:"schema"`
	Decision     State             `json:"decision"`
	Cases        []CorpusIndexCase `json:"cases"`
	States       map[string]int    `json:"states"`
	ProofChoices map[string]int    `json:"proof_choices"`
	Indicators   map[string]int    `json:"indicators"`
	Receipts     int               `json:"receipts"`
	Rotations    int               `json:"rotations"`
	Replays      int               `json:"replays"`
	Revocations  int               `json:"revocations"`
	Conformance  bool              `json:"conformance"`
}

func (r CandidateReceipt) Signable() ReceiptSignable {
	return ReceiptSignable{
		Schema: r.Schema, ReceiptID: r.ReceiptID, RequestID: r.RequestID,
		Tuple: r.Tuple, IssuerKeyID: r.IssuerKeyID,
	}
}

func (r CandidateReceipt) MarshalJSON() ([]byte, error) {
	type alias CandidateReceipt
	return json.Marshal(alias(r))
}
