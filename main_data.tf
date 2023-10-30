data "aws_route53_zone" "zone" {
  for_each = { for zone, config in var.emails : zone => zone }

  name = each.key
}

data "aws_region" "current" {}

data "aws_ecr_authorization_token" "token" {}

data "aws_caller_identity" "current" {}

locals {
  domains = keys(var.emails)
}
