package core

import (
	"fmt"
	"strings"

	"github.com/d2lang/d2/d2ast"
)

// Header is the inkline block of a file's vars:
//
//	vars: { inkline: { diagram: flowchart; design_system: carbon; theme: light } }
type Header struct {
	Diagram      string
	DesignSystem string
	Theme        string
	// Orientation applies to swimlanes: vertical (default; lanes are columns)
	// or horizontal (lanes are rows).
	Orientation string
}

var headerKeys = []string{"diagram", "design_system", "theme", "orientation"}

// ReadHeader reads and validates the inkline header from a parsed file.
func ReadHeader(ast *d2ast.Map) (Header, error) {
	h := Header{DesignSystem: "plain", Theme: "light", Orientation: "vertical"}
	block := findMap(ast, "vars", "inkline")
	if block == nil {
		return h, fmt.Errorf("missing inkline header: add vars: { inkline: { diagram: <type> } }")
	}
	for _, n := range block.Nodes {
		k := n.MapKey
		if k == nil || k.Key == nil || len(k.Key.Path) != 1 {
			continue
		}
		v := scalar(k)
		switch name := k.Key.Path[0].ScalarString(); name {
		case "diagram":
			h.Diagram = v
		case "design_system":
			h.DesignSystem = v
		case "theme":
			h.Theme = v
		case "orientation":
			h.Orientation = v
		default:
			return h, fmt.Errorf("unknown inkline header key %q%s; keys: %s", name, suggest(name), strings.Join(headerKeys, ", "))
		}
	}
	return h, h.validate()
}

// Override replaces the design system and theme, where given, and validates
// the result.
func (h *Header) Override(designSystem, theme string) error {
	if designSystem != "" {
		h.DesignSystem = designSystem
	}
	if theme != "" {
		h.Theme = theme
	}
	return h.validate()
}

func (h Header) validate() error {
	if h.Diagram == "" {
		return fmt.Errorf("inkline header has no diagram; one of: %s", strings.Join(DiagramNames(), ", "))
	}
	if _, ok := Lookup(h.Diagram); !ok {
		return fmt.Errorf("unknown diagram %q; one of: %s", h.Diagram, strings.Join(DiagramNames(), ", "))
	}
	if h.Theme != "light" && h.Theme != "dark" {
		return fmt.Errorf("unknown theme %q; one of: light, dark", h.Theme)
	}
	if h.Orientation != "horizontal" && h.Orientation != "vertical" {
		return fmt.Errorf("unknown orientation %q; one of: horizontal, vertical", h.Orientation)
	}
	return nil
}

// findMap returns the map at a path of single-segment keys, or nil. A key may
// be written more than once (two vars blocks); each is searched. Keys written
// as dotted paths (vars.inkline.diagram) are not supported here.
func findMap(m *d2ast.Map, path ...string) *d2ast.Map {
	if m == nil {
		return nil
	}
	for _, n := range m.Nodes {
		k := n.MapKey
		if k == nil || k.Key == nil || len(k.Edges) > 0 || len(k.Key.Path) != 1 ||
			k.Key.Path[0].ScalarString() != path[0] || k.Value.Map == nil {
			continue
		}
		if len(path) == 1 {
			return k.Value.Map
		}
		if found := findMap(k.Value.Map, path[1:]...); found != nil {
			return found
		}
	}
	return nil
}

// scalar returns a key's scalar value, whether written as `k: v` or `k: v {…}`.
func scalar(k *d2ast.Key) string {
	if v, ok := k.Value.Unbox().(d2ast.Scalar); ok && v != nil {
		return v.ScalarString()
	}
	if k.Primary.Unbox() != nil {
		return k.Primary.ScalarString()
	}
	return ""
}

// suggest offers the header key a misspelt one most likely meant, such as
// design_system for design-system or designSystem.
func suggest(name string) string {
	norm := strings.NewReplacer("-", "", "_", "").Replace(strings.ToLower(name))
	for _, k := range headerKeys {
		if strings.ReplaceAll(k, "_", "") == norm {
			return fmt.Sprintf(" (did you mean %s?)", k)
		}
	}
	return ""
}
