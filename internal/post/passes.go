package post

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	legendBox  = regexp.MustCompile(`<rect [^>]*?stroke="#DEE1EB"[^>]*?/>`)
	bareText   = regexp.MustCompile(`<text class="(text|text-bold)"`)
	fontSize   = regexp.MustCompile(`font-size:\s*([\d.]+)px`)
	markerHead = regexp.MustCompile(`<polygon points="[^"]*" [^>]*?class="connection"[^>]*?/>`)
	coords     = regexp.MustCompile(`([-\d.]+) ([-\d.]+)`)
)

// chevronHeads redraws D2's filled arrowheads (triangle and arrow) as open
// chevrons, Cloudscape's arrow style. Hollow triangles (which D2 strokes),
// diamonds and crow's feet mean something else and are left alone.
func chevronHeads(s string) string {
	return markerHead.ReplaceAllStringFunc(s, func(poly string) string {
		if get(poly, "stroke") != "" {
			return poly // a hollow triangle
		}
		pts := strings.Fields(get(poly, "points"))
		if len(pts) != 3 && len(pts) != 4 {
			return poly
		}
		xy := func(p string) (float64, float64) {
			x, y, _ := strings.Cut(p, ",")
			fx, _ := strconv.ParseFloat(x, 64)
			fy, _ := strconv.ParseFloat(y, 64)
			return fx, fy
		}
		x0, y0 := xy(pts[0])
		x1, y1 := xy(pts[1])
		x2, y2 := xy(pts[2])
		if len(pts) == 4 {
			if x3, y3 := xy(pts[3]); x3 <= 0 || y3 != y1 || x0 != 0 {
				return poly // a diamond, not a barbed arrow
			}
		}
		const w = 1.5
		return fmt.Sprintf(`<polyline points="%g,%g %g,%g %g,%g" fill="none" stroke="%s" stroke-width="%g" stroke-linecap="round" stroke-linejoin="round" />`,
			x0+w, y0+w, x1-w, y1, x2+w, y2-w, get(poly, "fill"), w)
	})
}

// clearLabels keeps connections off the labels D2 draws below a shape (actor
// circles, tiles). D2 runs a connection that leaves or enters such a node
// straight up or down to the shape itself, through its label; bent ones it
// already starts below the label. A straight end is moved to the label's foot.
func clearLabels(s string) string {
	type below struct{ left, right, top, bottom float64 }
	var labels []below
	for _, m := range groupOpen.FindAllStringIndex(s, -1) {
		g := s[m[0]:groupEnd(s, m[1])]
		end := strings.Index(g, "</g>")
		t := textTag.FindStringSubmatch(g)
		if end < 0 || t == nil {
			continue
		}
		b, ok := shapeBox(g[:end])
		if !ok {
			continue
		}
		y, _ := strconv.ParseFloat(t[2], 64)
		size := 16.0
		if f := fontSize.FindStringSubmatch(g); f != nil {
			size, _ = strconv.ParseFloat(f[1], 64)
		}
		if y-size > b.y+b.h-1 { // the label sits below the shape
			labels = append(labels, below{b.x, b.x + b.w, b.y + b.h, y + size*0.3})
		}
	}
	if len(labels) == 0 {
		return s
	}
	return connPath.ReplaceAllStringFunc(s, func(tag string) string {
		d := get(tag, "d")
		pts := coords.FindAllStringSubmatchIndex(d, -1)
		if len(pts) < 2 {
			return tag
		}
		pt := func(i int) (float64, float64) {
			x, _ := strconv.ParseFloat(d[pts[i][2]:pts[i][3]], 64)
			y, _ := strconv.ParseFloat(d[pts[i][4]:pts[i][5]], 64)
			return x, y
		}
		fix := func(i, j int, down bool) {
			x, y := pt(i)
			nx, ny := pt(j)
			if abs(x-nx) > 0.5 || (down && ny <= y) || (!down && ny >= y) {
				return // not a straight vertical end heading away from the node
			}
			for _, l := range labels {
				if x > l.left && x < l.right && y >= l.top-2 && y < l.bottom {
					at := pts[i]
					d = d[:at[4]] + strconv.FormatFloat(l.bottom+2, 'f', -1, 64) + d[at[5]:]
					return
				}
			}
		}
		fix(len(pts)-1, len(pts)-2, false) // entering from below
		fix(0, 1, true)                    // leaving downwards
		return attr("d").ReplaceAllString(tag, ` d="`+d+`"`)
	})
}

// themeLegend recolours D2's legend, which it draws in fixed light colours
// whatever the theme: the frame takes the canvas colour and the legend text,
// which D2 leaves uncoloured, the text colour. flat also removes the frame's
// border, shadow and rounded corners.
func themeLegend(s, canvas, text string, flat bool) string {
	s = legendBox.ReplaceAllStringFunc(s, func(r string) string {
		r = attr("fill").ReplaceAllString(r, ` fill="`+canvas+`"`)
		if flat {
			r = attr("stroke").ReplaceAllString(r, ` stroke="none"`)
			r = attr("rx").ReplaceAllString(r, ` rx="0"`)
			r = attr("style").ReplaceAllString(r, ` style="stroke-width: 0"`)
		}
		return r
	})
	return bareText.ReplaceAllString(s, `<text fill="`+text+`" class="$1"`)
}
