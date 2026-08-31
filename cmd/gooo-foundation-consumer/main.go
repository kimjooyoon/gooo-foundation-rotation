package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kimjooyoon/gooo-foundation-rotation/internal/protocol"
)

type consumerObservation struct {
	CaseID      string `json:"case_id"`
	Decision    string `json:"decision"`
	AxisResults []struct {
		Axis  string `json:"axis"`
		State string `json:"state"`
	} `json:"axis_results"`
}

type ConsumerReceipt struct {
	Schema              string                `json:"schema"`
	ConsumerIdentity    string                `json:"consumer_identity"`
	SourceDirectory     string                `json:"source_directory"`
	ObservedCases       int                   `json:"observed_cases"`
	ObservedDecisions   map[string]int        `json:"observed_decisions"`
	IndependentDigest   string                `json:"independent_digest"`
	Observations        []consumerObservation `json:"observations"`
	SelfApprovalAllowed bool                  `json:"self_approval_allowed"`
}

func main() {
	set := flag.NewFlagSet("consume", flag.ExitOnError)
	inputDir := set.String("input-dir", "evidence/cases", "verifier result directory")
	output := set.String("output", "evidence/independent-consumer-receipt.json", "consumer receipt output")
	set.Parse(os.Args[1:])
	paths, err := filepath.Glob(filepath.Join(*inputDir, "*", "verification.json"))
	if err != nil {
		fatal("find verifier results: %v", err)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		fatal("no verifier results found")
	}
	receipt := ConsumerReceipt{
		Schema:           protocol.ConsumerReceiptSchema,
		ConsumerIdentity: "independent-consumer:gooo-foundation-rotation",
		SourceDirectory:  *inputDir,
		ObservedCases:    len(paths), ObservedDecisions: map[string]int{}, Observations: []consumerObservation{},
		SelfApprovalAllowed: false,
	}
	canonicalInputs := make([]json.RawMessage, 0, len(paths))
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			fatal("read %s: %v", path, err)
		}
		var observation consumerObservation
		if err := json.Unmarshal(raw, &observation); err != nil {
			fatal("decode %s: %v", path, err)
		}
		if observation.CaseID == "" || observation.Decision == "" {
			fatal("incomplete observation %s", path)
		}
		receipt.ObservedDecisions[observation.Decision]++
		receipt.Observations = append(receipt.Observations, observation)
		canonicalInputs = append(canonicalInputs, raw)
	}
	canonical, err := protocol.CanonicalJSON(canonicalInputs)
	if err != nil {
		fatal("canonicalize observations: %v", err)
	}
	receipt.IndependentDigest = protocol.DigestBytes(canonical)
	if err := protocol.WriteJSON(*output, receipt); err != nil {
		fatal("write consumer receipt: %v", err)
	}
	fmt.Printf("consumer=%s cases=%d\n", receipt.ConsumerIdentity, receipt.ObservedCases)
}

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
