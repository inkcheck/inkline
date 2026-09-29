---
title: How it works
weight: 3
---

inkline has three layers. A file that uses them is still valid D2.

1. **Plain D2.** A inkline file can use any D2 feature.
2. **A style library.** The design system's D2 classes, one per [role](../roles).
3. **Preprocessing and post-processing.** The core expands each diagram's
   shorthand and assigns roles; the design system adds tokens, fonts and render
   settings, then adds SVG decorations that D2's style model cannot express.

```
source.d2 ─ core: header, shorthand, roles ─┐
design system: tokens + classes ────────────┴─> D2 ─> SVG ─> decorators ─> out.svg
```

## What the core does

The core handles diagram types. The design system handles styling. The core:

- reads the `vars.inkline` header
- expands shorthand, such as `"[*]"` states, nested trees, quadrant grids and swimlane ranking
- compiles the file once to find every object and connection
- writes a role overlay, such as `a.class: [process; hot]`, with the role first
- reports compile errors against the lines of your file

Run `inkline expand` on any file to see the output of the core.

## What a design system does

A design system handles how diagrams look. The core handles diagram syntax.
It supplies colour tokens, D2 classes for each role, fonts, layout settings and
SVG decorators. See [design systems](../design-systems).

## Why roles

Roles connect the core and the design systems. The core gives every object and
connection exactly one role, as the first entry of its class list. Classes you
add after the role take precedence. For example, `{class: [component; security]}`
is a component in the security colour.

Because of roles, a new diagram type needs no changes to the design systems. A
new design system needs no changes to the core.
