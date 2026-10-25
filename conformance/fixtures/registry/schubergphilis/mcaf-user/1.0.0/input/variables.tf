variable "create_policy" {
  type        = bool
  default     = null
  description = "Overrule whether the user role policy has to be created."
}

variable "create_iam_access_key" {
  type        = bool
  default     = true
  description = "Overrule whether the user IAM access keys have to be created"
}

variable "force_destroy" {
  type        = bool
  default     = false
  description = "Destroy the user even if it has non-terraform-managed IAM access keys, login profile or MFA devices"
}

variable "groups" {
  type        = set(string)
  default     = []
  description = "Set of group names to attach to the user."
}

variable "kms_key_id" {
  type        = string
  default     = null
  description = "The KMS key ID used to encrypt the SSM parameters."
}

variable "name" {
  type        = string
  description = "The name of the user."
}

variable "path" {
  type        = string
  default     = "/"
  description = "Path in which to create the user."
}

variable "permissions_boundary" {
  type        = string
  default     = null
  description = "The ARN of the policy that is used to set the permissions boundary for the user."
}

variable "policy" {
  type        = string
  default     = null
  description = "The policy to attach to the user."
}

variable "policy_arns" {
  type        = set(string)
  default     = []
  description = "A set of policy ARNs to attach to the user."
}

variable "postfix" {
  type        = bool
  default     = true
  description = "Postfix the user, policy and group names with Account, Policy and Group."
}

variable "ssm_ses_smtp_password_v4" {
  type        = bool
  default     = false
  description = "Store the user's SES SMTP password in the SSM Parameter Store."
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "A mapping of tags to assign to the user."
}
