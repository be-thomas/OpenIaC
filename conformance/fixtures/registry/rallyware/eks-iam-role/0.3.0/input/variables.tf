variable "service_account_name" {
  type        = string
  description = "Kubernetes ServiceAccount name"
}

variable "service_account_namespace" {
  type        = string
  description = "Kubernetes Namespace where service account is deployed"
}

variable "aws_account_number" {
  type        = string
  default     = null
  description = "AWS account number of EKS cluster owner. If an AWS account number is not provided, the current aws provider account number will be used."
}

variable "aws_partition" {
  type        = string
  default     = null
  description = "AWS partition: `aws`, `aws-cn`, or `aws-us-gov`"
}

variable "aws_iam_policy_document" {
  type        = string
  default     = "{}"
  description = "JSON string representation of the IAM policy for this service account"
}

variable "aws_iam_policy_arns" {
  type        = list(string)
  default     = null
  description = "List of exclusive IAM managed policy ARNs to attach to the IAM role"
}

variable "eks_cluster_oidc_issuer_url" {
  type        = string
  description = "OIDC issuer URL for the EKS cluster (initial \"https://\" may be omitted)"
}

variable "managed_iam_policy_enabled" {
  type        = bool
  description = "Create a managed IAM policy that can be reused. Set to `false` to use an inline IAM policy."
  default     = true
}

variable "additional_service_accounts" {
  type        = list(any)
  description = "Additional service accounts to handle"
  default     = []
}

variable "path" {
  type        = string
  default     = null
  description = "Path to the role"
}

variable "use_name_prefix" {
  type        = bool
  default     = false
  description = "Whether to use a name prefix for the IAM role"
}
