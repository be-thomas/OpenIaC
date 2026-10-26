# Private Repository
variable "create_private_repository" {
  description = "Whether to create a private repository or not."
  type        = bool
  default     = false
}

variable "scan_on_push" {
  description = "(Required) Indicates whether images are scanned after being pushed to the repository (true) or not scanned (false)."
  default     = true
  type        = bool
}

variable "name" {
  description = "(Required) Name of the repository."
  type        = string
}

variable "encryption_configuration" {
  description = "(Optional) Encryption configuration for the repository. "
  default = {
    encryption_type = "KMS"
    kms_key         = null
  }
  type = object({
    encryption_type = string
    kms_key         = any
  })
}

variable "image_tag_mutability" {
  description = "(Optional) The tag mutability setting for the repository. Must be one of: MUTABLE or IMMUTABLE. Defaults to MUTABLE."
  default     = "IMMUTABLE"
  type        = string
}

variable "other_tags" {
  description = " Any additional values for tags"
  type        = map(string)
  default     = {}
}

variable "aws_ecr_repository_policy" {
  description = "(Required) The policy document. This is a JSON formatted string."
  type        = string
  default     = ""
}

variable "aws_ecr_lifecycle_policy" {
  description = "(Required) The policy document. This is a JSON formatted string.."
  type        = string
  default     = ""
}

# Public Repository
variable "create_public_repository" {
  description = "Whether to create a public repository or not."
  type        = bool
  default     = false
}

variable "catalog_data" {
  description = "(Optional) Catalog data configuration for the repository. "
  type        = any
  default     = []
}

variable "timeouts" {
  description = "aws_ecrpublic_repository provides the following Timeouts configuration options:delete"
  type        = map(any)
  default     = {}
}

variable "create_ecrpublic_repository_policy" {
  description = "whether to create ecr public repository policy or not"
  type        = bool
  default     = false
}

variable "policy" {
  description = "(Required) The policy document. This is a JSON formatted string."
  type        = string
  default     = ""
}
