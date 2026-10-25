variable "name" {
  type        = string
  description = "Friendly name of the WebACL."
}

variable "description" {
  type        = string
  description = "Friendly description of the WebACL."
  default     = null
}

variable "scope" {
  type        = string
  description = "Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1 (N. Virginia)` on the AWS provider."
  default     = "REGIONAL"
}

variable "default_action" {
  type        = string
  description = "Set to `allow` for WAF to allow requests by default. Set to `block` for WAF to block requests by default."
  default     = "allow"
}

variable "custom_response_bodies" {
  type        = any
  description = "Defines custom response bodies that can be referenced by `custom_response` actions"
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Map of key-value pairs to associate with the resource."
  default     = {}
}

variable "rules" {
  type        = any
  description = "Rule blocks used to identify the web requests that you want to `allow`, `block`, or `count`"
  default     = []
}

variable "cloudwatch_metrics_enabled" {
  type        = bool
  description = "Whether to enable cloudwatch metrics"
  default     = false
}

variable "metric_name" {
  type        = string
  description = "The name of the metric"
  default     = "sample-name-1"
}

variable "sampled_requests_enabled" {
  type        = bool
  description = "Whether to enable sample requests"
  default     = false
}

variable "web_acl_resource_arn" {
  type        = string
  description = "The Amazon Resource Name (ARN) of the resource to associate with the web ACL. This must be an ARN of an Application Load Balancer, an Amazon API Gateway stage, or an Amazon Cognito User Pool."
  default     = null
}

variable "create_acl_association" {
  type        = bool
  description = "Whether to create acl association"
  default     = false
}

variable "ip_set_v4" {
  type        = any
  description = "IPV4 IP set"
  default     = []
}

variable "ip_set_v6" {
  type        = any
  description = "IPV6 IP set"
  default     = []
}

variable "ip_set_reference_statement" {
  type        = any
  description = "A rule statement used to detect web requests coming from particular IP addresses or address ranges."
  default     = {}
}

variable "enable_logging" {
  type        = bool
  description = "Whether to enable logging for the WAF WebACL"
  default     = false
}

variable "log_destination_configs" {
  type        = list(string)
  description = "The Amazon Kinesis Data Firehose, CloudWatch Log log group, or S3 bucket Amazon Resource Names (ARNs) that you want to associate with the web ACL."
  default     = []
}

variable "redacted_fields" {
  type        = list(any)
  description = "List of fields to redact from the logs. Currently only supports single_header type."
  default     = null
}

variable "logging_filter" {
  type = object({
    default_behavior = string, # Required: "KEEP" or "DROP"
    filters = list(object({
      behavior    = string,           # Required: "KEEP" or "DROP"
      requirement = string,           # Required: "MEETS_ALL" or "MEETS_ANY"
      conditions  = list(map(string)) # Map to hold action_condition OR label_name_condition
    }))
  })
  description = "Configuration for WAF logging filters. Determines which requests are logged."
  default     = null

  validation {
    condition     = var.logging_filter == null ? true : contains(["KEEP", "DROP"], var.logging_filter.default_behavior)
    error_message = "Default_behavior must be either \"KEEP\" or \"DROP\"."
  }
}
