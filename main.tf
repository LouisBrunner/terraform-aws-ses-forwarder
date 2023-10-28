data "aws_route53_zone" "zone" {
  zone_id = var.zone_id
}

data "aws_region" "current" {}

locals {
  domain = data.aws_route53_zone.zone.name
  emails = {
    for email, aliases in var.emails :
    "^${email}@${local.domain}$" => aliases
  }
}

resource "aws_route53_record" "mx" {
  zone_id = var.zone_id
  name    = local.domain
  type    = "MX"
  ttl     = "600"
  records = ["10 feedback-smtp.${data.aws_region.current.name}.amazonses.com"]
}

resource "aws_ses_domain_mail_from" "from" {
  domain           = aws_ses_domain_identity.identity.domain
  mail_from_domain = "${var.mail_from}.${local.domain}"
}

resource "aws_route53_record" "spf" {
  zone_id = var.zone_id
  name    = aws_ses_domain_mail_from.from.mail_from_domain
  type    = "TXT"
  ttl     = "600"
  records = ["v=spf1 include:amazonses.com -all"]
}

resource "aws_ses_domain_identity" "identity" {
  domain = local.domain
}

resource "aws_route53_record" "verification" {
  zone_id = var.zone_id
  name    = "_amazonses.${local.domain}"
  type    = "TXT"
  ttl     = "1800"
  records = [aws_ses_domain_identity.identity.verification_token]
}

resource "aws_ses_domain_identity_verification" "verification" {
  domain = aws_ses_domain_identity.identity.id

  depends_on = [aws_route53_record.verification]
}

resource "aws_ses_domain_dkim" "dkim" {
  domain = local.domain
}

resource "aws_route53_record" "dkim" {
  for_each = toset(aws_ses_domain_dkim.dkim.dkim_tokens)

  zone_id = var.zone_id
  name    = "${each.key}._domainkey.${local.domain}"
  type    = "CNAME"
  ttl     = "1800"
  records = ["${each.key}.dkim.amazonses.com"]
}

resource "aws_ses_email_identity" "users" {
  for_each = var.emails

  email = "${each.key}@${local.domain}"
}

resource "aws_sns_topic" "emails" {
  name         = "${var.prefix}-sns"
  display_name = "Email notifications for ${local.domain}"
}

resource "aws_sns_topic_subscription" "emails" {
  topic_arn = aws_sns_topic.emails.arn
  protocol  = "lambda"
  endpoint  = module.lambda.lambda_function_arn
}

resource "aws_ecr_repository" "repository" {
  name = "${var.prefix}-repository"
}

data "aws_ecr_authorization_token" "token" {
}

resource "null_resource" "docker_pull_push" {
  triggers = {
    lambda_version = var.lambda_version
  }

  provisioner "local-exec" {
    command = <<EOF
      docker pull ghcr.io/louisbrunner/aws-ses-forwarder:${var.lambda_version}
      docker tag ghcr.io/louisbrunner/aws-ses-forwarder:${var.lambda_version} ${aws_ecr_repository.repository.repository_url}:${var.lambda_version}
      echo ${data.aws_ecr_authorization_token.token.password} | docker login --username AWS --password-stdin ${aws_ecr_repository.repository.repository_url}
      docker push ${aws_ecr_repository.repository.repository_url}:${var.lambda_version}
    EOF
  }
}

module "lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "6.2.0"

  function_name = "${var.prefix}-lambda"
  description   = "Lambda function to forward SES emails to SNS"

  create_package = false
  publish        = true

  image_uri     = "${aws_ecr_repository.repository.repository_url}:${var.lambda_version}"
  package_type  = "Image"
  architectures = ["x86_64"]

  allowed_triggers = {
    AllowExecutionBySNS = {
      service    = "sns"
      source_arn = aws_sns_topic.emails.arn
    }
  }

  attach_policy_json = true
  policy_json        = data.aws_iam_policy_document.raw_email.json

  environment_variables = {
    CONFIG = jsonencode({
      "emails" = local.emails,
    })
  }
}

data "aws_iam_policy_document" "raw_email" {
  statement {
    actions = [
      "ses:SendRawEmail",
    ]

    resources = [
      aws_ses_domain_identity.identity.arn,
    ]
  }
}
