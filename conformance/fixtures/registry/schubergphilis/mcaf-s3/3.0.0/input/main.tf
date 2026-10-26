locals {
  account_id            = data.aws_caller_identity.default.account_id
  account_region        = var.region != null ? var.region : data.aws_region.default.region
  logging_permissions   = length(var.logging_source_bucket_arns) > 0 ? { create = true } : {}
  malware_iam_role_name = aws_s3_bucket.default.id
  policy                = var.policy != null ? var.policy : null

  # On/Off switches for optional resources and configuration
  bucket_key_enabled                       = var.kms_key_arn != null ? true : false
  cors_rule_enabled                        = var.cors_rule != null ? { create = true } : {}
  bucket_key_encryption_enforced           = var.bucket_key_encryption_enforced ? { create = true } : {}
  logging_partitioned_prefix_enabled       = try(var.logging.target_object_key_format.format_type, null) == "partitioned" ? { create = true } : {}
  logging_simple_prefix_enabled            = try(var.logging.target_object_key_format.format_type, null) == "simple" ? { create = true } : {}
  logging_target_object_key_format_enabled = try(var.logging.target_object_key_format, null) != null ? { create = true } : {}
  malware_protection_enabled               = var.malware_protection.enabled ? { create = true } : {}
  object_lock_enabled                      = var.object_lock_mode != null ? { create = true } : {}
  replication_configuration_enabled        = var.replication_configuration != null ? { create = true } : {}
}

data "aws_caller_identity" "default" {}

data "aws_region" "default" {}

################################################################################
# S3 Bucket & Bucket Policy
################################################################################

resource "aws_s3_bucket" "default" {
  region              = local.account_region
  bucket              = var.name
  bucket_prefix       = var.name_prefix
  force_destroy       = var.force_destroy
  object_lock_enabled = var.object_lock_mode != null ? true : false
  tags                = var.tags
}

data "aws_iam_policy_document" "ssl_policy" {
  statement {
    sid     = "AllowSSLRequestsOnly"
    actions = ["s3:*"]
    effect  = "Deny"
    resources = [
      aws_s3_bucket.default.arn,
      "${aws_s3_bucket.default.arn}/*"
    ]
    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}

data "aws_iam_policy_document" "logging_policy" {
  dynamic "statement" {
    for_each = local.logging_permissions

    content {
      sid     = "S3AccessLog"
      actions = ["s3:PutObject"]
      effect  = "Allow"
      resources = [
        "${aws_s3_bucket.default.arn}/*"
      ]
      principals {
        type        = "Service"
        identifiers = ["logging.s3.amazonaws.com"]
      }
      condition {
        test     = "ArnLike"
        variable = "aws:SourceArn"
        values   = var.logging_source_bucket_arns
      }
    }
  }
}

data "aws_iam_policy_document" "malware_protection_policy" {
  for_each = local.malware_protection_enabled

  statement {
    sid    = "NoReadExceptForClean"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion"
    ]
    resources = [
      aws_s3_bucket.default.arn,
      "${aws_s3_bucket.default.arn}/*"
    ]
    condition {
      test     = "StringNotEquals"
      variable = "s3:ExistingObjectTag/GuardDutyMalwareScanStatus"
      values   = ["NO_THREATS_FOUND"]
    }
    condition {
      test     = "ForAnyValue:ArnNotEquals"
      variable = "aws:PrincipalArn"
      values = [
        "arn:aws:iam::${local.account_id}:assumed-role/${module.s3_malware_protection_role["create"].name}/GuardDutyMalwareProtection",
        module.s3_malware_protection_role["create"].arn
      ]
    }
  }

  statement {
    sid    = "OnlyGuardDutyCanTag"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = ["s3:PutObjectTagging"]
    resources = [
      aws_s3_bucket.default.arn,
      "${aws_s3_bucket.default.arn}/*"
    ]
    condition {
      test     = "ForAnyValue:ArnNotEquals"
      variable = "aws:PrincipalArn"
      values = [
        "arn:aws:iam::${local.account_id}:assumed-role/${module.s3_malware_protection_role["create"].name}/GuardDutyMalwareProtection",
        module.s3_malware_protection_role["create"].arn
      ]
    }
  }
}

data "aws_iam_policy_document" "bucket_key_encryption_policy_enforced" {
  for_each = local.bucket_key_encryption_enforced

  statement {
    sid       = "EnforceBucketKeyEncryption"
    actions   = ["s3:PutObject"]
    effect    = "Deny"
    resources = ["${aws_s3_bucket.default.arn}/*"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "StringNotEquals"
      variable = "s3:x-amz-server-side-encryption-aws-kms-key-id"
      values   = [var.kms_key_arn]
    }
  }
}

data "aws_iam_policy_document" "combined" {
  source_policy_documents = compact([
    local.policy,
    data.aws_iam_policy_document.ssl_policy.json,
    data.aws_iam_policy_document.logging_policy.json,
    try(data.aws_iam_policy_document.malware_protection_policy["create"].json, ""),
    try(data.aws_iam_policy_document.bucket_key_encryption_policy_enforced["create"].json, "")
  ])
}

resource "aws_s3_bucket_policy" "default" {
  region = local.account_region
  bucket = aws_s3_bucket.default.id
  policy = data.aws_iam_policy_document.combined.json
}

################################################################################
# S3 Bucket Configuration
################################################################################

###
# Ownership & ACL
###

resource "aws_s3_bucket_ownership_controls" "default" {
  region = local.account_region
  bucket = aws_s3_bucket.default.id

  rule {
    object_ownership = var.object_ownership_type
  }
}

resource "aws_s3_bucket_acl" "default" {
  count = var.object_ownership_type == "ObjectWriter" ? 1 : 0

  region = local.account_region
  bucket = aws_s3_bucket.default.id
  acl    = var.access_control_policy != null ? null : var.acl

  dynamic "access_control_policy" {
    for_each = var.access_control_policy != null ? [var.access_control_policy] : []

    content {
      owner {
        id = access_control_policy.value.owner_id
      }

      dynamic "grant" {
        for_each = access_control_policy.value.grants
        content {
          grantee {
            type          = grant.value.grantee.type
            id            = grant.value.grantee.type == "CanonicalUser" ? grant.value.grantee.identifier : null
            uri           = grant.value.grantee.type == "Group" ? grant.value.grantee.identifier : null
            email_address = grant.value.grantee.type == "AmazonCustomerByEmail" ? grant.value.grantee.identifier : null
          }
          permission = grant.value.permission
        }
      }
    }
  }

  depends_on = [aws_s3_bucket_ownership_controls.default]
}

###
# CORS
###

resource "aws_s3_bucket_cors_configuration" "default" {
  for_each = local.cors_rule_enabled

  region = local.account_region
  bucket = aws_s3_bucket.default.bucket

  cors_rule {
    allowed_headers = var.cors_rule.allowed_headers
    allowed_methods = var.cors_rule.allowed_methods
    allowed_origins = var.cors_rule.allowed_origins
    expose_headers  = var.cors_rule.expose_headers
    max_age_seconds = var.cors_rule.max_age_seconds
  }
}

###
# Inventory
###

resource "aws_s3_bucket_inventory" "default" {
  for_each = var.inventory_configuration

  region                   = local.account_region
  bucket                   = aws_s3_bucket.default.id
  enabled                  = each.value.enabled
  included_object_versions = each.value.included_object_versions
  name                     = each.key
  optional_fields          = each.value.optional_fields

  destination {
    bucket {
      account_id = each.value.destination.account_id
      bucket_arn = each.value.destination.bucket_arn
      format     = each.value.destination.format
      prefix     = each.value.destination.prefix

      encryption {
        dynamic "sse_kms" {
          for_each = each.value.destination.encryption.encryption_type == "sse_kms" ? { create = true } : {}

          content {
            key_id = each.value.destination.encryption.kms_key_id
          }
        }

        dynamic "sse_s3" {
          for_each = each.value.destination.encryption.encryption_type == "sse_s3" ? { create = true } : {}

          content {
          }
        }
      }
    }
  }

  schedule {
    frequency = each.value.frequency
  }

  dynamic "filter" {
    for_each = each.value.filter_prefix != null ? { create = true } : {}

    content {
      prefix = each.value.filter_prefix
    }
  }
}

###
# Lifecycle
###

resource "aws_s3_bucket_lifecycle_configuration" "default" {
  #checkov:skip=CKV_AWS_300: Ensure S3 lifecycle configuration sets period for aborting failed uploads - consumer of the module should decide
  count = length(var.lifecycle_rule) > 0 ? 1 : 0

  region                                 = local.account_region
  bucket                                 = aws_s3_bucket.default.bucket
  transition_default_minimum_object_size = var.transition_default_minimum_object_size

  dynamic "rule" {
    for_each = var.lifecycle_rule

    content {
      id     = rule.value.id
      status = rule.value.enabled ? "Enabled" : "Disabled"

      # --------------------------------------------------------------
      # abort_incomplete_multipart_upload (max 1 block)
      # --------------------------------------------------------------
      dynamic "abort_incomplete_multipart_upload" {
        for_each = rule.value.abort_incomplete_multipart_upload != null ? [rule.value.abort_incomplete_multipart_upload] : []

        content {
          days_after_initiation = abort_incomplete_multipart_upload.value.days_after_initiation
        }
      }

      # ------------
      # expiration (max 1 block)
      # ------------
      dynamic "expiration" {
        for_each = rule.value.expiration != null ? [rule.value.expiration] : []

        content {
          date                         = expiration.value.date
          days                         = expiration.value.days
          expired_object_delete_marker = expiration.value.expired_object_delete_marker
        }
      }

      # --------------------------------------------------
      # filter (max 1 block)
      # --------------------------------------------------
      dynamic "filter" {
        for_each = rule.value.filter != null ? [rule.value.filter] : []

        content {
          prefix                   = filter.value.prefix
          object_size_greater_than = filter.value.object_size_greater_than
          object_size_less_than    = filter.value.object_size_less_than

          dynamic "tag" {
            for_each = filter.value.tag != null ? [filter.value.tag] : []

            content {
              key   = tag.value.key
              value = tag.value.value
            }
          }

          dynamic "and" {
            for_each = filter.value.and != null ? [filter.value.and] : []

            content {
              prefix                   = and.value.prefix
              object_size_greater_than = and.value.object_size_greater_than
              object_size_less_than    = and.value.object_size_less_than
              tags                     = and.value.tags
            }
          }
        }
      }

      dynamic "filter" {
        for_each = rule.value.filter == null ? { create = true } : {}

        content {}
      }

      # -------------------------------
      # noncurrent_version_expiration (max 1 block)
      # -------------------------------
      dynamic "noncurrent_version_expiration" {
        for_each = rule.value.noncurrent_version_expiration != null ? [rule.value.noncurrent_version_expiration] : []

        content {
          newer_noncurrent_versions = noncurrent_version_expiration.value.newer_noncurrent_versions
          noncurrent_days           = noncurrent_version_expiration.value.noncurrent_days
        }
      }

      # -------------------------------
      # noncurrent_version_transition (1..n blocks)
      # -------------------------------
      dynamic "noncurrent_version_transition" {
        for_each = rule.value.noncurrent_version_transition != null ? rule.value.noncurrent_version_transition : []

        content {
          newer_noncurrent_versions = noncurrent_version_transition.value.newer_noncurrent_versions
          noncurrent_days           = noncurrent_version_transition.value.noncurrent_days
          storage_class             = noncurrent_version_transition.value.storage_class
        }
      }

      # -----------
      # transition (1..n blocks)
      # -----------
      dynamic "transition" {
        for_each = rule.value.transition != null ? rule.value.transition : []

        content {
          date          = transition.value.date
          days          = transition.value.days
          storage_class = transition.value.storage_class
        }
      }
    }
  }
}

###
# Logging
###

resource "aws_s3_bucket_logging" "default" {
  count = var.logging != null ? 1 : 0

  region        = local.account_region
  bucket        = aws_s3_bucket.default.id
  target_bucket = var.logging.target_bucket
  target_prefix = var.logging.target_prefix


  dynamic "target_object_key_format" {
    for_each = local.logging_target_object_key_format_enabled

    content {
      dynamic "partitioned_prefix" {
        for_each = local.logging_partitioned_prefix_enabled

        content {
          partition_date_source = var.logging.target_object_key_format.partition_date_source
        }
      }

      dynamic "simple_prefix" {
        for_each = local.logging_simple_prefix_enabled

        content {}
      }
    }
  }

  lifecycle {
    precondition {
      condition     = var.logging.target_bucket != aws_s3_bucket.default.id || var.object_lock_mode == null
      error_message = "You're trying to enable server access logging and object locking on the same bucket! Object lock will prevent server access logs from written to the bucket. Either log to a different bucket or remove the object lock configuration."
    }
  }
}

###
# Object Lock
###

resource "aws_s3_bucket_object_lock_configuration" "default" {
  for_each = local.object_lock_enabled

  region = local.account_region
  bucket = aws_s3_bucket.default.bucket

  rule {
    default_retention {
      mode  = var.object_lock_mode
      years = var.object_lock_years
      days  = var.object_lock_days
    }
  }

  lifecycle {
    precondition {
      condition     = var.object_lock_mode == null || length(var.logging_source_bucket_arns) == 0
      error_message = "You're trying to allow (other buckets) logging to this bucket and enable object locking on the same bucket! Object lock will prevent server access logs from written to the bucket. Either remove the logging source buckets configuration or remove the object lock configuration."
    }
    precondition {
      condition     = var.object_lock_mode == null || var.object_lock_days != null || var.object_lock_years != null
      error_message = "When object_lock_mode is set, either object_lock_days or object_lock_years must also be set."
    }
  }
}

###
# Malware protection
###

resource "aws_guardduty_malware_protection_plan" "default" {
  for_each = local.malware_protection_enabled

  region = local.account_region
  role   = module.s3_malware_protection_role["create"].arn

  protected_resource {
    s3_bucket {
      bucket_name     = aws_s3_bucket.default.id
      object_prefixes = var.malware_protection.object_prefixes
    }
  }

  actions {
    tagging {
      status = "ENABLED"
    }
  }
}

data "aws_iam_policy_document" "s3_malware_protection_assume_role" {
  for_each = local.malware_protection_enabled

  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["malware-protection-plan.guardduty.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [local.account_id]
    }
    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:guardduty:${local.account_region}:${local.account_id}:malware-protection-plan/*"]
    }
  }
}

data "aws_iam_policy_document" "s3_malware_protection_policy" {
  for_each = local.malware_protection_enabled

  statement {
    sid    = "AllowManagedRuleToSendS3EventsToGuardDuty"
    effect = "Allow"
    actions = [
      "events:PutRule",
      "events:DeleteRule",
      "events:PutTargets",
      "events:RemoveTargets"
    ]
    resources = [
      "arn:aws:events:${local.account_region}:${local.account_id}:rule/DO-NOT-DELETE-AmazonGuardDutyMalwareProtectionS3*"
    ]
    condition {
      test     = "StringEquals"
      variable = "events:ManagedBy"
      values   = ["malware-protection-plan.guardduty.amazonaws.com"]
    }
  }

  statement {
    sid    = "AllowGuardDutyToMonitorEventBridgeManagedRule"
    effect = "Allow"
    actions = [
      "events:DescribeRule",
      "events:ListTargetsByRule"
    ]
    resources = [
      "arn:aws:events:${local.account_region}:${local.account_id}:rule/DO-NOT-DELETE-AmazonGuardDutyMalwareProtectionS3*"
    ]
  }

  statement {
    sid    = "AllowEnableS3EventBridgeEvents"
    effect = "Allow"
    actions = [
      "s3:PutBucketNotification",
      "s3:GetBucketNotification"
    ]
    resources = [
      aws_s3_bucket.default.arn
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:ResourceAccount"
      values   = [local.account_id]
    }
  }

  statement {
    sid    = "AllowPostScanTag"
    effect = "Allow"
    actions = [
      "s3:GetObjectTagging",
      "s3:GetObjectVersionTagging",
      "s3:PutObjectTagging",
      "s3:PutObjectVersionTagging"
    ]
    resources = [
      "${aws_s3_bucket.default.arn}/*"
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:ResourceAccount"
      values   = [local.account_id]
    }
  }

  statement {
    sid     = "AllowPutValidationObject"
    effect  = "Allow"
    actions = ["s3:PutObject"]
    resources = [
      "${aws_s3_bucket.default.arn}/malware-protection-resource-validation-object"
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:ResourceAccount"
      values   = [local.account_id]
    }
  }

  statement {
    sid     = "AllowCheckBucketOwnership"
    effect  = "Allow"
    actions = ["s3:ListBucket"]
    resources = [
      aws_s3_bucket.default.arn
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:ResourceAccount"
      values   = [local.account_id]
    }
  }

  statement {
    sid    = "AllowMalwareScan"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion"
    ]
    resources = [
      "${aws_s3_bucket.default.arn}/*"
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:ResourceAccount"
      values   = [local.account_id]
    }
  }

  dynamic "statement" {
    for_each = var.kms_key_arn != null ? { create = true } : {}

    content {
      sid    = "AllowDecryptForMalwareScan"
      effect = "Allow"
      actions = [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ]
      resources = [
        var.kms_key_arn
      ]
      condition {
        test     = "StringLike"
        variable = "kms:ViaService"
        values   = ["s3.${local.account_region}.amazonaws.com"]
      }
    }
  }
}

module "s3_malware_protection_role" {
  for_each = local.malware_protection_enabled

  source  = "schubergphilis/mcaf-role/aws"
  version = "~> 0.4"

  name                 = local.malware_iam_role_name
  assume_policy        = data.aws_iam_policy_document.s3_malware_protection_assume_role["create"].json
  create_policy        = true
  permissions_boundary = var.malware_protection.permissions_boundary
  role_policy          = data.aws_iam_policy_document.s3_malware_protection_policy["create"].json
}


###
# Notification / EventBridge
###

resource "aws_s3_bucket_notification" "eventbridge" {
  count = var.eventbridge_enabled ? 1 : 0

  region      = local.account_region
  bucket      = aws_s3_bucket.default.id
  eventbridge = var.eventbridge_enabled
}

###
# Replication
###

resource "aws_s3_bucket_replication_configuration" "default" {
  for_each = local.replication_configuration_enabled

  region = local.account_region
  role   = var.replication_configuration.iam_role_arn
  bucket = aws_s3_bucket.default.id

  dynamic "rule" {
    for_each = var.replication_configuration.rules

    content {
      id       = rule.value["id"]
      priority = rule.key
      status   = "Enabled"

      delete_marker_replication {
        status = "Disabled"
      }

      destination {
        bucket        = rule.value["dest_bucket"]
        storage_class = rule.value["dest_storage_class"]

        dynamic "encryption_configuration" {
          for_each = rule.value.replica_kms_key_arn != null ? { create = true } : {}

          content {
            replica_kms_key_id = rule.value.replica_kms_key_arn
          }
        }

        dynamic "metrics" {
          for_each = rule.value.metrics != null ? { create = true } : {}

          content {
            status = rule.value.metrics.status ? "Enabled" : "Disabled"

            event_threshold {
              minutes = rule.value.metrics.event_threshold_minutes
            }
          }
        }

        dynamic "replication_time" {
          for_each = rule.value.replication_time != null ? { create = true } : {}

          content {
            status = rule.value.replication_time.status ? "Enabled" : "Disabled"

            time {
              minutes = rule.value.replication_time.time_minutes
            }
          }
        }
      }

      dynamic "source_selection_criteria" {
        for_each = rule.value.source_selection_criteria != null ? { create = true } : {}

        content {
          replica_modifications {
            status = rule.value.source_selection_criteria.replica_modifications ? "Enabled" : "Disabled"
          }
          sse_kms_encrypted_objects {
            status = rule.value.source_selection_criteria.sse_kms_encrypted_objects ? "Enabled" : "Disabled"
          }
        }
      }

      filter {}
    }
  }

  depends_on = [aws_s3_bucket_versioning.default]
}

###
# Requester Pays configuration
###

resource "aws_s3_bucket_request_payment_configuration" "default" {
  region = local.account_region
  bucket = aws_s3_bucket.default.bucket
  payer  = var.request_payer
}

###
# Server-Side Encryption
###

resource "aws_s3_bucket_server_side_encryption_configuration" "default" {
  region = local.account_region
  bucket = aws_s3_bucket.default.bucket

  rule {
    blocked_encryption_types = var.blocked_encryption_types
    bucket_key_enabled       = local.bucket_key_enabled

    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms_key_arn
      sse_algorithm     = var.kms_key_arn != null ? "aws:kms" : "AES256"
    }
  }
}

###
# Public Access Block
###

resource "aws_s3_bucket_public_access_block" "default" {
  region                  = local.account_region
  bucket                  = aws_s3_bucket.default.id
  block_public_acls       = var.block_public_acls
  block_public_policy     = var.block_public_policy
  ignore_public_acls      = var.ignore_public_acls
  restrict_public_buckets = var.restrict_public_buckets
}

###
# Versioning
###

resource "aws_s3_bucket_versioning" "default" {
  region = local.account_region
  bucket = aws_s3_bucket.default.id

  versioning_configuration {
    status = var.versioning ? "Enabled" : "Suspended"
  }
}
