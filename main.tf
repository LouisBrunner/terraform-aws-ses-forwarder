resource "aws_ecr_repository" "repository" {
  name = "${var.prefix}-repository"
}

resource "aws_ecr_lifecycle_policy" "keep_few" {
  repository = aws_ecr_repository.repository.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        description  = "Keep only the last 3 images",
        selection = {
          tagStatus   = "any",
          countType   = "imageCountMoreThan",
          countNumber = 3,
        },
        action = {
          type = "expire",
        },
      },
    ],
  })
}


locals {
  source_version = var.debug_mode ? "latest" : var.lambda_version
  version        = var.debug_mode ? "v0.0.0-${formatdate("YYYYMMDD'T'hhmmssZ", timestamp())}" : var.lambda_version
  lambda_src     = var.debug_mode ? "local/aws-ses-forwarder:${local.source_version}" : "ghcr.io/louisbrunner/aws-ses-forwarder:${local.source_version}"
  lambda_dest    = "${aws_ecr_repository.repository.repository_url}:${local.version}"
}

resource "null_resource" "docker_pull_push" {
  triggers = {
    lambda_version = local.version
    debug_mode     = var.debug_mode
  }

  provisioner "local-exec" {
    command = <<EOF
      docker pull ${local.lambda_src}
      docker tag ${local.lambda_src} ${local.lambda_dest}
      echo ${data.aws_ecr_authorization_token.token.password} | docker login --username AWS --password-stdin ${aws_ecr_repository.repository.repository_url}
      docker push ${local.lambda_dest}
    EOF
  }
}

module "lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "6.2.0"

  function_name = "${var.prefix}-lambda"
  description   = "Lambda function to forward SES emails"

  create_package = false
  publish        = true

  image_uri     = local.lambda_dest
  package_type  = "Image"
  architectures = ["x86_64"]

  allowed_triggers = {
    "AllowExecutionBySNS" = {
      service    = "sns"
      source_arn = aws_sns_topic.emails.arn
    }
  }

  attach_policy_json = true
  policy_json        = data.aws_iam_policy_document.raw_email.json

  environment_variables = {
    CONFIG = jsonencode({
      "emails" = flatten([
        for domain, config in var.emails :
        [
          for entry in config :
          {
            regex      = "${trimsuffix(entry.regex, "$")}@${domain}$",
            forward_to = entry.forward_to,
          }
        ]
      ]),
    })
  }

  depends_on = [null_resource.docker_pull_push]
}

data "aws_iam_policy_document" "raw_email" {
  statement {
    actions = [
      "ses:SendRawEmail",
    ]

    resources = [for destinary in local.destinaries :
      aws_ses_email_identity.destinaries[destinary].arn
    ]
  }
}
