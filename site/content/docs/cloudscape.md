---
title: Cloudscape
weight: 7
---

The `cloudscape` design system applies AWS's
[Cloudscape Design System](https://cloudscape.design/) to diagrams. It also adds
the groupings that AWS architecture diagrams use. These are AWS Cloud, Regions,
Availability Zones, VPCs, subnets, security groups and Auto Scaling groups.

```d2
vars: { inkline: { diagram: architecture; design_system: cloudscape } }
```

{{< diagram name="aws" caption="examples/aws.d2: a three-tier service in the Cloudscape kit, with AWS groupings" >}}

Cloudscape is a design system for web application interfaces. It has no
diagram guidance of its own, so this kit is inkline's translation of it. inkline
generates the kit's colours from Cloudscape's published design tokens, so they
match Cloudscape in light and dark. Blue is the accent colour, as Cloudscape's
colour guidance asks.

| Aspect | Setting |
|---|---|
| Type | Open Sans, 14px, bold labels |
| Shapes | Rounded: 8px cards, 16px groups, pill-shaped labels and start/end steps |
| Borders | Cloudscape's neutral card border |
| Lines | Cloudscape's line grey, with open chevron arrowheads like its icons |
| Colours | Cloudscape tokens for canvas, surfaces, text and status; AWS's service-category colours for domains; AWS squid ink for emphasis |
| Services | 48px tiles filled with the domain colour, labelled underneath, as in AWS diagrams |
| People | Cloudscape avatars: its `user-profile` icon, outlined, on the avatar grey |

## AWS groupings

Add a grouping class to a container. Its role stays `location`, so the same
diagram also renders in other kits.

| Class | Draws | Use for |
|---|---|---|
| `aws-cloud` | solid dark border | the AWS Cloud boundary |
| `region` | dashed teal | an AWS Region |
| `availability-zone` | dashed teal | an Availability Zone |
| `vpc` | solid purple | a VPC |
| `public-subnet` | solid green, tinted | a public subnet |
| `private-subnet` | solid teal, tinted | a private subnet |
| `security-group` | solid red | a security group |
| `auto-scaling-group` | dashed orange | an Auto Scaling group |
| `data-center` | solid grey | an on-premises data centre |

```d2
aws: AWS Cloud {
  class: aws-cloud
  vpc: VPC 10.0.0.0/16 {
    class: vpc
    public: Public subnet {
      class: public-subnet
      alb: Load balancer {class: network}
    }
  }
}
```

## Domain colours

Cloudscape uses the same classes as Carbon, so a diagram renders in either kit
without changes. The colours follow the service categories of AWS's
architecture icons, as AWS's own diagrams do. On architecture services they fill the tile.

| Class | Colour | AWS category |
|---|---|---|
| `compute` | orange | Compute, Containers |
| `storage` | green | Storage (S3, EBS, EFS) |
| `data` | magenta | Database |
| `network` | purple | Networking & Content Delivery |
| `security` | red | Security, Identity & Compliance |
| `application` | pink | Application Integration |
| `observability` | pink | Management & Governance |
| `devops` | magenta | Developer Tools |
| `backend` | grey | General |

## Icons

AWS Architecture Icons are not openly licensed, so inkline does not include them.
Download the official asset package from
[AWS](https://aws.amazon.com/architecture/icons/) and reference an icon with
D2's `icon` key. The icon sits in the service's tile. inkline draws a single-colour
icon from any other set in white on the tile colour.

```d2
lambda: Orders function {class: compute; icon: ./aws-icons/Arch_AWS-Lambda_48.svg}
```

## Licences

Cloudscape's design tokens and its user-profile icon are under the Apache
License 2.0 (Copyright Amazon.com, Inc. or its affiliates). Open Sans is under
the SIL Open Font License 1.1. The licence texts are in the repository's
`THIRD_PARTY_NOTICES.md`. inkline is not affiliated with or endorsed by Amazon.
