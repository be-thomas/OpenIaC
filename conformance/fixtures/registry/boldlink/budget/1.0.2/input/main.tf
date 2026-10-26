resource "aws_budgets_budget" "main" {
  name              = var.name
  name_prefix       = var.name_prefix
  account_id        = var.account_id
  budget_type       = var.budget_type
  limit_amount      = length(keys(var.auto_adjust_data)) > 0 ? null : var.limit_amount
  limit_unit        = var.limit_unit
  time_period_end   = var.time_period_end
  time_period_start = var.time_period_start
  time_unit         = var.time_unit

  dynamic "auto_adjust_data" {
    for_each = length(keys(var.auto_adjust_data)) > 0 ? [var.auto_adjust_data] : []
    content {
      auto_adjust_type = lookup(var.auto_adjust_data, "auto_adjust_type", null)

      dynamic "historical_options" {
        for_each = try([var.auto_adjust_data["historical_options"]], [])
        content {
          budget_adjustment_period = try(historical_options.value.budget_adjustment_period, null)
        }
      }
    }
  }

  # RI Utilization plans require a service cost filter to be set
  dynamic "cost_filter" {
    for_each = var.cost_filter
    content {
      name   = try(cost_filter.value.name, null)
      values = try(cost_filter.value.values, null)
    }
  }

  dynamic "planned_limit" {
    for_each = var.limit_amount != null || var.limit_unit != null ? [] : var.planned_limit
    content {
      start_time = planned_limit.value.start_time
      amount     = planned_limit.value.amount
      unit       = planned_limit.value.unit
    }
  }

  # Cost types must be defined for RI budgets because the settings conflict with the defaults
  dynamic "cost_types" {
    for_each = length(keys(var.cost_types)) > 0 ? [var.cost_types] : [] #var.cost_types
    content {
      include_credit             = try(cost_types.value.include_credit, true)
      include_discount           = try(cost_types.value.include_discount, true)
      include_other_subscription = try(cost_types.value.include_other_subscription, true)
      include_recurring          = try(cost_types.value.include_recurring, true)
      include_refund             = try(cost_types.value.include_refund, true)
      include_subscription       = try(cost_types.value.include_subscription, true)
      include_support            = try(cost_types.value.include_support, true)
      include_tax                = try(cost_types.value.include_tax, true)
      include_upfront            = try(cost_types.value.include_upfront, true)
      use_amortized              = try(cost_types.value.use_amortized, false)
      use_blended                = try(cost_types.value.use_blended, false)
    }
  }

  dynamic "notification" {
    for_each = var.notification
    content {
      comparison_operator        = notification.value.comparison_operator
      threshold                  = notification.value.threshold
      threshold_type             = notification.value.threshold_type
      notification_type          = notification.value.notification_type
      subscriber_email_addresses = try(notification.value.subscriber_email_addresses, null)
      subscriber_sns_topic_arns  = try(notification.value.subscriber_sns_topic_arns, null)
    }
  }
}
