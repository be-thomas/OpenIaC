### Cloudwatch resources
resource "aws_kms_key" "cloudwatch" {
  count                   = var.install_cloudwatch_agent && var.create_kms_key ? 1 : 0
  description             = "${var.name} Log Group KMS key"
  enable_key_rotation     = var.enable_key_rotation
  policy                  = local.kms_policy
  deletion_window_in_days = var.key_deletion_window_in_days
  tags                    = var.tags
}

resource "aws_cloudwatch_log_group" "main" {
  count             = var.install_cloudwatch_agent ? 1 : 0
  name              = "/aws/asg/${var.name}"
  retention_in_days = var.retention_in_days
  kms_key_id        = var.create_kms_key ? aws_kms_key.cloudwatch[0].arn : var.kms_key_id
  tags              = var.tags
}

### IAM Resources
resource "aws_iam_instance_profile" "main" {
  count = var.create_instance_profile ? 1 : 0
  name  = "${var.name}-iam-role"
  path  = var.iam_role_path
  role  = aws_iam_role.main[0].name
}

resource "aws_iam_role" "main" {
  count              = var.create_instance_profile ? 1 : 0
  description        = "${var.name} EC2 IAM Role"
  name               = "${var.name}-iam-role"
  path               = var.iam_role_path
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_policy" "main" {
  count       = var.create_instance_profile ? 1 : 0
  name        = "${var.name}-iam-policy"
  description = "${var.name} EC2 IAM role policy"
  path        = var.iam_role_path
  policy      = data.aws_iam_policy_document.asg.json
}

resource "aws_iam_role_policy_attachment" "main" {
  count      = var.create_instance_profile ? 1 : 0
  role       = aws_iam_role.main[0].name
  policy_arn = aws_iam_policy.main[0].arn
}

### For adding custom permissions to the role created above
resource "aws_iam_policy" "additional" {
  count       = var.additional_role_policy_document != null && var.create_instance_profile ? 1 : 0
  name        = "${var.name}-additional-policy"
  description = "${var.name} additional IAM role policy"
  path        = var.iam_role_path
  policy      = var.additional_role_policy_document
}

resource "aws_iam_role_policy_attachment" "additional" {
  count      = var.additional_role_policy_document != null && var.create_instance_profile ? 1 : 0
  role       = aws_iam_role.main[0].name
  policy_arn = aws_iam_policy.additional[0].arn
}

## Managed Policy to allow cloudwatch agent to write metrics to CloudWatch
resource "aws_iam_role_policy_attachment" "cloudwatchagentserverpolicy" {
  count      = var.create_instance_profile && var.install_cloudwatch_agent ? 1 : 0
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
  role       = aws_iam_role.main[0].name
}

## Managed Policy to allow ssm agent to communicate with SSM Manager
resource "aws_iam_role_policy_attachment" "ssm" {
  count      = var.create_instance_profile && var.install_ssm_agent ? 1 : 0
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.main[0].name
}

## Configure CloudWatch agent to set the retention policy for log groups that it sends log events to.
resource "aws_iam_role_policy" "logs_policy" {
  count = var.create_instance_profile && var.install_cloudwatch_agent ? 1 : 0
  name  = "CloudWatchAgentPutLogsRetention"
  role  = aws_iam_role.main[0].name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:PutRetentionPolicy",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

### Security Group
resource "aws_security_group" "main" {
  count       = var.create_launch_template ? 1 : 0
  name        = var.name
  description = "ASG Group Security Group"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.security_group_ingress
    content {
      description      = "Rule to allow port ${try(ingress.value.from_port, "")} inbound traffic"
      from_port        = try(ingress.value.from_port, null)
      to_port          = try(ingress.value.to_port, null)
      protocol         = try(ingress.value.protocol, null)
      cidr_blocks      = try(ingress.value.cidr_blocks, [])
      ipv6_cidr_blocks = try(ingress.value.ipv6_cidr_blocks, [])
    }
  }

  dynamic "egress" {
    for_each = var.security_group_egress
    content {
      description = "Rule to allow outbound traffic"
      from_port   = try(egress.value.from_port, 0)
      to_port     = try(egress.value.to_port, 0)
      protocol    = try(egress.value.protocol, -1)
      cidr_blocks = try(egress.value.cidr_blocks, ["0.0.0.0/0"])
    }
  }

  tags = var.tags
}

### ASG resources
resource "aws_autoscaling_group" "main" {
  name                 = var.name
  name_prefix          = var.name_prefix
  max_size             = var.max_size
  min_size             = var.min_size
  availability_zones   = var.vpc_zone_identifier != null ? null : var.availability_zones
  capacity_rebalance   = var.capacity_rebalance
  default_cooldown     = var.default_cooldown
  launch_configuration = var.launch_configuration

  dynamic "launch_template" {
    for_each = var.use_mixed_instances_policy ? [] : [1]

    content {
      name    = local.launch_template
      version = local.launch_template_version
    }
  }

  dynamic "mixed_instances_policy" {
    for_each = var.use_mixed_instances_policy ? [var.mixed_instances_policy] : []
    content {
      dynamic "instances_distribution" {
        for_each = try([mixed_instances_policy.value.instances_distribution], [])
        content {
          on_demand_allocation_strategy            = try(instances_distribution.value.on_demand_allocation_strategy, null)
          on_demand_base_capacity                  = try(instances_distribution.value.on_demand_base_capacity, null)
          on_demand_percentage_above_base_capacity = try(instances_distribution.value.on_demand_percentage_above_base_capacity, null)
          spot_allocation_strategy                 = try(instances_distribution.value.spot_allocation_strategy, null)
          spot_instance_pools                      = try(instances_distribution.value.spot_instance_pools, null)
          spot_max_price                           = try(instances_distribution.value.spot_max_price, null)
        }
      }

      launch_template {
        launch_template_specification {
          launch_template_id = local.launch_template_id
          version            = local.launch_template_version
        }

        dynamic "override" {
          for_each = try(mixed_instances_policy.value.override, [])

          content {
            dynamic "instance_requirements" {
              for_each = try([override.value.instance_requirements], [])

              content {
                dynamic "accelerator_count" {
                  for_each = try([instance_requirements.value.accelerator_count], [])

                  content {
                    max = try(accelerator_count.value.max, null)
                    min = try(accelerator_count.value.min, null)
                  }
                }

                accelerator_manufacturers = try(instance_requirements.value.accelerator_manufacturers, null)
                accelerator_names         = try(instance_requirements.value.accelerator_names, null)

                dynamic "accelerator_total_memory_mib" {
                  for_each = try([instance_requirements.value.accelerator_total_memory_mib], [])

                  content {
                    max = try(accelerator_total_memory_mib.value.max, null)
                    min = try(accelerator_total_memory_mib.value.min, null)
                  }
                }

                accelerator_types      = try(instance_requirements.value.accelerator_types, null)
                allowed_instance_types = try(instance_requirements.value.allowed_instance_types, null)
                bare_metal             = try(instance_requirements.value.bare_metal, null)

                dynamic "baseline_ebs_bandwidth_mbps" {
                  for_each = try([instance_requirements.value.baseline_ebs_bandwidth_mbps], [])

                  content {
                    max = try(baseline_ebs_bandwidth_mbps.value.max, null)
                    min = try(baseline_ebs_bandwidth_mbps.value.min, null)
                  }
                }

                burstable_performance   = try(instance_requirements.value.burstable_performance, null)
                cpu_manufacturers       = try(instance_requirements.value.cpu_manufacturers, null)
                excluded_instance_types = try(instance_requirements.value.excluded_instance_types, null)
                instance_generations    = try(instance_requirements.value.instance_generations, null)
                local_storage           = try(instance_requirements.value.local_storage, null)
                local_storage_types     = try(instance_requirements.value.local_storage_types, null)

                dynamic "memory_gib_per_vcpu" {
                  for_each = try([instance_requirements.value.memory_gib_per_vcpu], [])

                  content {
                    max = try(memory_gib_per_vcpu.value.max, null)
                    min = try(memory_gib_per_vcpu.value.min, null)
                  }
                }

                dynamic "memory_mib" {
                  for_each = try([instance_requirements.value.memory_mib], [])

                  content {
                    max = try(memory_mib.value.max, null)
                    min = try(memory_mib.value.min, null)
                  }
                }

                dynamic "network_bandwidth_gbps" {
                  for_each = try([instance_requirements.value.network_bandwidth_gbps], [])

                  content {
                    max = try(network_bandwidth_gbps.value.max, null)
                    min = try(network_bandwidth_gbps.value.min, null)
                  }
                }

                dynamic "network_interface_count" {
                  for_each = try([instance_requirements.value.network_interface_count], [])

                  content {
                    max = try(network_interface_count.value.max, null)
                    min = try(network_interface_count.value.min, null)
                  }
                }

                on_demand_max_price_percentage_over_lowest_price = try(instance_requirements.value.on_demand_max_price_percentage_over_lowest_price, null)
                require_hibernate_support                        = try(instance_requirements.value.require_hibernate_support, null)
                spot_max_price_percentage_over_lowest_price      = try(instance_requirements.value.spot_max_price_percentage_over_lowest_price, null)

                dynamic "total_local_storage_gb" {
                  for_each = try([instance_requirements.value.total_local_storage_gb], [])

                  content {
                    max = try(total_local_storage_gb.value.max, null)
                    min = try(total_local_storage_gb.value.min, null)
                  }
                }

                dynamic "vcpu_count" {
                  for_each = try([instance_requirements.value.vcpu_count], [])

                  content {
                    max = try(vcpu_count.value.max, null)
                    min = try(vcpu_count.value.min, null)
                  }
                }
              }
            }

            instance_type = try(override.value.instance_type, null)

            dynamic "launch_template_specification" {
              for_each = try([override.value.launch_template_specification], [])

              content {
                launch_template_id = try(launch_template_specification.value.launch_template_id, null)
              }
            }

            weighted_capacity = try(override.value.weighted_capacity, null)
          }
        }
      }
    }
  }

  dynamic "initial_lifecycle_hook" {
    for_each = var.initial_lifecycle_hooks
    content {
      name                    = initial_lifecycle_hook.value.name
      default_result          = try(initial_lifecycle_hook.value.default_result, null)
      heartbeat_timeout       = try(initial_lifecycle_hook.value.heartbeat_timeout, null)
      lifecycle_transition    = initial_lifecycle_hook.value.lifecycle_transition
      notification_metadata   = try(initial_lifecycle_hook.value.notification_metadata, null)
      notification_target_arn = try(initial_lifecycle_hook.value.notification_target_arn, null)
      role_arn                = try(initial_lifecycle_hook.value.role_arn, null)
    }
  }

  health_check_grace_period = var.health_check_grace_period
  health_check_type         = var.health_check_type
  desired_capacity          = var.desired_capacity
  desired_capacity_type     = var.desired_capacity_type
  force_delete              = var.force_delete
  load_balancers            = var.load_balancers
  vpc_zone_identifier       = var.vpc_zone_identifier
  target_group_arns         = var.target_group_arns
  termination_policies      = var.termination_policies
  suspended_processes       = var.suspended_processes
  placement_group           = var.placement_group
  metrics_granularity       = var.metrics_granularity
  enabled_metrics           = var.enabled_metrics
  wait_for_capacity_timeout = var.wait_for_capacity_timeout
  min_elb_capacity          = var.min_elb_capacity
  wait_for_elb_capacity     = var.wait_for_elb_capacity
  protect_from_scale_in     = var.protect_from_scale_in
  service_linked_role_arn   = var.service_linked_role_arn
  max_instance_lifetime     = var.max_instance_lifetime

  dynamic "instance_refresh" {
    for_each = length(var.instance_refresh) > 0 ? [var.instance_refresh] : []
    content {
      strategy = instance_refresh.value.strategy
      dynamic "preferences" {
        for_each = try([instance_refresh.value.preferences], [])
        content {
          checkpoint_delay             = try(preferences.value.checkpoint_delay, null)
          checkpoint_percentages       = try(preferences.value.checkpoint_percentages, null)
          instance_warmup              = try(preferences.value.instance_warmup, null)
          min_healthy_percentage       = try(preferences.value.min_healthy_percentage, null)
          skip_matching                = try(preferences.value.skip_matching, null)
          auto_rollback                = try(preferences.value.auto_rollback, null)
          scale_in_protected_instances = try(preferences.value.scale_in_protected_instances, null)
          standby_instances            = try(preferences.value.standby_instances, null)
        }
      }
      triggers = try(instance_refresh.value.triggers, null)
    }
  }

  dynamic "warm_pool" {
    for_each = length(var.warm_pool) > 0 ? [var.warm_pool] : []
    content {
      pool_state = lookup(warm_pool.value, "pool_state", null)
      min_size   = lookup(warm_pool.value, "min_size", null)
      dynamic "instance_reuse_policy" {
        for_each = lookup(warm_pool.value, "instance_reuse_policy", [])
        content {
          reuse_on_scale_in = try(instance_reuse_policy.value.reuse_on_scale_in, null)
        }
      }
    }
  }

  timeouts {
    delete = try(var.timeouts["delete"], "10m")
  }

  tag {
    key                 = "Name"
    value               = var.name
    propagate_at_launch = true
  }

  dynamic "tag" {
    for_each = var.tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      name,
      id,
    ]
  }
}

### Launch Template
resource "aws_launch_template" "main" {
  count                                = var.create_launch_template && var.external_launch_template_name == null ? 1 : 0
  name                                 = local.launch_template
  name_prefix                          = var.launch_template_name_prefix
  description                          = var.launch_template_description
  ebs_optimized                        = var.ebs_optimized
  image_id                             = var.image_id
  instance_type                        = var.instance_type
  user_data                            = var.install_ssm_agent || var.install_cloudwatch_agent ? data.template_cloudinit_config.config.rendered : var.user_data
  vpc_security_group_ids               = length(var.network_interfaces) > 0 ? [] : compact(concat(length(aws_security_group.main) > 0 ? [aws_security_group.main[0].id] : [], var.security_group_ids))
  default_version                      = var.default_version
  update_default_version               = var.update_default_version
  disable_api_termination              = var.disable_api_termination
  instance_initiated_shutdown_behavior = var.instance_initiated_shutdown_behavior
  kernel_id                            = var.kernel_id
  ram_disk_id                          = var.ram_disk_id

  dynamic "block_device_mappings" {
    for_each = var.block_device_mappings
    content {
      device_name  = block_device_mappings.value.device_name
      no_device    = try(block_device_mappings.value.no_device, null)
      virtual_name = try(block_device_mappings.value.virtual_name, null)

      dynamic "ebs" {
        for_each = flatten([try(block_device_mappings.value.ebs, [])])
        content {
          delete_on_termination = try(ebs.value.delete_on_termination, null)
          encrypted             = try(ebs.value.encrypted, null)
          kms_key_id            = try(ebs.value.kms_key_id, ebs.value.kms_key_arn, null)
          iops                  = try(ebs.value.iops, null)
          throughput            = try(ebs.value.throughput, null)
          snapshot_id           = try(ebs.value.snapshot_id, null)
          volume_size           = try(ebs.value.volume_size, null)
          volume_type           = try(ebs.value.volume_type, null)
        }
      }
    }
  }

  dynamic "capacity_reservation_specification" {
    for_each = length(var.capacity_reservation_specification) > 0 ? [var.capacity_reservation_specification] : []
    content {
      capacity_reservation_preference = try(capacity_reservation_specification.value.capacity_reservation_preference, null)

      dynamic "capacity_reservation_target" {
        for_each = try([capacity_reservation_specification.value.capacity_reservation_target], [])
        content {
          capacity_reservation_id                 = try(capacity_reservation_target.value.capacity_reservation_id, null)
          capacity_reservation_resource_group_arn = try(capacity_reservation_target.value.capacity_reservation_resource_group_arn, null)
        }
      }
    }
  }

  dynamic "cpu_options" {
    for_each = length(var.cpu_options) > 0 ? [var.cpu_options] : []
    content {
      core_count       = cpu_options.value.core_count
      threads_per_core = cpu_options.value.threads_per_core
    }
  }

  dynamic "credit_specification" {
    for_each = length(var.credit_specification) > 0 ? [var.credit_specification] : []
    content {
      cpu_credits = credit_specification.value.cpu_credits
    }
  }

  dynamic "enclave_options" {
    for_each = length(var.enclave_options) > 0 ? [var.enclave_options] : []
    content {
      enabled = enclave_options.value.enabled
    }
  }

  dynamic "hibernation_options" {
    for_each = length(var.hibernation_options) > 0 ? [var.hibernation_options] : []
    content {
      configured = hibernation_options.value.configured
    }
  }

  dynamic "instance_market_options" {
    for_each = length(var.instance_market_options) > 0 ? [var.instance_market_options] : []
    content {
      market_type = instance_market_options.value.market_type

      dynamic "spot_options" {
        for_each = try([instance_market_options.value.spot_options], [])
        content {
          block_duration_minutes         = try(spot_options.value.block_duration_minutes, null)
          instance_interruption_behavior = try(spot_options.value.instance_interruption_behavior, null)
          max_price                      = try(spot_options.value.max_price, null)
          spot_instance_type             = try(spot_options.value.spot_instance_type, null)
          valid_until                    = try(spot_options.value.valid_until, null)
        }
      }
    }
  }

  dynamic "license_specification" {
    for_each = length(var.license_specifications) > 0 ? [var.license_specifications] : []
    content {
      license_configuration_arn = license_specifications.value.license_configuration_arn
    }
  }

  metadata_options {
    http_endpoint               = lookup(var.metadata_options, "http_endpoint", "enabled")
    http_put_response_hop_limit = lookup(var.metadata_options, "http_put_response_hop_limit", 10)
    http_tokens                 = lookup(var.metadata_options, "http_tokens", "required")
    http_protocol_ipv6          = lookup(var.metadata_options, "http_protocol_ipv6", null)
    instance_metadata_tags      = lookup(var.metadata_options, "instance_metadata_tags", null)
  }

  monitoring {
    enabled = var.enable_monitoring
  }

  iam_instance_profile {
    name = var.create_instance_profile ? aws_iam_instance_profile.main[0].name : var.iam_instance_profile
  }

  dynamic "network_interfaces" {
    for_each = var.network_interfaces
    content {
      associate_carrier_ip_address = try(network_interfaces.value.associate_carrier_ip_address, null)
      associate_public_ip_address  = try(network_interfaces.value.associate_public_ip_address, null)
      delete_on_termination        = try(network_interfaces.value.delete_on_termination, false)
      description                  = try(network_interfaces.value.description, null)
      device_index                 = try(network_interfaces.value.device_index, null)
      interface_type               = try(network_interfaces.value.interface_type, null)
      ipv4_prefix_count            = try(network_interfaces.value.ipv4_prefixes, null) != null ? null : try(network_interfaces.value.ipv4_prefix_count, null)
      ipv4_prefixes                = try(network_interfaces.value.ipv4_prefixes, null)
      ipv4_addresses               = try(network_interfaces.value.ipv4_addresses, null)
      ipv4_address_count           = try(network_interfaces.value.ipv4_addresses, null) != null ? null : try(network_interfaces.value.ipv4_address_count, null)
      ipv6_prefix_count            = try(network_interfaces.value.ipv6_prefixes, []) != [] ? null : try(network_interfaces.value.ipv6_prefix_count, null)
      ipv6_prefixes                = try(network_interfaces.value.ipv6_prefixes, [])
      ipv6_addresses               = try(network_interfaces.value.ipv6_addresses, null)
      ipv6_address_count           = try(network_interfaces.value.ipv6_addresses, null) != null ? null : try(network_interfaces.value.ipv6_address_count, null)
      network_interface_id         = try(network_interfaces.value.network_interface_id, null)
      network_card_index           = try(network_interfaces.value.network_card_index, null)
      private_ip_address           = try(network_interfaces.value.private_ip_address, null)
      security_groups              = compact(concat(length(aws_security_group.main) > 0 ? [aws_security_group.main[0].id] : [], var.security_group_ids))
      subnet_id                    = try(network_interfaces.value.subnet_id, null)
    }
  }

  dynamic "placement" {
    for_each = length(var.placement) > 0 ? [var.placement] : []
    content {
      affinity                = lookup(placement.value, "affinity", null)
      availability_zone       = lookup(placement.value, "availability_zone", null)
      group_name              = lookup(placement.value, "group_name", null)
      host_id                 = lookup(placement.value, "host_id", null)
      host_resource_group_arn = lookup(placement.value, "host_resource_group_arn", null)
      spread_domain           = lookup(placement.value, "spread_domain", null)
      tenancy                 = lookup(placement.value, "tenancy", null)
      partition_number        = lookup(placement.value, "partition_number", null)
    }
  }

  dynamic "private_dns_name_options" {
    for_each = length(var.private_dns_name_options) > 0 ? [var.private_dns_name_options] : []
    content {
      enable_resource_name_dns_aaaa_record = lookup(private_dns_name_options.value, "enable_resource_name_dns_aaaa_record", null)
      enable_resource_name_dns_a_record    = lookup(private_dns_name_options.value, "enable_resource_name_dns_a_record", null)
      hostname_type                        = private_dns_name_options.value.hostname_type
    }
  }

  dynamic "tag_specifications" {
    for_each = var.tag_specifications
    content {
      resource_type = tag_specifications.value.resource_type
      tags          = var.tags
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_schedule" "main" {
  for_each               = var.schedules
  scheduled_action_name  = each.key
  autoscaling_group_name = aws_autoscaling_group.main.name
  min_size               = lookup(each.value, "min_size", null)
  max_size               = lookup(each.value, "max_size", null)
  desired_capacity       = lookup(each.value, "desired_capacity", null)
  start_time             = lookup(each.value, "start_time", null)
  end_time               = lookup(each.value, "end_time", null)
  time_zone              = lookup(each.value, "time_zone", null)
  recurrence             = lookup(each.value, "recurrence", null)
  # Syntax=>>[Minute] [Hour] [Day_of_Month] [Month_of_Year] [Day_of_Week]
}

## Autoscaling Policies Resources
## Below are the resources to trigger autoscaling events and report to an email address (default)
resource "aws_autoscaling_policy" "main" {
  for_each                  = var.autoscaling_policy
  name                      = try(each.value.name, each.key)
  autoscaling_group_name    = aws_autoscaling_group.main.name
  adjustment_type           = lookup(each.value, "adjustment_type", null)
  policy_type               = lookup(each.value, "policy_type", null)
  estimated_instance_warmup = lookup(each.value, "estimated_instance_warmup", null)
  min_adjustment_magnitude  = lookup(each.value, "min_adjustment_magnitude", null)
  cooldown                  = lookup(each.value, "cooldown", null)
  scaling_adjustment        = lookup(each.value, "scaling_adjustment", null)
  metric_aggregation_type   = lookup(each.value, "metric_aggregation_type", null)

  dynamic "step_adjustment" {
    for_each = try([each.value.step_adjustment], [])
    content {
      scaling_adjustment          = step_adjustment.value.scaling_adjustment
      metric_interval_lower_bound = lookup(step_adjustment.value, "metric_interval_lower_bound", null)
      metric_interval_upper_bound = lookup(step_adjustment.value, "metric_interval_upper_bound", null)
    }
  }

  dynamic "target_tracking_configuration" {
    for_each = try([each.value.target_tracking_configuration], [])
    content {
      target_value     = target_tracking_configuration.value.target_value
      disable_scale_in = try(target_tracking_configuration.value.disable_scale_in, null)

      dynamic "predefined_metric_specification" {
        for_each = try([target_tracking_configuration.value.predefined_metric_specification], [])
        content {
          predefined_metric_type = predefined_metric_specification.value.predefined_metric_type
          resource_label         = try(predefined_metric_specification.value.resource_label, null)
        }
      }

      dynamic "customized_metric_specification" {
        for_each = try([target_tracking_configuration.value.customized_metric_specification], [])

        content {

          dynamic "metric_dimension" {
            for_each = try([customized_metric_specification.value.metric_dimension], [])
            content {
              name  = metric_dimension.value.name
              value = metric_dimension.value.value
            }
          }

          metric_name = try(customized_metric_specification.value.metric_name, null)
          namespace   = try(customized_metric_specification.value.namespace, null)
          statistic   = try(customized_metric_specification.value.statistic, null)
          unit        = try(customized_metric_specification.value.unit, null)

          dynamic "metrics" {
            for_each = try(customized_metric_specification.value.metrics, [])

            content {
              expression = try(metrics.value.expression, null)
              id         = metrics.value.id
              label      = try(metrics.value.label, null)

              dynamic "metric_stat" {
                for_each = try([metrics.value.metric_stat], [])

                content {
                  dynamic "metric" {
                    for_each = try([metric_stat.value.metric], [])

                    content {
                      dynamic "dimensions" {
                        for_each = try(metric.value.dimensions, [])

                        content {
                          name  = dimensions.value.name
                          value = dimensions.value.value
                        }
                      }

                      metric_name = metric.value.metric_name
                      namespace   = metric.value.namespace
                    }
                  }

                  stat = metric_stat.value.stat
                  unit = try(metric_stat.value.unit, null)
                }
              }

              return_data = try(metrics.value.return_data, null)
            }
          }
        }
      }
    }
  }

  dynamic "predictive_scaling_configuration" {
    for_each = try([each.value.predictive_scaling_configuration], [])
    content {
      max_capacity_breach_behavior = try(predictive_scaling_configuration.value.max_capacity_breach_behavior, null)
      max_capacity_buffer          = try(predictive_scaling_configuration.value.max_capacity_buffer, null)
      mode                         = try(predictive_scaling_configuration.value.mode, null)
      scheduling_buffer_time       = try(predictive_scaling_configuration.value.scheduling_buffer_time, null)

      dynamic "metric_specification" {
        for_each = try([predictive_scaling_configuration.value.metric_specification], [])
        content {
          target_value = metric_specification.value.target_value

          dynamic "predefined_load_metric_specification" {
            for_each = try([metric_specification.value.predefined_load_metric_specification], [])
            content {
              predefined_metric_type = predefined_load_metric_specification.value.predefined_metric_type
              resource_label         = predefined_load_metric_specification.value.resource_label
            }
          }

          dynamic "predefined_metric_pair_specification" {
            for_each = try([metric_specification.value.predefined_metric_pair_specification], [])
            content {
              predefined_metric_type = predefined_metric_pair_specification.value.predefined_metric_type
              resource_label         = predefined_metric_pair_specification.value.resource_label
            }
          }

          dynamic "predefined_scaling_metric_specification" {
            for_each = try([metric_specification.value.predefined_scaling_metric_specification], [])
            content {
              predefined_metric_type = predefined_scaling_metric_specification.value.predefined_metric_type
              resource_label         = predefined_scaling_metric_specification.value.resource_label
            }
          }
        }
      }
    }
  }
}

resource "aws_autoscaling_notification" "main" {
  count = var.enable_asg_events_notify ? 1 : 0
  group_names = [
    aws_autoscaling_group.main.name,
  ]

  notifications = var.sns_notifications
  topic_arn     = var.create_asg_sns_topic ? aws_sns_topic.main[0].arn : var.topic_arn
}

resource "aws_sns_topic" "main" {
  count             = var.enable_asg_events_notify && var.create_asg_sns_topic ? 1 : 0
  name              = var.sns_topic_name
  kms_master_key_id = var.sns_kms_master_key_id
  tags              = var.tags
}
