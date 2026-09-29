---
title: Carbon
weight: 6
---

The `carbon` design system follows IBM's *Technical diagrams kit* (v2), which is
part of the IBM Design Language. Its colours come from the
[Carbon Design System](https://carbondesignsystem.com/). This site uses the
same design language.

```d2
vars: { inkline: { diagram: architecture; design_system: carbon; theme: dark } }
```

## What it sets

| Aspect | Setting |
|---|---|
| Type | IBM Plex Sans; 14/18 semibold node labels, 12/16 connection labels |
| Nodes | 48px tall, 1px outline, square, or 8px corners for logical elements |
| Light | White canvas, Cool Gray 10 surfaces, Cool Gray 70 outlines |
| Dark | Cool Gray 100 canvas, Cool Gray 90 surfaces, Cool Gray 50 outlines |
| Layout | ELK, so connections are orthogonal; dagre for treeview's curves |

IBM Plex Sans ships with inkline under the SIL Open Font License. D2 draws node
labels in its bold style, so the design system maps bold to Plex SemiBold.

## The domain colour key

Add a domain class after the role. It sets the primary colour. The outline,
the side bar and the colour block use this colour.

| Class | Carbon colour |
|---|---|
| `security` | Red 50 |
| `devops` | Magenta 50 |
| `application` | Purple 50 |
| `data` | Blue 60 |
| `storage` | Blue 60 |
| `network` | Cyan 50 |
| `observability` | Teal 50 |
| `vpc` | Teal 70 |
| `compute` | Green 60 |
| `backend` | Cool Gray 50 |

```d2
api: Orders API {class: [component; compute]}
fw: Edge firewall {class: [component; security]}
```

{{< callout type="info" >}}
The kit asks for a legend in every diagram, even when the diagram uses the
recommended colour key. Add one with D2's `vars.d2-legend`. Give each entry the
`swatch` class and its domain. For example, `sec: Security {class: [swatch; security]}`
draws a solid Red 50 swatch.
{{< /callout >}}

## Decorations

D2 styles cannot express these parts of the kit. inkline draws them onto the
rendered SVG.

- **Side bars** on locations, lanes, states and participants.
- **Top bars** on flowchart and swimlane steps.
- **Colour blocks** on architecture components and nodes, with the icon centred in them.
- **Actors** have the Carbon user icon. It sits in a colour block in sequence
  diagrams and inside the circle in architecture diagrams.
- **Pill labels** on connections.
- **Strike-through** for `removed`.

## Icons

Put 24px Carbon icons in `kits/carbon/icons/` and reference them with
D2's `icon` key. On components and nodes, the colour-block decorator centres the
icon in the block.

```d2
lb: Load balancer {class: [component; network]; icon: ./icons/load-balancer.svg}
```

inkline redraws a single-colour icon in the canvas colour: white in the light
theme and dark in the dark theme. A full-colour icon is drawn as it is.

## Licences

IBM Plex Sans ships under the SIL Open Font License 1.1. The Carbon user icon
ships under the Apache License 2.0. Both licence texts are in the repository's
`THIRD_PARTY_NOTICES.md`. inkline itself is licensed under the Apache License
2.0. inkline is not affiliated with or endorsed by IBM.
