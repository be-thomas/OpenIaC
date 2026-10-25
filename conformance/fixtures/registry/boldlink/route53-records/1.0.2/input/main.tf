resource "aws_route53_record" "main" {
  zone_id                          = var.zone_id
  name                             = var.name
  type                             = var.type
  ttl                              = length(keys(var.alias)) > 0 ? null : var.ttl
  records                          = length(keys(var.alias)) > 0 ? null : var.records
  allow_overwrite                  = var.allow_overwrite
  health_check_id                  = var.health_check_id
  set_identifier                   = var.set_identifier
  multivalue_answer_routing_policy = var.multivalue_answer_routing_policy

  dynamic "alias" {
    for_each = length(keys(var.alias)) > 0 ? [var.alias] : []
    content {
      name                   = alias.value.name
      zone_id                = alias.value.zone_id
      evaluate_target_health = alias.value.evaluate_target_health
    }
  }

  ## Only one policy block should be supplied
  dynamic "geolocation_routing_policy" {
    for_each = length(keys(var.geolocation_routing_policy)) > 0 && var.multivalue_answer_routing_policy == null ? [var.geolocation_routing_policy] : []
    content {
      continent   = try(geolocation_routing_policy.value.continent, null)
      country     = try(geolocation_routing_policy.value.country, null)
      subdivision = try(geolocation_routing_policy.value.subdivision, null)
    }
  }

  dynamic "failover_routing_policy" {
    for_each = length(keys(var.failover_routing_policy)) > 0 && var.multivalue_answer_routing_policy == null ? [var.failover_routing_policy] : []
    content {
      type = failover_routing_policy.value.type
    }
  }

  dynamic "latency_routing_policy" {
    for_each = length(keys(var.latency_routing_policy)) > 0 && var.multivalue_answer_routing_policy == null ? [var.latency_routing_policy] : []
    content {
      region = latency_routing_policy.value.region
    }
  }

  dynamic "weighted_routing_policy" {
    for_each = length(keys(var.weighted_routing_policy)) > 0 && var.multivalue_answer_routing_policy == null ? [var.weighted_routing_policy] : []
    content {
      weight = weighted_routing_policy.value.weight
    }
  }

  dynamic "cidr_routing_policy" {
    for_each = length(keys(var.cidr_routing_policy)) > 0 && var.multivalue_answer_routing_policy == null ? [var.cidr_routing_policy] : []
    content {
      collection_id = cidr_routing_policy.value.collection_id
      location_name = cidr_routing_policy.value.location_name
    }
  }
}
