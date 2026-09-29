package core

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/inkcheck/inkline/internal/ids"
)

func expand(t *testing.T, src string) string {
	t.Helper()
	s := Source{Path: "in.d2", FS: fstest.MapFS{}, Text: src}
	f, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	out, err := f.Expand("")
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func roles(out string) string {
	return out[strings.Index(out, "# --- inkline roles ---"):]
}

func mustContain(t *testing.T, out string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in:\n%s", w, out)
		}
	}
}

func TestHeader(t *testing.T) {
	for src, wantErr := range map[string]string{
		"a -> b": "missing inkline header",
		"vars: {inkline: {design_system: carbon}}":                   "has no diagram",
		"vars: {inkline: {diagram: pie}}":                            `unknown diagram "pie"`,
		"vars: {inkline: {diagram: flowchart; theme: sepia}}":        `unknown theme "sepia"`,
		"vars: {inkline: {diagram: flowchart; colour: red}}":         `unknown inkline header key "colour"`,
		"vars: {inkline: {diagram: flowchart; design-system: ibm}}":  `did you mean design_system?`,
		"vars: {inkline: {diagram: swimlanes; orientation: sloped}}": `unknown orientation "sloped"`,
	} {
		_, err := Parse(Source{Path: "in.d2", Text: src})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("%s: got %v, want %q", src, err, wantErr)
		}
	}
	f, err := Parse(Source{Path: "in.d2", Text: "vars: {inkline: {diagram: state}}"})
	if err != nil {
		t.Fatal(err)
	}
	if h := f.Header; h.DesignSystem != "plain" || h.Theme != "light" || h.Orientation != "vertical" {
		t.Errorf("defaults: got %+v", h)
	}
}

func TestRolesKeepAuthorClassesAfterRole(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: flowchart}}
classes: {hot: {style.fill: red}}
a: A {class: hot}
b: B {class: [decision; hot]}
g: G {c}
a -> b: {class: optional}
b -> g.c`)
	mustContain(t, roles(out),
		"a.class: [process; hot]",
		"b.class: [decision; hot]",
		"g.class: [group]",
		"g.c.class: [process]",
		"(a -> b)[0].class: [optional]",
		"(b -> g.c)[0].class: [flow]",
	)
	// The author's class keys are gone from the source: D2 ignores a class
	// list assigned to an object that already has a class.
	src := out[strings.Index(out, "# --- source ---"):strings.Index(out, "# --- inkline roles ---")]
	if strings.Contains(src, "class: hot") || strings.Contains(src, "class: optional") {
		t.Errorf("author classes not stripped:\n%s", src)
	}
	mustContain(t, src, "classes: {hot:")
}

func TestState(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: state}}
"[*]" -> idle
idle -> busy
busy -> "[*]"
job: {
  "[*]" -> run
  run -> "[*]"
}`)
	mustContain(t, out, "vf_start -> idle", "busy -> vf_end", "vf_start -> run")
	mustContain(t, roles(out),
		"vf_start.class: [start]", "vf_end.class: [end]", "idle.class: [state]",
		"job.class: [composite]", "job.vf_start.class: [start]", "job.vf_end.class: [end]")
}

func TestClassStereotypeAndTopLevelOnly(t *testing.T) {
	out := roles(expand(t, `vars: {inkline: {diagram: class}}
Repo: {class: interface; +find(id string): Order}
Order: {+id: string}
Order -> Repo: {class: dependency}`))
	mustContain(t, out, "Repo.class: [interface]", `Repo.label: "<<interface>>\nRepo"`,
		"Order.class: [class]", "(Order <-> Repo)[0].class: [dependency]")
	if strings.Contains(out, "Order.id") {
		t.Errorf("fields must not get roles:\n%s", out)
	}
}

func TestERColumnEdges(t *testing.T) {
	out := roles(expand(t, `vars: {inkline: {diagram: er}}
customer: {id: int {constraint: primary_key}}
order: {customer_id: int {constraint: foreign_key}}
customer.id -> order.customer_id: {class: one-to-many}`))
	mustContain(t, out, "customer.class: [entity]", "(customer.id <-> order.customer_id)[0].class: [one-to-many]")
}

func TestSequence(t *testing.T) {
	out := roles(expand(t, `vars: {inkline: {diagram: sequence}}
alice: {class: actor}
alice -> bob: hi
bob -> alice: ok {class: reply}`))
	mustContain(t, out, "shape: sequence_diagram", "alice.class: [actor]", "bob.class: [participant]",
		"(alice -> bob)[0].class: [message]", "(bob -> alice)[0].class: [reply]")
}

// D2 counts a connection and its mirror image as one key.
func TestMirroredConnectionsShareAnIndex(t *testing.T) {
	out := roles(expand(t, `vars: {inkline: {diagram: class}}
A: {+x: int}
B: {+y: int}
A -> B: {class: inheritance}
B -> A: {class: dependency}`))
	mustContain(t, out, "(A <-> B)[0].class: [inheritance]", "(B <-> A)[1].class: [dependency]")

	out = roles(expand(t, `vars: {inkline: {diagram: flowchart}}
a -> b
b <- a
b -> a
a -> b: {class: optional}`))
	mustContain(t, out, "(a -> b)[0].class: [flow]", "(b <- a)[1].class: [flow]",
		"(b -> a)[0].class: [flow]", "(a -> b)[2].class: [optional]")
}

func TestObjectsWithoutRoleKeepTheirClasses(t *testing.T) {
	out := roles(expand(t, `vars: {inkline: {diagram: sequence}}
classes: {hot: {style.fill: red}}
alice -> bob: hi
alice.t1.class: hot
alice.t1 -> bob.t1: span
retry: {
  class: hot
  alice -> bob: again
}
bob."a note": {class: hot}`))
	mustContain(t, out, "alice.t1.class: [hot]", "retry.class: [hot]", "bob.a note.class: [hot]")
	if strings.Contains(out, "bob.t1.class") {
		t.Errorf("an object with neither role nor class needs no class list:\n%s", out)
	}
}

func TestTreeviewDottedAttributes(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: treeview}}
vars: {quiet: {style.opacity: 0.4}}
root: Root {
  a: A {b: B}
  a.style.fill: red
  c: C {...${quiet}}
}
root.style.bold: true`)
	mustContain(t, out, "root/a: {\n  style.fill: red\n}", "root: {\n  style.bold: true\n}", "...${quiet}")
	mustContain(t, roles(out), "root.class: [root]", "root/a.class: [branch]", "root/c.class: [leaf]")
	if strings.Contains(out, "style.fill\"") || strings.Count(out, "root -- root/a\n") != 1 {
		t.Errorf("an attribute is not a node, and a node is linked once:\n%s", out)
	}

	f, err := Parse(Source{Path: "in.d2", FS: fstest.MapFS{}, Text: "vars: {inkline: {diagram: treeview}}\nroot: {...@sub}\n"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Expand(""); err == nil || !strings.Contains(err.Error(), "imports are not allowed inside") {
		t.Errorf("import inside a node: got %v", err)
	}
}

func TestTreeview(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: treeview}}
root: Root {
  a: A {
    style.italic: true
    a1: A1
  }
  b: B
}`)
	mustContain(t, out, "direction: right", "root -- root/a", "root/a -- root/a/a1", "style.italic: true")
	mustContain(t, roles(out), "root.class: [root]", "root/a.class: [branch]", "root/a/a1.class: [leaf]", "root/b.class: [leaf]")
}

func TestQuadrant(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: quadrant}}
title: T
x: Reach
q1: Expand {a: A}
q3: Drop`)
	mustContain(t, out, "vf_chart: T {", `q2: ""`, "q4: \"\"", "vf_x: Reach {near: bottom-center}")
	// Grid cells fill in declaration order: q2 q1 / q3 q4.
	if i, j := strings.Index(out, "q2:"), strings.Index(out, "q1:"); i > j {
		t.Errorf("q2 must precede q1")
	}
	mustContain(t, roles(out), "vf_chart.class: [frame]", "vf_chart.q1.class: [quadrant]", "vf_chart.q1.a.class: [item]", "vf_x.class: [axis]")
}

func TestSwimlaneRanks(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: swimlanes}}
a: A {s; t}
b: B {u}
a.s -> b.u -> a.t
a.t -> a.s`)
	layout := out[strings.Index(out, "# --- inkline lane layout ---"):strings.Index(out, "# --- source ---")]
	// s (rank 0) -> u (rank 1) -> t (rank 2); the t -> s loop is ignored.
	mustContain(t, layout, "grid-columns: 2",
		"a: {\n  grid-rows: 3\n  s: {}\n  vf_gap_1: \"\" {class: gap; }\n  t: {}\n}",
		"b: {\n  grid-rows: 3\n  vf_gap_2: \"\" {class: gap; }\n  u: {}\n  vf_gap_3: \"\" {class: gap; }\n}")
	mustContain(t, roles(out), "a.class: [lane]", "a.s.class: [step]")
	if strings.Contains(roles(out), ids.GapPrefix) {
		t.Errorf("gap cells must not get roles")
	}
}

func TestHeaderInSecondVarsBlock(t *testing.T) {
	f, err := Parse(Source{Path: "in.d2", Text: "vars: {d2-config: {pad: 10}}\nvars: {inkline: {diagram: state}}"})
	if err != nil || f.Header.Diagram != "state" {
		t.Errorf("got %+v, %v", f, err)
	}
}

func TestOverrideValidates(t *testing.T) {
	f, err := Parse(Source{Path: "in.d2", Text: "vars: {inkline: {diagram: state}}"})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Header.Override("carbon", "dark"); err != nil || f.Header.DesignSystem != "carbon" || f.Header.Theme != "dark" {
		t.Errorf("got %+v, %v", f.Header, err)
	}
	if err := f.Header.Override("", "sepia"); err == nil || !strings.Contains(err.Error(), `unknown theme "sepia"`) {
		t.Errorf("got %v", err)
	}
}

func TestSequenceGroupsTakeNoRole(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: sequence}}
alice -> bob: hi
retry: {
  alice -> bob: again
}
`)
	mustContain(t, roles(out), "alice.class: [participant]", "retry.(alice -> bob)[0].class: [message]")
	if strings.Contains(roles(out), "retry.class") {
		t.Errorf("a group is not a participant:\n%s", roles(out))
	}
}

func TestImportedObjectsTakeTheirRole(t *testing.T) {
	s := Source{Path: "in.d2", FS: fstest.MapFS{"lib.d2": {Data: []byte("d: D {class: hot}\nd -> x: {class: hot}\n")}}, Text: `vars: {inkline: {diagram: flowchart}}
classes: {hot: {style.fill: red}}
...@lib
e: E {class: hot}
`}
	f, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	out, err := f.Expand("")
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, roles(out), "d.class: null\nd.class: [process; hot]", "e.class: [process; hot]",
		"(d -> x)[0]: {class: null}\n(d -> x)[0].class: [flow; hot]")
	if strings.Contains(roles(out), "e.class: null") {
		t.Errorf("only imported objects need their class reset:\n%s", roles(out))
	}
}

func TestQuadrantKeysReachIntoTheChart(t *testing.T) {
	out := expand(t, `vars: {inkline: {diagram: quadrant}}
q3.c -> q1.a: moves
q1: Do {a: A}
q3: Drop {c: C}
q1.a.style.bold: true
q1: {d: D}`)
	mustContain(t, out, ids.Chart+".q3.c -> "+ids.Chart+".q1.a: moves", ids.Chart+".q1.a.style.bold: true")
	mustContain(t, roles(out), ids.Chart+".q1.d.class: [item]", ids.Chart+".(q3.c -> q1.a)[0].class: [link]")
	if strings.Contains(roles(out), "\nq1") || strings.Contains(roles(out), "\nq3") {
		t.Errorf("a quadrant outside the chart:\n%s", roles(out))
	}
	// The connection is written after the chart, so the grid keeps its order.
	if strings.Index(out, "moves") < strings.Index(out, "grid-rows") {
		t.Errorf("connection written before the chart:\n%s", out)
	}
}

func TestQuadrantFrameKeepsClearOfAuthorKeys(t *testing.T) {
	out := expand(t, "vars: {inkline: {diagram: quadrant}}\nchart: Mine\nq1: {a}\n")
	mustContain(t, roles(out), "chart.class: [item]", "vf_chart.class: [frame]")
}

// Compile errors name the line of the author's file, whatever the prelude's
// length and however the source was reformatted or rewritten.
func TestCompileErrorsNameTheSourceLine(t *testing.T) {
	const prelude = "classes: {\n  x: {}\n}\n"
	for name, tc := range map[string]struct{ src, want string }{
		"plain": {`vars: {inkline: {diagram: flowchart}}


a -> b

g: {
  c: {shape: nonsense}
}
`, "in.d2:7: "},
		"rewritten in place": {`vars: {inkline: {diagram: state}}

"[*]" -> a
a: {
  b.shape: nonsense
}
`, "in.d2:5: "},
		"treeview": {`vars: {inkline: {diagram: treeview}}
root: {
  kid: {
    shape: nonsense
  }
}
`, "in.d2:4: "},
		"quadrant": {`vars: {inkline: {diagram: quadrant}}
title: T
q1: {
  a
  b: {shape: nonsense}
}
`, "in.d2:5: "},
	} {
		f, err := Parse(Source{Path: "in.d2", FS: fstest.MapFS{}, Text: tc.src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Expand(prelude)
		if err == nil || !strings.HasPrefix(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want prefix %q", name, err, tc.want)
		}
	}
}
