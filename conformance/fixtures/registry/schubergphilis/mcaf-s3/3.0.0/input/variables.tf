variable "access_control_policy" {
  type = object({
    owner_id = string
    grants = list(object({
      grantee = object({
        type       = string # Allowed values: "CanonicalUser", "Group", "AmazonCustomerByEmail"
        identifier = string # Maps to id, uri, or email_address based on the grantee type
      })
      permission = string
    }))
  })
  default     = null
  description = "The access control policy permissions for an S3 bucket object per grantee."
  validation {
    condition = (
      var.access_control_policy == null ?
      true :
      alltrue([
        for grant in var.access_control_policy.grants : (
          grant.grantee.type == "CanonicalUser" ||
          grant.grantee.type == "Group" ||
          grant.grantee.type == "AmazonCustomerByEmail"
        )
      ])
    )
    error_message = "Every grantee 'type' in grants must be one of 'CanonicalUser', 'Group', or 'AmazonCustomerByEmail'."
  }
}

variable "acl" {
  type        = string
  default     = "private"
  description = "The canned ACL to apply, defaults to `private`."
}

variable "block_public_acls" {
  type        = bool
  default     = true
  description = "Whether Amazon S3 should block public ACLs for this bucket."
}

variable "block_public_policy" {
  type        = bool
  default     = true
  description = "Whether Amazon S3 should block public bucket policies for this bucket."
}

variable "blocked_encryption_types" {
  type        = list(string)
  default     = ["SSE-C"]
  description = "List of encryption types to block for S3 bucket objects. Valid values: `NONE` (unblocks all encryption types), `SSE-C` (blocks SSE-C encrypted uploads)."

  validation {
    condition     = alltrue([for v in var.blocked_encryption_types : contains(["NONE", "SSE-C"], v)])
    error_message = "Valid values for blocked_encryption_types are \"NONE\" and \"SSE-C\"."
  }
}

variable "bucket_key_encryption_enforced" {
  type        = bool
  default     = false
  description = "Enforces the default key encryption for all objects in the bucket"
}

variable "cors_rule" {
  type = object({
    allowed_headers = list(string)
    allowed_methods = list(string)
    allowed_origins = list(string)
    expose_headers  = list(string)
    max_age_seconds = number
  })
  default     = null
  description = "The CORS rule for the S3 bucket"
}

variable "eventbridge_enabled" {
  type        = bool
  default     = false
  description = "Whether to enable Amazon EventBridge notifications."
}

variable "force_destroy" {
  type        = bool
  default     = false
  description = "A boolean that indicates all objects should be deleted when deleting the bucket."
}

variable "ignore_public_acls" {
  type        = bool
  default     = true
  description = "Whether Amazon S3 should ignore public ACLs for this bucket."
}

variable "inventory_configuration" {
  type = map(object({
    enabled                  = optional(bool, true)
    filter_prefix            = optional(string, null)
    frequency                = optional(string, "Weekly")
    included_object_versions = optional(string, "Current")
    optional_fields          = optional(list(string), null)
    destination = object({
      account_id = string
      bucket_arn = string
      format     = optional(string, "Parquet")
      prefix     = optional(string, null)
      encryption = optional(object({
        encryption_type = string
        kms_key_id      = optional(string, null)
        }), {
        encryption_type = "sse_s3"
      })
    })
  }))
  default     = {}
  description = "Bucket inventory configuration settings"
}

variable "kms_key_arn" {
  type        = string
  default     = null
  description = "The KMS key ARN used for the bucket encryption."
}

variable "lifecycle_rule" {
  type = list(object({
    id      = string
    enabled = optional(bool, true)
    abort_incomplete_multipart_upload = optional(object({
      days_after_initiation = number
    }))
    expiration = optional(object({
      date                         = optional(string)
      days                         = optional(number)
      expired_object_delete_marker = optional(bool)
    }))
    filter = optional(object({
      prefix                   = optional(string)
      object_size_greater_than = optional(number)
      object_size_less_than    = optional(number)
      tag = optional(object({
        key   = string
        value = string
      }))
      # 'and' block for combining multiple predicates
      and = optional(object({
        object_size_greater_than = optional(number)
        object_size_less_than    = optional(number)
        prefix                   = optional(string)
        tags                     = optional(map(string))
      }))
    }))
    noncurrent_version_expiration = optional(object({
      newer_noncurrent_versions = optional(number)
      noncurrent_days           = optional(number)
    }))
    noncurrent_version_transition = optional(list(object({
      newer_noncurrent_versions = optional(number)
      noncurrent_days           = optional(number)
      storage_class             = string
    })))
    transition = optional(list(object({
      date          = optional(string)
      days          = optional(number)
      storage_class = string
    })))
  }))
  default     = []
  description = "List of lifecycle configuration settings."
}

variable "logging" {
  type = object({
    target_bucket = string
    target_prefix = string
    target_object_key_format = optional(object({
      format_type           = optional(string)                 # "simple" or "partitioned"
      partition_date_source = optional(string, "DeliveryTime") # Required if format_type is "partitioned", default is DeliveryTime
    }))
  })
  default     = null
  description = "Logging configuration, logging is disabled by default."
  validation {
    condition = var.logging == null ? true : (
      # target_object_key_format should be null or have a valid format_type
      var.logging.target_object_key_format == null ? true : (
        # target_object_key_format.format_type must be "simple" or "partitioned"
        contains(["simple", "partitioned"], var.logging.target_object_key_format.format_type) &&
        (
          # If simple, partition_date_source doesn't matter
          var.logging.target_object_key_format.format_type == "simple" ||
          (
            # If partitioned, partition_date_source must be "DeliveryTime" or "EventTime"
            var.logging.target_object_key_format.format_type == "partitioned" &&
            contains(["DeliveryTime", "EventTime"], var.logging.target_object_key_format.partition_date_source)
          )
        )
      )
    )
    error_message = "When logging is enabled: target_object_key_format.format_type must be 'simple' or 'partitioned'. If set to partitioned, target_object_key_format.partition_date_source must be 'DeliveryTime' or 'EventTime'."
  }
}

variable "logging_source_bucket_arns" {
  type        = list(string)
  default     = []
  description = "Configures which source buckets are allowed to log to this bucket."
}

variable "malware_protection" {
  type = object({
    enabled              = optional(bool, false)
    object_prefixes      = optional(list(string), [])
    permissions_boundary = optional(string, null)
  })
  default     = {}
  description = "AWS GuardDuty malware protection bucket protection settings."
}

variable "name" {
  type        = string
  default     = null
  description = "The Name of the bucket. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`."
  validation {
    condition     = var.name != null ? length(var.name) <= 63 : true
    error_message = "The name must be less than or equal to 63 characters in length"
  }
}

variable "name_prefix" {
  type        = string
  default     = null
  description = "Creates a unique bucket name beginning with the specified prefix. Conflicts with `name`."
  validation {
    condition     = var.name_prefix != null ? length(var.name_prefix) <= 37 : true
    error_message = "The name prefix must be less than or equal to 37 characters in length"
  }
}

variable "object_lock_days" {
  type        = number
  default     = null
  description = "The number of days that you want to specify for the default retention period."
}

variable "object_lock_mode" {
  type        = string
  default     = null
  description = "The default object Lock retention mode to apply to new objects."

  validation {
    condition = (
      var.object_lock_mode == null
      ? true
      : contains(["COMPLIANCE", "GOVERNANCE"], var.object_lock_mode)
    )
    error_message = "If set, object lock mode should be COMPLIANCE or GOVERNANCE"
  }
}

variable "object_lock_years" {
  type        = number
  default     = null
  description = "The number of years that you want to specify for the default retention period."
}

variable "object_ownership_type" {
  type        = string
  default     = "BucketOwnerEnforced"
  description = "The object ownership type for the objects in S3 Bucket, defaults to BucketOwnerEnforced."
}

variable "policy" {
  type        = string
  default     = null
  description = "A valid bucket policy JSON document."
}

variable "region" {
  type        = string
  default     = null
  description = "The AWS region where the bucket will be created. If omitted, the default provider region is used."
}

variable "replication_configuration" {
  type = object({
    iam_role_arn = string
    rules = map(object({
      id                  = string
      dest_bucket         = string
      dest_storage_class  = optional(string, null)
      replica_kms_key_arn = optional(string, null)
      metrics = optional(object({
        status                  = optional(bool, false)
        event_threshold_minutes = optional(number, 15)
      }))
      replication_time = optional(object({
        status       = optional(bool, false)
        time_minutes = optional(number, 15)
      }))
      source_selection_criteria = optional(object({
        replica_modifications     = optional(bool, false)
        sse_kms_encrypted_objects = optional(bool, false)
      }))
    }))
  })
  default     = null
  description = "Bucket replication configuration settings, specify the rules map keys as integers as these are used to determine the priority of the rules in case of conflict."
}

variable "request_payer" {
  type        = string
  default     = "BucketOwner"
  description = "The request payer for the bucket, defaults to BucketOwner. Valid values: BucketOwner, Requester."

  validation {
    condition     = contains(["BucketOwner", "Requester"], var.request_payer)
    error_message = "Allowed values for request_payer are 'BucketOwner' or 'Requester'."
  }
}

variable "restrict_public_buckets" {
  type        = bool
  default     = true
  description = "Whether Amazon S3 should restrict public bucket policies for this bucket."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "A mapping of tags to assign to the bucket."
}

variable "transition_default_minimum_object_size" {
  type        = string
  default     = null
  description = "The default minimum object size behavior applied to the lifecycle configuration. Valid values: all_storage_classes_128K (default), varies_by_storage_class"
  validation {
    condition     = var.transition_default_minimum_object_size != null ? contains(["all_storage_classes_128K", "varies_by_storage_class"], var.transition_default_minimum_object_size) : true
    error_message = "Allowed values for transition_default_minimum_object_size are \"all_storage_classes_128K\", \"varies_by_storage_class\"."
  }
}

variable "versioning" {
  type        = bool
  default     = true
  description = "Versioning is a means of keeping multiple variants of an object in the same bucket."
}

