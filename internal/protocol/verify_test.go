package protocol

import "testing"

func TestPrecedenceIsExplicitAndNonScoring(t *testing.T) {
	graph, err := ParseSemanticGraph([]byte("graph x version=0.1.0\nstates CLOSED UNKNOWN REFUTED precedence=REFUTED>UNKNOWN>CLOSED\naxis FOUNDATION meaning=anchor\naxis COHERENCE meaning=chain\naxis REGRESSION meaning=corpus\ncell 01 axis=FOUNDATION signal=a state=CLOSED\ncell 02 axis=FOUNDATION signal=b state=UNKNOWN\ncell 03 axis=FOUNDATION signal=c state=REFUTED\ncell 04 axis=COHERENCE signal=a state=CLOSED\ncell 05 axis=COHERENCE signal=b state=UNKNOWN\ncell 06 axis=COHERENCE signal=c state=REFUTED\ncell 07 axis=REGRESSION signal=a state=CLOSED\ncell 08 axis=REGRESSION signal=b state=UNKNOWN\ncell 09 axis=REGRESSION signal=c state=REFUTED\ncell 10 axis=REGRESSION signal=d state=CLOSED\ncell 11 axis=REGRESSION signal=e state=UNKNOWN\ncell 12 axis=REGRESSION signal=f state=REFUTED\n"))
	if err != nil {
		t.Fatal(err)
	}
	if graph.Combine(Closed, Unknown) != Unknown || graph.Combine(Unknown, Refuted) != Refuted || graph.Combine(Closed, Closed) != Closed {
		t.Fatalf("precedence is not REFUTED > UNKNOWN > CLOSED")
	}
}

func TestUnknownDetailHasSixCoordinates(t *testing.T) {
	detail := unknown("FOUNDATION", "STEP", "reason", "CLASS", "NEXT", []string{"blocker"})
	if detail.Stage == "" || detail.Step == "" || detail.Reason == "" || detail.UnknownClass == "" || detail.NextOperation == "" || detail.BlockedBy == nil {
		t.Fatalf("unknown detail is incomplete: %#v", detail)
	}
}
