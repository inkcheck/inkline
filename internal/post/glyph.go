package post

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

var (
	// paint matches a colour in a fill or stroke attribute or style rule.
	paint  = regexp.MustCompile(`((?:fill|stroke)(?:="|:\s*))([^";}]+)`)
	svgURI = "data:image/svg+xml;base64,"
)

// glyph returns the SVG of an <image> when it is a single-colour icon, as
// most icon sets are (Carbon, Cloudscape, Fluent), so a decorator can
// recolour it. Full-colour icons, such as AWS's and Azure's, which carry
// their own background, return "" and are drawn as they are.
func glyph(img string) string {
	href := get(img, "href")
	if !strings.HasPrefix(href, svgURI) {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(href[len(svgURI):])
	if err != nil {
		return ""
	}
	src := string(b)
	if strings.Contains(src, "Gradient") || strings.Contains(src, "<image") {
		return ""
	}
	colours := map[string]bool{}
	for _, m := range paint.FindAllStringSubmatch(src, -1) {
		if c := paintColour(m[2]); c != "" {
			colours[c] = true
		}
	}
	if len(colours) > 1 {
		return ""
	}
	return src
}

// paintColour normalises a paint value; "" means it paints nothing or
// inherits.
func paintColour(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "", "none", "transparent", "currentcolor", "inherit":
		return ""
	}
	return v
}

// inlineGlyph draws a single-colour SVG icon in the given colour, size×size
// at x, y, keeping which parts are filled and which are stroked.
func inlineGlyph(src string, x, y, size float64, colour string) string {
	inner := svgInner.FindStringSubmatch(src)
	if inner == nil {
		return ""
	}
	root := svgOpen.FindString(src)
	vb := get(root, "viewBox")
	if vb == "" {
		vb = fmt.Sprintf("0 0 %s %s", or(get(root, "width"), "24"), or(get(root, "height"), "24"))
	}
	recolour := func(s string) string {
		return paint.ReplaceAllStringFunc(s, func(m string) string {
			p := paint.FindStringSubmatch(m)
			if paintColour(p[2]) == "" && !strings.EqualFold(strings.TrimSpace(p[2]), "currentColor") {
				return m
			}
			return p[1] + colour
		})
	}
	// The root's own paint decides the default for parts that set none.
	fill := `fill="` + colour + `"`
	if f := get(root, "fill"); f != "" && paintColour(f) == "" && !strings.EqualFold(f, "currentColor") {
		fill = `fill="` + f + `"`
	}
	stroke := ""
	if st := get(root, "stroke"); st != "" && (paintColour(st) != "" || strings.EqualFold(st, "currentColor")) {
		stroke = fmt.Sprintf(` stroke="%s"`, colour)
		for _, a := range []string{"stroke-width", "stroke-linecap", "stroke-linejoin"} {
			if v := get(root, a); v != "" {
				stroke += fmt.Sprintf(` %s="%s"`, a, v)
			}
		}
	}
	return fmt.Sprintf(`<svg x="%g" y="%g" width="%g" height="%g" viewBox="%s" %s%s>%s</svg>`,
		x, y, size, size, vb, fill, stroke, recolour(inner[1]))
}
