# ECS Service
resource "aws_ecs_service" "service" {
  name                               = var.name
  cluster                            = var.cluster
  task_definition                    = join("", aws_ecs_task_definition.this[*].arn)
  desired_count                      = var.desired_count
  deployment_minimum_healthy_percent = var.tasks_minimum_healthy_percent
  deployment_maximum_percent         = var.tasks_maximum_percent
  launch_type                        = length(var.capacity_provider_strategy) > 0 ? null : var.launch_type
  enable_execute_command             = var.enable_execute_command
  force_new_deployment               = var.force_new_deployment
  triggers                           = var.triggers
  propagate_tags                     = var.propagate_tags
  tags                               = var.tags

  deployment_controller {
    type = var.deployment_controller_type
  }

  dynamic "capacity_provider_strategy" {
    for_each = var.capacity_provider_strategy
    content {
      capacity_provider = capacity_provider_strategy.value.capacity_provider
      weight            = capacity_provider_strategy.value.weight
      base              = capacity_provider_strategy.value.base
    }
  }

  dynamic "network_configuration" {
    for_each = var.network_mode == "awsvpc" ? [var.network_configuration] : []
    content {
      subnets          = network_configuration.value.subnets
      assign_public_ip = try(network_configuration.value.assign_public_ip, null)
      security_groups  = [aws_security_group.service.id]
    }
  }

  dynamic "load_balancer" {
    for_each = length(var.load_balancer) == 0 ? [] : [for lb in var.load_balancer : lb]
    content {
      container_name   = load_balancer.value.container_name
      container_port   = load_balancer.value.container_port
      target_group_arn = try(load_balancer.value.target_group_arn, (var.load_balancer_type == "application" ? aws_lb_target_group.main_alb[0].arn : aws_lb_target_group.main_nlb[0].arn))
    }
  }
}

# ECS Task Definition
resource "aws_ecs_task_definition" "this" {
  count                    = local.create_task_definition ? 1 : 0
  family                   = var.family
  task_role_arn            = join("", aws_iam_role.task_role[*].arn)
  execution_role_arn       = join("", aws_iam_role.task_execution_role[*].arn)
  network_mode             = var.network_mode
  requires_compatibilities = var.requires_compatibilities
  cpu                      = var.cpu
  memory                   = var.memory
  volume {
    name = var.volume_name
  }
  container_definitions = var.container_definitions
  tags                  = var.tags
}

# IAM Roles
resource "aws_iam_role" "task_role" {
  count              = local.create_task_definition && var.task_assume_role_policy != "" ? 1 : 0
  name               = "${var.name}-ecs-task-role"
  assume_role_policy = var.task_assume_role_policy
  tags               = var.tags
}

resource "aws_iam_role_policy" "task_role_policy" {
  count  = var.task_role_policy != "" ? 1 : 0
  name   = "${aws_iam_role.task_role[0].name}-policy"
  role   = aws_iam_role.task_role[0].name
  policy = var.task_role_policy
}

resource "aws_iam_role" "task_execution_role" {
  count              = local.create_task_definition && var.task_execution_role_policy != "" ? 1 : 0
  description        = "${var.name} task execution role"
  name               = "${var.name}-task-execution-role"
  assume_role_policy = var.task_execution_assume_role_policy
  tags               = var.tags
}

resource "aws_iam_role_policy" "task_execution_role_policy" {
  count  = var.task_execution_role_policy != "" ? 1 : 0
  name   = "${aws_iam_role.task_execution_role[0].name}-policy"
  role   = aws_iam_role.task_execution_role[0].name
  policy = var.task_execution_role_policy
}

# Cloudwatch Log Group
resource "aws_kms_key" "cloudwatch_log_group" {
  count                   = var.kms_key_id == null ? 1 : 0
  description             = "KMS key for encrypting/decrypting ecs service cloudwatch log group"
  enable_key_rotation     = var.enable_key_rotation
  policy                  = local.kms_policy
  deletion_window_in_days = var.key_deletion_window_in_days
  tags                    = var.tags
}

resource "aws_cloudwatch_log_group" "main" {
  name              = "/aws/ecs-service/${var.name}"
  retention_in_days = var.retention_in_days
  kms_key_id        = var.kms_key_id == null ? aws_kms_key.cloudwatch_log_group[0].arn : var.kms_key_id
  tags              = var.tags
}

# Load Balancer
resource "aws_lb" "main" {
  count                      = var.create_load_balancer ? 1 : 0
  name                       = var.name
  internal                   = var.internal
  idle_timeout               = var.load_balancer_type == "application" ? var.idle_timeout : null
  load_balancer_type         = var.load_balancer_type
  subnets                    = var.alb_subnets
  security_groups            = [aws_security_group.lb[0].id]
  drop_invalid_header_fields = var.drop_invalid_header_fields
  enable_deletion_protection = var.enable_deletion_protection
  tags                       = var.tags

  dynamic "access_logs" {
    for_each = length(keys(var.access_logs)) > 0 ? [var.access_logs] : []
    content {
      bucket  = access_logs.value.bucket
      enabled = lookup(access_logs.value, "enabled", null)
      prefix  = lookup(access_logs.value, "prefix", null)
    }
  }
}

## WAF Association
resource "aws_wafv2_web_acl_association" "main" {
  count        = var.associate_with_waf && var.create_load_balancer ? 1 : 0
  resource_arn = aws_lb.main[0].arn
  web_acl_arn  = var.web_acl_arn
}

resource "aws_wafregional_web_acl_association" "main" {
  count        = var.associate_with_wafregional && var.create_load_balancer && var.load_balancer_type == "application" ? 1 : 0
  resource_arn = aws_lb.main[0].arn
  web_acl_id   = var.wafregional_acl_id
}

# lb target group
resource "aws_lb_target_group" "main_alb" {
  count       = var.load_balancer_type == "application" && var.create_load_balancer ? 1 : 0
  name        = var.name
  port        = var.tg_port
  protocol    = var.tg_protocol
  target_type = var.target_type
  vpc_id      = var.vpc_id
  health_check {
    matcher           = var.matcher
    path              = var.path
    protocol          = var.tg_protocol
    healthy_threshold = var.healthy_threshold
    interval          = var.interval
  }

  depends_on = [aws_lb.main]
  tags       = var.tags
}

resource "aws_lb_target_group" "main_nlb" {
  count       = var.load_balancer_type == "network" && var.create_load_balancer ? 1 : 0
  name        = var.name
  port        = var.tg_port
  protocol    = var.tg_protocol
  target_type = var.target_type
  vpc_id      = var.vpc_id
  health_check {
    protocol          = var.tg_protocol
    healthy_threshold = var.healthy_threshold
    interval          = var.interval
  }
  tags       = var.tags
  depends_on = [aws_lb.main]
}

#http redirect listener
resource "aws_lb_listener" "http_redirect" {
  count             = var.load_balancer_type == "application" && var.create_load_balancer ? 1 : 0
  load_balancer_arn = aws_lb.main[0].id
  port              = var.listener_port
  protocol          = var.listener_protocol

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
  tags = var.tags
  lifecycle {
    create_before_destroy = true
  }
}

## Forward redirected traffic to target group
resource "aws_lb_listener" "https" {
  count             = var.load_balancer_type == "application" && var.create_load_balancer ? 1 : 0
  load_balancer_arn = aws_lb.main[0].id
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = var.ssl_policy
  certificate_arn   = var.acm_certificate_arn != null ? var.acm_certificate_arn : aws_acm_certificate.main[0].arn

  default_action {
    type             = var.default_type
    target_group_arn = aws_lb_target_group.main_alb[0].arn
  }
  tags = var.tags
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lb_listener" "nlb" {
  count             = var.load_balancer_type == "network" && var.create_load_balancer ? 1 : 0
  load_balancer_arn = aws_lb.main[0].id
  port              = var.tg_port
  protocol          = "TLS"
  ssl_policy        = var.ssl_policy
  certificate_arn   = var.acm_certificate_arn != null ? var.acm_certificate_arn : aws_acm_certificate.main[0].arn
  default_action {
    type             = var.default_type
    target_group_arn = aws_lb_target_group.main_nlb[0].arn
  }
  tags = var.tags
  lifecycle {
    create_before_destroy = true
  }
}


# NOTE: Self-signed certificates are usually used only in development environments or applications deployed internally to an organization.
# Please use ACM generated certificate in production. Specify the value of `acm_certificate_arn` to provide this
resource "tls_private_key" "default" {
  count     = var.create_load_balancer && var.acm_certificate_arn == null ? 1 : 0
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "default" {
  count           = var.create_load_balancer && var.acm_certificate_arn == null ? 1 : 0
  private_key_pem = tls_private_key.default[0].private_key_pem

  subject {
    common_name  = var.self_signed_cert_common_name
    organization = var.self_signed_cert_organization
  }

  validity_period_hours = 72

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "main" {
  count            = var.create_load_balancer && var.acm_certificate_arn == null ? 1 : 0
  private_key      = tls_private_key.default[0].private_key_pem
  certificate_body = tls_self_signed_cert.default[0].cert_pem

  lifecycle {
    create_before_destroy = true
  }
  tags = var.tags
}

# Alb Security group
resource "aws_security_group" "lb" {
  count                  = var.create_load_balancer ? 1 : 0
  name                   = "${var.name}-lb-security-group"
  vpc_id                 = var.vpc_id
  description            = "${var.name} Load balancer security group"
  revoke_rules_on_delete = true
  tags                   = var.tags

  lifecycle {
    # Necessary if changing 'name' or 'name_prefix' properties.
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "lb_ingress" {
  count                    = var.create_load_balancer && length(var.lb_ingress_rules) > 0 ? length(var.lb_ingress_rules) : 0
  security_group_id        = aws_security_group.lb[0].id
  type                     = "ingress"
  description              = try(var.lb_ingress_rules[count.index]["description"], null)
  from_port                = try(var.lb_ingress_rules[count.index]["from_port"], null)
  protocol                 = try(var.lb_ingress_rules[count.index]["protocol"], null)
  to_port                  = try(var.lb_ingress_rules[count.index]["to_port"], null)
  cidr_blocks              = try(var.lb_ingress_rules[count.index]["cidr_blocks"], [])
  source_security_group_id = try(var.lb_ingress_rules[count.index]["cidr_blocks"], null) == null ? try(var.lb_ingress_rules[count.index]["source_security_group_id"], null) : null
}

resource "aws_security_group_rule" "lb_egress" {
  count             = var.create_load_balancer ? 1 : 0
  security_group_id = aws_security_group.lb[0].id
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = "0"
  to_port           = "0"
  description       = "${var.name} Load Balancer security group egress rule"
  protocol          = "-1"
  type              = "egress"
}

# Service Security group
resource "aws_security_group" "service" {
  name                   = "${var.name}-security-group"
  vpc_id                 = var.vpc_id
  description            = "${var.name} Service security group"
  revoke_rules_on_delete = true
  tags                   = var.tags

  lifecycle {
    # Necessary if changing 'name' or 'name_prefix' properties.
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "service_with_lb_ingress" {
  count                    = var.create_load_balancer && length(var.lb_ingress_rules) > 0 ? 1 : 0
  security_group_id        = aws_security_group.service.id
  type                     = "ingress"
  description              = "Security group for ${var.name} ecs service accessible through a load-balancer"
  source_security_group_id = aws_security_group.lb[0].id
  from_port                = try(var.load_balancer[count.index]["container_port"], null)
  to_port                  = try(var.load_balancer[count.index]["container_port"], null)
  protocol                 = "-1"
}

resource "aws_security_group_rule" "service_ingress" {
  count             = var.create_load_balancer && length(var.lb_ingress_rules) > 0 ? 0 : length(var.service_ingress_rules)
  security_group_id = aws_security_group.service.id
  type              = "ingress"
  description       = try(var.service_ingress_rules[count.index]["description"], null)
  from_port         = try(var.service_ingress_rules[count.index]["from_port"], null)
  protocol          = try(var.service_ingress_rules[count.index]["protocol"], null)
  to_port           = try(var.service_ingress_rules[count.index]["to_port"], null)
  cidr_blocks       = try(var.service_ingress_rules[count.index]["cidr_blocks"], [])
  self              = try(var.service_ingress_rules[count.index]["self"], null)
}

resource "aws_security_group_rule" "service_ingress_sg" {
  count                    = var.create_load_balancer && length(var.lb_ingress_rules) > 0 ? 0 : length(var.service_ingress_rules_sg)
  security_group_id        = aws_security_group.service.id
  type                     = "ingress"
  description              = try(var.service_ingress_rules_sg[count.index]["description"], null)
  from_port                = try(var.service_ingress_rules_sg[count.index]["from_port"], null)
  protocol                 = try(var.service_ingress_rules_sg[count.index]["protocol"], null)
  to_port                  = try(var.service_ingress_rules_sg[count.index]["to_port"], null)
  source_security_group_id = try(var.service_ingress_rules_sg[count.index]["source_security_group_id"], null)
  self                     = try(var.service_ingress_rules_sg[count.index]["self"], null)
}

resource "aws_security_group_rule" "service_egress" {
  security_group_id = aws_security_group.service.id
  type              = "egress"
  from_port         = "0"
  to_port           = "0"
  cidr_blocks       = ["0.0.0.0/0"]
  description       = "${var.name} service security group egress rule"
  protocol          = "-1"
}

# Autoscaling
resource "aws_appautoscaling_target" "this" {
  count              = var.enable_autoscaling ? 1 : 0
  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${var.cluster}/${aws_ecs_service.service.name}"
  role_arn           = var.autoscale_role_arn
  scalable_dimension = var.scalable_dimension
  service_namespace  = var.service_namespace
  depends_on = [
    aws_ecs_service.service
  ]
  tags = var.tags
}

resource "aws_appautoscaling_policy" "stepscaling" {
  count              = var.enable_autoscaling ? length(var.step_scaling_policies) : 0
  policy_type        = var.step_scaling_policies[count.index].policy_type
  name               = "${var.name}-${var.step_scaling_policies[count.index].name}"
  service_namespace  = var.service_namespace
  resource_id        = aws_appautoscaling_target.this[0].resource_id
  scalable_dimension = aws_appautoscaling_target.this[0].scalable_dimension

  dynamic "step_scaling_policy_configuration" {
    for_each = var.step_scaling_policies[count.index].step_scaling_policy_configuration != null ? [var.step_scaling_policies[count.index].step_scaling_policy_configuration] : []
    content {
      adjustment_type         = step_scaling_policy_configuration.value.adjustment_type
      cooldown                = step_scaling_policy_configuration.value.cooldown
      metric_aggregation_type = step_scaling_policy_configuration.value.metric_aggregation_type

      dynamic "step_adjustment" {
        for_each = step_scaling_policy_configuration.value.step_adjustments
        content {
          metric_interval_lower_bound = step_adjustment.value.metric_interval_lower_bound
          metric_interval_upper_bound = step_adjustment.value.metric_interval_upper_bound
          scaling_adjustment          = step_adjustment.value.scaling_adjustment
        }
      }
    }
  }
  depends_on = [
    aws_appautoscaling_target.this
  ]
}

resource "aws_appautoscaling_policy" "targetscaling" {
  count              = var.enable_autoscaling ? length(var.target_scaling_policies) : 0
  policy_type        = var.target_scaling_policies[count.index].policy_type
  name               = "${var.name}-${var.target_scaling_policies[count.index].name}"
  service_namespace  = var.service_namespace
  resource_id        = aws_appautoscaling_target.this[0].resource_id
  scalable_dimension = aws_appautoscaling_target.this[0].scalable_dimension

  target_tracking_scaling_policy_configuration {
    target_value       = var.target_scaling_policies[count.index].target_tracking_scaling_policy_configuration.target_value
    scale_in_cooldown  = var.target_scaling_policies[count.index].target_tracking_scaling_policy_configuration.scale_in_cooldown
    scale_out_cooldown = var.target_scaling_policies[count.index].target_tracking_scaling_policy_configuration.scale_out_cooldown
    predefined_metric_specification {
      predefined_metric_type = var.target_scaling_policies[count.index].target_tracking_scaling_policy_configuration.predefined_metric_specification.predefined_metric_type
    }
  }

  depends_on = [
    aws_appautoscaling_target.this
  ]
}

resource "aws_appautoscaling_scheduled_action" "this" {
  for_each           = { for action in var.scheduled_actions : action.name => action }
  name               = each.value.name
  service_namespace  = aws_appautoscaling_target.this[0].service_namespace
  resource_id        = aws_appautoscaling_target.this[0].resource_id
  scalable_dimension = aws_appautoscaling_target.this[0].scalable_dimension
  schedule           = each.value.schedule
  timezone           = each.value.timezone

  scalable_target_action {
    min_capacity = each.value.min_capacity
    max_capacity = each.value.max_capacity
  }
}
