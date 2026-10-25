resource "aws_cloudwatch_event_rule" "main" {
  name                = var.name
  name_prefix         = var.name_prefix
  schedule_expression = var.schedule_expression
  event_bus_name      = var.event_bus_name
  event_pattern       = var.event_pattern
  description         = var.description
  role_arn            = var.role_arn
  is_enabled          = var.is_enabled
  tags                = var.tags
}

resource "aws_cloudwatch_event_target" "main" {
  rule           = aws_cloudwatch_event_rule.main.name
  arn            = var.arn
  event_bus_name = var.event_bus_name
  target_id      = var.target_id
  input          = var.input
  input_path     = var.input_path
  role_arn       = var.target_role_arn

  dynamic "run_command_targets" {
    for_each = var.run_command_targets
    content {
      key    = run_command_targets.value.key
      values = run_command_targets.value.values
    }
  }

  dynamic "ecs_target" {
    for_each = var.ecs_target
    content {

      dynamic "capacity_provider_strategy" {
        for_each = try([ecs_target.value.capacity_provider_strategy], [])
        content {
          base              = try(capacity_provider_strategy.value.base, null)
          capacity_provider = capacity_provider_strategy.value.capacity_provider
          weight            = try(capacity_provider_strategy.value.weight, null)
        }
      }

      dynamic "network_configuration" {
        for_each = try([ecs_target.value.network_configuration], [])
        content {
          subnets          = network_configuration.value.subnets
          security_groups  = try(network_configuration.value.security_groups, [])
          assign_public_ip = try(network_configuration.value.assign_public_ip, false)
        }
      }

      dynamic "placement_constraint" {
        for_each = try([ecs_target.value.placement_constraint], [])
        content {
          type       = placement_constraint.value.type
          expression = try(placement_constraint.value.expression, null)
        }
      }

      group                   = try(ecs_target.value.group, null)
      launch_type             = try(ecs_target.value.launch_type, null)
      platform_version        = try(ecs_target.value.platform_version, null)
      task_count              = try(ecs_target.value.task_count, null)
      task_definition_arn     = try(ecs_target.value.task_definition_arn, null)
      tags                    = try(ecs_target.value.tags, null)
      propagate_tags          = try(ecs_target.value.propagate_tags, null)
      enable_execute_command  = try(ecs_target.value.enable_execute_command, null)
      enable_ecs_managed_tags = try(ecs_target.value.enable_ecs_managed_tags, null)
    }
  }

  dynamic "batch_target" {
    for_each = var.batch_target
    content {
      job_definition = batch_target.value.job_definition
      job_name       = batch_target.value.job_name
      array_size     = try(batch_target.value.array_size, null)
      job_attempts   = try(batch_target.value.job_attempts, null)
    }
  }

  dynamic "kinesis_target" {
    for_each = var.kinesis_target
    content {
      partition_key_path = try(kinesis_target.value.partition_key_path, null)
    }
  }

  dynamic "redshift_target" {
    for_each = var.redshift_target
    content {
      database            = redshift_target.value.database
      db_user             = try(redshift_target.value.db_user, null)
      secrets_manager_arn = try(redshift_target.value.secrets_manager_arn, null)
      sql                 = try(redshift_target.value.sql, null)
      statement_name      = try(redshift_target.value.statement_name, null)
      with_event          = try(redshift_target.value.with_event, null)
    }
  }

  dynamic "sqs_target" {
    for_each = var.sqs_target
    content {
      message_group_id = try(sqs_target.value.message_group_id, null)
    }
  }

  dynamic "http_target" {
    for_each = var.http_target
    content {
      path_parameter_values   = try(http_target.value.path_parameter_values, null)
      query_string_parameters = try(http_target.value.query_string_parameters, null)
      header_parameters       = try(http_target.value.header_parameters, null)
    }
  }

  dynamic "input_transformer" {
    for_each = var.input_transformer
    content {
      input_paths    = input_transformer.value.input_paths
      input_template = try(input_transformer.value.input_template, null)
    }
  }

  dynamic "retry_policy" {
    for_each = var.retry_policy
    content {
      maximum_event_age_in_seconds = try(retry_policy.value.maximum_event_age_in_seconds, null)
      maximum_retry_attempts       = try(retry_policy.value.maximum_retry_attempts, null)
    }
  }

  dynamic "dead_letter_config" {
    for_each = var.dead_letter_config
    content {
      arn = try(dead_letter_config.value.arn, null)
    }
  }
}

resource "aws_cloudwatch_event_permission" "main" {
  count          = length(var.event_permissions) > 0 ? length(var.event_permissions) : 0
  principal      = var.event_permissions[count.index]["principal"]
  statement_id   = var.event_permissions[count.index]["statement_id"]
  action         = try(var.event_permissions[count.index]["action"], null)
  event_bus_name = try(var.event_permissions[count.index]["event_bus_name"], null)

  dynamic "condition" {
    for_each = try([var.event_permissions[count.index]["condition"]], [])
    content {
      key   = condition.value.key
      type  = condition.value.type
      value = condition.value.value
    }
  }
}
