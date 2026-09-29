# Kits

Each directory here is a *kit*. A kit packages one design system as a
`manifest.json`, D2 style files and assets. The inkline binary embeds the kits,
and `embed.go` lists them.

| Kit | Based on |
|---|---|
| `plain` | D2's default styling. It holds the structural defaults that every kit extends. |
| `carbon` | IBM's Carbon Design System and Technical diagrams kit |
| `cloudscape` | AWS's Cloudscape Design System, plus AWS architecture groupings. Its colour tokens come from `@cloudscape-design/design-tokens`. |
| `fluent` | Microsoft's Fluent 2, with Selawik in place of Segoe UI. Its colour tokens come from `@fluentui/react-theme`. |

```
<name>/
  manifest.json
  tokens-light.d2, tokens-dark.d2   colour vars + d2-config theme overrides
  base.d2                           shared styles, modifiers, marker classes
  diagrams/<type>.d2                role classes for one diagram type
  fonts/, icons/
```

## What inkline builds for a diagram

For a diagram of type `T` in theme `light`, inkline builds a prelude. It takes the
design systems in order, from the base design system (`plain`) to the most
specific, and adds:

1. `tokens[light]`, then `styles`, then `diagrams/T.d2` of each design system
2. render settings: layout engine, theme, padding

inkline moves every top-level `vars` block from those files above all classes.
This is needed because D2 cannot resolve a spread (`...${carbon.node}`) that
references other vars when the spread is merged into a class defined before
those vars.

The author's source comes next, followed by the role overlay. The overlay gives
every object and connection its class list: `[role, author classes…]`. Later
classes take precedence, so author modifiers and domain classes override the
role.

## Writing role classes

Define one class per role in `diagrams/T.d2`. The roles are listed in
[../spec/roles.md](../spec/roles.md). Classes with the same name merge, so a
role you leave out keeps the styling from its parent design system.

These D2 behaviours affect how you write the files:

- **A spread takes precedence over the map's own keys.** `{...${carbon.node}; style.fill: red}`
  keeps the spread's fill. Limit shared maps to keys that no class overrides
  (sizes, stroke width, type), and set fill, stroke and corners in each class.
- **`sql_table` and `class` shapes:** `fill` applies to the header and `stroke`
  to the body.
- **D2 draws an arrowhead only at an end that has an arrow.** The core turns
  `er` and `class` relationships into `<->`, so their roles set both ends. Use
  `none` for an end that has no arrowhead.

## Manifest

```json
{
  "name": "carbon",
  "extends": "plain",
  "tokens": {"light": "tokens-light.d2", "dark": "tokens-dark.d2"},
  "styles": ["base.d2"],
  "layout": {"treeview": "dagre"},
  "pad": 48,
  "fonts": {"regular": "fonts/…ttf", "italic": "…", "bold": "…", "semibold": "…"},
  "decorators": [
    {"type": "bar", "side": "left", "size": 4, "classes": ["lane"], "diagrams": ["swimlanes"]}
  ]
}
```

`name` must match the kit's directory. `extends` defaults to `plain`. `layout` overrides the layout engine that the
core sets for each diagram type. The value is `dagre` or `elk`, and both are
open source. D2 has a third engine, TALA. It is proprietary and needs a paid
licence, so inkline does not support it.

Fonts are .ttf files that inkline loads into a D2 font family. D2 draws node
labels in its bold style, so Carbon maps `bold` to Plex SemiBold.

## Decorators

Decorators draw shapes that D2's style model cannot express. D2 writes an
object's classes onto its SVG `<g>`. A decorator selects elements by class,
and optionally only in some `diagrams`. It takes its colour from the element's
stroke.

| type     | draws                                                      |
|----------|------------------------------------------------------------|
| `bar`    | a band on the `left` (default), `top` or `bottom` of a rectangular node; a left bar moves the label to the right |
| `block`  | a `size`-wide square at the left of a node, with its icon centred in it; the label moves past it |
| `icon`   | the decorator's `icon`, centred in a node, rectangle or circle |
| `avatar` | a `size`-wide badge at the left of a node, holding its icon; a `circle` (default) or a `square`; the label moves past it |
| `fill`   | fills a node with its stroke colour (Carbon's legend `swatch`) |
| `legend` | takes no classes; removes the legend's border, shadow and rounded frame |
| `chevron` | takes no classes; draws filled arrowheads as open chevrons and keeps hollow triangles, diamonds and crow's feet |
| `pill`   | a rounded outline around a connection label                |
| `strike` | a diagonal line through a node                             |

inkline checks every decorator when it loads a manifest. An unknown `type`, a
`side` other than `left`, `top` or `bottom`, or a `shape` other than `circle` or
`square` is an error.

An avatar's `fill` and a decorator's `icon_color` take a colour, or one per
theme: `{"light": "#e6e6e6", "dark": "#333333"}`. A `fill` of `"stroke"` uses
the node's stroke colour. `outline` gives the node a neutral border, by theme,
after a block or avatar has taken its domain colour. `icon_stroke` draws the
icon as an outline with strokes of that width.

D2 draws the legend frame and text in fixed light colours. inkline recolours them
to match the theme, for every design system.

inkline clips bars and blocks to the node's outline, so they stay inside rounded
corners.

`block`, `icon` and `avatar` take an `icon`, which is an SVG file in the design system.
inkline inlines it as vector paths and fills it with the canvas colour, so the
same file shows up on the solid shape in both themes. In a block, a node's own
D2 `icon` takes precedence.
