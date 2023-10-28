variable "prefix" {
  description = "The prefix name to use for all resources"
}

variable "zone_id" {
  description = "The zone id of the Route53 domain"
}

variable "lambda_version" {
  description = "The version of the lambda module to use (default: latest)"
  default     = "latest"
}

variable "mail_from" {
  description = "The subdomain to use as MAIL FROM"
  default     = "mail"
}

variable "emails" {
  description = "The mapping from email accounts to their respective aliases, e.g. {info = [\"camille@example.com\"]}"
  type        = map(list(string))
}

terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}
