resource "aws_route53_zone" "main" {
  name              = var.name
  comment           = var.comment
  delegation_set_id = length(var.vpc) == 0 ? var.delegation_set_id : null
  force_destroy     = var.force_destroy

  dynamic "vpc" {
    for_each = var.vpc
    content {
      vpc_id     = vpc.value.vpc_id
      vpc_region = lookup(vpc.value, "vpc_region", null)
    }
  }

  tags = var.tags
}

# Route53 records
resource "aws_route53_record" "main" {
  count                            = length(var.records)
  zone_id                          = aws_route53_zone.main.zone_id
  name                             = try(var.records[count.index]["name"], null)
  type                             = try(var.records[count.index]["type"], null)
  ttl                              = try([var.records[count.index]["alias"]], []) != [] ? null : try(var.records[count.index]["ttl"], 300)
  records                          = length(keys(lookup(var.records[count.index], "alias", {}))) > 0 ? null : lookup(var.records[count.index], "records", [])
  set_identifier                   = try(var.records[count.index]["set_identifier"], null)
  health_check_id                  = try(var.records[count.index]["health_check_id"], null)
  multivalue_answer_routing_policy = try(var.records[count.index]["multivalue_answer_routing_policy"], null)
  allow_overwrite                  = try(var.records[count.index]["allow_overwrite"], null)

  dynamic "alias" {
    for_each = try([var.records[count.index]["alias"]], [])
    content {
      name                   = try(alias.value.name, null)
      zone_id                = try(alias.value.zone_id, null)
      evaluate_target_health = try(alias.value.evaluate_target_health, false)
    }
  }

  dynamic "failover_routing_policy" {
    for_each = try(var.records[count.index]["multivalue_answer_routing_policy"], null) == null ? try([var.records[count.index]["failover_routing_policy"]], []) : []
    content {
      type = lookup(failover_routing_policy.value, "type", null)
    }
  }

  dynamic "geolocation_routing_policy" {
    for_each = try(var.records[count.index]["multivalue_answer_routing_policy"], null) == null ? try([var.records[count.index]["geolocation_routing_policy"]], []) : []
    content {
      continent   = try(geolocation_routing_policy.value.continent, null)
      country     = try(geolocation_routing_policy.value.country, null)
      subdivision = try(geolocation_routing_policy.value.subdivision, null)
    }
  }

  dynamic "latency_routing_policy" {
    for_each = try(var.records[count.index]["multivalue_answer_routing_policy"], null) == null ? try([var.records[count.index]["latency_routing_policy"]], []) : []
    content {
      region = lookup(latency_routing_policy.value, "region", null)
    }
  }

  dynamic "weighted_routing_policy" {
    for_each = try(var.records[count.index]["multivalue_answer_routing_policy"], null) == null ? try([var.records[count.index]["weighted_routing_policy"]], []) : []
    content {
      weight = lookup(weighted_routing_policy.value, "weight", null)
    }
  }
}

resource "aws_cloudwatch_log_group" "main" {
  count             = length(var.vpc) == 0 ? 1 : 0
  provider          = aws.us-east-1
  name              = "/aws/route53/${var.name}"
  retention_in_days = var.retention_in_days
  kms_key_id        = var.kms_key_id
  tags              = var.tags
}

resource "aws_cloudwatch_log_resource_policy" "main" {
  count           = length(var.vpc) == 0 ? 1 : 0
  provider        = aws.us-east-1
  policy_document = data.aws_iam_policy_document.route53_query_logging_policy.json
  policy_name     = "${var.name}-query-logging-policy"
}

resource "aws_route53_query_log" "main" {
  count                    = length(var.vpc) == 0 ? 1 : 0
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.main[0].arn
  zone_id                  = aws_route53_zone.main.zone_id
  depends_on               = [aws_cloudwatch_log_resource_policy.main]
}
