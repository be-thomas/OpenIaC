################################################################################
# API Gateway
################################################################################

variable "name" {
  description = "Name of the API gateway"
  type        = string
}

variable "stage" {
  description = "Name of the API gateway stage"
  type        = string
}

variable "body_template" {
  description = "OpenAPI specification template"
  type        = string
  default     = ""
}

variable "stage_variables" {
  description = "Stage variables for the API gateway stage"
  type        = map(any)
  default     = {}
}

variable "types" {
  description = "Name of the API gateway stage"
  type        = list(string)
  default     = ["REGIONAL"]
}

variable "throttling_burst_limit" {
  description = "The API request burst limit, the maximum rate limit over a time ranging from one to a few seconds, depending upon whether the underlying token bucket is at its full capacity."
  type        = number
  default     = 1000
}

variable "throttling_rate_limit" {
  description = "The API request steady-state rate limit"
  type        = number
  default     = 1000
}

variable "metrics_enabled" {
  description = "Specifies whether Amazon CloudWatch metrics are enabled for this method."
  type        = bool
  default     = false
}

variable "logging_level" {
  description = "Specifies the logging level for this method."
  type        = string
  default     = "ERROR"
}

variable "data_trace_enabled" {
  description = "Specifies the logging level for this method."
  type        = bool
  default     = false
}

variable "cache_cluster_enabled" {
  description = "Specifies whether a cache cluster is enabled for the stage."
  type        = bool
  default     = false
}

variable "cache_data_encrypted" {
  description = "Specifies whether a cache data needs to be encrypted."
  type        = bool
  default     = true
}

variable "cache_cluster_size" {
  description = "The size of the cache cluster for the stage, if enabled"
  type        = string
  default     = "0.5"
}

variable "cache_ttl" {
  description = "The cache TTL of the stage"
  type        = number
  default     = 300
}

variable "log_group_name" {
  description = "Custom log group name."
  type        = string
  default     = ""
}

variable "log_format" {
  description = "Cloudwatch log format"
  type        = map(string)
  default = {
    requestId         = "$context.requestId"
    extendedRequestId = "$context.extendedRequestId"
    ip                = "$context.identity.sourceIp"
    caller            = "$context.identity.caller"
    user              = "$context.identity.user"
    requestTime       = "$context.requestTime"
    httpMethod        = "$context.httpMethod"
    resourcePath      = "$context.resourcePath"
    status            = "$context.status"
    protocol          = "$context.protocol"
    responseLength    = "$context.responseLength"
  }
}

variable "log_retention_in_days" {
  description = "Cloudwatch log retention in days"
  type        = number
  default     = 7
}

variable "log_kms_key_id" {
  description = "The ARN of the KMS Key to use when encrypting log data"
  type        = string
  default     = ""
}


variable "caching_enabled" {
  description = "Specifies whether caching is enabled for the API gateway method"
  type        = bool
  default     = false
}

variable "vpc_links" {
  description = "List of VPC links for REST APIs. key in the map denotes the vpc link name"
  type = map(
    object({
      description = string
      name        = optional(string)
      target_arns = list(string)
    })
  )
  default = {}
}

variable "vpc_links_v2" {
  description = "VPC links for API Gateway v2 HTTP APIs. The map key denotes the VPC link name."
  type = map(
    object({
      name               = optional(string)
      security_group_ids = list(string)
      subnet_ids         = list(string)
    })
  )
  default = {}
}

variable "enable_global_apigw_logging" {
  description = "Enable global apigw logging"
  type        = bool
  default     = false
}

variable "enable_resource_policy" {
  description = "Enable resource policy to allow other resources to Invoke this API"
  type        = bool
  default     = false
}

variable "resource_policy_json" {
  description = "Resource policy as JSON to attach"
  type        = string
  default     = null
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

variable "put_rest_api_mode" {
  description = "Mode when updating the REST API (overwrite or merge). Overwrite ensures deleted routes are removed from AWS Console."
  type        = string
  default     = "overwrite"
}
