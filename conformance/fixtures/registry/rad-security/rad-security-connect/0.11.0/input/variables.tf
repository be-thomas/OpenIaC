variable "rad-security_assumed_role_arn" {
  type        = string
  description = "Rad Security Role that will assume the rad-security-connect IAM role you create to interact with resources in your account"
  default     = "arn:aws:iam::955322216602:role/rad-security-connector"
}

variable "rad-security_deprecated_assumed_role_arn" {
  type        = string
  description = "Deprecated Rad Security Role that will assume the rad-security-connect IAM role you create to interact with resources in your account. This role will be removed in the future."
  default     = "arn:aws:iam::955322216602:role/ksoc-connector"
}

variable "enable_eks_audit_logs_pipeline" {
  type        = bool
  description = "Enable EKS Audit Logs Pipeline (CloudWatch Logs -> FireHose -> S3)"
  default     = false
}

variable "eks_audit_logs_bucket_versioning_enabled" {
  type        = bool
  description = "Enable versioning for the S3 bucket that will store EKS audit logs"
  default     = true
}

variable "eks_audit_logs_bucket_object_age" {
  type        = number
  description = "The number of days to retain the objects in the S3 bucket that will store EKS audit logs"
  default     = 7
}

variable "rad-security_eks_audit_logs_assumed_role_arn" {
  type        = string
  description = "Rad Security Role dedicated for EKS audit logs that will be allowed to assume"
  default     = "arn:aws:iam::955322216602:role/ksoc-data-pipeline"
}

variable "eks_audit_logs_filter_pattern" {
  type        = string
  default     = "{ $.stage = \"ResponseComplete\" && $.requestURI != \"/version\" && $.requestURI != \"/version?*\" && $.requestURI != \"/metrics\" && $.requestURI != \"/metrics?*\" && $.requestURI != \"/logs\" && $.requestURI != \"/logs?*\" && $.requestURI != \"/swagger*\" && $.requestURI != \"/livez*\" && $.requestURI != \"/readyz*\" && $.requestURI != \"/healthz*\" }"
  description = "The Cloudwatch Log Subscription Filter pattern"
}

variable "secondary_region" {
  type        = bool
  description = "Enable this if running in a another region. It will disable the creation of global resources."
  default     = false
}

variable "eks_audit_log_firehose_role_arn" {
  type        = string
  description = "The ARN of the IAM role that will be used to write to the Firehose. Required for secondary regions."
  default     = ""
}

variable "eks_audit_log_cloudwatch_role_arn" {
  type        = string
  description = "The ARN of the IAM role that CloudWatch Logs will use to send data to the Firehose. Required for secondary regions."
  default     = ""
}

variable "tags" {
  description = "A set of tags to associate with the resources in this module."
  type        = map(string)
  default     = {}
}

variable "aws_external_id" {
  description = "External ID to use when connecting an AWS account with RAD Security"
  type        = string
  default     = ""
}

variable "readonly_s3_buckets" {
  description = "List of S3 buckets containing logs (CloudTrail, VPC flow logs, etc.) that RAD Security will be allowed to read"
  type        = list(string)
  default     = []
}

variable "enable_ecr_image_pull" {
  description = "Allow RAD Security to pull container images from ECR in this account, so it can generate SBOMs and scan images running in ECS/Fargate"
  type        = bool
  default     = false
}

variable "ecr_repository_arns" {
  description = "ECR repository ARNs RAD Security is allowed to pull images from when enable_ecr_image_pull is true. Defaults to all repositories in the account"
  type        = list(string)
  default     = ["*"]
}
