---
name: inkline
description: Write and render technical diagrams with inkline. A inkline diagram is a D2 file with a `vars.inkline` header, drawn in a design system such as Carbon (IBM's technical-diagram style), Cloudscape (AWS's, with AWS Cloud/Region/VPC/subnet groupings) or Fluent (Microsoft's Fluent 2). Use this skill whenever the user wants an architecture, flowchart, sequence, state, class, ER, block, swimlane, quadrant or tree diagram, asks for a diagram "in Carbon", "IBM style", "Cloudscape", "AWS style", "Fluent" or "Microsoft style", draws AWS infrastructure, converts a Mermaid or plain D2 diagram, edits a `.d2` file containing `inkline:`, or mentions inkline at all. Use it even if they only say "draw", "diagram", "visualise the flow" or "sketch the system".
---

# inkline

inkline turns a D2 file into an SVG diagram with a consistent visual style. You
give each object a role that says what it is. The design system sets its shape,
colour and font. A inkline file is ordinary D2 plus a header that names the
diagram type, so all D2 syntax still works.

## Workflow

1. **Pick the diagram type** from what the user describes (table below).
2. **Write the file** with the header, then the content in that type's
   shorthand. Read `references/diagrams.md` for the type you picked. It has the
   syntax, the roles and a complete example for each type.
3. **Use roles and modifiers** to say what things are. Avoid setting colours
   and shapes by hand (see below).
4. **Render it** and fix any errors: `inkline render file.d2 -o file.svg`.
5. **Check the result** if you can view images, and tidy the layout.

## The header

```d2
vars: { inkline: { diagram: flowchart; design_system: carbon; theme: light } }
```

| Key | Values | Notes |
|---|---|---|
| `diagram` | `architecture` `flowchart` `sequence` `state` `class` `er` `block` `swimlanes` `quadrant` `treeview` | required |
| `design_system` | `carbon`, `cloudscape`, `fluent`, `plain` | default `plain`; use `carbon` unless the user wants an AWS look (`cloudscape`), a Microsoft look (`fluent`) or D2's own (`plain`) |
| `theme` | `light`, `dark` | default `light` |
| `orientation` | `vertical`, `horizontal` | swimlanes only; lanes as columns (default) or rows |

## Choosing the type

| The user describes… | Use |
|---|---|
| systems, services, networks, cloud regions, deployments, what talks to what | `architecture` |
| a procedure with steps and decisions | `flowchart` |
| a process owned by several teams or systems, with handoffs | `swimlanes` |
| messages between participants over time, an API call chain | `sequence` |
| a lifecycle: statuses and the transitions between them | `state` |
| classes, interfaces and their relationships | `class` |
| tables, columns and keys | `er` |
| a layered stack or a grid of components | `block` |
| a 2×2 prioritisation (impact/effort, reach/engagement) | `quadrant` |
| a hierarchy or taxonomy | `treeview` |

## Roles

Every object and connection has a **role**, which is its first class. Each
diagram type has its own roles. An object without a role gets the type's
default. For example, a flowchart node is a `process` and an architecture node
is a `component`. Add a role with `class`:

```d2
check: Payment OK? {class: decision}
db: Orders {class: [node; data]}         # role first, then modifiers
```

List the role first. Later classes override it, which is how modifiers work:

- **Modifiers** (all types): `multiple`, `added`, `changed`, `removed`.
- **Domain colours**: `security`, `devops`, `application`, `data`, `storage`,
  `network`, `observability`, `vpc`, `compute`, `backend`. Carbon also has
  `prescribed` for architecture. See `references/carbon.md`.
- **Cloudscape** has the same domain classes. It also has AWS groupings for
  containers (`aws-cloud`, `region`, `availability-zone`, `vpc`,
  `public-subnet`, `private-subnet`, `security-group`, `auto-scaling-group`,
  `data-center`). For any AWS architecture, read `references/cloudscape.md`.
- **Fluent** has the same domain classes. It shows them as square badges on
  architecture services. See `references/fluent.md`.

Avoid setting `style.fill`, `shape` and similar keys directly. The design
system already gives every role its shape, size, colour and font, in light and
dark. Hand-set styles conflict with it and do not change when the theme
changes. Use them only when the user asks for a specific look that no role or
modifier provides.

## Icons

Add an icon to a service with D2's `icon` key and the path to an SVG file:
`cf: CloudFront {class: network; icon: ./icons/cloudfront.svg}`. The kit
places it in Carbon's colour block, Cloudscape's tile or Fluent's badge. inkline
recolours a single-colour icon (Carbon, Cloudscape, Fluent UI System Icons and
most other sets) to sit on the domain colour. It draws a full-colour icon with
its own background (AWS and Azure architecture icons) as it is. inkline bundles no
service icons, so the user must supply them. AWS and Azure icons are free to
download, but they are not openly licensed.

## Rendering and checking

```sh
inkline render diagram.d2 -o diagram.svg          # flags and file in any order
inkline render diagram.d2 --theme dark -o diagram-dark.svg
inkline expand diagram.d2                          # the plain D2 inkline generates
inkline list                                       # diagram types and design systems
inkline help                                       # the header format, in brief
```

- **Errors** name the line in your file, such as `flow.d2:12: unknown shape`.
  An error that starts with `expanded D2:` refers to a line of the text inkline
  generated. Run `inkline expand` to see that text.
- **`-o` belongs to `render`.** `inkline expand` prints to standard output and
  rejects `-o`.
- **Output** is a self-contained SVG. Fonts are embedded and icons are inlined.
- **If `inkline` is not on the PATH**, install it with
  `brew install --cask inkcheck/tap/inkline` or
  `go install github.com/inkcheck/inkline/cmd/inkline@latest`. Inside the inkline
  repository, `make build` produces `bin/inkline`.

## Pitfalls

- **`$` starts a D2 variable.** Put single quotes around any label that
  contains one: `check: 'Over $500?'`. Otherwise D2 fails with "substitutions
  must begin on {".
- **Quote pseudo-states** in state diagrams: `"[*]" -> idle`. D2 rejects an
  unquoted `[*]` at the end of an edge.
- **Treeview uses nesting only.** Write children inside their parent's map. Do
  not use `->` inside the tree. D2 keywords (`class`, `style`, `icon`) stay
  attributes.
- **Quadrant keys are fixed**: `q1` top right, `q2` top left, `q3` bottom
  left, `q4` bottom right, plus optional `x`, `y` and `title`. Items go inside
  a quadrant. There are no coordinates.
- **Swimlanes**: every step sits inside a lane (a top-level container). inkline
  ranks the steps along the flow and gives each rank its own row. Two steps in
  one lane never share a row, so parallel work in a lane is staggered.
- **Class and ER relationships** start at the child, implementer, whole or
  dependent end: `child -> parent {class: inheritance}`,
  `customer.id -> order.customer_id {class: one-to-many}`.
- **Sequence** messages appear in file order. Write them in the order they
  happen. A group is a top-level container that holds messages. It takes no
  role, so do not give it `participant` or `actor`.
- **Imports** (`...@shared`, `...@../lib/actors`) resolve relative to the file.
  Set the `class` of a connection in the file that uses it. A connection that
  sets `class` inside an imported file does not get a role.
- **The `vf_` prefix is reserved** for objects inkline generates, such as
  `vf_start`, `vf_end` and `vf_chart`. Do not use it for your own keys.
- **Outside labels take no space in the layout.** `label.near: outside-*`
  can overlap the shapes next to it. Stagger them with an invisible point (see
  "Compact decisions" in `references/diagrams.md`).
- **The layout engine is fixed.** inkline ignores `d2-config.layout-engine`.
- **Grids stretch nodes to their row and column**, and draw straight arrows
  through other shapes. Avoid grids for flowcharts with branches.
- **Keep IDs short** and put the readable text in the label:
  `lb: Load balancer`. Connections refer to IDs, so short IDs keep
  `cloud.vpc.lb -> db` easy to read.

## Converting from Mermaid

Map the Mermaid keyword to the inkline type (`flowchart`/`graph` → `flowchart`,
`sequenceDiagram` → `sequence`, `stateDiagram-v2` → `state`, `classDiagram` →
`class`, `erDiagram` → `er`, `block-beta` → `block`, `architecture-beta` →
`architecture`, `quadrantChart` → `quadrant`). Then rewrite nodes as D2 keys
with labels, and turn Mermaid node shapes into roles: `{}` becomes `decision`,
`([ ])` becomes `start` or `end`, `[( )]` becomes `store`. inkline has no pie,
Gantt, journey, git-graph, mindmap or chart types. If the user asks for one,
tell them inkline does not support it. Do not approximate it with another type.
