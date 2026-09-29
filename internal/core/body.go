package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2format"
	"github.com/d2lang/d2/d2parser"
)

// body is the D2 a transform writes. It records which line of the author's
// file each key came from, so compile errors can name that line, and role
// hints (absolute object ID -> role) for objects whose role follows from the
// shorthand rather than from their final shape.
type body struct {
	text  strings.Builder
	n     int         // lines written
	lines map[int]int // body line -> source line, both zero-based
	hints map[string]string
}

func newBody() *body {
	return &body{lines: map[int]int{}, hints: map[string]string{}}
}

func (b *body) String() string { return b.text.String() }

func (b *body) write(s string) {
	b.text.WriteString(s)
	b.n += strings.Count(s, "\n")
}

// source writes the author's whole file.
func (b *body) source(m *d2ast.Map) {
	text := d2format.Format(m)
	b.mapLines(m.Nodes, text)
	b.write(text)
}

// node writes one node of the author's file.
func (b *body) node(n d2ast.Node) {
	text := d2format.Format(n) + "\n"
	if k, ok := n.(*d2ast.Key); ok {
		b.mapLines([]d2ast.MapNodeBox{{MapKey: k}}, text)
	}
	b.write(text)
}

// printf writes generated D2 that stands for the author's key from.
func (b *body) printf(from *d2ast.Key, format string, args ...any) {
	if from != nil {
		b.lines[b.n] = from.Range.Start.Line
	}
	b.write(fmt.Sprintf(format, args...))
}

// mapLines pairs the keys of the author's nodes with the keys of their
// formatted text, which is about to be written.
func (b *body) mapLines(orig []d2ast.MapNodeBox, text string) {
	formatted, err := d2parser.Parse("", strings.NewReader(text), nil)
	if err != nil {
		return
	}
	var pair func(orig, formatted []d2ast.MapNodeBox)
	pair = func(orig, formatted []d2ast.MapNodeBox) {
		if len(orig) != len(formatted) {
			return
		}
		for i := range orig {
			o, f := orig[i].MapKey, formatted[i].MapKey
			if o == nil || f == nil {
				continue
			}
			b.lines[b.n+f.Range.Start.Line] = o.Range.Start.Line
			if o.Value.Map != nil && f.Value.Map != nil {
				pair(o.Value.Map.Nodes, f.Value.Map.Nodes)
			}
		}
	}
	pair(orig, formatted.Nodes)
}

// sourceLine returns the line of the author's file behind a line of the body:
// that of the nearest key at or above it, plus the distance to it.
func (b *body) sourceLine(line int) (int, bool) {
	if line < 0 || line >= b.n {
		return 0, false
	}
	mapped := make([]int, 0, len(b.lines))
	for l := range b.lines {
		mapped = append(mapped, l)
	}
	sort.Ints(mapped)
	i := sort.SearchInts(mapped, line+1) - 1
	if i < 0 {
		return 0, false
	}
	return b.lines[mapped[i]] + line - mapped[i], true
}

// locate rewrites the positions of compile errors, which refer to a document
// with the body at line offset, as lines of the author's file.
func (b *body) locate(err error, path string, offset int) error {
	var pe *d2parser.ParseError
	if !errors.As(err, &pe) {
		return err
	}
	msgs := make([]string, len(pe.Errors))
	for i, e := range pe.Errors {
		msg := strings.TrimPrefix(e.Message, e.Range.String()+": ")
		line, ok := b.sourceLine(e.Range.Start.Line - offset)
		switch {
		case e.Range.Path != path:
			msgs[i] = e.Message
		case ok:
			msgs[i] = fmt.Sprintf("%s:%d: %s", path, line+1, msg)
		default:
			msgs[i] = fmt.Sprintf("%s: %s (in the expanded D2; see inkline expand)", path, msg)
		}
	}
	return errors.New(strings.Join(msgs, "\n"))
}
