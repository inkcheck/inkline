package post

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	rectTag  = regexp.MustCompile(`<rect [^>]*?/>`)
	textTag  = regexp.MustCompile(`<text x="([-\d.]+)" y="([-\d.]+)"`)
	tspanTag = regexp.MustCompile(`<tspan x="[-\d.]+"`)
	imageTag = regexp.MustCompile(`<image [^>]*?/>`)
	connPath = regexp.MustCompile(`<path [^>]*?class="connection"[^>]*?/>`)
	ellipse  = regexp.MustCompile(`<ellipse [^>]*?/>`)
	shapeTag = regexp.MustCompile(`<(?:rect|ellipse|path|polygon) [^>]*?/>`)
	svgOpen  = regexp.MustCompile(`<svg[^>]*>`)
	svgInner = regexp.MustCompile(`(?s)<svg[^>]*>(.*)</svg>`)

	attrs sync.Map // attribute name -> *regexp.Regexp
)

// attr matches an attribute of a tag, capturing its value.
func attr(name string) *regexp.Regexp {
	if re, ok := attrs.Load(name); ok {
		return re.(*regexp.Regexp)
	}
	re := regexp.MustCompile(` ` + name + `="([^"]*)"`)
	attrs.Store(name, re)
	return re
}

func get(tag, name string) string {
	if m := attr(name).FindStringSubmatch(tag); m != nil {
		return m[1]
	}
	return ""
}

func num(tag, name string) float64 {
	f, _ := strconv.ParseFloat(get(tag, name), 64)
	return f
}

func set(tag, name string, v float64) string {
	return attr(name).ReplaceAllString(tag, fmt.Sprintf(` %s="%s"`, name, strconv.FormatFloat(v, 'f', -1, 64)))
}

type box struct{ x, y, w, h float64 }

// shapeBox is the bounding box of a node's shape: a rect, an ellipse or a path.
func shapeBox(shape string) (box, bool) {
	if r := rectTag.FindString(shape); r != "" {
		return box{num(r, "x"), num(r, "y"), num(r, "width"), num(r, "height")}, true
	}
	if e := ellipse.FindString(shape); e != "" {
		rx, ry := num(e, "rx"), num(e, "ry")
		return box{num(e, "cx") - rx, num(e, "cy") - ry, 2 * rx, 2 * ry}, true
	}
	return box{}, false
}

// groupEnd returns the index just past the </g> closing the group opened before i.
func groupEnd(s string, i int) int {
	depth := 1
	for depth > 0 {
		o := strings.Index(s[i:], "<g")
		c := strings.Index(s[i:], "</g>")
		if c < 0 {
			return len(s)
		}
		if o >= 0 && o < c {
			depth++
			i += o + 2
			continue
		}
		depth--
		i += c + 4
	}
	return i
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func orDefault(v, d float64) float64 {
	if v == 0 {
		return d
	}
	return v
}

func orZero(s string) string {
	if s == "" {
		return "0"
	}
	return s
}
