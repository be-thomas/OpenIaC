variable "allow_self_invocation" {
  default     = false
  description = "If true, allows this Lambda function to invoke itself. Useful for recursive invocations"
  type        = bool
}

variable "architectures" {
  default     = ["x86_64"]
  description = "Instruction set architecture for your Lambda function. Valid values are [x86_64] and [arm64]. Default is [x86_64]"
  type        = list(string)
}

variable "dead_letter_arn" {
  description = "The arn for the SNS topic that handles dead letters"
  type        = string
}

variable "description" {
  default     = ""
  description = "Description of what your Lambda Function does"
  type        = string
}

variable "destinations_on_failure_arn" {
  description = "Add destination to function. An ARN of SQS queue, SNS topic, Lambda function, or EventBridge event bus. Added as on_failure aysynchronous invocation"
  type        = string
  default     = ""
}

variable "destinations_on_success_arn" {
  description = "Add destination to function. An ARN of SQS queue, SNS topic, Lambda function, or EventBridge event bus. Added as on_success aysynchronous invocation"
  type        = string
  default     = ""
}

variable "environment_variables" {
  default = {
    DEFAULT = "default"
  }
  description = "The map of environment variables to give to the Lambda function"
  type        = map(any)
}

variable "handler" {
  description = "The handler for the lambda function"
  type        = string
}

variable "kms_key_arn" {
  description = "The arn of the KMS key used to encrypt the environment variables"
  type        = string
}

variable "layers" {
  default     = []
  description = "The list of lambda layers to be attached"
  type        = list(string)
}

variable "log_retention_days" {
  default     = 7
  description = "The number of days you want to retain log events in the specified log group"
  type        = number
}

variable "memory_size" {
  default     = 128
  description = "The memory allocation for the function"
  type        = number
}

variable "name" {
  description = "The name of the function"
  type        = string
}

variable "policy_arns" {
  default     = []
  description = "A list of additional policy arns to attach to the function's role"
  type        = list(string)
}


variable "provisioned_concurrency" {
  default     = false
  description = "If true, allows this Lambda function to create alias and manage provisioned_concurrency_config"
  type        = bool
}

variable "provisioned_concurrent_executions" {
  default     = 1
  description = "Amount of capacity to allocate. Must be greater than or equal to 1"
  type        = number
}

variable "runtime" {
  description = "The runtime the function should use"
  type        = string
}

variable "s3_bucket" {
  description = "The name or id of the S3 bucket that contains the function package"
  type        = string
}

variable "s3_object_key" {
  description = "The key of the function package in the s3_bucket"
  type        = string
}

variable "tags" {
  default     = {}
  description = "Key-value map of tags"
  type        = map(any)
}

variable "timeout" {
  default     = 3
  description = "The timeout to apply to the function"
  type        = number
}

variable "vpc_config" {
  description = "Provide this to allow your function to access your VPC. Both 'subnet_ids' and 'security_group_ids' are a list of strings are required to enable vpc)."
  type        = map(list(string))
  default     = {}
}

variable "vpc_enabled" {
  default     = true
  description = "Lambda is put in VPC if this bool is true and vpc_config is passed in"
  type        = bool
}
