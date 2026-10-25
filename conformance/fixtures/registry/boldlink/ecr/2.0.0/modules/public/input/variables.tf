
variable "create_public_repository" {
  description = "Whether to create a public repository or not."
  type        = bool
  default     = false
}

variable "repository_name" {
  description = "(Required) Name of the repository."
  type        = string
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
