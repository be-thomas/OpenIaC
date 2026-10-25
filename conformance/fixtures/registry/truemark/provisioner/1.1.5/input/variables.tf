variable "name" {
  description = "Name of the IAM user/role to create"
  type        = string
}

variable "description" {
  description = "Set a description for the purpose of the IAM user role"
  type        = string
  default     = "User access role created by TrueMark terraform module terraform-aws-provisioner."
}

variable "create_user" {
  description = "Default is True to create an IAM user"
  type        = bool
  default     = false
}

variable "create_role" {
  description = "Default is False. Set to True to create an IAM role"
  type        = bool
  default     = true
}

variable "s3_bucket" {
  description = "Name of the terraform bucket holding terraform state information. Defaults to <account>-terraform."
  default     = null
  type        = string
}

variable "s3_prefix" {
  description = "Optional path prefix inside the terraform S3 bucket to grant access to."
  default     = null
  type        = string
}

variable "policy" {
  description = "The policy to create and apply to the IAM user."
  default     = null
  type        = string
}

variable "policies" {
  description = "The policies to create and apply to the IAM user."
  default     = []
  type        = list(string)
}

variable "policy_arns" {
  description = "Additional policies to attach to the IAM user."
  default     = []
  type        = list(string)
}

variable "path" {
  default = "/"
  type    = string
}

variable "trusted_account" {
  default = null
  type    = string
}

variable "max_session_duration" {
  default     = "7200"
  description = "The maximum session duration (in seconds) that you want to set for the specified role. If you do not specify a value for this setting, the default maximum of two hour is applied. This setting can have a value from 1 hour to 12 hours."
}
