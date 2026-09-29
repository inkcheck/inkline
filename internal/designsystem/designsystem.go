// Package designsystem loads a design system: D2 style files (layer 2) plus the
// render settings and SVG decorators it needs (layer 3). Design systems form a
// chain through "extends"; every chain ends at "plain".
package designsystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/d2lang/d2/d2format"
	"github.com/d2lang/d2/d2parser"
	"github.com/d2lang/d2/d2renderers/d2fonts"

	"github.com/inkcheck/inkline/internal/post"
	"github.com/inkcheck/inkline/kits"
)

// Manifest is a design system's manifest.json.
type Manifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Extends names the parent design system. Defaults to "plain".
	Extends string `json:"extends"`
	// Tokens maps a theme (light, dark) to a D2 file defining colour vars and
	// d2-config theme overrides.
	Tokens map[string]string `json:"tokens"`
	// Styles are D2 files loaded for every diagram type. diagrams/<type>.d2 is
	// loaded after them when present.
	Styles []string `json:"styles"`
	// Layout overrides the core's layout engine per diagram type.
	Layout map[string]string `json:"layout"`
	Pad    *int              `json:"pad"`
	// Fonts maps a d2 font flag (regular, italic, bold, semibold) to a .ttf file.
	Fonts      map[string]string `json:"fonts"`
	Decorators []post.Decorator  `json:"decorators"`

	dir string
}

// DesignSystem is a resolved chain of manifests, base first.
type DesignSystem struct {
	Name  string
	chain []*Manifest
}

// Load resolves a bundled design system and its parents.
func Load(name string) (*DesignSystem, error) {
	ds := &DesignSystem{Name: name}
	seen := map[string]bool{}
	for n := name; n != ""; {
		if seen[n] {
			return nil, fmt.Errorf("design system %q extends itself", n)
		}
		seen[n] = true
		m, err := readManifest(n)
		if err != nil {
			return nil, err
		}
		ds.chain = append([]*Manifest{m}, ds.chain...)
		switch {
		case n == "plain":
			n = ""
		case m.Extends == "":
			n = "plain"
		default:
			n = m.Extends
		}
	}
	return ds, nil
}

// Names lists the bundled design systems.
func Names() []string {
	entries, _ := fs.ReadDir(kits.FS, ".")
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

func readManifest(name string) (*Manifest, error) {
	b, err := fs.ReadFile(kits.FS, path.Join(name, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("unknown design system %q; one of: %s", name, strings.Join(Names(), ", "))
	}
	m := &Manifest{dir: name}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, fmt.Errorf("%s/manifest.json: %w", name, err)
	}
	if m.Name != name {
		return nil, fmt.Errorf("%s/manifest.json: name is %q, want %q", name, m.Name, name)
	}
	for _, d := range m.Decorators {
		if err := d.Validate(); err != nil {
			return nil, fmt.Errorf("%s/manifest.json: %w", name, err)
		}
	}
	return m, nil
}

// Prelude returns the D2 that precedes a diagram's source: tokens, styles and
// d2-config, from the base design system to the most specific.
//
// Every file's top-level vars are hoisted above all classes. D2 fails to resolve
// a spread variable that references other variables (...${carbon.node}, whose
// style uses ${color.x}) when it is merged into a class first defined before
// those variables were declared, which is exactly what extending a parent's
// classes does.
func (ds *DesignSystem) Prelude(diagram, theme, layout string) (string, error) {
	var vars, b strings.Builder
	pad := -1
	for _, m := range ds.chain {
		files := []string{}
		if f, ok := m.Tokens[theme]; ok {
			files = append(files, f)
		}
		files = append(files, m.Styles...)
		files = append(files, path.Join("diagrams", diagram+".d2"))
		for _, f := range files {
			src, err := fs.ReadFile(kits.FS, path.Join(m.dir, f))
			if errors.Is(err, fs.ErrNotExist) && strings.HasPrefix(f, "diagrams/") {
				continue
			}
			if err != nil {
				return "", fmt.Errorf("design system %s: %w", m.Name, err)
			}
			name := m.Name + "/" + f
			v, rest, err := splitVars(name, string(src))
			if err != nil {
				return "", err
			}
			if v != "" {
				fmt.Fprintf(&vars, "# --- %s (vars) ---\n%s\n", name, v)
			}
			fmt.Fprintf(&b, "# --- %s ---\n%s\n", name, rest)
		}
		if l, ok := m.Layout[diagram]; ok {
			layout = l
		}
		if m.Pad != nil {
			pad = *m.Pad
		}
	}

	themeID := 0
	if theme == "dark" {
		themeID = 200
	}
	fmt.Fprintf(&b, "# --- render settings ---\nvars: {\n  d2-config: {\n    layout-engine: %s\n    theme-id: %d\n", layout, themeID)
	if pad >= 0 {
		fmt.Fprintf(&b, "    pad: %d\n", pad)
	}
	b.WriteString("  }\n}\n")
	return vars.String() + b.String(), nil
}

// splitVars separates a D2 file's top-level vars blocks from everything else.
func splitVars(name, src string) (vars, rest string, err error) {
	ast, err := d2parser.Parse(name, strings.NewReader(src), nil)
	if err != nil {
		return "", "", fmt.Errorf("design system file %s: %w", name, err)
	}
	var v, r strings.Builder
	for _, n := range ast.Nodes {
		out := &r
		if k := n.MapKey; k != nil && k.Key != nil && len(k.Edges) == 0 &&
			len(k.Key.Path) == 1 && k.Key.Path[0].ScalarString() == "vars" {
			out = &v
		}
		out.WriteString(d2format.Format(n.Unbox()) + "\n")
	}
	return v.String(), r.String(), nil
}

// Decorators returns the SVG decorators of the whole chain, base first, with
// their icon files loaded.
func (ds *DesignSystem) Decorators() ([]post.Decorator, error) {
	var out []post.Decorator
	for _, m := range ds.chain {
		for _, d := range m.Decorators {
			if d.Icon != "" {
				b, err := fs.ReadFile(kits.FS, path.Join(m.dir, d.Icon))
				if err != nil {
					return nil, fmt.Errorf("design system %s: decorator icon: %w", m.Name, err)
				}
				d.IconSVG = string(b)
			}
			out = append(out, d)
		}
	}
	return out, nil
}

// Font returns the chain's font family, or nil when no design system bundles
// fonts. The most specific design system wins per style.
//
// D2 draws node labels in its bold style, so a design system whose labels are
// semibold (IBM Plex Sans) maps "bold" to the semibold file.
func (ds *DesignSystem) Font() (*d2fonts.FontFamily, error) {
	fonts := map[string]*Manifest{}
	for _, m := range ds.chain {
		for kind := range m.Fonts {
			fonts[kind] = m
		}
	}
	if len(fonts) == 0 {
		return nil, nil
	}
	ttf := map[string][]byte{}
	for kind, m := range fonts {
		b, err := fs.ReadFile(kits.FS, path.Join(m.dir, m.Fonts[kind]))
		if err != nil {
			return nil, fmt.Errorf("design system %s: %w", m.Name, err)
		}
		ttf[kind] = b
	}
	return d2fonts.AddFontFamily(ds.Name, ttf["regular"], ttf["italic"], ttf["bold"], ttf["semibold"])
}
