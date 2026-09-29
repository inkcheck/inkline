package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inkcheck/inkline/internal/designsystem"
)

// Every example renders in every kit and theme: a kit that fails to compile
// for some diagram type shows up here.
func TestExamplesRenderInEveryKit(t *testing.T) {
	examples, err := filepath.Glob("../../examples/*.d2")
	if err != nil || len(examples) == 0 {
		t.Fatalf("no examples: %v", err)
	}
	out := t.TempDir()
	for _, kit := range designsystem.Names() {
		for _, theme := range []string{"light", "dark"} {
			for _, ex := range examples {
				name := strings.TrimSuffix(filepath.Base(ex), ".d2")
				t.Run(kit+"/"+theme+"/"+name, func(t *testing.T) {
					svg := filepath.Join(out, kit+"-"+theme+"-"+name+".svg")
					if err := run([]string{"render", ex, "--design-system", kit, "--theme", theme, "-o", svg}); err != nil {
						t.Fatal(err)
					}
					if b, err := os.ReadFile(svg); err != nil || !strings.Contains(string(b), "</svg>") {
						t.Fatalf("no SVG written: %v", err)
					}
				})
			}
		}
	}
}

func TestCommandLine(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) string {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	in := write("a.d2", "vars: {inkline: {diagram: flowchart}}\na -> b\n")
	write("shared.d2", "x: Shared\n")
	nested := write("sub/c.d2", "vars: {inkline: {diagram: flowchart}}\n...@../shared\nx -> y\n")

	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"render", "-h"}, ""},
		{[]string{"render", nested}, ""},
		{[]string{"render", "--theme", "dark", "--", in}, ""},
		{[]string{"render", "--", in, "--theme"}, "expected one input file"},
		{[]string{"frobnicate"}, `unknown command "frobnicate"`},
		{[]string{"render", in, "--theme", "sepia"}, `unknown theme "sepia"`},
		{[]string{"render", in, "--design-system", "nope"}, `unknown design system "nope"`},
		{[]string{"expand", in, "-o", "out.svg"}, "flag provided but not defined"},
		{[]string{"render"}, "expected one input file"},
	} {
		err := run(tc.args)
		if (tc.want == "") != (err == nil) || err != nil && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: got %v, want %q", tc.args, err, tc.want)
		}
	}
}
