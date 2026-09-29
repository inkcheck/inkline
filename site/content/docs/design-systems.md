---
title: Design systems
weight: 5
---

A design system is a directory of data files. It holds a manifest, D2 style
files, fonts and icons. inkline embeds it in the binary.

```
<name>/
  manifest.json
  tokens-light.d2, tokens-dark.d2   colour vars + d2-config theme overrides
  base.d2                           shared styles, modifiers, marker classes
  diagrams/<type>.d2                role classes for one diagram type
  fonts/, icons/
```

Design systems form a chain through `extends`. Every chain ends at `plain`.
`plain` supplies the structural defaults, such as the D2 shape for each role,
and uses D2's default styles. A role you leave out keeps the parent's styling.

## What inkline assembles

For diagram `T` in theme `light`, from the base design system to the most specific:

1. `tokens[light]`, then `styles`, then `diagrams/T.d2`
2. render settings: layout engine, theme, padding
3. your source
4. the role overlay

## Manifest

```json
{
  "name": "carbon",
  "extends": "plain",
  "tokens": {"light": "tokens-light.d2", "dark": "tokens-dark.d2"},
  "styles": ["base.d2"],
  "layout": {"treeview": "dagre"},
  "pad": 48,
  "fonts": {"regular": "fonts/IBMPlexSans-Regular.ttf", "bold": "fonts/IBMPlexSans-SemiBold.ttf"},
  "decorators": [
    {"type": "bar", "side": "left", "size": 4, "classes": ["lane"], "diagrams": ["swimlanes"]}
  ]
}
```

`name` must match the kit's directory. `layout` overrides the core's layout engine for a diagram type. The value is
`dagre` or `elk`. Both are open source. D2's third engine, TALA, is proprietary
and needs a paid licence, so inkline does not support it.

## Decorators

Decorators draw elements that D2's style model cannot express. D2 writes an
element's classes onto its SVG group. A decorator selects elements by class and
can be limited to certain diagram types. It takes its colour from the element's
stroke.

| Type | Draws |
|---|---|
| `bar` | a band on the `left`, `top` or `bottom` of a node; left bars push the label right |
| `block` | a square at the left of a node with its icon centred in it; the label moves past it |
| `icon` | an icon centred in a node, rectangle or circle |
| `avatar` | a round or square badge at the left of a node, holding its icon; the label moves past it |
| `fill` | fills a node with its stroke colour, for legend swatches |
| `legend` | removes the legend's border, shadow and rounded frame |
| `chevron` | draws filled arrowheads as open chevrons |
| `pill` | a rounded outline around a connection label |
| `strike` | a diagonal line through a node |

inkline clips bars and blocks to the node's outline, so they follow rounded corners.

inkline checks a manifest when it loads it. The `name` must match the kit's
directory, and an unknown decorator `type`, `side` or `shape` is an error.

## D2 behaviour

{{< callout type="warning" >}}
These D2 behaviours affect how the style files are written. The source has a
comment on each one.
{{< /callout >}}

- **A spread takes precedence over a map's own keys.** `{...${carbon.node}; style.fill: red}`
  keeps the spread's fill. Put only keys that no class overrides in shared maps.
  Set fill, stroke and corners in each class.
- **Variables must be declared before the classes that spread them**, so inkline
  hoists every `vars` block above all classes.
- **`sql_table` and `class` shapes** use `fill` for the header and `stroke` for the body.
- **Arrowheads appear only at ends that have an arrow.** The core makes `er`
  and `class` relationships bidirectional, and each role sets both ends.
