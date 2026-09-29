package route

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2layouts/d2elklayout"
	"github.com/d2lang/d2/d2lib"
	"github.com/d2lang/d2/d2renderers/d2svg"
	"github.com/d2lang/d2/d2target"
	d2log "github.com/d2lang/d2/lib/log"
	"github.com/d2lang/d2/lib/textmeasure"
)

// lanes is a swimlane matrix as the core lays it out: lanes are grid cells,
// each lane a grid with one cell per rank, padded with gap cells. Ranks line
// up across lanes because every cell shares the flow-axis size: the height in
// vertical lanes, the width (set by the core) in horizontal ones.
const lanes = `
classes: {gap: {label: ""; height: 48; width: %d; style.opacity: 0}}
%s: 3
customer: {
  %s: 5
  submit: {width: 160; height: 48}
  vf_gap_1: "" {class: gap}
  vf_gap_2: "" {class: gap}
  vf_gap_3: "" {class: gap}
  pay: {width: 160; height: 48}
}
sales: {
  %s: 5
  vf_gap_4: "" {class: gap}
  review: {width: 160; height: 48}
  ok: {shape: diamond; width: 160; height: 48}
  vf_gap_5: "" {class: gap}
  vf_gap_6: "" {class: gap}
}
warehouse: {
  %s: 5
  vf_gap_7: "" {class: gap}
  vf_gap_8: "" {class: gap}
  vf_gap_9: "" {class: gap}
  pick: {width: 160; height: 48}
  ship: {width: 160; height: 48}
}
customer.submit -> sales.review -> sales.ok
sales.ok -> warehouse.pick: Yes
sales.ok -> customer.submit: No
warehouse.pick -> customer.pay
customer.pay -> warehouse.ship
`

func layout(t *testing.T, horizontal bool, outer, inner string, gapWidth int) *d2target.Diagram {
	t.Helper()
	ctx := d2log.With(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	doc := fmt.Sprintf(lanes, gapWidth, outer, inner, inner, inner)
	diagram, _, err := d2lib.Compile(ctx, doc, &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: func(string) (d2graph.LayoutGraph, error) { return d2elklayout.DefaultLayout, nil },
		RouterResolver: func(string) (d2graph.RouteEdges, error) { return Lanes(horizontal), nil },
	}, &d2svg.RenderOpts{})
	if err != nil {
		t.Fatal(err)
	}
	return diagram
}

func TestLanesRouteAtRightAnglesAroundSteps(t *testing.T) {
	for _, tc := range []struct {
		name, outer, inner string
		gapWidth           int
	}{
		{"vertical", "grid-columns", "grid-rows", 8},
		{"horizontal", "grid-rows", "grid-columns", 160},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := layout(t, tc.name == "horizontal", tc.outer, tc.inner, tc.gapWidth)
			var steps []d2target.Shape
			for _, s := range d.Shapes {
				// Steps are the lanes' visible children.
				if strings.Count(s.ID, ".") == 1 && !strings.Contains(s.ID, "vf_gap_") {
					steps = append(steps, s)
				}
			}
			if len(steps) != 6 {
				t.Fatalf("want 6 steps, got %d", len(steps))
			}
			for _, c := range d.Connections {
				for i := 1; i < len(c.Route); i++ {
					a, b := c.Route[i-1], c.Route[i]
					if a.X != b.X && a.Y != b.Y {
						t.Errorf("%s: segment %d is diagonal: %v -> %v", c.ID, i, *a, *b)
					}
					for _, s := range steps {
						if s.ID == c.Src || s.ID == c.Dst {
							continue
						}
						if crosses(a.X, a.Y, b.X, b.Y, s) {
							t.Errorf("%s: segment %d crosses step %s", c.ID, i, s.ID)
						}
					}
				}
			}
		})
	}
}

// crosses reports whether an axis-aligned segment passes through a shape's
// interior.
func crosses(x0, y0, x1, y1 float64, s d2target.Shape) bool {
	left, top := float64(s.Pos.X), float64(s.Pos.Y)
	right, bottom := left+float64(s.Width), top+float64(s.Height)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return x1 > left && x0 < right && y1 > top && y0 < bottom
}
