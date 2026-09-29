# The Cloudscape design system

`design_system: cloudscape` adapts AWS's [Cloudscape](https://cloudscape.design/)
for diagrams. It uses Open Sans, Cloudscape colour tokens (light and dark),
rounded cards with neutral borders, pill labels, open chevron arrowheads and
Cloudscape avatars for people. Use it when the user wants an AWS look or an
AWS architecture diagram.

## Architecture: tiles and AWS groupings

Architecture services (`component`, the default) are 48px tiles filled with
their domain colour, with the label underneath, as in AWS diagrams. Wrap them
in AWS groupings by adding a grouping class to a container. The container's
role stays `location`.

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

Nest them in the same order as AWS: AWS Cloud → Region → VPC → Availability
Zone → subnet → Auto Scaling or security group → services. Regional services
that are outside a VPC (S3, Cognito, DynamoDB) sit in the Region. External
systems and users sit outside the AWS Cloud.

```d2
vars: { inkline: { diagram: architecture; design_system: cloudscape } }

users: Users {class: actor}
aws: AWS Cloud {
  class: aws-cloud
  region: us-east-1 {
    class: region
    vpc: VPC 10.0.0.0/16 {
      class: vpc
      az: Availability Zone a {
        class: availability-zone
        public: Public subnet {
          class: public-subnet
          alb: Load balancer {class: network}
        }
        private: Private subnet {
          class: private-subnet
          app: Orders service {class: compute}
          db: Aurora {class: data}
        }
      }
    }
    s3: Assets bucket {class: storage}
  }
}
users -> aws.region.vpc.az.public.alb: HTTPS
aws.region.vpc.az.public.alb -> aws.region.vpc.az.private.app
aws.region.vpc.az.private.app -> aws.region.vpc.az.private.db: SQL
aws.region.vpc.az.private.app -> aws.region.s3
```

## Domain colours and modifiers

The classes are the same as in Carbon, so a diagram moves between kits
unchanged. The colours follow AWS's service categories:

- `compute` orange (EC2, ECS, Lambda)
- `storage` green (S3, EBS, EFS)
- `data` magenta (RDS, Aurora, DynamoDB)
- `network` purple (ELB, CloudFront, API Gateway, Route 53)
- `security` red (IAM, Cognito, WAF)
- `application` pink (SQS, SNS, EventBridge)
- `observability` pink (CloudWatch)
- `devops` magenta
- `backend` grey

Pick the class from the AWS service's category. In this kit `vpc` is the
grouping. Modifiers: `multiple`, `added` (double outline), `changed` (dashed),
`removed` (struck through).

For a legend, give each entry the `swatch` class and its domain:
`vars.d2-legend.data: Data {class: [swatch; data]}`.

## Icons

inkline does not bundle AWS Architecture Icons, because they are not openly
licensed. If the user has the official asset package, add an icon with D2's
`icon` key and it sits in the tile: `fn: Orders function {class: compute; icon: ./aws-icons/Arch_AWS-Lambda_48.svg}`.
Otherwise leave icons out. The tiles and domain colours show what each service
is.
