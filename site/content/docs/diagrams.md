---
title: Diagram types
weight: 2
---

inkline supports ten diagram types. Each type has its own shorthand and its own set
of [roles](../roles). Each example below is a file in `examples/`, rendered by inkline
in the Carbon design system. Switch the site between light and dark to see both
themes.

## architecture

Systems, their parts and the connections between them. The terms follow
IBM's Unified Method Framework. These are actors, components, nodes, locations
and zones.

{{< diagram name="architecture" caption="architecture: colour blocks show the domain colour, zones are dotted and VSI 2 is a multiple" >}}

```d2
vars: { inkline: { diagram: architecture; design_system: carbon } }

user: User {class: actor}
cloud: IBM Cloud {
  vpc: VPC A {
    class: vpc
    lb: Load balancer {class: network}
    vsi: VSI 1 {class: [compute; multiple]}
  }
}
user -> cloud.vpc.lb -> cloud.vpc.vsi
```

## flowchart

Steps, decisions and the paths between them.

{{< diagram name="flowchart" caption="flowchart: process nodes have a top bar, decisions are small dark diamonds and edge labels are pills" >}}

## sequence

Messages between participants over time. inkline sets `shape: sequence_diagram`.

{{< diagram name="sequence" caption="sequence: actors, messages, replies and an async message" >}}

## state

States and transitions. Write `"[*]"` for the initial and final state, as in
Mermaid. Each scope has its own initial and final state.

{{< diagram name="state" caption="state: pseudo-states are drawn as bullets and composite states are nested" >}}

```d2
vars: { inkline: { diagram: state; design_system: carbon } }

"[*]" -> pending: create
pending -> stable: propagate
cleanup: Cleanup {
  "[*]" -> deleting
  deleting -> deleted: propagate
  deleted -> "[*]"
}
stable -> cleanup: delete
```

## class

Classes, interfaces and UML relationships. Keys inside a class are its fields
and methods. `interface`, `abstract` and `enum` add a stereotype to the label.

{{< diagram name="class" caption="class: realization, composition and dependency" >}}

## er

Entities and relationships. Keys inside an entity are its columns. The
cardinality roles draw crow's feet at both ends.

{{< diagram name="er" caption="er: one-to-many relationships between three tables" >}}

```d2
vars: { inkline: { diagram: er; design_system: carbon } }

customer: {
  id: int {constraint: primary_key}
  email: string {constraint: unique}
}
order: {
  id: int {constraint: primary_key}
  customer_id: int {constraint: foreign_key}
}
customer.id -> order.customer_id: places {class: one-to-many}
```

## block

Blocks in a grid, for layer diagrams and component maps. Use D2's `grid-columns`
and `grid-rows`.

{{< diagram name="block" caption="block: a three-column grid with a nested group" >}}

## swimlanes

A process across lanes. inkline ranks the steps along the flow. A handoff always
moves forward, and steps at the same stage line up across lanes. Connections run
at right angles through the gaps between stages and lanes, so they do not cross
a step. A handoff turns shortly before its target. A loop back, such as *No*,
runs up the gutter beside its lane.

{{< diagram name="swimlanes" caption="swimlanes: lanes as columns, steps aligned by rank, right-angled handoffs" >}}

## quadrant

A categorical 2×2 chart. Each item sits inside a quadrant and has no
coordinates. `q1` is top right, `q2` top left, `q3` bottom left and `q4` bottom
right.

{{< diagram name="quadrant" caption="quadrant: four named quadrants with items" >}}

```d2
vars: { inkline: { diagram: quadrant; design_system: carbon } }

title: Reach and engagement of campaigns
x: Reach →
y: Engagement →
q1: We should expand { campaign_a: Campaign A }
q2: Need to promote { campaign_b: Campaign B }
```

## treeview

A dendrogram. Write the tree as nested maps. inkline flattens it into nodes
joined by links.

{{< diagram name="treeview" caption="treeview: nested maps drawn as a left-to-right tree" >}}

```d2
vars: { inkline: { diagram: treeview; design_system: carbon } }

second_floor: Second floor {
  location: Location {
    tri_building: triBuilding
    tri_floor: triFloor
  }
  geography: Geography
}
```
