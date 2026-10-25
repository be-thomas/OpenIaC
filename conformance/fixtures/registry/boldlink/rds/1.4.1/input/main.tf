#####
resource "aws_db_instance" "this" {
  allocated_storage                     = var.allocated_storage
  max_allocated_storage                 = var.max_allocated_storage
  allow_major_version_upgrade           = var.allow_major_version_upgrade
  auto_minor_version_upgrade            = var.auto_minor_version_upgrade
  apply_immediately                     = var.apply_immediately
  backup_retention_period               = var.backup_retention_period
  backup_window                         = var.backup_window
  ca_cert_identifier                    = var.ca_cert_identifier
  copy_tags_to_snapshot                 = var.copy_tags_to_snapshot
  db_subnet_group_name                  = var.create_db_subnet_group ? aws_db_subnet_group.this[0].id : var.db_subnet_group_name
  delete_automated_backups              = var.delete_automated_backups
  deletion_protection                   = var.deletion_protection
  domain                                = var.domain
  domain_iam_role_name                  = var.domain_iam_role_name
  enabled_cloudwatch_logs_exports       = var.enabled_cloudwatch_logs_exports
  engine                                = var.engine
  engine_version                        = var.engine_version
  final_snapshot_identifier             = var.final_snapshot_identifier
  iam_database_authentication_enabled   = var.iam_database_authentication_enabled
  identifier                            = lower(var.identifier)
  identifier_prefix                     = var.identifier_prefix
  publicly_accessible                   = var.publicly_accessible
  instance_class                        = var.instance_class
  kms_key_id                            = var.kms_key_id
  maintenance_window                    = var.maintenance_window
  monitoring_interval                   = var.monitoring_interval
  monitoring_role_arn                   = var.create_monitoring_role && var.monitoring_interval > 0 ? aws_iam_role.this[0].arn : var.monitoring_role_arn
  availability_zone                     = var.availability_zone
  multi_az                              = var.multi_az
  nchar_character_set_name              = var.nchar_character_set_name
  license_model                         = var.license_model
  replica_mode                          = var.replica_mode
  db_name                               = var.db_name
  option_group_name                     = var.option_group_name != null ? var.option_group_name : join("", aws_db_option_group.this.*.name)
  parameter_group_name                  = var.parameter_group_name != null ? var.parameter_group_name : join("", aws_db_parameter_group.this.*.id)
  username                              = var.username
  password                              = var.manage_master_user_password ? null : var.password
  manage_master_user_password           = var.password == null ? var.manage_master_user_password : null
  master_user_secret_kms_key_id         = var.manage_master_user_password ? var.master_user_secret_kms_key_id : null
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_kms_key_id       = var.performance_insights_kms_key_id
  performance_insights_retention_period = var.performance_insights_retention_period
  port                                  = var.port
  replicate_source_db                   = var.replicate_source_db
  vpc_security_group_ids                = local.vpc_security_group_ids == null ? null : compact(concat(aws_security_group.this.*.id, local.vpc_security_group_ids))
  skip_final_snapshot                   = var.skip_final_snapshot
  snapshot_identifier                   = var.snapshot_identifier
  storage_encrypted                     = var.storage_encrypted
  storage_type                          = var.storage_type
  timezone                              = var.timezone # Currently only supported by Microsoft SQL Server.
  tags                                  = var.tags

  dynamic "restore_to_point_in_time" {
    for_each = var.restore_to_point_in_time != null ? [var.restore_to_point_in_time] : []
    content {
      restore_time                  = lookup(restore_to_point_in_time.value, "restore_time", null)
      source_db_instance_identifier = lookup(restore_to_point_in_time.value, "source_db_instance_identifier", null)
      source_dbi_resource_id        = lookup(restore_to_point_in_time.value, "source_dbi_resource_id", null)
      use_latest_restorable_time    = lookup(restore_to_point_in_time.value, "use_latest_restorable_time", null)
    }
  }

  # This will not recreate the resource if the S3 object changes in some way. It's only used to initialize the database
  dynamic "s3_import" {
    for_each = var.s3_import == null ? [] : [var.s3_import]
    content {
      bucket_name           = s3_import.value.bucket_name
      bucket_prefix         = lookup(s3_import.value, "bucket_prefix", null)
      ingestion_role        = s3_import.value.ingestion_role
      source_engine         = "mysql"
      source_engine_version = s3_import.value.source_engine_version
    }
  }

  timeouts {
    create = lookup(var.timeouts, "create", null)
    update = lookup(var.timeouts, "update", null)
    delete = lookup(var.timeouts, "delete", null)
  }

  lifecycle {
    ignore_changes = [enabled_cloudwatch_logs_exports]
  }

}

# Subnet Group
resource "aws_db_subnet_group" "this" {
  count       = var.create_db_subnet_group ? 1 : 0
  name        = lower("${var.identifier}-subnetgroup")
  subnet_ids  = var.subnet_ids
  description = "${var.identifier} RDS Subnet Group"
  tags        = var.tags
}

# Option Group
resource "aws_db_option_group" "this" {
  count                = var.create_option_group ? 1 : 0
  name                 = lower("${var.identifier}-option-group")
  name_prefix          = var.name_prefix
  engine_name          = var.engine
  major_engine_version = var.major_engine_version
  dynamic "option" {
    for_each = var.options
    content {
      option_name                    = lookup(option.value, "option_name")
      port                           = lookup(option.value, "port", null)
      version                        = lookup(option.value, "version", null)
      db_security_group_memberships  = lookup(option.value, "db_security_group_memberships", null)
      vpc_security_group_memberships = lookup(option.value, "vpc_security_group_memberships", null)
      dynamic "option_settings" {
        for_each = lookup(option.value, "option_settings", [])
        content {
          name  = option_settings.value.name
          value = option_settings.value.value
        }
      }
    }
  }
}

# Security group
resource "aws_security_group" "this" {
  count       = var.create_security_group ? 1 : 0
  name        = "${var.identifier}-rds-security-group"
  vpc_id      = var.vpc_id
  description = "${var.identifier} RDS instance Security Group"

  dynamic "ingress" {
    for_each = var.security_group_ingress
    content {
      description      = try(ingress.value.description, "Rule to allow port ${try(ingress.value.from_port, "")} inbound traffic")
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
      description      = try(egress.value.description, "Rule to allow outbound traffic")
      from_port        = try(egress.value.from_port, 0)
      to_port          = try(egress.value.to_port, 0)
      protocol         = try(egress.value.protocol, -1)
      cidr_blocks      = try(egress.value.cidr_blocks, ["0.0.0.0/0"])
      ipv6_cidr_blocks = try(egress.value.ipv6_cidr_blocks, [])
    }
  }

  tags = var.tags
}

# enhanced monitoring IAM role
resource "aws_iam_role" "this" {
  count              = var.create_monitoring_role ? 1 : 0
  name               = "${var.identifier}-enhanced-monitoring-role"
  assume_role_policy = var.assume_role_policy
  description        = "enhanced monitoring iam role for ${var.identifier} rds instance."
  tags               = var.tags
}

resource "aws_iam_role_policy_attachment" "this" {
  count      = var.create_monitoring_role ? 1 : 0
  role       = aws_iam_role.this[0].name
  policy_arn = var.policy_arn
}

## Parameter Group
resource "aws_db_parameter_group" "this" {
  count  = var.create_parameter_group ? 1 : 0
  name   = "${var.identifier}-parameter-group"
  family = var.parameter_group_family

  dynamic "parameter" {
    for_each = try([var.parameters], [])
    content {
      name         = parameter.value.name
      value        = parameter.value.value
      apply_method = lookup(parameter.value, "apply_method", "immediate")
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}
