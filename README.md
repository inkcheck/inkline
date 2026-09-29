# INKLINE

inkline renders technical diagrams from text. It is built on
[D2](https://d2lang.com), and it styles each diagram with a design system that
you can swap out. A inkline file is a D2 file with a header that names its
diagram type.

inkline ships with three design systems:

- Carbon follows IBM's Technical diagrams kit.
- Cloudscape is AWS's design system, with AWS architecture groupings.
- Fluent is Microsoft's Fluent 2.

```d2
vars: { inkline: { diagram: architecture; design_system: carbon; theme: light } }

user: User {class: actor}
cloud: IBM Cloud {
  vpc: VPC A {
    class: vpc
    lb: Load balancer {class: network}
    vsi: VSI 1 {class: [compute; multiple]}
  }
}
user -> cloud.vpc.lb -> cloud.vpc.vsi
```

```sh
brew install --cask inkcheck/tap/inkline                   # macOS and Linux
go install github.com/inkcheck/inkline/cmd/inkline@latest    # or from source

make                                                   # list the make targets
make build                                             # bin/inkline; D2 is a Go library, no d2 CLI needed
bin/inkline render examples/architecture.d2              # -> examples/architecture.svg
bin/inkline render --theme dark -o out.svg examples/state.d2
bin/inkline expand examples/state.d2                     # the plain D2 inkline generates
bin/inkline list
```

## Diagram types

`architecture`, `block`, `class`, `er`, `flowchart`, `quadrant`, `sequence`,
`state`, `swimlanes`, `treeview`. The `examples/` directory has one of each.
[spec/roles.md](spec/roles.md) describes the header, and the shorthand and
roles for each type.

## How it works

inkline works in three layers. The source stays valid D2 in each of them.

1. **Plain D2.** A inkline file can use anything D2 supports.
2. **Style library.** Each design system defines D2 classes, one per semantic
   role.
3. **Preprocessing and post-processing.** The *core* expands each diagram's
   shorthand (`"[*]"` states, tree nesting, quadrants, lane ranking). It also
   gives every object and connection a role. The *design system* adds its
   tokens, fonts and render settings. After rendering, it decorates the SVG
   with shapes that D2's styles cannot draw (Carbon side bars, colour blocks,
   pill labels). D2 compiles, lays out and renders the diagram in-process
   through its Go packages.

```
source.d2 ─ core: header, shorthand, roles ─┐
design system: tokens + classes ────────────┴─> d2 ─> SVG ─> design system decorators ─> out.svg
```

Roles are the interface between the core and the design system. The core
handles diagram types and their grammar, and has no styling. A design system
styles the roles, and has no knowledge of diagram grammar. To write a design
system, see [kits/README.md](kits/README.md).

## Docs site

`site/` is a Hugo site that uses the Hextra theme. It is styled with the
Carbon Design System, using IBM Plex, Carbon colour tokens, square corners and
Carbon focus rings. It embeds every example, rendered in both themes.

```sh
make serve     # render the examples, then serve on http://localhost:1414/inkline/
make site      # build into site/public
```

The site is published at <https://inkcheck.github.io/inkline/>. The workflow in
`.github/workflows/pages.yml` builds and deploys it on every push to `main`.
To turn it on, set **Settings > Pages > Source** to **GitHub Actions** in the
repository.

## Layout

```
cmd/inkline/           CLI
internal/core/         header, diagram specs, shorthand transforms, lane ranking, role overlay
internal/designsystem/ design-system loading (chain, prelude, fonts)
internal/ids/          IDs of the objects inkline generates (the vf_ prefix)
internal/post/         SVG decorators
internal/render/       in-process compile, layout and SVG render via D2's packages
internal/route/        right-angle routing for connections between swimlanes
kits/                  design systems: plain (structural defaults), carbon, cloudscape, fluent
spec/roles.md          roles per diagram type: the core/design-system contract
examples/              one diagram per type
scripts/               render site diagrams, notices, releases
site/                  Hugo + Hextra docs site, themed with Carbon
skills/inkline/           Agent skill: teaches Claude to write and render inkline diagrams
context/               local reference material, e.g. technical design systems (gitignored)
```

## Known limits

- inkline outputs SVG only. It does not produce PNG or PDF.
- D2 is pinned in `go.mod` (v0.9.0). The dagre and ELK layouts both run
  in-process.
- Layers, scenarios and steps (D2 boards) do not get roles yet.
- A connection that sets its own `class` inside an imported file does not get
  a role. Set the class in the importing file.
- Swimlane steps align by rank. A lane cannot hold two steps at the same rank,
  so parallel work in one lane is staggered.
- ELK decides positions, so inkline does not snap them to the kit's 8px grid.

## Agent skill

`skills/inkline/` is an [Agent Skill](https://docs.claude.com/en/docs/agents-and-tools/agent-skills).
It teaches Claude to pick a diagram type, write the inkline file with the right
roles and render it. To use it in Claude Code, copy or symlink it into
`~/.claude/skills/` for all projects, or into a project's `.claude/skills/`:

```sh
ln -s "$PWD/skills/inkline" ~/.claude/skills/inkline
```

## Licences

inkline is licensed under the [Apache License 2.0](LICENSE). See also
[NOTICE](NOTICE).

inkline redistributes these third-party components:

- IBM Plex Sans, Open Sans and Selawik (SIL Open Font License 1.1).
- Carbon and Cloudscape icons and design tokens (Apache License 2.0).
- Fluent UI icons and design tokens (MIT License).
- D2 (Mozilla Public License 2.0) and its dependencies.

[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) holds every licence text.
Run `make notices` to regenerate it after dependency changes.

inkline is not affiliated with or endorsed by IBM, Amazon or Microsoft. Their
product names are used only to identify the typefaces and design systems that
the `carbon`, `cloudscape` and `fluent` design systems follow.
