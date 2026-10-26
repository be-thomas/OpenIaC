locals {
  cluster_id = coalesce(var.cluster_id, var.name)

  cluster_size = var.replication_enabled ? var.cluster_size >= 2 ? var.cluster_size : 2 : var.cluster_size

  num_nodes = var.cluster_mode_enabled ? var.num_node_groups * var.replicas_per_node_group : local.cluster_size

  parameters = var.cluster_mode_enabled ? concat(
    var.parameters,
    [{
      name  = "cluster-enabled"
      value = "yes"
    }]
  ) : var.parameters
}

resource "aws_elasticache_parameter_group" "this" {
  count = var.enabled && var.parameter_group_name == "" && !var.use_serverless || var.parameter_group_name == null ? 1 : 0

  name   = var.name
  family = var.elasticache_parameter_group_family

  dynamic "parameter" {
    for_each = local.parameters

    content {
      name  = parameter.value.name
      value = parameter.value.value
    }
  }
}

resource "aws_elasticache_subnet_group" "this" {
  count = var.enabled && var.create_elasticache_subnet_group ? 1 : 0

  name       = format("%s-subnet-group", var.name)
  subnet_ids = var.subnets
}

resource "aws_elasticache_replication_group" "this" {
  count = var.enabled && !var.use_serverless ? 1 : 0

  replication_group_id = var.replication_group_id == "" ? local.cluster_id : var.replication_group_id
  description          = "${var.engine} replication group"

  engine         = var.engine
  engine_version = var.engine_version
  port           = var.port

  node_type          = var.instance_type
  num_cache_clusters = var.cluster_mode_enabled ? null : local.cluster_size

  maintenance_window         = var.maintenance_window
  auto_minor_version_upgrade = var.auto_minor_version_upgrade
  data_tiering_enabled       = var.data_tiering_enabled

  preferred_cache_cluster_azs = var.replication_enabled && !var.cluster_mode_enabled ? var.preferred_cache_cluster_azs : null

  parameter_group_name = var.parameter_group_name != "" ? var.parameter_group_name : aws_elasticache_parameter_group.this[0].name
  subnet_group_name    = try(aws_elasticache_subnet_group.this[0].name, var.subnet_group_name)
  security_group_ids   = var.security_groups

  multi_az_enabled           = var.replication_enabled ? true : false
  at_rest_encryption_enabled = var.at_rest_encryption_enabled
  transit_encryption_enabled = var.transit_encryption_enabled
  transit_encryption_mode    = var.transit_encryption_mode
  automatic_failover_enabled = var.replication_enabled ? true : false

  notification_topic_arn = var.notification_topic_arn

  apply_immediately = var.apply_immediately

  auth_token                 = var.transit_encryption_enabled ? var.auth_token : null
  auth_token_update_strategy = var.auth_token_update_strategy
  kms_key_id                 = var.at_rest_encryption_enabled ? var.kms_key_id : null

  num_node_groups         = var.cluster_mode_enabled ? var.num_node_groups : null
  replicas_per_node_group = var.cluster_mode_enabled ? var.replicas_per_node_group : null

  user_group_ids = length(var.user_group_id) != 0 ? [var.user_group_id] : null

  snapshot_retention_limit = var.instance_type != "cache.t1.micro" ? var.snapshot_retention_limit : 0
  snapshot_window          = var.snapshot_window
  snapshot_arns            = var.snapshot_arns
  snapshot_name            = var.snapshot_name

  dynamic "log_delivery_configuration" {
    for_each = var.log_delivery_configuration

    content {
      destination      = log_delivery_configuration.value.destination
      destination_type = log_delivery_configuration.value.destination_type
      log_format       = log_delivery_configuration.value.log_format
      log_type         = log_delivery_configuration.value.log_type
    }
  }

  tags = var.tags
}

resource "aws_elasticache_serverless_cache" "this" {
  count = var.enabled && var.use_serverless ? 1 : 0

  name                 = var.name
  description          = "${var.name} ElastiCache ${var.engine} Serverless"
  engine               = var.engine
  major_engine_version = split(".", var.engine_version)[0]

  cache_usage_limits {
    data_storage {
      maximum = var.max_data_storage
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = var.max_ecpu_per_second
    }
  }

  user_group_id = var.user_group_id != "" ? var.user_group_id : null

  kms_key_id         = var.kms_key_id
  security_group_ids = var.security_groups
  subnet_ids         = var.subnets

  daily_snapshot_time      = var.daily_snapshot_time
  snapshot_arns_to_restore = var.snapshot_arns_to_restore
  snapshot_retention_limit = var.snapshot_retention_limit

  tags = var.tags
}
