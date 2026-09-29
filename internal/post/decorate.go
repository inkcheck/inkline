package post

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	fillClass = regexp.MustCompile(` fill-[A-Z]+[0-9]+`)
	viewBox   = regexp.MustCompile(`viewBox="([^"]+)"`)
	iconClass = regexp.MustCompile(` class="[^"]*"`)
	// ownFill matches an icon element's own fill, other than none.
	ownFill = regexp.MustCompile(`\sfill="(?:[^"n]|n[^o])[^"]*"`)
)

func (d Decorator) apply(group, canvas, text string, labels []box, clips *int) string {
	switch d.Type {
	case "bar", "block", "icon", "avatar", "strike":
		return d.decorateNode(group, canvas, text, clips)
	case "pill":
		return pill(group, canvas, labels)
	case "fill":
		return fillShape(group, or(d.IconColor.at(d.theme), "#ffffff"))
	}
	return group
}

// fillShape paints a node's shape in its stroke colour, so a domain class,
// which sets only the outline, reads as a solid swatch. A single-colour icon
// on it is redrawn in the given colour.
func fillShape(group, iconColor string) string {
	group = imageTag.ReplaceAllStringFunc(group, func(img string) string {
		if g := glyph(img); g != "" {
			return inlineGlyph(g, num(img, "x"), num(img, "y"), num(img, "width"), iconColor)
		}
		return img
	})
	end := strings.Index(group, "</g>")
	if end < 0 {
		return group
	}
	shape := shapeTag.ReplaceAllStringFunc(group[:end], func(tag string) string {
		if c := get(tag, "stroke"); c != "" {
			tag = attr("fill").ReplaceAllString(tag, ` fill="`+c+`"`)
			// D2 also colours shapes through a fill-XX class, which CSS
			// applies over the attribute.
			return fillClass.ReplaceAllString(tag, "")
		}
		return tag
	})
	return shape + group[end:]
}

// decorateNode draws over a node, between its shape and its label. Bars,
// blocks and strikes need a rectangle; icons also centre in circles.
func (d Decorator) decorateNode(group, canvas, text string, clips *int) string {
	shapeEnd := strings.Index(group, "</g>")
	if shapeEnd < 0 {
		return group
	}
	shape := group[:shapeEnd]
	// Double borders and multiples draw several rects; the outline that
	// frames the node is the left-most one.
	var main string
	for _, r := range rectTag.FindAllString(shape, -1) {
		if main == "" || num(r, "x") < num(main, "x") {
			main = r
		}
	}
	if main == "" {
		if d.Type == "icon" {
			if e := ellipse.FindString(shape); e != "" {
				cx, cy, rx, ry := num(e, "cx"), num(e, "cy"), num(e, "rx"), num(e, "ry")
				icon := d.icon(cx-rx, cy-ry, 2*rx, 2*ry, 24, or(d.IconColor.at(d.theme), canvas))
				return group[:shapeEnd] + icon + group[shapeEnd:]
			}
		}
		return group // not a rectangle (diamond, …): nothing to align to
	}
	b := box{num(main, "x"), num(main, "y"), num(main, "width"), num(main, "height")}
	color := get(main, "stroke")

	// Bars, blocks and strikes are clipped to the node's outline.
	var clip string
	clipID := func() string {
		*clips++
		id := fmt.Sprintf("vf-clip-%d", *clips)
		clip = fmt.Sprintf(`<clipPath id="%s"><rect x="%g" y="%g" width="%g" height="%g" rx="%s" /></clipPath>`,
			id, b.x, b.y, b.w, b.h, orZero(get(main, "rx")))
		return id
	}

	var deco string
	shift := 0.0 // how far the label moves right
	switch d.Type {
	case "bar":
		size := orDefault(d.Size, 4)
		r := box{b.x, b.y, size, b.h}
		switch d.Side {
		case "top":
			r = box{b.x, b.y, b.w, size}
		case "bottom":
			r = box{b.x, b.y + b.h - size, b.w, size}
		default:
			shift = size + 6
		}
		deco = fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g" fill="%s" clip-path="url(#%s)" />`, r.x, r.y, r.w, r.h, color, clipID())
	case "icon":
		deco = d.icon(b.x, b.y, b.w, b.h, 24, or(d.IconColor.at(d.theme), canvas))
	case "avatar":
		size := orDefault(d.Size, 28)
		inset := (b.h - size) / 2
		fill := or(d.Fill.at(d.theme), text)
		if fill == "stroke" {
			fill = color
		}
		ax, ay := b.x+inset, b.y+inset
		img := imageTag.FindString(group)
		if g := glyph(img); g != "" {
			// A single-colour icon sits on the badge, recoloured.
			group = strings.Replace(group, img, "", 1)
			shapeEnd = strings.Index(group, "</g>")
			deco = d.badge(ax, ay, size, fill) + inlineGlyph(g, ax+size*0.2, ay+size*0.2, size*0.6, or(d.IconColor.at(d.theme), canvas))
		} else if img != "" {
			// The node's own icon takes the badge's place.
			group = imageTag.ReplaceAllStringFunc(group, func(img string) string {
				img = set(img, "width", size)
				img = set(img, "height", size)
				img = set(img, "x", ax)
				return set(img, "y", ay)
			})
			shapeEnd = strings.Index(group, "</g>")
		} else {
			deco = d.badge(ax, ay, size, fill) + d.icon(ax, ay, size, size, size*0.6, or(d.IconColor.at(d.theme), canvas))
		}
		// A left-aligned label moves past the avatar; a centred one re-centres
		// in the space that is left.
		shift = inset + size
		if t := textTag.FindStringSubmatch(group[shapeEnd:]); t != nil {
			if x, _ := strconv.ParseFloat(t[1], 64); abs(x-(b.x+b.w/2)) < 1 {
				shift = (inset + size) / 2
			}
		}
	case "block":
		size := orDefault(d.Size, 48)
		deco = fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g" fill="%s" clip-path="url(#%s)" />`, b.x, b.y, size, b.h, color, clipID())
		if !imageTag.MatchString(group) {
			deco += d.icon(b.x, b.y, size, b.h, 24, or(d.IconColor.at(d.theme), canvas))
		}
		// A left-aligned label moves past the block; a centred one re-centres
		// in the space that is left.
		shift = size + 10
		if t := textTag.FindStringSubmatch(group[shapeEnd:]); t != nil {
			if x, _ := strconv.ParseFloat(t[1], 64); abs(x-(b.x+b.w/2)) < 1 {
				shift = size / 2
			}
		}
		// The icon moves into the block; a single-colour one is recoloured.
		group = imageTag.ReplaceAllStringFunc(group, func(img string) string {
			icon := 24.0
			if g := glyph(img); g != "" {
				return inlineGlyph(g, b.x+(size-icon)/2, b.y+(b.h-icon)/2, icon, or(d.IconColor.at(d.theme), canvas))
			}
			img = set(img, "width", icon)
			img = set(img, "height", icon)
			img = set(img, "x", b.x+(size-icon)/2)
			return set(img, "y", b.y+(b.h-icon)/2)
		})
		shapeEnd = strings.Index(group, "</g>")
	case "strike":
		deco = fmt.Sprintf(`<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="%s" stroke-width="1" clip-path="url(#%s)" />`,
			b.x, b.y, b.x+b.w, b.y+b.h, color, clipID())
	}
	if c := d.Outline.at(d.theme); c != "" && (d.Type == "block" || d.Type == "avatar") {
		// The block or badge has the domain colour; the node's border goes
		// neutral.
		shape := strings.ReplaceAll(group[:shapeEnd], ` stroke="`+color+`"`, ` stroke="`+c+`"`)
		group = shape + group[shapeEnd:]
		shapeEnd = strings.Index(group, "</g>")
	}
	rest := group[shapeEnd:]
	if shift != 0 {
		rest = textTag.ReplaceAllStringFunc(rest, func(t string) string {
			m := textTag.FindStringSubmatch(t)
			x, _ := strconv.ParseFloat(m[1], 64)
			return fmt.Sprintf(`<text x="%g" y="%s"`, x+shift, m[2])
		})
		// Each line of a multi-line label carries its own x.
		rest = tspanTag.ReplaceAllStringFunc(rest, func(t string) string {
			return set(t, "x", num(t, "x")+shift)
		})
	}
	return group[:shapeEnd] + clip + deco + rest
}

// pill outlines a connection's label with a rounded box on the canvas colour.
func pill(group, canvas string, labels []box) string {
	t := textTag.FindStringSubmatchIndex(group)
	if t == nil {
		return group
	}
	x, _ := strconv.ParseFloat(group[t[2]:t[3]], 64)
	y, _ := strconv.ParseFloat(group[t[4]:t[5]], 64)
	color := "#161616"
	if p := connPath.FindString(group); p != "" {
		color = get(p, "stroke")
	}
	// D2 masks each connection label out of its line; that mask box is the
	// label's exact extent. The text's x is its centre, its y the baseline.
	for _, l := range labels {
		if x >= l.x && x <= l.x+l.w && y >= l.y && y <= l.y+l.h+4 {
			const padX, padY = 8.0, 2.0
			h := l.h + 2*padY
			r := fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g" rx="%g" fill="%s" stroke="%s" stroke-width="1" />`,
				l.x-padX, l.y-padY, l.w+2*padX, h, h/2, canvas, color)
			return group[:t[0]] + r + group[t[0]:]
		}
	}
	return group
}

// icon draws the decorator's icon, size px square, centred in the box, in the
// given colour.
func (d Decorator) icon(x, y, w, h, size float64, fill string) string {
	if d.IconSVG == "" {
		return ""
	}
	inner := svgInner.FindStringSubmatch(d.IconSVG)
	if inner == nil {
		return ""
	}
	vb := "0 0 32 32"
	if m := viewBox.FindStringSubmatch(d.IconSVG); m != nil {
		vb = m[1]
	}
	body, paint := inner[1], fmt.Sprintf(`fill="%s"`, fill)
	if d.IconStroke > 0 {
		// An outline icon: strokes in the colour, and only parts marked
		// "filled" filled, as Cloudscape's icon CSS does.
		paint = fmt.Sprintf(`fill="none" stroke="%s" stroke-width="%g" stroke-linejoin="round" stroke-linecap="round"`, fill, d.IconStroke)
		body = strings.ReplaceAll(body, `class="filled"`, `fill="`+fill+`"`)
		body = iconClass.ReplaceAllString(body, "")
	} else {
		// A filled icon: its own fills (Fluent's are #212121) give way to the
		// colour; fill="none" parts stay open.
		body = ownFill.ReplaceAllString(body, "")
	}
	return fmt.Sprintf(`<svg x="%g" y="%g" width="%g" height="%g" viewBox="%s" %s>%s</svg>`,
		x+(w-size)/2, y+(h-size)/2, size, size, vb, paint, body)
}

// badge is the avatar decorator's round or square badge.
func (d Decorator) badge(x, y, size float64, fill string) string {
	if d.Shape == "square" {
		return fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g" rx="4" fill="%s" />`, x, y, size, size, fill)
	}
	return fmt.Sprintf(`<circle cx="%g" cy="%g" r="%g" fill="%s" />`, x+size/2, y+size/2, size/2, fill)
}
