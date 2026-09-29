# The Carbon design system

`design_system: carbon` follows IBM's *Technical diagrams kit* with
[Carbon](https://carbondesignsystem.com/) colours. It uses IBM Plex Sans, 48px
nodes, Cool Gray neutrals and square corners (8px for logical architecture
elements), with light and dark themes. Use it unless the user asks for plain
D2 styling.

## Domain colours

Add a domain class after the role. It sets the element's primary colour. This
colours the outline and the side bar or colour block that the design system
draws.

| Class | Colour | Use for |
|---|---|---|
| `security` | Red 50 | firewalls, identity, secrets, WAF |
| `devops` | Magenta 50 | CI/CD, pipelines, registries |
| `application` | Purple 50 | apps, services, APIs, frontends |
| `data` | Blue 60 | databases, caches, queues holding data |
| `storage` | Blue 60 | object and block storage (same colour as data in the kit's key) |
| `network` | Cyan 50 | load balancers, gateways, DNS, the internet |
| `observability` | Teal 50 | logging, metrics, tracing |
| `vpc` | Teal 70 | VPCs and private networks (usually on a container) |
| `compute` | Green 60 | VMs, containers, clusters, functions |
| `backend` | Cool Gray 50 | back-office and legacy systems |

The kit asks for a legend in every diagram, even when it uses this colour key.
Add one with D2's `vars.d2-legend` whenever colours carry meaning. Give each
entry the `swatch` class and its domain class. This draws a solid swatch in
the domain's colour:

```d2
vars: {
  inkline: { diagram: architecture; design_system: carbon }
  d2-legend: {
    app: Application {class: [swatch; application]}
    sec: Security {class: [swatch; security]}
    dat: Data {class: [swatch; data]}
  }
}
web: Web app {class: application}
fw: Firewall {class: [component; security]}
db: Orders {class: [node; data]}
web -> fw -> db
```

## Modifiers

| Class | Draws | Meaning |
|---|---|---|
| `multiple` | offset copies behind the node | several instances |
| `added` | a double outline | new in this change |
| `changed` | a dashed outline | modified, passive or optional |
| `removed` | a strike through the node | removed or disabled |
| `prescribed` | square corners (architecture) | a prescribed element (logical elements have 8px corners) |

## Architecture roles

- `system` is a dark pill. Use it for the target system of the diagram, or for
  an external system (a SaaS such as Stripe) shown as a single box.
- `actor` is a person or external client, drawn as a circle with the user icon.
- `node` is infrastructure (a VM, a database server). `component` is software.

## What the design system draws

- Colour blocks on architecture components and nodes, with the node's icon
  centred in them.
- Side bars on locations, lanes, states and sequence participants.
- Top bars on flowchart and swimlane steps. Start and end steps are green.
- Pill-shaped labels on flowchart, swimlane and architecture connections.
- The Carbon user icon in actors.

## Icons

Use D2's `icon` key with a path relative to the diagram, or a URL. Carbon icons
are 24px. On components and nodes the icon goes in the colour block. inkline
redraws a single-colour icon in the canvas colour: white in the light theme and
dark in the dark theme. A full-colour icon is drawn as it is.

```d2
lb: Load balancer {class: [component; network]; icon: ./icons/load-balancer.svg}
```
