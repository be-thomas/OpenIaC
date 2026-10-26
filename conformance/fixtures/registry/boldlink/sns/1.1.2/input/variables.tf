#########################
## sns topic
#########################

variable "name" {
  type        = string
  description = "(Optional) The name of the topic. Topic names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and must be between 1 and 256 characters long. For a FIFO (first-in-first-out) topic, the name must end with the `.fifo` suffix. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`"
  default     = null
}

variable "key_deletion_window" {
  type        = number
  description = "The waiting period, specified in number of days. After the waiting period ends, AWS KMS deletes the KMS key. If you specify a value, it must be between 7 and 30, inclusive. If you do not specify a value, it defaults to 30."
  default     = 7
}

variable "name_prefix" {
  type        = string
  description = "(Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`"
  default     = null
}

variable "display_name" {
  type        = string
  description = "(Optional) The display name for the topic"
  default     = null
}

variable "create_sns_topic_policy" {
  description = "Whether to create sns topic policy or not"
  type        = bool
  default     = false
}

variable "policy" {
  type        = string
  description = "(Optional) The fully-formed AWS policy as JSON."
  default     = null
}

variable "delivery_policy" {
  type        = string
  description = "(Optional) The SNS delivery policy."
  default     = null
}

variable "application_success_feedback_role_arn" {
  type        = string
  description = "(Optional) The IAM role permitted to receive success feedback for this topic"
  default     = null
}

variable "application_success_feedback_sample_rate" {
  type        = string
  description = "(Optional) Percentage of success to sample"
  default     = null
}

variable "application_failure_feedback_role_arn" {
  type        = string
  description = "(Optional) IAM role for failure feedback"
  default     = null
}

variable "http_success_feedback_role_arn" {
  type        = string
  description = "(Optional) The IAM role permitted to receive success feedback for this topic"
  default     = null
}

variable "http_success_feedback_sample_rate" {
  type        = string
  description = "(Optional) Percentage of success to sample"
  default     = null
}

variable "http_failure_feedback_role_arn" {
  type        = string
  description = "(Optional) IAM role for failure feedback"
  default     = null
}

variable "kms_master_key_arn" {
  type        = string
  description = "(Optional) The ARN of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK."
  default     = null
}

variable "fifo_topic" {
  type        = bool
  description = "(Optional) Boolean indicating whether or not to create a FIFO (first-in-first-out) topic (default is `false`)."
  default     = false
}

variable "content_based_deduplication" {
  type        = bool
  description = "(Optional) Enables content-based deduplication for FIFO topics."
  default     = false
}

variable "lambda_success_feedback_role_arn" {
  type        = string
  description = "(Optional) The IAM role permitted to receive success feedback for this topic"
  default     = null
}

variable "lambda_success_feedback_sample_rate" {
  type        = number
  description = "(Optional) Percentage of success to sample"
  default     = null
}

variable "lambda_failure_feedback_role_arn" {
  type        = string
  description = "(Optional) IAM role for failure feedback"
  default     = null
}

variable "sqs_success_feedback_role_arn" {
  type        = string
  description = "(Optional) The IAM role permitted to receive success feedback for this topic"
  default     = null
}

variable "sqs_success_feedback_sample_rate" {
  type        = number
  description = "(Optional) Percentage of success to sample"
  default     = null
}

variable "sqs_failure_feedback_role_arn" {
  type        = string
  description = "(Optional) IAM role for failure feedback"
  default     = null
}

variable "firehose_success_feedback_role_arn" {
  type        = string
  description = "(Optional) The IAM role permitted to receive success feedback for this topic"
  default     = null
}

variable "firehose_success_feedback_sample_rate" {
  type        = number
  description = "(Optional) Percentage of success to sample"
  default     = null
}

variable "firehose_failure_feedback_role_arn" {
  type        = string
  description = "(Optional) IAM role for failure feedback"
  default     = null
}

variable "tags" {
  type        = map(string)
  description = "(Optional) Key-value map of resource tags. If configured with a provider [default_tags configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level."
  default     = {}
}

#########################
## sns topic subscription
#########################

variable "sns_topic_subscriptions" {
  type        = any
  description = "Resource block for sns topic subscriptions"
  default     = {}
}

######################
### Encryption
######################
variable "create_kms_key" {
  type        = bool
  description = "Choose whether to create kms key for sns"
  default     = false
}

variable "enable_key_rotation" {
  type        = bool
  description = "Choose whether to enable key rotation"
  default     = true
}
