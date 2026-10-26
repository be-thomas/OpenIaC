resource "aws_api_gateway_usage_plan" "usage_plan" {
  name        = var.name
  description = "Usage plan created for ${var.name}"

  dynamic "api_stages" {
    for_each = var.stages

    content {
      api_id = api_stages.value.api_id
      stage  = api_stages.value.stage
    }

  }

  dynamic "quota_settings" {
    for_each = var.quota_settings_unlimited == true ? [1] : []
    content {
      limit  = var.limit
      offset = var.offset
      period = var.period
    }
  }

  dynamic "throttle_settings" {
    for_each = var.enable_throttling == true ? [1] : []
    content {
      burst_limit = var.burst_limit
      rate_limit  = var.rate_limit
    }
  }

  tags = {
    Name = var.name
  }
}

resource "aws_api_gateway_api_key" "key" {
  for_each = var.api_keys

  name        = each.value.key_name
  description = "API Key for ${each.value.key_name}"
  enabled     = each.value.enabled

  tags = {
    Name = each.value.key_name
  }

}

resource "aws_api_gateway_usage_plan_key" "main" {
  for_each = var.api_keys

  key_id        = aws_api_gateway_api_key.key[each.value.key_name].id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.usage_plan.id
}
