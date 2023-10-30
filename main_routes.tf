locals {
  mail_from_domain = var.mail_from != "" ? {
    for domain in local.domains : domain => aws_ses_domain_mail_from.from[domain].mail_from_domain
    } : {
    for domain in local.domains : domain => domain
  }
}

resource "aws_route53_record" "mx" {
  for_each = { for domain in local.domains : domain => domain }

  zone_id = data.aws_route53_zone.zone[each.key].zone_id
  name    = local.mail_from_domain[each.key]
  type    = "MX"
  ttl     = "600"
  records = ["10 inbound-smtp.${data.aws_region.current.name}.amazonses.com"]
}

resource "aws_route53_record" "spf" {
  for_each = { for domain in local.domains : domain => domain }

  zone_id = data.aws_route53_zone.zone[each.key].zone_id
  name    = local.mail_from_domain[each.key]
  type    = "TXT"
  ttl     = "600"
  records = ["v=spf1 include:amazonses.com -all"]
}

resource "aws_route53_record" "verification" {
  for_each = { for domain in local.domains : domain => domain }

  zone_id = data.aws_route53_zone.zone[each.key].zone_id
  name    = "_amazonses.${aws_ses_domain_identity.identity[each.key].id}"
  type    = "TXT"
  ttl     = "1800"
  records = [aws_ses_domain_identity.identity[each.key].verification_token]
}

locals {
  dkim_token_with_domain = flatten([
    for domain in local.domains : [
      for dkim_token in aws_ses_domain_dkim.dkim[domain].dkim_tokens : {
        domain     = domain
        dkim_token = dkim_token
      }
    ]
  ])
}

resource "aws_route53_record" "dkim" {
  for_each = { for data in local.dkim_token_with_domain : "${data.domain}_${data.dkim_token}" => data }

  zone_id = data.aws_route53_zone.zone[each.value.domain].zone_id
  name    = "${each.value.dkim_token}._domainkey"
  type    = "CNAME"
  ttl     = "1800"
  records = ["${each.value.dkim_token}.dkim.amazonses.com"]
}
