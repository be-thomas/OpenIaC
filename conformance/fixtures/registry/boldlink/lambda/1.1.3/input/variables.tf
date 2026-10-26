variable "additional_lambda_permissions" {
  type        = any
  description = "Add additional iam policy statements for the lambda created by this module"
  default     = {}
}

variable "function_name" {
  type        = string
  description = "(Required) Unique name for your Lambda Function."
}

variable "architectures" {
  type        = list(string)
  description = "(Optional) Instruction set architecture for your Lambda function. Valid values are `[\"x86_64\"]` and `[\"arm64\"]`. Default is `[\"x86_64\"]`. Removing this attribute, function's architecture stay the same."
  default     = ["x86_64"]
}

variable "code_signing_config_arn" {
  type        = string
  description = "(Optional) To enable code signing for this function, specify the ARN of a code-signing configuration. A code-signing configuration includes a set of signing profiles, which define the trusted publishers for this function."
  default     = null
}

variable "description" {
  type        = string
  description = "(Optional) Description of what your Lambda Function does."
  default     = null
}

variable "filename" {
  type        = string
  description = "(Optional) Path to the function's deployment package within the local filesystem. Conflicts with `image_uri`, `s3_bucket`, `s3_key`, and `s3_object_version`."
  default     = null
}

variable "handler" {
  type        = string
  description = "(Optional) Function entrypoint in your code."
  default     = null
}

variable "image_uri" {
  type        = string
  description = "(Optional) ECR image URI containing the function's deployment package. Conflicts with `filename`, `s3_bucket`, `s3_key`, and `s3_object_version`."
  default     = null
}

variable "kms_key_arn" {
  type        = string
  description = "(Optional) Amazon Resource Name (ARN) of the AWS Key Management Service (KMS) key that is used to encrypt environment variables. If this configuration is not provided when environment variables are in use, AWS Lambda uses a default service key. If this configuration is provided when environment variables are not in use, the AWS Lambda API does not save this configuration and Terraform will show a perpetual difference of adding the key. To fix the perpetual difference, remove this configuration."
  default     = null
}

variable "layers" {
  type        = map(any)
  description = "(Optional) Map of Lambda Layers to create and attach to your Lambda Function"
  default     = {}
}

variable "additional_layers" {
  type        = list(string)
  description = "(Optional) List of Lambda Layer Version (maximum of 5) to attach to your Lambda Function"
  default     = []
}

variable "memory_size" {
  type        = number
  description = "(Optional) Amount of memory in MB your Lambda Function can use at runtime. Defaults to `128`"
  default     = 128
}

variable "package_type" {
  type        = string
  description = "(Optional) Lambda deployment package type. Valid values are `Zip` and `Image`. Defaults to `Zip`"
  default     = "Zip"
}

variable "publish" {
  type        = bool
  description = "(Optional) Whether to publish creation/change as new Lambda Function Version. Defaults to `false`"
  default     = false
}

variable "reserved_concurrent_executions" {
  type        = number
  description = "(Optional) Amount of reserved concurrent executions for this lambda function. A value of `0` disables lambda from being triggered and `-1` removes any concurrency limitations. Defaults to Unreserved Concurrency Limits `-1`"
  default     = -1
}

variable "runtime" {
  type        = string
  description = "(Optional) Identifier of the function's runtime."
  default     = null
}

variable "s3_bucket" {
  type        = string
  description = "(Optional) S3 bucket location containing the function's deployment package. Conflicts with filename and image_uri. This bucket must reside in the same AWS region where you are creating the Lambda function."
  default     = null
}

variable "s3_key" {
  type        = string
  description = "(Optional) S3 key of an object containing the function's deployment package. Conflicts with `filename` and `image_uri`."
  default     = null
}

variable "s3_object_version" {
  type        = string
  description = "(Optional) Object version containing the function's deployment package. Conflicts with filename and image_uri."
  default     = null
}

variable "source_code_hash" {
  type        = string
  description = "(Optional) Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the package file specified with either filename or s3_key. The usual way to set this is `filebase64sha256(\"file.zip\")` (Terraform 0.11.12 and later) or `base64sha256(file(\"file.zip\"))` (Terraform 0.11.11 and earlier), where \"file.zip\" is the local filename of the lambda function source archive."
  default     = null
}

variable "tags" {
  type        = map(string)
  description = "(Optional) Map of tags to assign to the object."
  default     = {}
}

variable "timeout" {
  type        = number
  description = "(Optional) Amount of time your Lambda Function has to run in seconds."
  default     = 7
}

variable "dead_letter_config" {
  type        = map(any)
  description = "Dead letter queue configuration that specifies the queue or topic where Lambda sends asynchronous events when they fail processing."
  default     = {}
}

variable "environment" {
  type        = map(any)
  description = "(Optional) Map of environment variables that are accessible from the function code during execution."
  default     = {}
}

variable "ephemeral_storage" {
  type        = map(any)
  description = "Configuration block for the size of the Lambda function Ephemeral storage(/tmp) represented in MB."
  default     = {}
}

variable "file_system_config" {
  type        = map(any)
  description = "Connection settings for an EFS file system. Before creating or updating Lambda functions with `file_system_config`, EFS mount targets must be in available lifecycle state."
  default     = {}
}

variable "image_config" {
  type        = map(any)
  description = "Container image configuration values that override the values in the container image Dockerfile."
  default     = {}
}

variable "tracing_config" {
  type        = map(any)
  description = "Configuration block for whether to to sample and trace a subset of incoming requests with AWS X-Ray"
  default     = {}
}

variable "timeouts" {
  type        = map(string)
  description = "Timeouts configuration for creating the resource"
  default     = {}
}

variable "enable_key_rotation" {
  type        = bool
  description = "Choose whether to enable key rotation for cloudwatch kms key"
  default     = true
}

variable "key_deletion_window_in_days" {
  type        = number
  description = "The number of days before the cloudwatch kms key is deleted"
  default     = 7
}

variable "retention_in_days" {
  type        = number
  description = "Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire."
  default     = 1827
}

variable "lambda_permissions" {
  type        = list(any)
  description = "Configuration to give external source(s) (like an EventBridge Rule, SNS, or S3) permission to access the Lambda function."
  default     = []
}

variable "layer_permission" {
  type        = map(any)
  description = "Configuration to allow sharing of Lambda Layers to another account by account ID, to all accounts in AWS organization or even to all AWS accounts."
  default     = {}
}

## Alias
variable "alias" {
  type        = map(any)
  description = "Configuration for lambda alias"
  default     = {}
}

## Lambda Invocation
variable "create_lambda_invocation" {
  description = "Whether to create lambda invocation resource"
  type        = bool
  default     = false
}

variable "input" {
  description = "(Required) JSON payload to the lambda function."
  type        = any
  default     = ""
}

variable "lifecycle_scope" {
  description = "(Optional) Lifecycle scope of the resource to manage. Valid values are CREATE_ONLY and CRUD. Defaults to CREATE_ONLY. CREATE_ONLY will invoke the function only on creation or replacement. CRUD will invoke the function on each lifecycle event, and augment the input JSON payload with additional lifecycle information."
  type        = string
  default     = "CREATE_ONLY"
}

variable "qualifier" {
  description = "(Optional) Qualifier (i.e., version) of the lambda function. Defaults to $LATEST."
  type        = string
  default     = "$LATEST"
}

variable "terraform_key" {
  description = "(Optional) The JSON key used to store lifecycle information in the input JSON payload. Defaults to tf. This additional key is only included when lifecycle_scope is set to CRUD."
  type        = string
  default     = "tf"
}

variable "triggers" {
  description = "(Optional) Map of arbitrary keys and values that, when changed, will trigger a re-invocation. "
  type        = map(any)
  default     = {}
}


## Security group
variable "security_group_ids" {
  description = "List of security group IDs associated with the Lambda function."
  type        = list(string)
  default     = []
}

variable "subnet_ids" {
  description = "List of subnet IDs associated with the Lambda function."
  type        = list(string)
  default     = null
}

variable "vpc_id" {
  description = "(Optional, Forces new resource) VPC ID. Defaults to the region's default VPC."
  type        = string
  default     = null
}

variable "create_security_group" {
  description = "Whether to create a Security Group for the lambda function."
  default     = false
  type        = bool
}

variable "security_group_ingress_rules" {
  description = "(Optional) Ingress rules to add to the security group"
  type        = any
  default     = {}
}

variable "security_group_egress_rules" {
  description = "(Optional) Egress rules to add to the security group"
  type        = any
  default     = {}
}

variable "create_kms_key" {
  description = "Whether to create KMS  key or not"
  type        = bool
  default     = false
}
