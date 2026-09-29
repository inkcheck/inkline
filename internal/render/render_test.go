package render

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSVG(t *testing.T) {
	doc := `vars: {d2-config: {layout-engine: elk; theme-id: 0; pad: 48}}
classes: {box: {style.fill: "#abcdef"}}
a: Hello {class: box}
b: World
a -> b: hi`
	svg, err := SVG(context.Background(), doc, Options{Path: "in.d2", FS: fstest.MapFS{}})
	if err != nil {
		t.Fatal(err)
	}
	s := string(svg)
	for _, want := range []string{"<svg", "Hello", "World", `fill="#abcdef"`, "</svg>"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in rendered SVG", want)
		}
	}
}

func TestUnknownLayoutEngine(t *testing.T) {
	doc := "vars: {d2-config: {layout-engine: spring}}\na -> b"
	_, err := SVG(context.Background(), doc, Options{Path: "in.d2", FS: fstest.MapFS{}})
	if err == nil || !strings.Contains(err.Error(), `unknown layout engine "spring"`) {
		t.Errorf("got %v, want unknown layout engine", err)
	}
}
