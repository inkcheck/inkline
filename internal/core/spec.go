// Package core turns a inkline source file into plain D2: it expands per-diagram
// shorthand and gives every object and connection a semantic role (see
// spec/roles.md). It knows nothing about how roles look; that is the design
// system's job.
package core

import (
	"fmt"
	"sort"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2graph"

	"github.com/inkcheck/inkline/internal/ids"
)

// Spec describes one diagram type. The first role of each list is the default.
type Spec struct {
	Name       string
	Nodes      []string
	Containers []string
	Edges      []string
	// TopLevelOnly limits object roles to top-level objects, for diagrams whose
	// nested keys are not objects (class fields, table columns, sequence spans).
	TopLevelOnly bool
	// Layout is the layout engine used unless the file or design system says otherwise.
	Layout string
	// AnalysisShape is given to every top-level object while the core compiles the
	// source to find its objects. Diagrams whose keys only compile inside a
	// structural shape (sql_table columns with constraints) need it; the final
	// shape comes from the role's class.
	AnalysisShape string
	// DefaultRole, when set, overrides the default-role rule for objects.
	DefaultRole func(obj *d2graph.Object) string
	// Transform expands the diagram's shorthand, writing the D2 to compile.
	// Nil writes the source as it is.
	Transform func(ast *d2ast.Map, b *body) error
	// Settings is D2 that applies to the diagram as a whole.
	Settings string
	// Skip reports objects that take no role.
	Skip func(obj *d2graph.Object) bool
	// Annotate returns further D2 for an object, given its key and role.
	Annotate func(obj *d2graph.Object, id, role string) string
	// Lanes lays the diagram out as a matrix of lanes and ranks (see
	// laneLayout); its connections need the lane router.
	Lanes bool
}

func topLevel(obj *d2graph.Object) bool { return obj.Parent == obj.Graph.Root }

// sequenceGroup reports a sequence diagram's groups: objects that frame
// messages rather than exchange them. The core compiles the source before it
// is a sequence diagram, so D2's own test does not apply.
func sequenceGroup(obj *d2graph.Object) bool {
	for _, e := range obj.Graph.Edges {
		if e.Src == obj || e.Dst == obj {
			return false
		}
	}
	return obj.ContainsAnyEdge(obj.Graph.Edges)
}

var specs = map[string]*Spec{
	"flowchart": {
		Nodes:      []string{"process", "start", "end", "decision", "io", "subprocess", "store", "note"},
		Containers: []string{"group"},
		Edges:      []string{"flow", "optional"},
		Layout:     "elk",
	},
	"sequence": {
		Nodes:        []string{"participant", "actor"},
		Containers:   []string{"participant", "actor"},
		Edges:        []string{"message", "reply", "async"},
		TopLevelOnly: true,
		Layout:       "elk",
		Settings:     "shape: sequence_diagram\n",
		Skip:         sequenceGroup,
	},
	"state": {
		Nodes:      []string{"state", "start", "end", "choice", "fork", "join", "note"},
		Containers: []string{"composite"},
		Edges:      []string{"transition"},
		Layout:     "elk",
		DefaultRole: func(obj *d2graph.Object) string {
			switch obj.ID {
			case ids.Start:
				return "start"
			case ids.End:
				return "end"
			}
			return ""
		},
		Transform: inPlace(expandPseudoStates),
	},
	"class": {
		Nodes:         []string{"class", "interface", "abstract", "enum"},
		Containers:    []string{"class", "interface", "abstract", "enum"},
		Edges:         []string{"association", "inheritance", "realization", "composition", "aggregation", "dependency"},
		TopLevelOnly:  true,
		AnalysisShape: "class",
		Layout:        "elk",
		Transform:     inPlace(makeUndirected),
		Annotate: func(obj *d2graph.Object, id, role string) string {
			if role == "class" {
				return ""
			}
			return fmt.Sprintf("%s.label: %s\n", id, value("<<"+role+">>\n"+obj.Label.Value))
		},
	},
	"er": {
		Nodes:         []string{"entity"},
		Containers:    []string{"entity"},
		Edges:         []string{"relation", "one-to-one", "one-to-many", "many-to-one", "many-to-many"},
		TopLevelOnly:  true,
		AnalysisShape: "sql_table",
		Layout:        "elk",
		Transform:     inPlace(makeUndirected),
	},
	"block": {
		Nodes:      []string{"block"},
		Containers: []string{"group"},
		Edges:      []string{"link"},
		Layout:     "elk",
	},
	"architecture": {
		Nodes:      []string{"component", "actor", "system", "node", "junction"},
		Containers: []string{"location", "zone", "subsystem"},
		Edges:      []string{"connection", "async"},
		Layout:     "elk",
	},
	"swimlanes": {
		Nodes:      []string{"step", "start", "end", "decision"},
		Containers: []string{"lane"},
		Edges:      []string{"flow", "optional"},
		Layout:     "elk",
		Lanes:      true,
		DefaultRole: func(obj *d2graph.Object) string {
			if topLevel(obj) {
				return "lane"
			}
			return ""
		},
	},
	"quadrant": {
		Nodes:      []string{"item", "quadrant", "axis", "frame"},
		Containers: []string{"quadrant", "frame"},
		Edges:      []string{"link"},
		Layout:     "elk",
		DefaultRole: func(obj *d2graph.Object) string {
			switch {
			case topLevel(obj) && obj.ID == ids.Chart:
				return "frame"
			case topLevel(obj) && (obj.ID == ids.XAxis || obj.ID == ids.YAxis):
				return "axis"
			case topLevel(obj.Parent) && obj.Parent.ID == ids.Chart && isQuadrantKey(obj.ID):
				return "quadrant"
			}
			return "item"
		},
		Transform: wrapQuadrants,
	},
	"treeview": {
		Nodes:      []string{"leaf", "branch", "root"},
		Containers: []string{"leaf", "branch", "root"},
		Edges:      []string{"link"},
		// dagre draws curved links, which is the dendrogram look.
		Layout:    "dagre",
		Transform: flattenTree,
	},
}

func init() {
	for name, s := range specs {
		s.Name = name
	}
}

// Lookup returns the spec for a diagram type.
func Lookup(name string) (*Spec, bool) {
	s, ok := specs[name]
	return s, ok
}

// DiagramNames lists the supported diagram types, sorted.
func DiagramNames() []string {
	names := make([]string, 0, len(specs))
	for n := range specs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
