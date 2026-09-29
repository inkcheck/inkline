---
title: Fluent
weight: 8
---

The `fluent` design system applies Microsoft's
[Fluent 2](https://fluent2.microsoft.design/) to diagrams.

```d2
vars: { inkline: { diagram: architecture; design_system: fluent } }
```

{{< diagram name="architecture-fluent" caption="examples/architecture.d2 in the Fluent kit" >}}

Fluent is a design system for application interfaces. It has no diagram
guidance of its own, so this kit is inkline's translation of it. inkline generates the
kit's colours from Fluent's published web themes, so they match Fluent in light
and dark. Brand blue is used for accents. These are the system being drawn,
decisions and highlighted items.

| Aspect | Setting |
|---|---|
| Type | Selawik, 14px, semibold labels |
| Canvas | Fluent's neutral background 2, with cards on background 1 |
| Shapes | 4px cards and 8px groups, Fluent's medium and extra-large corner radii |
| Borders | Fluent's neutral stroke 2 |
| Lines | Fluent's neutral foreground 3, with filled arrowheads |
| Colours | Fluent tokens for canvas, surfaces, text and status; Fluent's shared palette for domains |
| Services | Cards with a square badge in the domain colour |
| People | Fluent's neutral avatar, with its `person` icon |
| Start and end | Success-coloured pills |

{{< diagram name="flowchart-fluent" caption="examples/flowchart.d2 in the Fluent kit" >}}

## Domain colours

Fluent uses the same classes as Carbon and Cloudscape, so a diagram renders in
any of these kits without changes. On architecture services the colours fill the
badge. On other elements they set the outline.

| Class | Fluent palette colour |
|---|---|
| `compute` | dark orange |
| `storage` | green |
| `data` | purple |
| `network` | teal |
| `security` | red |
| `application` | berry |
| `observability` | steel |
| `devops` | grape |
| `vpc` | royal blue |
| `backend` | neutral grey |

## Icons

An icon set with D2's `icon` key goes in the badge. inkline draws a single-colour
icon in white on the domain colour. A full-colour icon with its own background,
such as Azure's or AWS's architecture icons, replaces the badge. The
[Fluent UI System Icons](https://github.com/microsoft/fluentui-system-icons)
are MIT licensed and work with this kit.

```d2
api: Orders API {class: application; icon: ./icons/ic_fluent_server_24_regular.svg}
```

## Type

Fluent's typeface, Segoe UI, is not licensed for redistribution. The kit uses
Selawik instead. Selawik is Microsoft's open-source substitute for Segoe UI and
has matching metrics.

## Licences

Fluent's design tokens and its person icon are MIT licensed (Copyright
Microsoft Corporation). Selawik is under the SIL Open Font License 1.1. The
licence texts are in the repository's `THIRD_PARTY_NOTICES.md`. inkline is not
affiliated with or endorsed by Microsoft.
