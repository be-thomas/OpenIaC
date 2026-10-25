variable "agent_role_arns" {
  type        = list(string)
  default     = null
  description = "IAM role ARNs used by Terraform Cloud Agent to assume role in the created account"
}

variable "auth_method" {
  type        = string
  default     = "iam_role_oidc"
  description = "Configures how the workspace authenticates with the AWS account (can be iam_user, iam_role, or iam_role_oidc)"

  validation {
    condition     = lower(var.auth_method) == "iam_user" || lower(var.auth_method) == "iam_role" || lower(var.auth_method) == "iam_role_oidc"
    error_message = "The auth_method value must be either \"iam_user\", \"iam_role\", or \"iam_role_oidc\"."
  }
}

variable "oidc_settings" {
  type = object({
    audience              = optional(string, "aws.workload.identity")
    oidc_project_filter   = optional(string, "*")
    oidc_workspace_filter = string
    provider_arn          = string
    site_address          = optional(string, "app.terraform.io")
  })
  default     = null
  description = "OIDC settings to use if \"auth_method\" is set to \"iam_role_oidc\""
}

variable "path" {
  type        = string
  default     = "/"
  description = "Path in which to create the IAM role or user"
}

variable "permissions_boundary_arn" {
  type        = string
  default     = null
  description = "ARN of the policy that is used to set the permissions boundary for the IAM role or IAM user"
}

variable "policy" {
  type        = string
  default     = null
  description = "The policy to attach to the pipeline role or user"
}

variable "policy_arns" {
  type        = set(string)
  default     = []
  description = "A set of policy ARNs to attach to the pipeline user"
}

variable "postfix" {
  type        = string
  default     = true
  description = "Whether to postfix the IAM resources with e.g. `User` and `Role`"
}

variable "role_name" {
  type        = string
  default     = null
  description = "The IAM role name for a new pipeline role"
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "A mapping of tags to assign to resource"
}

variable "terraform_organization" {
  type        = string
  default     = null
  description = "The Terraform Enterprise organization to create the workspace in. If omitted, organization must be defined in the provider config."
}

variable "username" {
  type        = string
  default     = null
  description = "The username for a new pipeline user"
}

variable "workspace_id" {
  type        = string
  default     = null
  description = "The TFE workspace ID to attach variables to"

  validation {
    condition     = (var.workspace_id != null) != (var.variable_set_id != null)
    error_message = "Exactly one of workspace_id or variable_set_id must be specified."
  }
}

variable "variable_set_id" {
  type        = string
  default     = null
  description = "The TFE variable set ID to attach variables to"
}
