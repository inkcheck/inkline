---
title: CLI
weight: 9
---

```sh
inkline render [flags] <file.d2>   # render to SVG
inkline expand [flags] <file.d2>   # print the plain D2 inkline generates
inkline list                       # list diagram types and design systems
inkline help                       # the header format and the flags
inkline version                    # print the version
```

Flags and the filename can come in any order.

| Flag | Meaning |
|---|---|
| `-o <path>` | `render` only: output file; defaults to the input with an `.svg` extension |
| `--design-system <name>` | override the file's header |
| `--theme <light\|dark>` | override the file's header |
| `-h` | print the usage for the command |

inkline checks a flag's value the same way it checks the header. An unknown theme
or design system is an error.

## Examples

```sh
inkline render examples/architecture.d2
inkline render examples/state.d2 --theme dark -o state-dark.svg
inkline render examples/flowchart.d2 --design-system plain
inkline expand examples/swimlanes.d2 | less
```

Use `inkline expand` to debug a diagram. It prints the design system's prelude,
your source after shorthand expansion, and the role overlay. It prints them in
the order D2 reads them.

## Errors

An error in your file names the file and the line, such as
`orders.d2:12: unknown shape "cilinder"`. The line is the one in your file,
even when inkline has rewritten the source.

An error that starts with `expanded D2:` comes from the text inkline generated. Its
line number refers to the output of `inkline expand`.

## Imports

D2 imports work as they do in D2. A path is relative to the importing file and
can point to a parent directory, such as `...@../shared/actors`. Give roles to
imported connections in the importing file. A connection that sets its own
`class` in an imported file keeps that class and does not get a role.

## Output

inkline writes SVG only. Each diagram is self-contained. inkline measures fonts at
render time and inlines icons as data URIs, so a rendered file needs no other
files.

D2 compiles, lays out and renders in the same process through its Go packages.
You do not need to install a `d2` executable or match its version.
