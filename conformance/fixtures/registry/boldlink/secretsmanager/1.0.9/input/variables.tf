# Secret
variable "description" {
  description = "(Optional) Description of the secret."
  type        = string
  default     = null
}

variable "kms_key_id" {
  description = "(Optional) ARN or Id of the AWS KMS key to be used to encrypt the secret values in the versions stored in this secret. If you don't specify this value, then Secrets Manager defaults to using the AWS account's default KMS key (the one named aws/secretsmanager).If the default KMS key with that name doesn't yet exist, then AWS Secrets Manager creates it for you automatically the first time."
  type        = string
  default     = null
}

variable "name_prefix" {
  description = " (Optional) Creates a unique name beginning with the specified prefix. Conflicts with name."
  type        = string
  default     = null
}

variable "name" {
  description = "(Optional) Friendly name of the new secret. The secret name can consist of uppercase letters, lowercase letters, digits, and any of the following characters: /_+=.@- Conflicts with name_prefix."
  type        = string
  default     = null
}

variable "recovery_window_in_days" {
  description = "(Optional) Number of days that AWS Secrets Manager waits before it can delete the secret. This value can be 0 to force deletion without recovery or range from 7 to 30 days. The default value is 30."
  type        = number
  default     = 30
}

variable "force_overwrite_replica_secret" {
  description = "(Optional) Accepts boolean value to specify whether to overwrite a secret with the same name in the destination Region."
  type        = bool
  default     = false

}

#Tags
variable "tags" {
  description = " A map of tags to assign to the object. If configured with a provider default_tags configuration block present, tags with matching keys will overwrite those defined at the provider-level."
  type        = map(string)
  default     = {}
}

# Secret_Rotation
variable "enable_secretsmanager_secret_rotation" {
  description = "Whether to enable secrets rotation"
  type        = bool
  default     = false
}

variable "rotation_lambda_arn" {
  description = " (Required) Specifies the ARN of the Lambda function that can rotate the secret."
  type        = string
  default     = ""
}

variable "automatically_after_days" {
  description = "(Required) Specifies the number of days between automatic scheduled rotations of the secret."
  type        = number
  default     = 30
}

# Secret_Version
variable "enable_secretsmanager_secret_version" {
  description = "Whether to enable secret version"
  type        = bool
  default     = true
}

variable "secrets" {
  description = "Map of secrets to store."
  type        = any
  default     = {}
}

# Secret_policy
variable "secret_policy" {
  description = "(Required) Valid JSON document representing a resource policy."
  type        = string
  default     = null
}

variable "block_public_policy" {
  description = "(Optional) Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret."
  type        = bool
  default     = false
}
