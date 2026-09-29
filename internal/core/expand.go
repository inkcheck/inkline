package core

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2compiler"
	"github.com/d2lang/d2/d2format"
	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2parser"
)

// Source is one inkline file to expand.
type Source struct {
	// Path is the file's path. Imports resolve against its directory within
	// FS, or on disk when FS is nil.
	Path string
	FS   fs.FS
	Text string
}

// File is a parsed inkline file.
type File struct {
	Source
	Header Header
	Spec   *Spec

	ast *d2ast.Map
}

// Parse parses a inkline file and reads its header.
func Parse(src Source) (*File, error) {
	ast, err := d2parser.Parse(src.Path, strings.NewReader(src.Text), nil)
	if err != nil {
		return nil, err
	}
	h, err := ReadHeader(ast)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src.Path, err)
	}
	spec, _ := Lookup(h.Diagram)
	return &File{Source: src, Header: h, Spec: spec, ast: ast}, nil
}

// Expand returns plain D2: prelude, the author's (transformed) source, and a
// role overlay that gives every object and connection its role.
//
// The prelude comes from the design system; it is compiled together with the
// source so the overlay sees the final shapes (a class may set shape: sql_table).
func (f *File) Expand(prelude string) (string, error) {
	spec := f.Spec
	b := newBody()
	if spec.Transform == nil {
		b.source(f.ast)
	} else if err := spec.Transform(f.ast, b); err != nil {
		return "", err
	}
	body := b.String()

	// Compile once to learn every object's final classes and shape.
	analysis := prelude + "\n" + body
	if spec.AnalysisShape != "" {
		analysis += "\n*.shape: " + spec.AnalysisShape + "\n"
	}
	g, _, err := d2compiler.Compile(f.Path, strings.NewReader(analysis), &d2compiler.CompileOptions{FS: f.FS})
	if err != nil {
		return "", b.locate(err, f.Path, strings.Count(prelude, "\n")+1)
	}

	// D2 ignores a class list assigned to an object that already has a class,
	// so the author's class keys are removed and the overlay assigns each
	// object's complete list, role first, exactly once.
	bodyAST, err := d2parser.Parse(f.Path, strings.NewReader(body), nil)
	if err != nil {
		return "", err
	}
	stripClasses(bodyAST)

	var doc strings.Builder
	doc.WriteString(prelude)
	if spec.Lanes {
		// Lane layout must precede the source; see laneLayout.
		doc.WriteString("\n# --- inkline lane layout ---\n" + laneLayout(g, f.Header))
	}
	doc.WriteString("\n# --- source ---\n" + d2format.Format(bodyAST))
	doc.WriteString("\n# --- inkline roles ---\n" + f.overlay(g, b.hints))
	return doc.String(), nil
}

// stripClasses removes class assignments from objects and connections. Style
// definitions (classes, vars) are left alone.
func stripClasses(m *d2ast.Map) {
	kept := m.Nodes[:0]
	for _, n := range m.Nodes {
		k := n.MapKey
		if k == nil {
			kept = append(kept, n)
			continue
		}
		if isClassKey(k) {
			continue
		}
		if k.Value.Map != nil && !isStyleDefinition(k) {
			stripClasses(k.Value.Map)
		}
		kept = append(kept, n)
	}
	m.Nodes = kept
}

func isClassKey(k *d2ast.Key) bool {
	kp := k.Key
	if len(k.Edges) > 0 {
		kp = k.EdgeKey
	}
	return kp != nil && len(kp.Path) > 0 && kp.Path[len(kp.Path)-1].ScalarString() == "class"
}

func isStyleDefinition(k *d2ast.Key) bool {
	if k.Key == nil || len(k.Edges) > 0 || len(k.Key.Path) != 1 {
		return false
	}
	name := k.Key.Path[0].ScalarString()
	return name == "classes" || name == "vars"
}

// overlay assigns roles and applies diagram-level settings.
func (f *File) overlay(g *d2graph.Graph, hints map[string]string) string {
	spec := f.Spec
	var b strings.Builder
	b.WriteString(spec.Settings)

	lowerHints := map[string]string{}
	for k, v := range hints {
		lowerHints[strings.ToLower(k)] = v
	}

	for _, obj := range g.Objects {
		// Objects that take no role keep the author's classes.
		classes := obj.Classes
		takesRole := !(spec.TopLevelOnly && !topLevel(obj) || spec.Skip != nil && spec.Skip(obj))
		role := ""
		if takesRole {
			role = objectRole(obj, spec, lowerHints)
			classes = withRole(role, classes)
		}
		if len(classes) == 0 {
			continue
		}
		id := obj.AbsID()
		if len(obj.Classes) > 0 && slices.ContainsFunc(obj.References, func(r d2graph.Reference) bool { return f.imported(r.MapKey) }) {
			// stripClasses does not reach into imports.
			fmt.Fprintf(&b, "%s.class: null\n", id)
		}
		fmt.Fprintf(&b, "%s.class: %s\n", id, classList(classes))
		if takesRole && spec.Annotate != nil {
			b.WriteString(spec.Annotate(obj, id, role))
		}
	}

	seen := map[string]int{}
	for _, e := range g.Edges {
		role := spec.Edges[0]
		for _, c := range e.Classes {
			if slices.Contains(spec.Edges, c) {
				role = c
				break
			}
		}
		id := edgeKey(e, seen)
		if len(e.Classes) > 0 && slices.ContainsFunc(e.References, func(r d2graph.EdgeReference) bool { return f.imported(r.MapKey) }) {
			// Written as a map: D2 reads (a -> b)[0].class: null as removing
			// the connection.
			fmt.Fprintf(&b, "%s: {class: null}\n", id)
		}
		fmt.Fprintf(&b, "%s.class: %s\n", id, classList(withRole(role, e.Classes)))
	}
	return b.String()
}

// imported reports whether a key is written in an imported file.
func (f *File) imported(k *d2ast.Key) bool {
	return k != nil && k.Range.Path != f.Path
}

// edgeKey returns the D2 key of a connection, like Edge.AbsID but keeping the
// table columns that sql_table connections attach to: D2 records those
// connections between the tables, while their key names the columns.
// seen counts earlier connections with the same key, giving the index. D2
// counts a connection and its mirror image (a -> b and b <- a, a -- b and
// b -- a) as one key, so both count together.
func edgeKey(e *d2graph.Edge, seen map[string]int) string {
	src, dst := endpoint(e.Src, e.SrcTableColumnIndex), endpoint(e.Dst, e.DstTableColumnIndex)
	var common []string
	for len(src) > 1 && len(dst) > 1 && strings.EqualFold(src[0], dst[0]) {
		common = append(common, src[0])
		src, dst = src[1:], dst[1:]
	}
	scope := ""
	if len(common) > 0 {
		scope = strings.Join(common, ".") + "."
	}
	s, d, arrow := strings.Join(src, "."), strings.Join(dst, "."), e.ArrowString()
	mirror := map[string]string{"->": "<-", "<-": "->"}[arrow]
	if mirror == "" {
		mirror = arrow
	}
	count := strings.ToLower(scope) + min(strings.ToLower(s+" "+arrow+" "+d), strings.ToLower(d+" "+mirror+" "+s))
	i := seen[count]
	seen[count]++
	return fmt.Sprintf("%s(%s %s %s)[%d]", scope, s, arrow, d, i)
}

func endpoint(obj *d2graph.Object, column *int) []string {
	ida := obj.AbsIDArray()
	if column != nil && obj.SQLTable != nil {
		ida = append(ida, key(obj.SQLTable.Columns[*column].Name.Label))
	}
	return ida
}

func objectRole(obj *d2graph.Object, spec *Spec, hints map[string]string) string {
	for _, c := range obj.Classes {
		if slices.Contains(spec.Nodes, c) || slices.Contains(spec.Containers, c) {
			return c
		}
	}
	if r, ok := hints[strings.ToLower(obj.AbsID())]; ok {
		return r
	}
	if spec.DefaultRole != nil {
		if r := spec.DefaultRole(obj); r != "" {
			return r
		}
	}
	if obj.IsContainer() {
		return spec.Containers[0]
	}
	return spec.Nodes[0]
}

// withRole puts role first and keeps the author's other classes after it, so
// modifiers and design-system classes override the role's styling.
func withRole(role string, classes []string) []string {
	out := []string{role}
	for _, c := range classes {
		if c != role {
			out = append(out, c)
		}
	}
	return out
}

func classList(classes []string) string {
	quoted := make([]string, len(classes))
	for i, c := range classes {
		quoted[i] = value(c)
	}
	return "[" + strings.Join(quoted, "; ") + "]"
}
