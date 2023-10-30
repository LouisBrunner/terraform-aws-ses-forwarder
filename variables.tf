variable "prefix" {
  description = "The prefix name to use for all resources"
}

variable "lambda_version" {
  description = "The version of the lambda module to use (default: latest)"
  default     = "v1.0.0"
}

variable "mail_from" {
  description = "The subdomain to use as MAIL FROM"
  default     = ""
}

variable "emails" {
  description = "The mapping from email accounts to their respective aliases, e.g. {\"example.com\" = {\"^info$\" = [\"camille@gmail.com\"]}}"
  type        = map(map(list(string)))
}

variable "scan_enabled" {
  description = "Whether to scan emails for spam and viruses"
  type        = bool
  default     = false
}

variable "debug_mode" {
  description = "Developer setting which allows easier debugging for the lambda function"
  type        = bool
  default     = false
}

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    null = {
      source  = "hashicorp/null"
      version = ">= 2.0"
    }
  }
}
