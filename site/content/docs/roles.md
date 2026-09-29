---
title: Roles
weight: 4
---

A role states what an element is. A design system sets how each role looks.
The core gives every object and connection exactly one role, as the first entry
of its D2 `class` list.

```d2
db: Orders {class: store}              # you pick the role
api: Orders API {class: [node; data]}  # role, then a design-system class
```

An element without a role gets its diagram's default role (marked \* below).
D2 applies classes in order, and later classes take precedence. Classes you add
after the role override the role's styling.

## Modifiers

Modifiers work in every diagram type. inkline does not treat them as roles.

| Modifier | Meaning |
|---|---|
| `multiple` | several instances of the element |
| `added` | new in this change |
| `changed` | modified or passive |
| `removed` | removed or disabled |

The `carbon` design system adds the architecture modifier `prescribed` and
these domain colours: `security`, `devops`, `application`, `data`, `network`,
`observability`, `vpc`, `compute` and `backend`.

## Roles by diagram

### architecture
| Kind | Roles |
|---|---|
| node | `component`\*, `actor`, `system`, `node`, `junction` |
| container | `location`\*, `zone`, `subsystem` |
| connection | `connection`\*, `async` |

### flowchart
| Kind | Roles |
|---|---|
| node | `process`\*, `start`, `end`, `decision`, `io`, `subprocess`, `store`, `note` |
| container | `group`\* |
| connection | `flow`\*, `optional` |

### sequence
| Kind | Roles |
|---|---|
| participant | `participant`\*, `actor` |
| connection | `message`\*, `reply`, `async` |

A group, which is a top-level container that holds messages, takes no role.

### state
| Kind | Roles |
|---|---|
| node | `state`\*, `start`, `end`, `choice`, `fork`, `join`, `note` |
| container | `composite`\* |
| connection | `transition`\* |

### class
| Kind | Roles |
|---|---|
| class | `class`\*, `interface`, `abstract`, `enum` |
| connection | `association`\*, `inheritance`, `realization`, `composition`, `aggregation`, `dependency` |

Write connections from the child, implementer, whole or dependent end.

### er
| Kind | Roles |
|---|---|
| entity | `entity`\* |
| connection | `relation`\*, `one-to-one`, `one-to-many`, `many-to-one`, `many-to-many` |

### block
| Kind | Roles |
|---|---|
| node | `block`\* |
| container | `group`\* |
| connection | `link`\* |

### swimlanes
| Kind | Roles |
|---|---|
| lane | `lane`\* |
| node | `step`\*, `start`, `end`, `decision` |
| connection | `flow`\*, `optional` |

### quadrant
| Kind | Roles |
|---|---|
| chart | `frame` |
| quadrant | `quadrant` |
| node | `item`\* |
| axis | `axis` |

### treeview
| Kind | Roles |
|---|---|
| node | `root`, `branch`, `leaf` |
| connection | `link`\* |

inkline assigns roles by depth. The top level is `root`, a node with children is a
`branch` and every other node is a `leaf`.
