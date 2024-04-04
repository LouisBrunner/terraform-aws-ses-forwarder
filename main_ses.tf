resource "aws_ses_domain_mail_from" "from" {
  for_each = var.mail_from != "" ? { for domain in local.domains : domain => domain } : {}

  domain           = aws_ses_domain_identity.identity[each.key].domain
  mail_from_domain = "${var.mail_from}.${each.key}"
}

resource "aws_ses_domain_identity" "identity" {
  for_each = { for domain in local.domains : domain => domain }

  domain = each.key
}

locals {
  destinaries = toset(flatten([for domain, config in var.emails :
    [for entry in config : entry.forward_to]
  ]))
}

resource "aws_ses_email_identity" "destinaries" {
  for_each = { for destinaries in local.destinaries : destinaries => destinaries }

  email = each.key
}

resource "aws_ses_domain_identity_verification" "verification" {
  for_each = { for domain in local.domains : domain => domain }

  domain = aws_ses_domain_identity.identity[each.key].id

  depends_on = [aws_route53_record.verification]
}

resource "aws_ses_domain_dkim" "dkim" {
  for_each = { for domain in local.domains : domain => domain }

  domain = each.key
}

resource "aws_ses_receipt_rule_set" "rule" {
  rule_set_name = "${var.prefix}-rule"
}

resource "aws_ses_active_receipt_rule_set" "rule" {
  rule_set_name = aws_ses_receipt_rule_set.rule.rule_set_name
}

resource "aws_ses_receipt_rule" "rule" {
  name          = "${var.prefix}-rule"
  rule_set_name = aws_ses_receipt_rule_set.rule.rule_set_name
  recipients    = [for domain in local.domains : domain]
  enabled       = true
  scan_enabled  = var.scan_enabled

  sns_action {
    topic_arn = aws_sns_topic.emails.arn
    encoding  = "Base64"
    position  = 1
  }
}

resource "aws_sns_topic" "emails" {
  name_prefix = var.prefix
  policy      = data.aws_iam_policy_document.sns_access.json
}

data "aws_iam_policy_document" "sns_access" {
  policy_id = "__default_policy_ID"

  statement {
    sid = "__default_statement_ID"

    actions = [
      "sns:Publish",
    ]

    resources = [
      "*",
    ]

    principals {
      type        = "Service"
      identifiers = ["ses.amazonaws.com"]
    }
  }
}

resource "aws_sns_topic_subscription" "emails_and_lambda" {
  endpoint  = module.lambda.lambda_function_arn
  protocol  = "lambda"
  topic_arn = aws_sns_topic.emails.arn
}
