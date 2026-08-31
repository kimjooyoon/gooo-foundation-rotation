package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kimjooyoon/gooo-foundation-rotation/internal/protocol"
)

func main() {
	if len(os.Args) < 2 {
		fatal("usage: gooo-foundation-rotation <compile|verify|conformance>")
	}
	switch os.Args[1] {
	case "compile":
		compile(os.Args[2:])
	case "verify":
		verify(os.Args[2:])
	case "conformance":
		conformance(os.Args[2:])
	default:
		fatal("unknown command %q", os.Args[1])
	}
}

func compile(args []string) {
	set := flag.NewFlagSet("compile", flag.ExitOnError)
	source := set.String("source", "semantic/foundation-authorization.gooo", "semantic graph source")
	output := set.String("output", "semantic-ir.json", "compiled semantic graph output")
	set.Parse(args)
	raw, err := os.ReadFile(*source)
	if err != nil {
		fatal("read semantic graph: %v", err)
	}
	graph, err := protocol.ParseSemanticGraph(raw)
	if err != nil {
		fatal("parse semantic graph: %v", err)
	}
	record := struct {
		Schema       string                 `json:"schema"`
		SourceDigest string                 `json:"source_digest"`
		Cells        int                    `json:"cells"`
		Axes         []protocol.Axis        `json:"axes"`
		Precedence   []protocol.State       `json:"precedence"`
		Graph        protocol.SemanticGraph `json:"graph"`
	}{
		Schema:       "gooo/foundation-authorization/semantic-ir/v1",
		SourceDigest: protocol.DigestBytes(raw),
		Cells:        len(graph.Rules), Axes: graph.Axes, Precedence: graph.Precedence, Graph: graph,
	}
	if err := protocol.WriteJSON(*output, record); err != nil {
		fatal("write compiled graph: %v", err)
	}
}

func verify(args []string) {
	set := flag.NewFlagSet("verify", flag.ExitOnError)
	casePath := set.String("case", "", "case JSON")
	graphPath := set.String("graph", "semantic/foundation-authorization.gooo", "semantic graph source")
	trustPath := set.String("trust", "fixtures/trusted-issuers.json", "trusted issuer fixture")
	output := set.String("output", "verification-result.json", "verification result output")
	set.Parse(args)
	if *casePath == "" {
		fatal("--case is required")
	}
	graph, trust, err := loadAuthority(*graphPath, *trustPath)
	if err != nil {
		fatal("load authority: %v", err)
	}
	var c protocol.ConformanceCase
	if err := protocol.ReadJSON(*casePath, &c); err != nil {
		fatal("load case: %v", err)
	}
	result, err := protocol.VerifyCase(graph, trust, c)
	if err != nil {
		fatal("verify case: %v", err)
	}
	if err := protocol.WriteJSON(*output, result); err != nil {
		fatal("write verification result: %v", err)
	}
	fmt.Printf("case=%s decision=%s\n", c.CaseID, result.Decision)
}

func conformance(args []string) {
	set := flag.NewFlagSet("conformance", flag.ExitOnError)
	casesDir := set.String("cases-dir", "fixtures/cases", "directory containing cases")
	graphPath := set.String("graph", "semantic/foundation-authorization.gooo", "semantic graph source")
	trustPath := set.String("trust", "fixtures/trusted-issuers.json", "trusted issuer fixture")
	outputDir := set.String("output-dir", "evidence/cases", "conformance output directory")
	set.Parse(args)
	graph, trust, err := loadAuthority(*graphPath, *trustPath)
	if err != nil {
		fatal("load authority: %v", err)
	}
	paths, err := filepath.Glob(filepath.Join(*casesDir, "*.json"))
	if err != nil {
		fatal("find cases: %v", err)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		fatal("no conformance cases found")
	}
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fatal("create conformance output: %v", err)
	}
	index := protocol.CorpusIndex{
		Schema:   "gooo/foundation-authorization/conformance-index/v1",
		Decision: protocol.Closed, Cases: []protocol.CorpusIndexCase{},
		States: map[string]int{}, ProofChoices: map[string]int{}, Indicators: map[string]int{},
		Conformance: true,
	}
	for _, casePath := range paths {
		var c protocol.ConformanceCase
		if err := protocol.ReadJSON(casePath, &c); err != nil {
			fatal("load %s: %v", casePath, err)
		}
		result, err := protocol.VerifyCase(graph, trust, c)
		if err != nil {
			fatal("verify %s: %v", c.CaseID, err)
		}
		if err := checkExpected(c, result); err != nil {
			fatal("case %s: %v", c.CaseID, err)
		}
		caseDir := filepath.Join(*outputDir, c.CaseID)
		if err := os.MkdirAll(caseDir, 0o755); err != nil {
			fatal("create %s: %v", caseDir, err)
		}
		if err := protocol.WriteJSON(filepath.Join(caseDir, "verification.json"), result); err != nil {
			fatal("write %s: %v", c.CaseID, err)
		}
		index.Cases = append(index.Cases, protocol.CorpusIndexCase{
			CaseID: c.CaseID, Decision: result.Decision, ProofChoice: c.ProofChoice, Indicator: c.Indicator,
			ReceiptCount: len(c.Chain.Receipts), RotationCount: result.RotationCount,
			ReplayCount: result.ReplayCount, RevocationCount: result.RevocationCount, OutputFiles: 1,
		})
		index.States[string(result.Decision)]++
		index.ProofChoices[string(c.ProofChoice)]++
		index.Indicators[c.Indicator]++
		index.Receipts += len(c.Chain.Receipts)
		index.Rotations += result.RotationCount
		index.Replays += result.ReplayCount
		index.Revocations += result.RevocationCount
	}
	if err := protocol.WriteJSON(filepath.Join(*outputDir, "conformance-index.json"), index); err != nil {
		fatal("write conformance index: %v", err)
	}
	fmt.Printf("cases=%d decision=%s\n", len(index.Cases), index.Decision)
}

func loadAuthority(graphPath, trustPath string) (protocol.SemanticGraph, protocol.TrustStore, error) {
	raw, err := os.ReadFile(graphPath)
	if err != nil {
		return protocol.SemanticGraph{}, protocol.TrustStore{}, err
	}
	graph, err := protocol.ParseSemanticGraph(raw)
	if err != nil {
		return protocol.SemanticGraph{}, protocol.TrustStore{}, err
	}
	var trust protocol.TrustStore
	if err := protocol.ReadJSON(trustPath, &trust); err != nil {
		return protocol.SemanticGraph{}, protocol.TrustStore{}, err
	}
	return graph, trust, nil
}

func checkExpected(c protocol.ConformanceCase, result protocol.VerificationResult) error {
	if result.Decision != c.Expected.Decision {
		return fmt.Errorf("decision=%s expected=%s", result.Decision, c.Expected.Decision)
	}
	want := map[protocol.Axis]protocol.State{
		protocol.Foundation: c.Expected.Foundation,
		protocol.Coherence:  c.Expected.Coherence,
		protocol.Regression: c.Expected.Regression,
	}
	for _, axis := range result.AxisResults {
		if axis.State != want[axis.Axis] {
			return fmt.Errorf("%s=%s expected=%s", axis.Axis, axis.State, want[axis.Axis])
		}
	}
	return nil
}

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
