# The Fluent design system

`design_system: fluent` adapts Microsoft's [Fluent 2](https://fluent2.microsoft.design/)
for diagrams. It uses Selawik (Fluent's open substitute for Segoe UI), Fluent
colour tokens (light and dark), 4px cards, 8px groups and
neutral Fluent avatars for people. Brand blue is kept for accents: the
`system` in an architecture diagram, decisions, choices and quadrant items.
Use it when the user wants a Microsoft, Fluent, Office or Windows look.

## Architecture: cards with badges

Services (`component`, the default, and `node`) are white cards with a square
badge on the left, filled with the domain colour. Give every service a domain
class. A service without one shows a grey badge. An icon set with D2's `icon`
key goes in the badge. Single-colour icons are drawn white on the badge.
Full-colour icons (Azure, AWS) replace the badge.

```d2
vars: { inkline: { diagram: architecture; design_system: fluent } }

user: Customer {class: actor}
shop: Online shop {class: system}
cloud: Cloud {
  class: location
  api: Orders API {class: application}
  fn: Workers {class: [compute; multiple]}
  db: Orders DB {class: data}
  vault: Key vault {class: security}
}
user -> shop -> cloud.api
cloud.api -> cloud.fn: queue {class: async}
cloud.fn -> cloud.db: SQL
cloud.fn -> cloud.vault
```

Containers (`location`, `zone`, `subsystem`) are Fluent's subtle surfaces,
with the label at the top left. There are no Azure-specific groupings. Use
`location` for subscriptions, resource groups and virtual networks. Add the
`vpc` class to a virtual network to make it stand out.

## Domain colours

The classes are the same as in the other kits. The colours come from Fluent's
shared palette: `compute` dark orange, `storage` green, `data` purple,
`network` teal, `security` red, `application` berry, `observability` steel,
`devops` grape, `vpc` royal blue, `backend` grey.

## Other diagram types

The other diagram types use the shared roles unchanged. Steps, states and
participants are Fluent cards. Flowchart and swimlane start and end steps are
green success pills. Decisions are brand-blue diamonds. This kit has no syntax
of its own, so a diagram written for Carbon renders in Fluent when you change
`design_system`.
