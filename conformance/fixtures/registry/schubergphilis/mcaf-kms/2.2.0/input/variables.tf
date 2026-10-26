variable "custom_key_store_id" {
  type        = string
  default     = null
  description = "ID of the KMS Custom Key Store where the key will be stored instead of KMS (eg CloudHSM)"
}

variable "customer_master_key_spec" {
  type        = string
  default     = "SYMMETRIC_DEFAULT"
  description = "Specifies whether the key contains a symmetric key or an asymmetric key pair and the encryption algorithms or signing algorithms that the key supports."

  validation {
    condition     = contains(["SYMMETRIC_DEFAULT", "RSA_2048", "RSA_3072", "RSA_4096", "HMAC_256", "ECC_NIST_P256", "ECC_NIST_P384", "ECC_NIST_P521", "ECC_SECG_P256K1"], var.customer_master_key_spec)
    error_message = "Allowed values for customer_master_key_spec are \"SYMMETRIC_DEFAULT\", \"RSA_2048\", \"RSA_3072\", \"RSA_4096\", \"HMAC_256\", \"ECC_NIST_P256\", \"ECC_NIST_P384\", \"ECC_NIST_P521\", \"ECC_SECG_P256K1\"."
  }
}

variable "deletion_window_in_days" {
  type        = number
  default     = 30
  description = "Delay in days after which the key is deleted"
}

variable "description" {
  type        = string
  default     = null
  description = "The description of the key as viewed in AWS console"
}

variable "enable_key_rotation" {
  type        = bool
  default     = true
  description = "Specifies whether key rotation is enabled"
}

variable "key_usage" {
  type        = string
  default     = "ENCRYPT_DECRYPT"
  description = "Specifies the intended use of the key."

  validation {
    condition     = contains(["ENCRYPT_DECRYPT", "SIGN_VERIFY", "GENERATE_VERIFY_MAC"], var.key_usage)
    error_message = "Allowed values for key_usage are \"ENCRYPT_DECRYPT\", \"SIGN_VERIFY\", \"GENERATE_VERIFY_MAC\"."
  }
}

variable "multi_region" {
  type        = bool
  default     = false
  description = "Indicates whether the KMS key is a multi-Region (`true`) or regional (`false`) key."
}

variable "name" {
  type        = string
  description = "The name of the key"
}

variable "path" {
  type        = string
  default     = "/"
  description = "Path of the KMS key alias"

  validation {
    condition     = length(regex("^(\\/|\\/.+\\/)$", var.path)) > 0
    error_message = "The path must be '/' or start and end with '/' (e.g., '/custom/')."
  }
}

variable "policy" {
  type        = string
  default     = null
  description = "A valid KMS policy JSON document"
}

variable "default_policy" {
  type = object({
    enable                    = optional(bool, true)
    override_policy_documents = optional(list(string), [])
    source_policy_documents   = optional(list(string), [])

    iam_all_principals_read  = optional(bool, true)
    iam_arns_administrator   = optional(list(string), [])
    iam_arns_decrypt         = optional(list(string), [])
    iam_arns_decrypt_encrypt = optional(list(string), [])
    iam_arns_owner           = optional(list(string), [])
    iam_arns_service_grant   = optional(list(string), [])
    iam_arns_sign_verify     = optional(list(string), [])
    iam_aws_config_read      = optional(bool, true)
  })
  default     = {}
  description = "Configuration object for defining the KMS key policy and permissions. Use `override_policy_documents` to add statements that override the default policy, or `source_policy_documents` to add statements that are merged with the default policy"
}

variable "region" {
  type        = string
  default     = null
  description = "The AWS region where resources will be created; if omitted the default provider region is used"
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "A mapping of tags to assign to the key"
}

variable "xks_key_id" {
  type        = string
  default     = null
  description = "Identifies the external key that serves as key material for the KMS key in an external key store"

  validation {
    condition     = var.xks_key_id == null || var.custom_key_store_id != null
    error_message = "xks_key_id can only be set when custom_key_store_id references an external key store."
  }
}
