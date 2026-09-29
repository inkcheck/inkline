---
title: Getting started
weight: 1
---

## Install

inkline is a single binary. It includes D2 as a library, so you do not need to
install anything else. The macOS build is signed and notarised.

```sh
brew install --cask inkcheck/tap/inkline
```

Or with Go:

```sh
go install github.com/inkcheck/inkline/cmd/inkline@latest
```

## Write a diagram

Every inkline file declares its diagram type in a `vars` block. The block is
ordinary D2, so the file is valid D2.

```d2 {filename="orders.d2"}
vars: { inkline: { diagram: architecture; design_system: carbon; theme: light } }

user: Customer {class: actor}
shop: Storefront {
  web: Web app {class: application}
  api: Orders API {class: compute}
  db: Orders database {class: [node; data]}
}
user -> shop.web -> shop.api -> shop.db: Query
```

## Render it

```sh
inkline render orders.d2              # writes orders.svg
inkline render orders.d2 --theme dark
inkline expand orders.d2              # the plain D2 inkline generates
```

{{< diagram name="architecture" caption="examples/architecture.d2 in the Carbon design system" >}}

## What the header does

| Key | Values | Meaning |
|---|---|---|
| `diagram` | one of the [ten types](../diagrams) | required; picks the shorthand and the roles |
| `design_system` | `carbon`, `cloudscape`, `fluent`, `plain` | default `plain` |
| `theme` | `light`, `dark` | default `light` |
| `orientation` | `vertical`, `horizontal` | swimlanes only |

## Next

- [Diagram types](../diagrams) shows an example of each type.
- [Roles](../roles) explains how to mark what each element is. The design system sets how each role looks.
