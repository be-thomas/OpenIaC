# # DocDB cluster Resource
locals {
  cluster_identifier = var.cluster_identifier != null ? var.cluster_identifier : replace(var.cluster_identifier_prefix, "-", "")
}

resource "aws_docdb_cluster" "this" {
  apply_immediately               = var.apply_immediately
  availability_zones              = var.availability_zones
  backup_retention_period         = var.backup_retention_period
  cluster_identifier_prefix       = var.cluster_identifier == null ? var.cluster_identifier_prefix : null
  cluster_identifier              = var.cluster_identifier
  db_subnet_group_name            = aws_docdb_subnet_group.this.id
  db_cluster_parameter_group_name = var.db_cluster_parameter_group_name
  deletion_protection             = var.deletion_protection
  enabled_cloudwatch_logs_exports = var.enabled_cloudwatch_logs_exports
  engine_version                  = var.engine_version
  engine                          = var.engine
  final_snapshot_identifier       = "${local.cluster_identifier}-final-snapshot"
  kms_key_id                      = var.kms_key_id
  master_password                 = var.master_password
  master_username                 = var.master_username
  port                            = var.port
  preferred_backup_window         = var.preferred_backup_window
  preferred_maintenance_window    = var.preferred_maintenance_window
  skip_final_snapshot             = var.skip_final_snapshot
  snapshot_identifier             = var.snapshot_identifier
  storage_encrypted               = var.storage_encrypted
  tags                            = var.tags
  vpc_security_group_ids          = compact(concat(var.vpc_security_group_ids, [aws_security_group.this[0].id]))

  timeouts {
    create = try(var.cluster_timeouts["create"], "90m")
    update = try(var.cluster_timeouts["update"], "90m")
    delete = try(var.cluster_timeouts["delete"], "90m")
  }
}

# DocDB Parameter Group
resource "aws_docdb_cluster_parameter_group" "this" {
  count       = var.create_cluster_parameter_group ? 1 : 0
  name        = var.name
  name_prefix = var.name_prefix
  family      = var.family
  description = "${var.cluster_identifier} parameter group"
  dynamic "parameter" {
    for_each = var.cluster_parameters
    content {
      name         = lookup(parameter.value, "name")
      value        = lookup(parameter.value, "value")
      apply_method = lookup(parameter.value, "apply_method", "immediate")
    }
  }

  tags = var.tags
}

# DocDB SubnetGroup Group
resource "aws_docdb_subnet_group" "this" {
  name        = "${local.cluster_identifier}-subnet-group"
  name_prefix = var.subnet_name_prefix
  description = "${local.cluster_identifier} subnet group."
  subnet_ids  = var.subnet_ids
  tags        = var.tags
}

# DocDB Cluster Instance
resource "aws_docdb_cluster_instance" "this" {
  count                        = var.instance_count
  apply_immediately            = var.apply_immediately
  auto_minor_version_upgrade   = var.auto_minor_version_upgrade
  availability_zone            = var.availability_zone
  cluster_identifier           = aws_docdb_cluster.this.id
  engine                       = var.engine
  identifier                   = "${var.identifier}-${count.index}"
  identifier_prefix            = var.identifier == null ? var.identifier_prefix : null
  instance_class               = var.instance_class
  preferred_maintenance_window = var.preferred_maintenance_window
  promotion_tier               = var.promotion_tier
  tags                         = var.tags

  timeouts {
    create = try(var.instance_timeouts["create"], "90m")
    update = try(var.instance_timeouts["update"], "90m")
    delete = try(var.instance_timeouts["delete"], "90m")
  }
}

# Security Group for DocDB cluster
resource "aws_security_group" "this" {
  count       = var.create_security_group ? 1 : 0
  name        = "${local.cluster_identifier}-security-group"
  vpc_id      = var.vpc_id
  description = "${local.cluster_identifier} Security Group"
  tags        = var.tags
}

resource "aws_security_group_rule" "ingress" {
  for_each                 = var.security_group_ingress_rules
  type                     = "ingress"
  description              = lookup(each.value, "description", null)
  from_port                = lookup(each.value, "from_port", 27017)
  to_port                  = lookup(each.value, "to_port", 27017)
  protocol                 = lookup(each.value, "protocol", "tcp")
  cidr_blocks              = lookup(each.value, "cidr_blocks", [])
  source_security_group_id = lookup(each.value, "source_security_group_id", null)
  self                     = lookup(each.value, "self", null)
  security_group_id        = aws_security_group.this[0].id
}

resource "aws_security_group_rule" "egress" {
  for_each                 = var.security_group_egress_rules
  type                     = "egress"
  description              = lookup(each.value, "description", null)
  from_port                = lookup(each.value, "from_port", 0)
  to_port                  = lookup(each.value, "to_port", 0)
  protocol                 = lookup(each.value, "protocol", "-1")
  cidr_blocks              = lookup(each.value, "cidr_blocks", [])
  source_security_group_id = lookup(each.value, "source_security_group_id", null)
  self                     = lookup(each.value, "self", null)
  security_group_id        = aws_security_group.this[0].id
}
