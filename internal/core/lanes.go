package core

import (
	"fmt"
	"strings"

	"github.com/d2lang/d2/d2graph"

	"github.com/inkcheck/inkline/internal/ids"
)

// laneCellWidth is the width of every step in horizontal lanes.
const laneCellWidth = 192

// laneLayout lays swimlanes out as a matrix: lanes are cells of a root grid and
// every lane is itself a grid with one cell per rank of the flow. A step's rank
// is its longest-path distance from the flow's start, so a handoff always moves
// forward and steps at the same moment line up across lanes.
//
// Grid cells take the order in which their objects are first declared, so the
// returned D2 must precede the author's source.
func laneLayout(g *d2graph.Graph, h Header) string {
	lanes := g.Root.ChildrenArray
	if len(lanes) == 0 {
		return ""
	}
	laneOf := map[*d2graph.Object]*d2graph.Object{}
	var steps []*d2graph.Object
	for _, l := range lanes {
		for _, s := range l.ChildrenArray {
			laneOf[s] = l
			steps = append(steps, s)
		}
	}
	rank := rankSteps(steps, g.Edges, laneOf)

	ranks := 1
	for _, r := range rank {
		ranks = max(ranks, r+1)
	}

	// Vertical lanes align because steps share a height (a design-system
	// concern); horizontal lanes need a shared width, which the core sets.
	outer, inner, size := "grid-columns", "grid-rows", ""
	if h.Orientation == "horizontal" {
		outer, inner, size = "grid-rows", "grid-columns", fmt.Sprintf("width: %d", laneCellWidth)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s: %d\n", outer, len(lanes))
	gaps := 0
	for _, l := range lanes {
		cells := make([]*d2graph.Object, ranks)
		for _, s := range l.ChildrenArray {
			cells[rank[s]] = s
		}
		fmt.Fprintf(&b, "%s: {\n  %s: %d\n", l.AbsID(), inner, ranks)
		for _, c := range cells {
			if c != nil {
				fmt.Fprintf(&b, "  %s: {%s}\n", c.ID, size)
				continue
			}
			gaps++
			fmt.Fprintf(&b, "  %s%d: \"\" {class: gap; %s}\n", ids.GapPrefix, gaps, size)
		}
		b.WriteString("}\n")
	}
	return b.String()
}

// rankSteps assigns each step the length of the longest path reaching it,
// ignoring edges that close a cycle. Two steps of one lane never share a rank:
// the later one moves down and its successors follow.
func rankSteps(steps []*d2graph.Object, edges []*d2graph.Edge, laneOf map[*d2graph.Object]*d2graph.Object) map[*d2graph.Object]int {
	out := map[*d2graph.Object][]*d2graph.Object{}
	for _, e := range edges {
		if laneOf[e.Src] != nil && laneOf[e.Dst] != nil && e.Src != e.Dst {
			out[e.Src] = append(out[e.Src], e.Dst)
		}
	}

	// Depth-first search in declaration order; an edge to a step still on the
	// stack is a back edge (a loop in the flow) and does not constrain ranks.
	const (
		unvisited = iota
		active
		done
	)
	state := map[*d2graph.Object]int{}
	back := map[[2]*d2graph.Object]bool{}
	var order []*d2graph.Object
	var visit func(s *d2graph.Object)
	visit = func(s *d2graph.Object) {
		state[s] = active
		for _, t := range out[s] {
			switch state[t] {
			case unvisited:
				visit(t)
			case active:
				back[[2]*d2graph.Object{s, t}] = true
			}
		}
		state[s] = done
		order = append(order, s)
	}
	for _, s := range steps {
		if state[s] == unvisited {
			visit(s)
		}
	}
	// order is a reverse topological order of the acyclic edges.
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	rank := map[*d2graph.Object]int{}
	for range len(steps) + 1 {
		for _, s := range order {
			for _, t := range out[s] {
				if !back[[2]*d2graph.Object{s, t}] && rank[t] < rank[s]+1 {
					rank[t] = rank[s] + 1
				}
			}
		}
		// Resolve collisions within a lane, then propagate again.
		moved := false
		taken := map[*d2graph.Object]map[int]bool{}
		for _, s := range order {
			l := laneOf[s]
			if taken[l] == nil {
				taken[l] = map[int]bool{}
			}
			for taken[l][rank[s]] {
				rank[s]++
				moved = true
			}
			taken[l][rank[s]] = true
		}
		if !moved {
			break
		}
	}
	return rank
}
