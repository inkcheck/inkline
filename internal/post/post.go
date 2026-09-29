// Package post applies a design system's SVG decorators to D2 output: the
// parts of a visual language D2's style model cannot express (Carbon side bars,
// colour blocks, pill labels, strike-through).
//
// D2 writes each object and connection as a <g> whose class attribute holds an
// encoded ID followed by the element's D2 classes, so decorators select by
// class: a role, a modifier, or a marker class an author adds. Decorations take
// their colour from the element's stroke, so domain colours carry through.
package post

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Decorator is one entry of a manifest's "decorators" list.
type Decorator struct {
	// Type is one of: bar, block, icon, avatar, fill, legend, chevron, pill,
	// strike.
	//   bar:    a solid band along one side of a node (Carbon side bar / top bar)
	//   block:  a solid square at the left of a node, holding its icon (Carbon colour block)
	//   icon:   an icon centred in a node (Carbon actor circles)
	//   avatar: a badge at the left of a node, holding the icon: round, or
	//           square with shape "square"; filled with fill ("stroke" = the
	//           node's domain colour) or the theme's text colour (Cloudscape
	//           avatars, Fluent service badges)
	//   fill:   fills a node with its stroke colour (legend swatches)
	//   legend: flattens the legend: no frame, shadow or rounded corners
	//           (takes no classes; applies to the whole diagram)
	//   chevron: draws filled arrowheads as open chevrons (takes no classes;
	//           hollow UML triangles, diamonds and crow's feet are kept)
	//   pill:   a rounded outline around a connection's label
	//   strike: a diagonal line through a node (Carbon "removed")
	Type string `json:"type"`
	// Classes selects the elements to decorate (any match).
	Classes []string `json:"classes"`
	// Diagrams limits the decorator to these diagram types; empty means all.
	// Roles are per diagram type, so the same class name can mean different
	// things in different diagrams.
	Diagrams []string `json:"diagrams,omitempty"`
	// Side places a bar: left (default), top or bottom.
	Side string `json:"side,omitempty"`
	// Size is the bar thickness, block width or avatar diameter, in px.
	Size float64 `json:"size,omitempty"`
	// Icon is an SVG file in the design system, drawn by block (when the node
	// has no icon of its own) and icon. It is filled with the canvas colour, so
	// it reads on the solid block or shape in both themes.
	Icon string `json:"icon,omitempty"`
	// IconStroke draws the icon as an outline, with strokes this wide in the
	// icon's own units (Cloudscape icons: 2 on a 16px grid). Zero fills it.
	IconStroke float64 `json:"icon_stroke,omitempty"`
	// Fill colours an avatar badge; IconColor colours the icon. Defaults: the
	// theme's text colour and the canvas colour.
	// Both take a colour, or one per theme: {"light": ..., "dark": ...}.
	// Fill "stroke" uses the node's stroke colour (its domain colour).
	Fill      Colour `json:"fill,omitempty"`
	IconColor Colour `json:"icon_color,omitempty"`
	// Shape of an avatar badge: circle (default) or square.
	Shape string `json:"shape,omitempty"`
	// Outline, by theme, restrokes a block's node after the block takes the
	// domain colour: the colour stays in the block, the border goes neutral.
	Outline Colour `json:"outline,omitempty"`
	// IconSVG is the icon file's content, loaded by the design system.
	IconSVG string `json:"-"`

	theme string
}

// Validate reports a decorator its manifest got wrong.
func (d Decorator) Validate() error {
	switch d.Type {
	case "bar", "block", "icon", "avatar", "fill", "legend", "chevron", "pill", "strike":
	default:
		return fmt.Errorf("unknown decorator type %q", d.Type)
	}
	switch d.Side {
	case "", "left", "top", "bottom":
	default:
		return fmt.Errorf("%s decorator: unknown side %q; one of: left, top, bottom", d.Type, d.Side)
	}
	switch d.Shape {
	case "", "circle", "square":
	default:
		return fmt.Errorf("%s decorator: unknown shape %q; one of: circle, square", d.Type, d.Shape)
	}
	return nil
}

// Colour is a colour for every theme, or one per theme. In JSON it is a string
// ("#fff") or an object ({"light": "#fff", "dark": "#000"}).
type Colour map[string]string

// UnmarshalJSON accepts a plain string as a colour for every theme.
func (c *Colour) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*c = Colour{"": one}
		return nil
	}
	var byTheme map[string]string
	if err := json.Unmarshal(b, &byTheme); err != nil {
		return err
	}
	*c = byTheme
	return nil
}

// at returns the colour for a theme, or the one for every theme.
func (c Colour) at(theme string) string {
	if v, ok := c[theme]; ok {
		return v
	}
	return c[""]
}

var (
	// D2 adds a style attribute to the group of an element with opacity.
	groupOpen = regexp.MustCompile(`<g class="([^"]+)"[^>]*>`)
	maskRect  = regexp.MustCompile(`<rect x="([-\d.]+)" y="([-\d.]+)" width="([-\d.]+)" height="([-\d.]+)" fill="black"></rect>`)
	canvasCSS = regexp.MustCompile(`\.fill-N7\{fill:(#[0-9a-fA-F]+);\}`)
	// The theme's text colour. D2 always writes it for its appendix; the
	// fill-N1 class is only emitted when a diagram uses it.
	textFill = regexp.MustCompile(`\.appendix text\.text\{fill:(#[0-9a-fA-F]+)\}`)
	textCSS  = regexp.MustCompile(`\.fill-N1\{fill:(#[0-9a-fA-F]+);\}`)
)

// Apply returns svg, a rendered diagram of the given type and theme, with
// decorators applied.
func Apply(svg []byte, diagram, theme string, all []Decorator) []byte {
	var decorators []Decorator
	flat, chevrons := false, false
	for _, d := range all {
		if len(d.Diagrams) == 0 || slices.Contains(d.Diagrams, diagram) {
			switch d.Type {
			case "legend":
				flat = true
				continue
			case "chevron":
				chevrons = true
				continue
			}
			d.theme = theme
			decorators = append(decorators, d)
		}
	}
	s := string(svg)
	canvas, text := "#ffffff", "#000000"
	if m := canvasCSS.FindStringSubmatch(s); m != nil {
		canvas = m[1]
	}
	if m := textFill.FindStringSubmatch(s); m != nil {
		text = m[1]
	} else if m := textCSS.FindStringSubmatch(s); m != nil {
		text = m[1]
	}
	s = themeLegend(s, canvas, text, flat)
	s = clearLabels(s)
	if chevrons {
		s = chevronHeads(s)
	}
	if len(decorators) == 0 {
		return []byte(s)
	}
	var labels []box
	for _, m := range maskRect.FindAllStringSubmatch(s, -1) {
		b := box{}
		b.x, _ = strconv.ParseFloat(m[1], 64)
		b.y, _ = strconv.ParseFloat(m[2], 64)
		b.w, _ = strconv.ParseFloat(m[3], 64)
		b.h, _ = strconv.ParseFloat(m[4], 64)
		labels = append(labels, b)
	}

	var out strings.Builder
	clips := 0
	pos := 0
	for {
		loc := groupOpen.FindStringSubmatchIndex(s[pos:])
		if loc == nil {
			break
		}
		start, openEnd := pos+loc[0], pos+loc[1]
		classes := strings.Fields(s[pos+loc[2] : pos+loc[3]])
		if len(classes) < 2 || classes[0] == "shape" {
			out.WriteString(s[pos:openEnd])
			pos = openEnd
			continue
		}
		end := groupEnd(s, openEnd)
		group := s[start:end]
		for _, d := range decorators {
			if slices.ContainsFunc(classes[1:], func(c string) bool { return slices.Contains(d.Classes, c) }) {
				group = d.apply(group, canvas, text, labels, &clips)
			}
		}
		out.WriteString(s[pos:start])
		out.WriteString(group)
		pos = end
	}
	out.WriteString(s[pos:])
	return []byte(out.String())
}
