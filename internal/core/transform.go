package core

import (
	"fmt"
	"strings"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2format"

	"github.com/inkcheck/inkline/internal/ids"
)

// pseudoState is how authors write a state diagram's initial/final state.
const pseudoState = "[*]"

// inPlace is a transform that rewrites the author's AST and writes it out.
func inPlace(rewrite func(*d2ast.Map)) func(*d2ast.Map, *body) error {
	return func(ast *d2ast.Map, b *body) error {
		rewrite(ast)
		b.source(ast)
		return nil
	}
}

// expandPseudoStates rewrites "[*]" in every scope: as a source it becomes that
// scope's start state, as a target its end state.
func expandPseudoStates(m *d2ast.Map) {
	for _, n := range m.Nodes {
		k := n.MapKey
		if k == nil {
			continue
		}
		for _, e := range k.Edges {
			if isPseudoState(e.Src) {
				e.Src = d2ast.MakeKeyPath([]string{ids.Start})
			}
			if isPseudoState(e.Dst) {
				e.Dst = d2ast.MakeKeyPath([]string{ids.End})
			}
		}
		if k.Value.Map != nil {
			expandPseudoStates(k.Value.Map)
		}
	}
}

// makeUndirected gives every relationship an arrowhead slot at both ends: D2
// draws an arrowhead only where the connection has an arrow, and ER
// cardinalities and UML composition/aggregation mark the source end. Each
// role sets both ends, using "none" where there is nothing to draw.
func makeUndirected(m *d2ast.Map) {
	for _, n := range m.Nodes {
		k := n.MapKey
		if k == nil {
			continue
		}
		for _, e := range k.Edges {
			e.SrcArrow, e.DstArrow = "<", ">"
		}
		if k.Value.Map != nil {
			makeUndirected(k.Value.Map)
		}
	}
}

func isPseudoState(kp *d2ast.KeyPath) bool {
	return kp != nil && len(kp.Path) == 1 && kp.Path[0].ScalarString() == pseudoState
}

// flattenTree turns nested object maps into a flat set of nodes joined by
// parent -> child links. Reserved D2 keys (class, style, icon, …) inside a node
// stay attributes of that node, as do spread substitutions.
func flattenTree(ast *d2ast.Map, b *body) error {
	b.printf(nil, "direction: right\n")

	// A node may be written more than once (a: A, then a.style.bold: true);
	// it is labelled and linked the first time.
	declared := map[string]bool{}

	var walk func(k *d2ast.Key, parent string, depth int) error
	walk = func(k *d2ast.Key, parent string, depth int) error {
		k = nodeKey(k)
		segs := k.Key.StringIDA()
		id := strings.Join(segs, ".")
		if parent != "" {
			id = parent + "/" + id
		}
		qid := key(id)
		seen := strings.ToLower(qid)

		label, labelled := segs[len(segs)-1], true
		if k.Primary.Unbox() != nil {
			label = k.Primary.ScalarString()
		} else if v, ok := k.Value.Unbox().(d2ast.Scalar); ok && v != nil {
			label = v.ScalarString()
		} else {
			labelled = !declared[seen]
		}
		if k.Value.Import != nil {
			return fmt.Errorf("treeview: %q cannot import its content; nesting defines the tree", label)
		}

		var attrs []d2ast.Node
		var children []*d2ast.Key
		if k.Value.Map != nil {
			for _, n := range k.Value.Map.Nodes {
				ck := n.MapKey
				switch {
				case n.Import != nil:
					return fmt.Errorf("treeview: imports are not allowed inside %q; nesting defines the tree", label)
				case n.Substitution != nil:
					attrs = append(attrs, n.Substitution)
				case ck == nil:
				case len(ck.Edges) > 0:
					return fmt.Errorf("treeview: connections are not allowed inside %q; nesting defines the tree", label)
				case isReserved(ck):
					attrs = append(attrs, ck)
				default:
					children = append(children, ck)
				}
			}
		}

		if labelled {
			b.printf(k, "%s: %s {\n", qid, value(label))
		} else {
			b.printf(k, "%s: {\n", qid)
		}
		for _, a := range attrs {
			ak, _ := a.(*d2ast.Key)
			b.printf(ak, "  %s\n", d2format.Format(a))
		}
		b.printf(k, "}\n")
		if parent != "" && !declared[seen] {
			b.printf(k, "%s -- %s\n", key(parent), qid)
		}

		switch {
		case depth == 0:
			b.hints[qid] = "root"
		case len(children) > 0:
			b.hints[qid] = "branch"
		case !declared[seen]:
			b.hints[qid] = "leaf"
		}
		declared[seen] = true
		for _, c := range children {
			if err := walk(c, id, depth+1); err != nil {
				return err
			}
		}
		return nil
	}

	for _, n := range ast.Nodes {
		k := n.MapKey
		if k == nil || len(k.Edges) > 0 || isReserved(k) {
			b.node(n.Unbox())
			continue
		}
		if err := walk(k, "", 0); err != nil {
			return err
		}
	}
	return nil
}

// nodeKey returns k as the key of a tree node. A dotted key that reaches into
// a node's attributes (a.style.fill: red) is that node holding the attribute
// (a: {style.fill: red}).
func nodeKey(k *d2ast.Key) *d2ast.Key {
	for i, s := range k.Key.Path {
		if _, ok := d2ast.ReservedKeywords[s.ScalarString()]; !ok || i == 0 {
			continue
		}
		attr := *k
		attr.Key = &d2ast.KeyPath{Range: k.Key.Range, Path: k.Key.Path[i:]}
		return &d2ast.Key{
			Range: k.Range,
			Key:   &d2ast.KeyPath{Range: k.Key.Range, Path: k.Key.Path[:i]},
			Value: d2ast.MakeValueBox(&d2ast.Map{Range: k.Range, Nodes: []d2ast.MapNodeBox{{MapKey: &attr}}}),
		}
	}
	return k
}

// wrapQuadrants moves q1..q4 into a 2×2 grid and turns x/y/title into labels.
// Keys and connections that reach into a quadrant (q1.a.style.bold: true,
// q1.a -> q3.b) follow it into the grid.
func wrapQuadrants(ast *d2ast.Map, b *body) error {
	quads := map[string][]*d2ast.Key{}
	var x, y, title *d2ast.Key
	var inside []*d2ast.Key

	for _, n := range ast.Nodes {
		k := n.MapKey
		if k != nil && k.Key != nil && len(k.Edges) == 0 && len(k.Key.Path) == 1 {
			switch name := k.Key.Path[0].ScalarString(); {
			case isQuadrantKey(name):
				quads[name] = append(quads[name], k)
				continue
			case name == "x":
				x = k
				continue
			case name == "y":
				y = k
				continue
			case name == "title":
				title = k
				continue
			}
		}
		if k != nil && intoChart(k) {
			inside = append(inside, k)
			continue
		}
		b.node(n.Unbox())
	}

	label := ""
	if title != nil {
		label = scalar(title)
	}
	b.printf(title, "%s: %s {\n  grid-rows: 2\n  grid-columns: 2\n  grid-gap: 0\n", ids.Chart, value(label))
	// Grid cells fill left to right, top to bottom: q2 q1 / q3 q4.
	for _, q := range []string{"q2", "q1", "q3", "q4"} {
		if len(quads[q]) == 0 {
			b.printf(nil, "%s: \"\"\n", q)
		}
		for _, k := range quads[q] {
			b.node(k)
		}
	}
	b.printf(nil, "}\n")
	if x != nil && scalar(x) != "" {
		b.printf(x, "%s: %s {near: bottom-center}\n", ids.XAxis, value(scalar(x)))
	}
	if y != nil && scalar(y) != "" {
		b.printf(y, "%s: %s {near: center-left}\n", ids.YAxis, value(scalar(y)))
	}
	for _, k := range inside {
		b.node(k)
	}
	return nil
}

// intoChart points a key that reaches into a quadrant at the quadrant's place
// in the chart, and reports whether it did.
func intoChart(k *d2ast.Key) bool {
	chart := func(kp *d2ast.KeyPath) bool {
		if kp == nil || len(kp.Path) == 0 || !isQuadrantKey(kp.Path[0].ScalarString()) {
			return false
		}
		kp.Path = append(d2ast.MakeKeyPath([]string{ids.Chart}).Path, kp.Path...)
		return true
	}
	if k.Key != nil {
		return chart(k.Key)
	}
	moved := false
	for _, e := range k.Edges {
		// Both ends are tried.
		src, dst := chart(e.Src), chart(e.Dst)
		moved = moved || src || dst
	}
	return moved
}

func isQuadrantKey(s string) bool {
	return s == "q1" || s == "q2" || s == "q3" || s == "q4"
}

// isReserved reports whether a key is a D2 keyword (style, class, vars, …)
// rather than an object.
func isReserved(k *d2ast.Key) bool {
	if k.Key == nil || len(k.Key.Path) == 0 {
		return false
	}
	_, ok := d2ast.ReservedKeywords[k.Key.Path[0].ScalarString()]
	return ok
}

// key and value quote s for use as a D2 key or value.
func key(s string) string   { return d2format.Format(d2ast.RawString(s, true)) }
func value(s string) string { return d2format.Format(d2ast.RawString(s, false)) }
