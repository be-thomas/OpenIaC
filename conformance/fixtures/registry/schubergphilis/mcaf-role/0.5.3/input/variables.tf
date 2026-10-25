variable "assume_policy" {
  type        = string
  default     = null
  description = "The assume policy to attach to the role."
}

variable "create_policy" {
  type        = bool
  default     = null
  description = "Overrule whether the IAM role policy has to be created."
}

variable "description" {
  type        = string
  default     = null
  description = "The description of the role."
}

variable "force_detach_policies" {
  type        = bool
  default     = false
  description = "Force detaching any policies the role has before destroying it."
}

variable "max_session_duration" {
  type        = number
  default     = 3600
  description = "The maximum session duration (in seconds) for the role."
}

variable "name_prefix" {
  type        = string
  default     = null
  description = "Name prefix of the role. Conflicts with `name`."
}

variable "name" {
  type        = string
  default     = null
  description = "Name of the role. Conflicts with `name_prefix`."

  validation {
    condition     = (var.name == null && var.name_prefix != null) || (var.name != null && var.name_prefix == null)
    error_message = "Exactly one of 'name' or 'name_prefix' must be defined."
  }
}

variable "path" {
  type        = string
  default     = "/"
  description = "Path to the role."

  validation {
    condition     = length(regex("^(\\/|\\/.+\\/)$", var.path)) > 0
    error_message = "The path must be '/' or start and end with '/' (e.g., '/custom/')."
  }
}

variable "permissions_boundary" {
  type        = string
  default     = null
  description = "The permissions boundary to set on the role."
}

variable "principal_type" {
  type        = string
  default     = ""
  description = "The type of assume role principal."
}

variable "principal_identifiers" {
  type        = list(string)
  default     = []
  description = "List of identifiers for principals."
}

variable "policy_arns" {
  type        = set(string)
  default     = []
  description = "A set of policy ARNs to attach to the role."
}

variable "postfix" {
  type        = bool
  default     = true
  description = "Postfix the role and policy `name` or `name_prefix` with `Role` and `Policy`."
}

variable "role_policy" {
  type        = string
  default     = null
  description = "The JSON encoded IAM policy to attach to the role."
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "A mapping of tags to assign to the role."
}
