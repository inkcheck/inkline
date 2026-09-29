// Command inkline renders inkline diagrams: D2 files with a inkline header.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/inkcheck/inkline/internal/core"
	"github.com/inkcheck/inkline/internal/designsystem"
	"github.com/inkcheck/inkline/internal/post"
	"github.com/inkcheck/inkline/internal/render"
	"github.com/inkcheck/inkline/internal/route"
)

const usageText = `inkline renders D2 diagrams with a inkline header.

Usage:
  inkline render [flags] <file.d2>   render to SVG
  inkline expand [flags] <file.d2>   print the expanded plain D2
  inkline list                       list diagram types and design systems
  inkline version                    print the version

A inkline file is D2 with a header naming its diagram type:

  vars: { inkline: { diagram: flowchart; design_system: carbon; theme: light } }

  diagram        %s (required)
  design_system  %s (default plain)
  theme          light or dark (default light)
  orientation    swimlanes only: vertical or horizontal

Say what things are with roles, e.g. {class: decision} or
{class: [component; data]}; see https://github.com/inkcheck/inkline.

Flags (override the file's header):
`

// usage fills in the diagram types and design systems, so the text cannot
// drift from what is bundled.
func usage() string {
	const indent = "                 "
	return fmt.Sprintf(usageText, wrap(core.DiagramNames(), indent, 56), wrap(designsystem.Names(), indent, 56))
}

// wrap joins words with commas, starting a new indented line at width.
func wrap(words []string, indent string, width int) string {
	var b strings.Builder
	line := 0
	for i, w := range words {
		if i < len(words)-1 {
			w += ","
		}
		switch {
		case i == 0:
		case line+1+len(w) > width:
			b.WriteString("\n" + indent)
			line = 0
		default:
			b.WriteString(" ")
			line++
		}
		b.WriteString(w)
		line += len(w)
	}
	return b.String()
}

// Set by the release build (see .goreleaser.yaml).
var (
	version = ""
	commit  = ""
	date    = ""
)

// versionString reports the release version, or the module version when
// installed with go install, or "dev".
func versionString() string {
	v := version
	if v == "" {
		v = "dev"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	if commit != "" {
		v += fmt.Sprintf(" (%s, %s)", commit, date)
	}
	return v
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "inkline:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage())
		return fmt.Errorf("missing command")
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "help", "-h", "--help":
		fmt.Print(usage())
		return nil
	case "version", "--version", "-v":
		fmt.Println("inkline", versionString())
		return nil
	case "list":
		fmt.Println("diagrams:      ", strings.Join(core.DiagramNames(), ", "))
		fmt.Println("design systems:", strings.Join(designsystem.Names(), ", "))
		return nil
	case "render", "expand":
	default:
		fmt.Fprint(os.Stderr, usage())
		return fmt.Errorf("unknown command %q", cmd)
	}

	fl := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fl.Usage = func() { fmt.Fprint(os.Stderr, usage()); fl.PrintDefaults() }
	out := new(string)
	if cmd == "render" {
		out = fl.String("o", "", "output SVG (default: <file>.svg)")
	}
	dsName := fl.String("design-system", "", "design system")
	theme := fl.String("theme", "", "theme: light or dark")
	// Parse flags and arguments in any order, so both of these work:
	//   inkline render -o out.svg in.d2
	//   inkline render in.d2 -o out.svg
	// Everything after -- is a file.
	var files []string
	for rest := args; len(rest) > 0; {
		if err := fl.Parse(rest); errors.Is(err, flag.ErrHelp) {
			return nil
		} else if err != nil {
			return err
		}
		parsed := rest[:len(rest)-fl.NArg()]
		rest = fl.Args()
		if len(parsed) > 0 && parsed[len(parsed)-1] == "--" {
			files = append(files, rest...)
			break
		}
		if len(rest) > 0 {
			files, rest = append(files, rest[0]), rest[1:]
		}
	}
	if len(files) != 1 {
		fl.Usage()
		return fmt.Errorf("expected one input file")
	}
	in := files[0]

	text, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	// Imports and icons resolve on disk, against the file's directory.
	f, err := core.Parse(core.Source{Path: filepath.ToSlash(in), Text: string(text)})
	if err != nil {
		return err
	}
	h := &f.Header
	if err := h.Override(*dsName, *theme); err != nil {
		return err
	}
	ds, err := designsystem.Load(h.DesignSystem)
	if err != nil {
		return err
	}
	prelude, err := ds.Prelude(h.Diagram, h.Theme, f.Spec.Layout)
	if err != nil {
		return err
	}
	doc, err := f.Expand(prelude)
	if err != nil {
		return err
	}
	if cmd == "expand" {
		fmt.Print(doc)
		return nil
	}

	if *out == "" {
		*out = strings.TrimSuffix(in, filepath.Ext(in)) + ".svg"
	}
	font, err := ds.Font()
	if err != nil {
		return err
	}
	opts := render.Options{Path: f.Path, IconPath: in, Font: font}
	if f.Spec.Lanes {
		opts.Router = route.Lanes(h.Orientation == "horizontal")
	}
	svg, err := render.SVG(context.Background(), doc, opts)
	if err != nil {
		return err
	}
	decorators, err := ds.Decorators()
	if err != nil {
		return err
	}
	return os.WriteFile(*out, post.Apply(svg, h.Diagram, h.Theme, decorators), 0o644)
}
