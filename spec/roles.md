# inkline semantic roles

Roles are the interface between the **core** and a **design system**. The core
handles diagram types. The design system handles styling. The core ensures that
every object and connection in a diagram carries exactly one role, as the
*first* entry of its D2 `class` list. A design system styles roles by defining D2
classes with the same names.

D2 applies classes in order, and later classes take precedence. Anything the
author adds after the role (domain colours, states) overrides the role's styling.

## The inkline header

Every inkline file declares its diagram type in a header. The header is plain D2
`vars`, so the file stays valid D2:

```d2
vars: {
  inkline: {
    diagram: architecture   # required
    design_system: carbon      # optional, default: plain
    theme: light            # optional, light | dark, default: light
    orientation: vertical   # swimlanes only: vertical | horizontal
  }
}
```

A file can have more than one `vars` block. The header can be in any of them.

Objects the core generates have IDs that start with `vf_`: `vf_start`,
`vf_end`, `vf_chart`, `vf_x`, `vf_y` and the `vf_gap_` lane cells. Authors
should not use the prefix for their own keys.

## Assigning roles

- To pick a role, list it in `class`, e.g. `db: Orders {class: store}`.
- Anything without a role gets its diagram's default role (marked * below).
- inkline resolves roles per diagram type. For example, `decision` is a role in
  `flowchart`. It has no meaning in `er`.

## Modifiers (all diagrams)

Modifiers are separate from roles. List a modifier after the role, and it adds
its styling on top of the role's.

| Modifier   | Meaning                                  |
|------------|------------------------------------------|
| `multiple` | Several instances of the element         |
| `added`    | Element is new in this change            |
| `changed`  | Element is modified or passive           |
| `removed`  | Element is removed or disabled           |

Design systems can add their own modifiers. The `carbon` design system adds
domain colours (`security`, `devops`, `application`, `data`, `storage`,
`network`, `observability`, `vpc`, `compute`, `backend`) and the architecture
modifier `prescribed`.

The `cloudscape` design system has the same domain classes, with `vpc` as a
grouping. It also adds AWS groupings for containers: `aws-cloud`, `region`,
`availability-zone`, `vpc`, `public-subnet`, `private-subnet`,
`security-group`, `auto-scaling-group`, `data-center`.

The `fluent` design system has the same domain classes, with colours from
Fluent's colour palette.

## Roles by diagram type

`*` marks the default role.

### flowchart
| Kind      | Roles |
|-----------|-------|
| node      | `process`*, `start`, `end`, `decision`, `io`, `subprocess`, `store`, `note` |
| container | `group`* |
| edge      | `flow`*, `optional` |

### sequence
The core sets `shape: sequence_diagram` on the diagram. Roles apply to top-level
participants and to messages. Spans, groups and notes keep D2's own semantics.
A group, which is a top-level container that holds messages and exchanges none,
takes no role.

| Kind        | Roles |
|-------------|-------|
| participant | `participant`*, `actor` |
| edge        | `message`*, `reply`, `async` |

### state
Write `"[*]"` (quoted) for the initial and final pseudo-state, as in Mermaid:
`"[*]" -> idle` and `done -> "[*]"`. Each scope gets its own start and end.

| Kind      | Roles |
|-----------|-------|
| node      | `state`*, `start`, `end`, `choice`, `fork`, `join`, `note` |
| container | `composite`* |
| edge      | `transition`* |

### class
Roles apply to top-level objects only. Their keys are fields and methods. The
core adds a `<<stereotype>>` line to the label of each `interface`, `abstract` and
`enum`.

| Kind   | Roles |
|--------|-------|
| class  | `class`*, `interface`, `abstract`, `enum` |
| edge   | `association`*, `inheritance`, `realization`, `composition`, `aggregation`, `dependency` |

### er
Roles apply to top-level objects only. Their keys are columns.

| Kind   | Roles |
|--------|-------|
| entity | `entity`* |
| edge   | `relation`*, `one-to-one`, `one-to-many`, `many-to-one`, `many-to-many` |

### block
Use D2's `grid-columns` / `grid-rows` for layout.

| Kind      | Roles |
|-----------|-------|
| node      | `block`* |
| container | `group`* |
| edge      | `link`* |

### architecture
The role names follow IBM's Unified Method Framework (UMF). UMF is general enough
to use with any design system. Elements are logical by default. Add the
`prescribed` modifier to mark an element as prescribed.

| Kind      | Roles |
|-----------|-------|
| node      | `component`*, `actor`, `system`, `node`, `junction` |
| container | `location`*, `zone`, `subsystem` |
| edge      | `connection`*, `async` |

### swimlanes
Top-level containers are lanes. They are laid out in order as columns (default)
or as rows (`vars.inkline.orientation: horizontal`).

The core ranks the steps along the flow, ignoring loops. Each step sits one rank
after the step before it, and steps at the same rank line up across lanes. Every
step in a lane needs a rank of its own. In horizontal lanes the core gives every
step the same width.

Connections between lanes are routed at right angles through the gaps between
ranks and lanes (`internal/route`). Connections within a lane are straight,
because a lane stretches its steps to a common width.

| Kind | Roles |
|------|-------|
| lane | `lane`* |
| node | `step`*, `start`, `end`, `decision` |
| edge | `flow`*, `optional` |

### quadrant
A quadrant chart is categorical. Each item sits *in* a quadrant and has no
coordinates. The reserved top-level keys are `q1` (top right), `q2` (top left),
`q3` (bottom left) and `q4` (bottom right). The optional keys are `x` and `y`
(axis labels) and `title`. The core wraps the quadrants in a 2×2 grid called
`vf_chart`.

| Kind     | Roles |
|----------|-------|
| chart    | `frame` |
| quadrant | `quadrant` |
| node     | `item`* |
| axis     | `axis` |

### treeview
A treeview is a dendrogram. Write the tree as nested maps. The core flattens it
into nodes joined by links. D2 attribute keys (`class`, `style`, `icon`, …) inside
a node stay attributes of that node, and the core does not treat them as
children.

| Kind | Roles |
|------|-------|
| node | `root` (top level), `branch` (has children), `leaf` (no children) |
| edge | `link`* |
