variable "name" {
  type        = string
  description = "(Optional) The name of the rule. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`."
  default     = null
}

variable "name_prefix" {
  type        = string
  description = "(Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`."
  default     = null
}

variable "schedule_expression" {
  type        = string
  description = "(Optional) The scheduling expression. For example, `cron(0 20 * * ? *)` or `rate(5 minutes)`. At least one of `schedule_expression` or `event_pattern` is required. Can only be used on the default event bus."
  default     = null
}

variable "event_bus_name" {
  type        = string
  description = "(Optional) The event bus to associate with this rule. If you omit this, the `default` event bus is used."
  default     = "default"
}

variable "event_pattern" {
  type        = string
  description = "(Optional) The event pattern described a JSON object. At least one of `schedule_expression` or `event_pattern` is required."
  default     = null
}

variable "description" {
  type        = string
  description = "(Optional) The description of the rule."
  default     = null
}

variable "role_arn" {
  type        = string
  description = "(Optional) The Amazon Resource Name (ARN) associated with the role that is used for target invocation."
  default     = null
}

variable "is_enabled" {
  type        = bool
  description = "(Optional) Whether the rule should be enabled (defaults to `true`)."
  default     = true
}

variable "tags" {
  type        = map(string)
  description = "A map of tags to assign to the resource."
  default     = {}
}

### Cloudwatch event Target
variable "target_id" {
  type        = string
  description = "(Optional) The unique target assignment ID. If missing, will generate a random, unique id."
  default     = null
}

variable "arn" {
  type        = string
  description = "(Required) The Amazon Resource Name (ARN) of the target."
}

variable "input" {
  type        = string
  description = "(Optional) Valid JSON text passed to the target. Conflicts with `input_path` and `input_transformer`."
  default     = null
}

variable "input_path" {
  type        = string
  description = "(Optional) The value of the JSONPath that is used for extracting part of the matched event when passing it to the target. Conflicts with `input` and `input_transformer`."
  default     = null
}

variable "target_role_arn" {
  type        = string
  description = "(Optional) The Amazon Resource Name (ARN) of the IAM role to be used for this target when the rule is triggered. Required if `ecs_target` is used or target in arn is EC2 instance, Kinesis data stream, Step Functions state machine, or Event Bus in different account or region."
  default     = null
}

variable "run_command_targets" {
  type        = list(any)
  description = "(Optional) List of Parameters used when you are using the rule to invoke Amazon EC2 Run Command."
  default     = []
}

variable "ecs_target" {
  type        = any
  description = "(Optional) Parameters used when you are using the rule to invoke Amazon ECS Task."
  default     = {}
}

variable "batch_target" {
  type        = any
  description = "(Optional) Parameters used when you are using the rule to invoke an Amazon Batch Job."
  default     = {}
}

variable "kinesis_target" {
  type        = any
  description = "(Optional) Parameters used when you are using the rule to invoke an Amazon Kinesis Stream."
  default     = {}
}

variable "redshift_target" {
  type        = any
  description = "(Optional) Parameters used when you are using the rule to invoke an Amazon Redshift Statement."
  default     = {}
}

variable "sqs_target" {
  type        = any
  description = "(Optional) Parameters used when you are using the rule to invoke an Amazon SQS Queue."
  default     = {}
}

variable "http_target" {
  type        = any
  description = "(Optional) Parameters used when you are using the rule to invoke an API Gateway REST endpoint."
  default     = {}
}

variable "input_transformer" {
  type        = any
  description = "(Optional) Parameters used when you are providing a custom input to a target based on certain event data. Conflicts with `input` and `input_path`."
  default     = []
}

variable "retry_policy" {
  type        = any
  description = "(Optional) Parameters used when you are providing retry policies."
  default     = {}
}

variable "dead_letter_config" {
  type        = any
  description = "(Optional) Parameters used when you are providing a dead letter config."
  default     = {}
}


variable "event_permissions" {
  type        = any
  description = "EventBridge permissions to support cross-account events in the current account default event bus."
  default     = []
}
