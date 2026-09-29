# inkline diagram types

This file gives the syntax, the roles and a complete example for each type.
Each example renders as it is. `*` marks the default role. Types:

- [architecture](#architecture)
- [flowchart](#flowchart)
- [sequence](#sequence)
- [state](#state)
- [class](#class)
- [er](#er)
- [block](#block)
- [swimlanes](#swimlanes)
- [quadrant](#quadrant)
- [treeview](#treeview)

## architecture

Systems and the connections between them. The terms follow IBM's Unified Method
Framework.

| Kind | Roles |
|---|---|
| node | `component`\*, `actor`, `system`, `node`, `junction` |
| container | `location`\*, `zone`, `subsystem` |
| connection | `connection`\*, `async` |

- Nest containers for places and boundaries: cloud → region → VPC → zone.
- `node` is a deployment node (a server, a VM). `component` is software.
- Add `prescribed` to give a prescribed element square corners. Logical
  elements keep rounded corners.
- `async` draws a dashed connection.

```d2
vars: { inkline: { diagram: architecture; design_system: carbon } }

user: User {class: actor}
public: Public network {
  internet: Internet {class: network}
}
cloud: IBM Cloud {
  region: Dallas region {
    class: prescribed
    vpc: VPC A {
      class: vpc
      zone1: Dallas zone 1 {
        class: zone
        lb: Load balancer {class: network}
        vsi1: VSI 1 {class: compute}
      }
      zone2: Dallas zone 2 {
        class: zone
        vsi2: VSI 2 {class: [compute; multiple]}
      }
    }
  }
  db: Documents database {class: [node; data]}
}
enterprise: Enterprise network {
  apps: Enterprise applications {class: application}
  dl: Direct Link {class: [network; added]}
}

user -> public.internet
public.internet -> cloud.region.vpc.zone1.lb
cloud.region.vpc.zone1.lb -> cloud.region.vpc.zone1.vsi1
cloud.region.vpc.zone1.lb -> cloud.region.vpc.zone2.vsi2
cloud.region.vpc.zone1.vsi1 -> cloud.db: Query
cloud.region.vpc.zone2.vsi2 -> cloud.db: Query
enterprise.dl -> cloud.region.vpc: {class: async}
enterprise.apps -> enterprise.dl
```

## flowchart

Steps, decisions and paths.

| Kind | Roles |
|---|---|
| node | `process`\*, `start`, `end`, `decision`, `io`, `subprocess`, `store`, `note` |
| container | `group`\* |
| connection | `flow`\*, `optional` |

- Label decision branches on the connection: `check -> ok: Yes`.
- `optional` is a dashed flow. Add `changed` to a node to mark an optional step.

```d2
vars: { inkline: { diagram: flowchart; design_system: carbon } }

start: Start {class: start}
prereq: Meet prerequisites
download: Download package
check: Any nondefault installation condition met? {class: decision}
default: Default installation
nondefault: Nondefault installation
configure: Configure
test: Test {class: changed}
setup: Setup {class: changed}
end: End {class: end}

start -> prereq -> check
prereq -> download: {class: optional}
check -> default: No
check -> nondefault: Yes
default -> configure
nondefault -> configure
configure -> test -> setup -> end
```

### Compact decisions

A long question makes the default diamond wide and flat. For a compact
flowchart, draw a small diamond and put the question underneath it. Wrap a
long question onto two lines with `\n`.

```d2
classes: {
  ask: {width: 36; height: 36; label.near: outside-bottom-center; style.font-color: ${color.text}}
}
check: "Within personal or group\nbudget, and overall budget?" {class: [decision; ask]}
```

`font-color` is needed because a decision's text colour is set for the
inside of the diamond. `${color.text}` follows the theme, so the label works
in light and dark. It exists in `carbon`, `cloudscape` and `fluent`. In
`plain`, leave `style.font-color` out.

The layout engine doesn't reserve space for a label outside its shape, so a
box beside the diamond can overlap the question. To move the box down one
rank, send the branch through an invisible point:

```d2
classes: {
  via: {label: ""; shape: circle; width: 1; height: 1; style: {opacity: 0}}
}
v1: {class: via}
check -- v1: No        # undirected, and carries the branch label
v1 -> refuse           # the arrowhead is on the last segment
```

- Stagger only the branches that collide. Render the diagram and look at it
  to find them.
- If one branch's line crosses another box, give that outcome a box of its
  own, for example a second "Refuse with reason".
- Keep the flowchart top-down. `direction: right` gives a thin strip whose
  text is too small, and the arrows cross the questions.

A complete example:

```d2
vars: { inkline: { diagram: flowchart; design_system: fluent } }

classes: {
  ask: {width: 36; height: 36; label.near: outside-bottom-center; style.font-color: ${color.text}}
  via: {label: ""; shape: circle; width: 1; height: 1; style: {opacity: 0}}
}

start: Request {class: start}
session: Valid session? {class: [decision; ask]}
domain: "Route allowed for\nrole and location?" {class: [decision; ask]}
budget: "Within budget?" {class: [decision; ask]}
run: Handle request
signin: Redirect to sign in {class: end}
refuse: Refuse with reason {class: end}
refuse2: Refuse with reason {class: end}
done: Respond {class: end}
v1: {class: via}
v2: {class: via}

start -> session
session -- v1: No
v1 -> signin
session -> domain: Yes
domain -- v2: No
v2 -> refuse
domain -> budget: Yes
budget -> refuse2: No
budget -> run: Yes
run -> done
```

## sequence

Messages between participants over time. inkline sets `shape: sequence_diagram`.

| Kind | Roles |
|---|---|
| participant | `participant`\*, `actor` |
| connection | `message`\*, `reply`, `async` |

- Participants appear left to right in the order they are first mentioned.
- Self-messages (`svc -> svc: …`) are allowed.
- D2 spans and notes work as usual inside participants.
- A group is a top-level container that holds messages:
  `retry: Retry { app -> svc: again }`. It takes no role.

```d2
vars: { inkline: { diagram: sequence; design_system: carbon } }

user: User {class: actor}
app: Mobile app
gw: API gateway
svc: Order service

user -> app: Place order
app -> gw: POST /orders
gw -> svc: create(order)
svc -> gw: 201 Created {class: reply}
gw -> app: order id {class: reply}
svc -> svc: publish OrderCreated {class: async}
```

## state

A lifecycle: states and transitions.

| Kind | Roles |
|---|---|
| node | `state`\*, `start`, `end`, `choice`, `fork`, `join`, `note` |
| container | `composite`\* |
| connection | `transition`\* |

- A quoted `"[*]"` is the start state when it is the source of a connection and
  the end state when it is the target. Each scope has its own, so a composite
  state gets its own start and end.
- Label transitions with the event: `pending -> stable: propagate`.

```d2
vars: { inkline: { diagram: state; design_system: carbon } }

"[*]" -> pending: create
pending -> stable: propagate
pending -> pending_blocked: block
pending_blocked -> pending: resolve
stable -> updating: update
updating -> stable: reconcile
stable -> suspended: suspend
suspended -> stable: restore
cleanup: Cleanup {
  "[*]" -> deleting
  deleting -> deleted: propagate
  deleted -> "[*]"
}
stable -> cleanup: delete
cleanup -> "[*]"
```

## class

Classes and their relationships. Keys inside a class are fields and methods.

| Kind | Roles |
|---|---|
| class | `class`\*, `interface`, `abstract`, `enum` |
| connection | `association`\*, `inheritance`, `realization`, `composition`, `aggregation`, `dependency` |

- Visibility prefixes: `+` public, `-` private, `#` protected.
- Quote types with brackets or stars: `lines: "[]Line"`, `db: "*sql.DB"`.
- inkline adds a `<<stereotype>>` line to `interface`, `abstract` and `enum`.
- Direction: from the child, implementer, whole or dependent to the parent,
  interface, part or dependency.

```d2
vars: { inkline: { diagram: class; design_system: carbon } }

Repository: {
  class: interface
  +find(id string): Order
  +save(o Order): error
}
Order: {
  +id: string
  +lines: "[]Line"
  +total(): Money
}
Line: {
  +sku: string
  +qty: int
}
SQLRepository: {
  -db: "*sql.DB"
  +find(id string): Order
  +save(o Order): error
}
SQLRepository -> Repository: {class: realization}
Order -> Line: {class: composition}
Repository -> Order: {class: dependency}
```

## er

Tables, columns and keys. Keys inside an entity are columns.

| Kind | Roles |
|---|---|
| entity | `entity`\* |
| connection | `relation`\*, `one-to-one`, `one-to-many`, `many-to-one`, `many-to-many` |

- Constraints: `{constraint: primary_key}`, `foreign_key`, `unique`, or a list `[primary_key; foreign_key]`.
- Connect columns (`customer.id -> order.customer_id`) so the line meets the right row.
- Plain `relation` has no crow's feet. Use a cardinality role to show them.

```d2
vars: { inkline: { diagram: er; design_system: carbon } }

customer: {
  id: int {constraint: primary_key}
  name: string
  email: string {constraint: unique}
}
order: {
  id: int {constraint: primary_key}
  customer_id: int {constraint: foreign_key}
  placed_at: timestamp
}
line: {
  order_id: int {constraint: [primary_key; foreign_key]}
  sku: string {constraint: primary_key}
  qty: int
}
customer.id -> order.customer_id: places {class: one-to-many}
order.id -> line.order_id: contains {class: one-to-many}
```

## block

A layered stack or grid of components. Use D2 grids.

| Kind | Roles |
|---|---|
| node | `block`\* |
| container | `group`\* |
| connection | `link`\* |

- Set `grid-columns: N` (or `grid-rows`) at the top level or inside a group.
- `grid-gap` sets the spacing.

```d2
vars: { inkline: { diagram: block; design_system: carbon } }

grid-columns: 3
grid-gap: 16
frontend: Frontend
api: API
auth: Auth
services: Services {
  grid-columns: 2
  orders: Orders
  billing: Billing
}
queue: Queue
db: Database {class: data}
```

## swimlanes

A process across teams or systems, with handoffs.

| Kind | Roles |
|---|---|
| lane | `lane`\* |
| node | `step`\*, `start`, `end`, `decision` |
| connection | `flow`\*, `optional` |

- Every step lives inside a lane (a top-level container). Lanes appear in file order.
- inkline ranks steps along the flow. Loops back to an earlier step (such as a
  *No* branch) are allowed.
- Set `orientation: horizontal` in the header to draw lanes as rows.

```d2
vars: { inkline: { diagram: swimlanes; design_system: carbon } }

customer: Customer {
  start: Submit order {class: start}
  pay: Pay invoice
}
sales: Sales {
  review: Review order
  ok: Approved? {class: decision}
}
warehouse: Warehouse {
  pick: Pick items
  ship: Ship {class: end}
}
customer.start -> sales.review -> sales.ok
sales.ok -> warehouse.pick: Yes
sales.ok -> customer.start: No {class: optional}
warehouse.pick -> customer.pay -> warehouse.ship
```

## quadrant

A categorical 2×2.

| Kind | Roles |
|---|---|
| chart | `frame` (generated as `vf_chart`) |
| quadrant | `quadrant` |
| node | `item`\* |
| axis | `axis` (generated from `x` and `y`) |

- `q1` top right, `q2` top left, `q3` bottom left, `q4` bottom right. Missing quadrants are drawn empty.
- Items sit inside a quadrant. There are no coordinates.

```d2
vars: { inkline: { diagram: quadrant; design_system: carbon } }

title: Reach and engagement of campaigns
x: Reach →
y: Engagement →
q1: We should expand {
  campaign_a: Campaign A
  campaign_f: Campaign F
}
q2: Need to promote {
  campaign_b: Campaign B
}
q3: Re-evaluate {
  campaign_c: Campaign C
}
q4: May be improved {
  campaign_d: Campaign D
  campaign_e: Campaign E
}
```

## treeview

A hierarchy, drawn left to right as a dendrogram.

| Kind | Roles |
|---|---|
| node | `root` (top level), `branch` (has children), `leaf` (no children), set by depth |
| connection | `link`\* |

- Nesting defines the tree. A connection inside the tree is an error.
- Several top-level keys make several roots.

```d2
vars: { inkline: { diagram: treeview; design_system: carbon } }

second_floor: Second floor {
  classification: Classification
  geography: Geography
  location: Location {
    tri_building: triBuilding
    tri_floor: triFloor
    tri_space: triSpace {
      is_parent_of: Is Parent Of
      has: has {
        as11: 02-AS11 {
          tri_people: triPeople {
            profile: My Profile
          }
          tri_routing: triRouting
        }
        bm26: 02-BM26
        bp04: 02-BP04
      }
    }
  }
  organization: Organization
  system: System
}
```
