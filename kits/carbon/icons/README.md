# Icons

Put icon SVGs here (Carbon UI icons, 24px; kit p3). Reference them from a
diagram with D2's `icon` key, relative to the diagram file or as a URL:

```d2
lb: Load balancer {class: network; icon: ../kits/carbon/icons/load-balancer.svg}
```

On `component` and `node` (architecture) the colour-block decorator centres the
icon, at 24px, in the 48px block. Icons are drawn as images, so they are not
recoloured: use white icons for colour blocks in the light theme and dark
icons in the dark theme.

## Bundled

`user.svg` is the Carbon `user` icon (IBM, Apache License 2.0). The design
system draws it in sequence actors' colour blocks and in architecture actor
circles.
