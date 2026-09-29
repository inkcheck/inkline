---
title: inkline
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <div class="hx:w-2 hx:h-2 hx:rounded-full hx:bg-primary-400"></div>
  <span>Built on D2</span>
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Technical diagrams from text,&nbsp;<br class="hx:sm:block hx:hidden" />styled by a design system
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  inkline renders D2 text files as SVG diagrams.&nbsp;<br class="hx:sm:block hx:hidden" />It ships with three design systems: Carbon, Cloudscape and Fluent.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6">
{{< hextra/hero-button text="Get started" link="docs/getting-started" >}}
</div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Ten diagram types"
    subtitle="Architecture, flowchart, sequence, state, class, ER, block, swimlanes, quadrant and treeview. Each type has its own shorthand."
    link="docs/diagrams"
  >}}
  {{< hextra/feature-card
    title="Design systems"
    subtitle="A design system sets colour tokens, type, shapes and SVG decorations. Change the design system and inkline restyles every diagram."
    link="docs/design-systems"
  >}}
  {{< hextra/feature-card
    title="Valid D2"
    subtitle="A inkline file is a valid D2 file. Run inkline expand to print the D2 that inkline generates."
    link="docs/how-it-works"
  >}}
  {{< hextra/feature-card
    title="Carbon and the technical diagrams kit"
    subtitle="IBM Plex Sans, 48px nodes, the domain colour key, side bars and colour blocks, in light and dark."
    link="docs/carbon"
  >}}
  {{< hextra/feature-card
    title="Cloudscape and AWS groupings"
    subtitle="AWS's Cloudscape tokens and Open Sans, service tiles, and AWS Cloud, Region, VPC and subnet groupings."
    link="docs/cloudscape"
  >}}
  {{< hextra/feature-card
    title="Fluent 2"
    subtitle="Microsoft's Fluent tokens and Selawik, cards with square service badges, and brand blue for accents."
    link="docs/fluent"
  >}}
  {{< hextra/feature-card
    title="Single binary"
    subtitle="D2 compiles, lays out and renders in the same process. inkline does not need a d2 executable, a browser or Node."
    link="docs/cli"
  >}}
  {{< hextra/feature-card
    title="Semantic roles"
    subtitle="Every node and connection has a role. A design system styles the roles and does not parse the diagram."
    link="docs/roles"
  >}}
{{< /hextra/feature-grid >}}
