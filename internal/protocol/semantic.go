package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

type SemanticRule struct {
	Cell   int
	Axis   Axis
	Signal string
	State  State
}

type SemanticTransition struct {
	From      string
	To        string
	Operation string
}

type SemanticGraph struct {
	Schema      string
	Version     string
	States      []State
	Precedence  []State
	Axes        []Axis
	Rules       []SemanticRule
	Transitions []SemanticTransition
}

func ParseSemanticGraph(raw []byte) (SemanticGraph, error) {
	graph := SemanticGraph{}
	for lineNumber, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		values := make(map[string]string)
		for _, field := range fields[1:] {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				values[parts[0]] = parts[1]
			}
		}
		switch fields[0] {
		case "graph":
			if len(fields) < 3 || values["version"] == "" {
				return graph, fmt.Errorf("semantic graph line %d has no version", lineNumber+1)
			}
			graph.Schema = fields[1]
			graph.Version = values["version"]
		case "states":
			if len(fields) < 4 || values["precedence"] == "" {
				return graph, fmt.Errorf("semantic graph line %d has invalid states", lineNumber+1)
			}
			for _, state := range fields[1:4] {
				graph.States = append(graph.States, State(state))
			}
			for _, state := range strings.Split(values["precedence"], ">") {
				graph.Precedence = append(graph.Precedence, State(state))
			}
		case "axis":
			if len(fields) < 2 || values["meaning"] == "" {
				return graph, fmt.Errorf("semantic graph line %d has invalid axis", lineNumber+1)
			}
			graph.Axes = append(graph.Axes, Axis(fields[1]))
		case "cell":
			if len(fields) < 2 || values["axis"] == "" || values["signal"] == "" || values["state"] == "" {
				return graph, fmt.Errorf("semantic graph line %d has invalid cell", lineNumber+1)
			}
			cell, err := strconv.Atoi(fields[1])
			if err != nil || cell < 1 {
				return graph, fmt.Errorf("semantic graph line %d has invalid cell number", lineNumber+1)
			}
			graph.Rules = append(graph.Rules, SemanticRule{Cell: cell, Axis: Axis(values["axis"]), Signal: values["signal"], State: State(values["state"])})
		case "transition":
			if len(fields) < 2 || values["operation"] == "" {
				return graph, fmt.Errorf("semantic graph line %d has invalid transition", lineNumber+1)
			}
			parts := strings.SplitN(fields[1], "->", 2)
			if len(parts) != 2 {
				return graph, fmt.Errorf("semantic graph line %d has invalid transition path", lineNumber+1)
			}
			graph.Transitions = append(graph.Transitions, SemanticTransition{From: parts[0], To: parts[1], Operation: values["operation"]})
		default:
			return graph, fmt.Errorf("semantic graph line %d has unknown directive %q", lineNumber+1, fields[0])
		}
	}
	if graph.Schema == "" || graph.Version == "" || len(graph.Rules) != 12 || len(graph.Axes) != 3 {
		return graph, fmt.Errorf("semantic graph denominator is not exact: cells=%d axes=%d", len(graph.Rules), len(graph.Axes))
	}
	return graph, nil
}

func (g SemanticGraph) Rule(axis Axis, signal string) (SemanticRule, bool) {
	for _, rule := range g.Rules {
		if rule.Axis == axis && rule.Signal == signal {
			return rule, true
		}
	}
	return SemanticRule{}, false
}

func (g SemanticGraph) StateFor(axis Axis, signal string) (State, error) {
	rule, ok := g.Rule(axis, signal)
	if !ok {
		return "", fmt.Errorf("semantic graph has no rule for %s/%s", axis, signal)
	}
	return rule.State, nil
}

func (g SemanticGraph) Combine(states ...State) State {
	for _, precedence := range g.Precedence {
		for _, state := range states {
			if state == precedence {
				return state
			}
		}
	}
	return Unknown
}

func (g SemanticGraph) AxisCellCounts() map[string]int {
	counts := map[string]int{}
	for _, rule := range g.Rules {
		counts[string(rule.Axis)]++
	}
	return counts
}
