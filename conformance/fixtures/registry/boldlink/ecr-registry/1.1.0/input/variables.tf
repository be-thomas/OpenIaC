variable "scanning_configuration" {
  description = "Scanning configuration settings"
  type        = any
  default     = {}
}

variable "enable_pull_through_cache_rule" {
  description = "Whether to enable pull through cache rule"
  type        = bool
  default     = false
}

variable "ecr_repository_prefix" {
  description = "(Required if pull-through cache is enabled) The repository name prefix to use when caching images from the source registry."
  type        = string
  default     = ""
}

variable "upstream_registry_url" {
  description = "(Required if pull-through cache is enabled) The registry URL of the upstream public registry to use as the source."
  type        = string
  default     = ""
}

variable "replication_configuration" {
  description = "Replication configuration settings"
  type        = any
  default     = {}
}

# Repository configuration with encryption always enabled
variable "repository_name" {
  description = "The name of the ECR repository to create. If empty, no repository is created."
  type        = string
  default     = ""
}

variable "image_tag_mutability" {
  description = "The tag mutability setting for the repository (MUTABLE or IMMUTABLE)"
  type        = string
  default     = "MUTABLE"
}

variable "encryption_type" {
  description = "Encryption type for the repository images. Valid values are AES256 (default: AWS/ECR shared key) or KMS."
  type        = string
  default     = "AES256"
}

variable "kms_key_id" {
  description = "If using KMS encryption, the ARN of the KMS key to use. Leave empty when using AES256."
  type        = string
  default     = ""
}

variable "tags" {
  description = "Tags to apply to created resources"
  type        = map(string)
  default     = {}
}

variable "repository_policy" {
  description = "The repository-level policy document (a JSON formatted string) to attach to the repository. Set to null if not required."
  type        = string
  default     = null
}
