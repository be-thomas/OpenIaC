locals {

  region     = data.aws_region.current.name
  partition  = data.aws_partition.current.partition
  account_id = data.aws_caller_identity.current.account_id
  dns_suffix = data.aws_partition.current.dns_suffix
  kms_policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [
        {
          Sid    = var.kms_policy_iam_statement_id
          Effect = "Allow"
          Principal = {
            "AWS" = ["arn:${local.partition}:iam::${local.account_id}:root"]
          }
          Action   = ["kms:*"]
          Resource = ["*"]
        },
        {
          Sid    = var.kms_policy_cloudwatch_statement_id
          Action = var.kms_cloudwatch_actions
          Effect = "Allow"
          Principal = {
            Service = ["logs.${local.region}.${local.dns_suffix}"],
            AWS     = var.kms_policy_principal
          }
          Resource = ["*"]
          Condition = {
            ArnLike = {
              "kms:EncryptionContext:aws:logs:arn" = ["arn:${local.partition}:logs:${local.region}:${local.account_id}:log-group:/aws/ecs/${var.name}-log-group"]
            }
          }
        }
      ]
    }
  )

  # Create default root volume if requested and no block device mappings specified
  default_root_volume = var.create_default_root_volume && length(var.block_device_mappings) == 0 ? [{
    device_name = var.default_root_device_name
    ebs = {
      volume_size           = lookup(var.ec2_volume_sizes, "root_volume", var.default_root_volume_size)
      volume_type           = var.default_ebs_volume_type
      encrypted             = var.default_ebs_encrypted
      kms_key_id            = var.default_ebs_kms_key_id
      delete_on_termination = true
    }
  }] : []

  # Combine user-specified mappings with default root volume
  all_block_device_mappings = concat(var.block_device_mappings, local.default_root_volume)

  # Process block device mappings with default EBS configuration
  processed_block_device_mappings = [
    for bdm in local.all_block_device_mappings : {
      device_name  = bdm.device_name
      no_device    = lookup(bdm, "no_device", null)
      virtual_name = lookup(bdm, "virtual_name", null)
      ebs = contains(keys(bdm), "ebs") && var.enable_default_ebs_configuration ? {
        delete_on_termination = lookup(bdm.ebs, "delete_on_termination", null)
        encrypted             = lookup(bdm.ebs, "encrypted", var.default_ebs_encrypted)
        iops                  = lookup(bdm.ebs, "iops", null)
        kms_key_id            = lookup(bdm.ebs, "kms_key_id", var.default_ebs_kms_key_id)
        snapshot_id           = lookup(bdm.ebs, "snapshot_id", null)
        throughput            = lookup(bdm.ebs, "throughput", null)
        volume_size           = lookup(bdm.ebs, "volume_size", null)
        volume_type           = lookup(bdm.ebs, "volume_type", var.default_ebs_volume_type)
      } : lookup(bdm, "ebs", null)
    }
  ]
}
