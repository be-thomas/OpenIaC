
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
