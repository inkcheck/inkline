package post

import (
	"encoding/base64"
	"strings"
	"testing"
)

const svg = `<style>.fill-N7{fill:#fafafa;}</style>` +
	`<g class="YQ== component data"><g class="shape" ><rect x="10" y="20" width="240" height="48" rx="8" stroke="#0f62fe" fill="#ffffff" /></g>` +
	`<text x="40" y="48" class="text-bold">A</text></g>` +
	`<g class="Yg== lane"><g class="shape" ><rect x="0" y="0" width="100" height="300" stroke="#4d5358" fill="#f2f4f8" /></g><text x="30" y="19">Lane</text></g>` +
	`<g class="KGEgLT4gYilbMF0= flow"><path d="M 0 0 L 0 100" stroke="#161616" fill="none" class="connection" /><text x="50" y="62" style="font-size:12px">Yes</text></g>` +
	`<mask id="m"><rect x="40" y="50" width="20" height="16" fill="black"></rect></mask>`

func TestDecorators(t *testing.T) {
	out := Apply([]byte(svg), "architecture", "light", []Decorator{
		{Type: "block", Size: 48, Classes: []string{"component"}},
		{Type: "bar", Size: 4, Classes: []string{"lane"}},
		{Type: "pill", Classes: []string{"flow"}},
		{Type: "strike", Classes: []string{"lane"}, Diagrams: []string{"state"}},
	})
	s := string(out)
	for _, want := range []string{
		// Colour block in the node's stroke colour, clipped to its rounded outline;
		// the label moves past it.
		`<clipPath id="vf-clip-1"><rect x="10" y="20" width="240" height="48" rx="8" /></clipPath><rect x="10" y="20" width="48" height="48" fill="#0f62fe" clip-path="url(#vf-clip-1)" /></g><text x="98" y="48"`,
		// Side bar; the label moves past it.
		`<rect x="0" y="0" width="4" height="300" fill="#4d5358" clip-path="url(#vf-clip-2)" /></g><text x="40" y="19"`,
		// Pill around the masked label box, on the canvas colour.
		`<rect x="32" y="48" width="36" height="20" rx="10" fill="#fafafa" stroke="#161616" stroke-width="1" /><text x="50" y="62"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s\nin %s", want, s)
		}
	}
	if strings.Contains(s, "<line") {
		t.Errorf("strike is scoped to state diagrams and must not apply")
	}
}

func TestIconDecorators(t *testing.T) {
	const icon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><path d="M1 1"/></svg>`
	const in = `<style>.fill-N7{fill:#121619;}</style>` +
		// A sequence actor: centred label, no icon of its own.
		`<g class="dXNlcg== actor"><g class="shape" ><rect x="0" y="0" width="192" height="48" stroke="#ffffff" fill="#21272a" /></g>` +
		`<text x="96" y="28" class="text-bold">User</text></g>` +
		// An architecture actor: a circle.
		`<g class="YQ== actor"><g class="shape" ><ellipse rx="24" ry="24" cx="100" cy="100" stroke="#000" fill="#000" /></g></g>`
	out := Apply([]byte(in), "sequence", "light", []Decorator{
		{Type: "block", Size: 48, Classes: []string{"actor"}, Diagrams: []string{"sequence"}, IconSVG: icon},
	})
	s := string(out)
	for _, want := range []string{
		// The icon sits centred in the block, in the canvas colour.
		`<svg x="12" y="12" width="24" height="24" viewBox="0 0 32 32" fill="#121619"><path d="M1 1"/></svg>`,
		// A centred label re-centres in the space right of the block.
		`<text x="120" y="28"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("block: missing %s\nin %s", want, s)
		}
	}

	out = Apply([]byte(in), "architecture", "light", []Decorator{
		{Type: "icon", Classes: []string{"actor"}, IconSVG: icon},
	})
	if want := `<svg x="88" y="88" width="24" height="24" viewBox="0 0 32 32" fill="#121619">`; !strings.Contains(string(out), want) {
		t.Errorf("icon: missing %s in circle\n%s", want, out)
	}
	// An icon is not clipped, so the rectangle it sits in gets no clip path.
	if strings.Contains(string(out), "<clipPath") {
		t.Errorf("icon: unused clip path\n%s", out)
	}
}

func TestLegend(t *testing.T) {
	const in = `<style>.d2-1 .fill-N7{fill:#121619;}.appendix text.text{fill:#ffffff}</style>` +
		// D2's legend frame, drawn in fixed light colours.
		`<rect x="1" y="2" width="3" height="4" rx="4" stroke="#DEE1EB" fill="#ffffff" style="stroke-width: 1px" />` +
		`<text class="text-bold" x="5" y="6">Legend</text>` +
		// A legend swatch: the domain class set only the stroke.
		`<g class="cw== swatch security"><g class="shape" ><rect x="0" y="0" width="120" height="120" stroke="#fa4d56" fill="#fa4d56" class=" fill-B6" /></g></g>` +
		`<text class="text" x="7" y="8">Security</text>`

	plain := Apply([]byte(in), "architecture", "light", nil)
	for _, want := range []string{
		`stroke="#DEE1EB" fill="#121619"`,                     // frame on the canvas colour
		`<text fill="#ffffff" class="text-bold" x="5" y="6">`, // legend text in the theme's text colour
		`class=" fill-B6"`,                                    // no fill decorator: swatch untouched
	} {
		if !strings.Contains(string(plain), want) {
			t.Errorf("themed legend: missing %s\n%s", want, plain)
		}
	}

	carbon := Apply([]byte(in), "architecture", "light", []Decorator{
		{Type: "fill", Classes: []string{"swatch"}},
		{Type: "legend"},
	})
	for _, want := range []string{
		`rx="0" stroke="none" fill="#121619" style="stroke-width: 0"`, // flat frame
		`stroke="#fa4d56" fill="#fa4d56" class="" />`,                 // swatch filled, fill class gone
	} {
		if !strings.Contains(string(carbon), want) {
			t.Errorf("flat legend: missing %s\n%s", want, carbon)
		}
	}
}

func TestAvatar(t *testing.T) {
	const icon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path d="M1 1"/></svg>`
	const in = `<style>.d2-1 .fill-N7{fill:#161d26;}.appendix text.text{fill:#c6c6cd}</style>` +
		`<g class="dXNlcg== actor"><g class="shape" ><rect x="0" y="0" width="160" height="48" rx="8" stroke="#8c8c94" fill="#161d26" /></g>` +
		`<text x="80" y="28" class="text-bold">User</text></g>`
	out := Apply([]byte(in), "sequence", "light", []Decorator{
		{Type: "avatar", Size: 28, Classes: []string{"actor"}, IconSVG: icon},
	})
	s := string(out)
	for _, want := range []string{
		`<circle cx="24" cy="24" r="14" fill="#c6c6cd" />`,                                      // round badge in the text colour, inset 10px
		`<svg x="15.6" y="15.6" width="16.8" height="16.8" viewBox="0 0 16 16" fill="#161d26">`, // icon on the canvas colour
		`<text x="99" y="28"`, // centred label re-centred right of the avatar
	} {
		if !strings.Contains(s, want) {
			t.Errorf("avatar: missing %s\n%s", want, s)
		}
	}
}

func TestClearLabels(t *testing.T) {
	// An actor circle with its label below, and a connection D2 ran straight
	// down from the circle's edge through the label.
	const in = `<g class="dXNlcg== actor"><g class="shape" ><ellipse rx="24" ry="24" cx="100" cy="36" stroke="#000" fill="#000" /></g>` +
		`<text x="100" y="79" class="text-bold" style="font-size:14px">Web client</text></g>` +
		`<g class="KHUgLT4gZylbMF0= connection"><path d="M 100 61.5 L 100 254" stroke="#161616" fill="none" class="connection" /></g>`
	out := Apply([]byte(in), "architecture", "light", nil)
	// The label's foot is 79 + 14*0.3 = 83.2; the connection starts 2px below it.
	if want := `d="M 100 85.2 L 100 254"`; !strings.Contains(string(out), want) {
		t.Errorf("missing %s\n%s", want, out)
	}
}

func TestOutlineIcon(t *testing.T) {
	// Cloudscape icons are strokes; parts marked "filled" are filled.
	d := Decorator{IconSVG: `<svg viewBox="0 0 16 16"><path d="M1 1Z"/><path d="M2 2" class="filled"/><path d="M3 3" class="stroke-linejoin-round"/></svg>`, IconStroke: 2}
	got := d.icon(0, 0, 32, 32, 16, "#ffffff")
	for _, want := range []string{
		`fill="none" stroke="#ffffff" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"`,
		`<path d="M2 2" fill="#ffffff"/>`,
		`<path d="M3 3"/>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %s", want, got)
		}
	}
}

func TestChevronsAndOutline(t *testing.T) {
	const in = `<g class="YQ== component data"><g class="shape" ><rect x="0" y="0" width="200" height="48" rx="16" stroke="#c925d1" fill="#ffffff" /></g><text x="100" y="28">A</text></g>` +
		`<marker id="m1"> <polygon points="0,0 10,6 0,12" fill="#424650" class="connection" stroke-width="1" /> </marker>` +
		`<marker id="m2"> <polygon points="0.5,0.5 12.5,7.5 0.5,14.5" stroke="#161616" fill="#ffffff" class="connection fill-N7" stroke-width="1" /> </marker>`
	out := Apply([]byte(in), "architecture", "dark", []Decorator{
		{Type: "block", Size: 48, Classes: []string{"component"}, Outline: Colour{"light": "#c6c6cd", "dark": "#424650"}},
		{Type: "chevron"},
	})
	s := string(out)
	for _, want := range []string{
		`fill="#c925d1" clip-path`, // the block keeps the domain colour
		`rx="16" stroke="#424650"`, // the card border goes neutral, per theme
		`<polyline points="1.5,1.5 8.5,6 1.5,10.5" fill="none" stroke="#424650"`, // filled head -> chevron
		`points="0.5,0.5 12.5,7.5 0.5,14.5" stroke="#161616"`,                    // hollow UML triangle kept
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s\n%s", want, s)
		}
	}
}

func TestGlyph(t *testing.T) {
	uri := func(svg string) string {
		return `<image href="data:image/svg+xml;base64,` + base64.StdEncoding.EncodeToString([]byte(svg)) + `" x="0" y="0" width="24" height="24" />`
	}
	mono := `<svg viewBox="0 0 24 24" fill="none"><path d="M1 1" fill="#212121"/><path d="M2 2" stroke="#212121"/></svg>`
	full := `<svg viewBox="0 0 48 48"><rect fill="#8C4FFF"/><path fill="#FFFFFF" d="M1 1"/></svg>`
	if glyph(uri(mono)) == "" {
		t.Fatal("single-colour icon not recognised")
	}
	if glyph(uri(full)) != "" {
		t.Fatal("full-colour icon treated as a glyph")
	}
	out := inlineGlyph(mono, 0, 0, 16, "#ffffff")
	if strings.Contains(out, "#212121") || !strings.Contains(out, `fill="none"`) {
		t.Fatalf("glyph not recoloured: %s", out)
	}
}

// D2 gives the group of an element with opacity a style attribute.
func TestDecoratorsApplyToGroupsWithStyle(t *testing.T) {
	const in = `<g class="YQ== process removed" style='opacity:0.4'><g class="shape" ><rect x="0" y="0" width="100" height="48" stroke="#161616" fill="#ffffff" /></g>` +
		`<text x="50" y="28">A</text></g>`
	out := string(Apply([]byte(in), "flowchart", "light", []Decorator{{Type: "strike", Classes: []string{"removed"}}}))
	if !strings.Contains(out, `<line x1="0" y1="0" x2="100" y2="48"`) {
		t.Errorf("no strike in %s", out)
	}
}

func TestMultiLineLabelMovesWithItsLines(t *testing.T) {
	const in = `<g class="YQ== state"><g class="shape" ><rect x="0" y="0" width="120" height="60" stroke="#000" fill="#fff" /></g>` +
		`<text x="20" y="25" class="text-bold"><tspan x="20" dy="0">Order</tspan><tspan x="20" dy="16.5">Service</tspan></text></g>`
	out := string(Apply([]byte(in), "state", "light", []Decorator{{Type: "bar", Size: 4, Classes: []string{"state"}}}))
	if want := `<text x="30" y="25" class="text-bold"><tspan x="30" dy="0">Order</tspan><tspan x="30" dy="16.5">Service</tspan>`; !strings.Contains(out, want) {
		t.Errorf("missing %s\nin %s", want, out)
	}
}

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		d    Decorator
		want string
	}{
		{Decorator{Type: "bar", Side: "top"}, ""},
		{Decorator{Type: "sparkle"}, `unknown decorator type "sparkle"`},
		{Decorator{Type: "bar", Side: "right"}, `unknown side "right"`},
		{Decorator{Type: "avatar", Shape: "hexagon"}, `unknown shape "hexagon"`},
	} {
		err := tc.d.Validate()
		if (tc.want == "") != (err == nil) || err != nil && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%+v: got %v, want %q", tc.d, err, tc.want)
		}
	}
}
