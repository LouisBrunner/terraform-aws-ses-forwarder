# terraform-aws-ses-forwarder [![Coverage Status](https://coveralls.io/repos/github/LouisBrunner/terraform-aws-ses-forwarder/badge.svg?branch=main)](https://coveralls.io/github/LouisBrunner/terraform-aws-ses-forwarder?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/LouisBrunner/terraform-aws-ses-forwarder)](https://goreportcard.com/report/github.com/LouisBrunner/terraform-aws-ses-forwarder)

AWS Lambda (written in Go) to forward emails using AWS SES.

## Usage

### Terraform module

```hcl
resource "aws_route53_zone" "domain" {
  name = "example.com"
}

module "email" {
  source  = "LouisBrunner/ses-forwarder/aws"
  version = "2.0.0"
  prefix  = "forwarder"

  emails = {
    "example.com" = [
      {regex = "^camille$", forward_to = ["camille@gmail.com"]}
    ]
  }
}
```

### Docker image

You can also use the Docker image directly:

```bash
docker pull ghcr.io/louisbrunner/terraform-aws-ses-forwarder:v2.0.0
```
