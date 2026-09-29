// Package route draws connections D2 leaves to a pluggable router: those that
// cross between nested layouts, such as swimlane steps in different lanes.
package route

import (
	"context"
	"math"
	"strings"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/lib/geo"
	"github.com/d2lang/d2/lib/label"

	"github.com/inkcheck/inkline/internal/ids"
)

// Lanes returns a router that draws connections between swimlanes at right
// angles. horizontal says the lanes are rows, with the flow running to the
// right; otherwise they are columns and the flow runs downwards.
//
// Swimlanes are a matrix: lanes along one axis, ranks of the flow along the
// other, with every rank lined up across lanes. The gaps between ranks
// (corridors) and between lanes (gutters) are therefore empty, so a route that
// runs only through them never crosses a step:
//
//   - forward to the next free rank: leave the step's far end, turn in the
//     corridor just before the target, and enter the target's near end;
//   - anything else (loops, skipped ranks): leave the step's side, run along
//     the gutter to that corridor, then turn into the target.
//
// D2 draws connections inside one lane itself; lanes stretch their steps to a
// common width, so those are straight.
func Lanes(horizontal bool) d2graph.RouteEdges {
	return func(ctx context.Context, g *d2graph.Graph, edges []*d2graph.Edge) error {
		return routeLanes(frame{vertical: !horizontal}, g, edges)
	}
}

func routeLanes(f frame, g *d2graph.Graph, edges []*d2graph.Edge) error {
	lanes := lanesOf(g, edges)
	for _, e := range edges {
		var pts []point
		if r, ok := f.direct(e); ok {
			pts = r
		} else {
			pts = f.around(e, lanes)
		}
		e.Route = e.Route[:0]
		for _, p := range pts {
			e.Route = append(e.Route, f.out(p))
		}
		e.IsCurve = false
		if e.Label.Value != "" {
			pos := label.InsideMiddleCenter.String()
			e.LabelPosition = &pos
		}
		// The ends already sit on the shapes' outlines: the middles of a
		// rectangle's sides are also a diamond's tips.
	}
	return nil
}

// point is in flow coordinates: f along the flow, c across it.
type point struct{ c, f float64 }

// box is an object's extent in flow coordinates.
type box struct{ c0, c1, f0, f1 float64 }

func (b box) cmid() float64 { return (b.c0 + b.c1) / 2 }
func (b box) fmid() float64 { return (b.f0 + b.f1) / 2 }

// frame maps between canvas and flow coordinates, so one routine serves both
// orientations.
type frame struct{ vertical bool }

func (f frame) box(o *d2graph.Object) box {
	x0, y0 := o.TopLeft.X, o.TopLeft.Y
	x1, y1 := x0+o.Width, y0+o.Height
	if f.vertical {
		return box{x0, x1, y0, y1}
	}
	return box{y0, y1, x0, x1}
}

func (f frame) out(p point) *geo.Point {
	if f.vertical {
		return geo.NewPoint(p.c, p.f)
	}
	return geo.NewPoint(p.f, p.c)
}

// direct routes a forward connection whose source is the last step in its
// lane before the corridor leading into the target.
func (f frame) direct(e *d2graph.Edge) ([]point, bool) {
	s, d := f.box(e.Src), f.box(e.Dst)
	cor := f.corridor(e.Dst)
	if d.f0 <= s.f1 || cor <= s.f1 {
		return nil, false
	}
	for _, o := range e.Src.Parent.ChildrenArray {
		if o == e.Src || isGap(o) {
			continue
		}
		if b := f.box(o); b.f0 >= s.f1 && b.f0 < cor && b.c0 <= s.cmid() && s.cmid() <= b.c1 {
			return nil, false // a step of the same lane is in the way
		}
	}
	if math.Abs(s.cmid()-d.cmid()) < 0.5 {
		return []point{{s.cmid(), s.f1}, {d.cmid(), d.f0}}, true
	}
	return []point{
		{s.cmid(), s.f1},
		{s.cmid(), cor},
		{d.cmid(), cor},
		{d.cmid(), d.f0},
	}, true
}

// around leaves the source's side, runs along the gutter beside its lane to
// the corridor before the target, and turns into the target.
func (f frame) around(e *d2graph.Edge, lanes []*d2graph.Object) []point {
	s, d := f.box(e.Src), f.box(e.Dst)
	cor := f.corridor(e.Dst)
	side := 1.0
	if d.cmid() < s.cmid() {
		side = -1
	}
	exit := s.c1
	if side < 0 {
		exit = s.c0
	}
	gut := f.gutter(e.Src.Parent, lanes, side)
	return []point{
		{exit, s.fmid()},
		{gut, s.fmid()},
		{gut, cor},
		{d.cmid(), cor},
		{d.cmid(), d.f0},
	}
}

// corridor is the flow position midway through the gap before a step: the
// space between it and the previous cell of its lane, or just inside the lane
// for the first step, clear of the lane's label.
func (f frame) corridor(o *d2graph.Object) float64 {
	b := f.box(o)
	prev := math.Inf(-1)
	for _, sib := range o.Parent.ChildrenArray {
		if sb := f.box(sib); sib != o && sb.f1 <= b.f0 {
			prev = math.Max(prev, sb.f1)
		}
	}
	if math.IsInf(prev, -1) {
		return b.f0 - 12
	}
	return (prev + b.f0) / 2
}

// gutter is the cross position midway between a lane and its neighbour on the
// given side, or just outside the lane at the edge of the diagram.
func (f frame) gutter(lane *d2graph.Object, lanes []*d2graph.Object, side float64) float64 {
	l := f.box(lane)
	edge, next := l.c1, math.Inf(1)
	if side < 0 {
		edge, next = l.c0, math.Inf(-1)
	}
	for _, other := range lanes {
		o := f.box(other)
		if side > 0 && o.c0 >= l.c1 {
			next = math.Min(next, o.c0)
		}
		if side < 0 && o.c1 <= l.c0 {
			next = math.Max(next, o.c1)
		}
	}
	if math.IsInf(next, 0) {
		return edge + side*20
	}
	return (edge + next) / 2
}

// lanesOf returns every lane of the graph: the parents of routed steps and
// their siblings, in order.
func lanesOf(g *d2graph.Graph, edges []*d2graph.Edge) []*d2graph.Object {
	seen := map[*d2graph.Object]bool{}
	var lanes []*d2graph.Object
	add := func(o *d2graph.Object) {
		if o != nil && !seen[o] {
			seen[o] = true
			lanes = append(lanes, o)
		}
	}
	for _, e := range edges {
		if p := e.Src.Parent; p != nil && p.Parent != nil {
			for _, l := range p.Parent.ChildrenArray {
				add(l)
			}
		}
	}
	return lanes
}

func isGap(o *d2graph.Object) bool { return strings.HasPrefix(o.ID, ids.GapPrefix) }
