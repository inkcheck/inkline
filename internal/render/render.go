// Package render turns expanded D2 into SVG using D2's Go packages, so
// inkline needs no d2 executable.
package render

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2layouts/d2dagrelayout"
	"github.com/d2lang/d2/d2layouts/d2elklayout"
	"github.com/d2lang/d2/d2lib"
	"github.com/d2lang/d2/d2renderers/d2fonts"
	"github.com/d2lang/d2/d2renderers/d2svg"
	"github.com/d2lang/d2/lib/imgbundler"
	"github.com/d2lang/d2/lib/label"
	d2log "github.com/d2lang/d2/lib/log"
	"github.com/d2lang/d2/lib/simplelog"
	"github.com/d2lang/d2/lib/textmeasure"
)

// Options configure one render.
type Options struct {
	// Path is the source file's path. D2 imports resolve against its
	// directory within FS, or on disk when FS is nil.
	Path string
	FS   fs.FS
	// IconPath is the source file's path on disk. Icon files named by the
	// diagram resolve against its directory.
	IconPath string
	// Font is the design system's font family, or nil for D2's default.
	Font *d2fonts.FontFamily
	// Router draws connections that cross between nested layouts (grid
	// cells, for example). Nil keeps D2's default: straight lines.
	Router d2graph.RouteEdges
}

// SVG renders expanded D2. Theme, padding and layout engine come from the
// d2-config the design system's prelude writes.
func SVG(ctx context.Context, doc string, opts Options) ([]byte, error) {
	// D2 logs through a logger it expects in the context, and warns loudly when
	// there is none. Errors come back as values, so the log goes nowhere.
	ctx = d2log.With(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("text ruler: %w", err)
	}
	renderOpts := &d2svg.RenderOpts{}
	diagram, _, err := d2lib.Compile(ctx, doc, &d2lib.CompileOptions{
		FS:             opts.FS,
		InputPath:      opts.Path,
		Ruler:          ruler,
		FontFamily:     opts.Font,
		LayoutResolver: layoutResolver,
		RouterResolver: func(string) (d2graph.RouteEdges, error) { return opts.Router, nil },
	}, renderOpts)
	if err != nil {
		// Positions are lines of doc, not of the author's file.
		return nil, fmt.Errorf("expanded D2: %w", err)
	}
	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, err
	}
	// Inline icon files, so the SVG stands alone.
	svg, err = imgbundler.BundleLocal(ctx, simplelog.Make(nil, nil, nil), opts.IconPath, svg, false)
	if err != nil {
		return nil, err
	}
	return append(svg, '\n'), nil
}

// layoutResolver maps a layout engine name to D2's layout. Only the open-source
// engines are offered: TALA, D2's third engine, is proprietary and needs a paid
// licence, so inkline neither bundles nor calls it.
func layoutResolver(engine string) (d2graph.LayoutGraph, error) {
	switch strings.ToLower(engine) {
	case "", "dagre":
		return d2dagrelayout.DefaultLayout, nil
	case "elk":
		return elkLayout, nil
	case "tala":
		return nil, fmt.Errorf("layout engine tala is not supported: it is proprietary (Terrastruct); use dagre or elk")
	}
	return nil, fmt.Errorf("unknown layout engine %q; one of: dagre, elk", engine)
}

// elkLayout is ELK, except that a shape with an explicit height keeps it.
// ELK adds the label's height to every shape with both an icon and a label,
// to fit the label above the icon; a kit that sets a height (a 48px
// service tile, a card that holds its icon in a badge) has already made
// room, so the addition is taken off first.
func elkLayout(ctx context.Context, g *d2graph.Graph) error {
	for _, obj := range g.Objects {
		if obj.HeightAttr != nil && obj.HasLabel() && obj.HasIcon() {
			obj.Height -= float64(obj.LabelDimensions.Height + label.PADDING)
		}
	}
	return d2elklayout.DefaultLayout(ctx, g)
}
