package designsystem

import (
	"strings"
	"testing"
)

func TestLoadResolvesTheChainBaseFirst(t *testing.T) {
	ds, err := Load("carbon")
	if err != nil {
		t.Fatal(err)
	}
	var chain []string
	for _, m := range ds.chain {
		chain = append(chain, m.Name)
	}
	if got := strings.Join(chain, " "); got != "plain carbon" {
		t.Errorf("chain: got %q", got)
	}
	if _, err := Load("nope"); err == nil || !strings.Contains(err.Error(), `unknown design system "nope"`) {
		t.Errorf("unknown design system: got %v", err)
	}
}

func TestEveryBundledDesignSystemLoads(t *testing.T) {
	names := Names()
	if len(names) == 0 {
		t.Fatal("no design systems")
	}
	for _, name := range names {
		ds, err := Load(name)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if _, err := ds.Decorators(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if _, err := ds.Font(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestPreludeHoistsVarsAboveStyles(t *testing.T) {
	ds, err := Load("carbon")
	if err != nil {
		t.Fatal(err)
	}
	out, err := ds.Prelude("flowchart", "dark", "elk")
	if err != nil {
		t.Fatal(err)
	}
	lastVars := strings.LastIndex(out, "(vars) ---")
	firstStyles := strings.Index(out, "# --- plain/base.d2 ---")
	if lastVars < 0 || firstStyles < 0 || lastVars > firstStyles {
		t.Errorf("vars at %d must precede styles at %d", lastVars, firstStyles)
	}
	for _, want := range []string{
		"# --- carbon/tokens-dark.d2",
		"# --- carbon/diagrams/flowchart.d2 ---",
		"layout-engine: elk",
		"theme-id: 200",
		"pad: 48",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, "tokens-light.d2") {
		t.Errorf("dark prelude loads light tokens")
	}
}

func TestSplitVars(t *testing.T) {
	vars, rest, err := splitVars("x.d2", "vars: {a: 1}\nclasses: {c: {shape: circle}}\nvars: {b: 2}\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(vars, "a: 1") || !strings.Contains(vars, "b: 2") || strings.Contains(vars, "classes") {
		t.Errorf("vars: got %q", vars)
	}
	if !strings.Contains(rest, "classes") || strings.Contains(rest, "vars") {
		t.Errorf("rest: got %q", rest)
	}
}

func TestFont(t *testing.T) {
	plain, err := Load("plain")
	if err != nil {
		t.Fatal(err)
	}
	if f, err := plain.Font(); err != nil || f != nil {
		t.Errorf("plain bundles no fonts: got %v, %v", f, err)
	}
	carbon, err := Load("carbon")
	if err != nil {
		t.Fatal(err)
	}
	if f, err := carbon.Font(); err != nil || f == nil {
		t.Errorf("carbon bundles fonts: got %v, %v", f, err)
	}
}

func TestDecoratorsLoadTheirIcons(t *testing.T) {
	ds, err := Load("carbon")
	if err != nil {
		t.Fatal(err)
	}
	decorators, err := ds.Decorators()
	if err != nil {
		t.Fatal(err)
	}
	icons := 0
	for _, d := range decorators {
		if d.Icon != "" {
			icons++
			if !strings.Contains(d.IconSVG, "<svg") {
				t.Errorf("%s: icon %s not loaded", d.Type, d.Icon)
			}
		}
	}
	if icons == 0 {
		t.Errorf("carbon has icon decorators")
	}
}
