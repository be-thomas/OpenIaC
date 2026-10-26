variable "name" {
  type        = string
  description = "The name of the stack"
}

variable "document_name" {
  type        = string
  description = "The name of the custom document"
  default     = null
}

variable "document_type" {
  type        = string
  description = "The type of the document. Valid document types include: `Automation`, `Command`, `Package`, `Policy`, and `Session`"
  default     = null
}

variable "document_format" {
  type        = string
  description = "The format of the document. Valid document types include: `JSON` and `YAML`"
  default     = "JSON"
}

variable "content" {
  type        = string
  description = "The JSON or YAML content of the document."
  default     = ""
}

variable "create_custom_document" {
  type        = bool
  description = "Specify whether to create ssm document with custom settings"
  default     = false
}

variable "s3_key_prefix" {
  type        = string
  description = "Specify S3 key prefix for the logs"
  default     = ""
}

variable "association_name" {
  type        = string
  description = "The descriptive name for the association."
  default     = null
}

variable "document_version" {
  type        = string
  description = "The document version you want to associate with the target(s). Can be a specific version or the default version."
  default     = null
}

variable "targets" {
  type        = list(any)
  description = "The configuration for the targets to associate the Document with. As of this version AWS currently supports a maximum of 5 targets."
  default     = []
}

variable "run_as_enabled" {
  type        = bool
  description = "If set to true, you must specify a user account that exists on the managed nodes you will be connecting to in the runAsDefaultUser input. Otherwise, sessions will fail to start."
  default     = false
}

variable "run_as_default_user" {
  type        = string
  description = "The name of the user account to start sessions with on Linux managed nodes when the runAsEnabled input is set to true. The user account you specify for this input must exist on the managed nodes you will be connecting to; otherwise, sessions will fail to start."
  default     = ""
}

variable "idle_session_timeout" {
  type        = number
  description = "The amount of time of inactivity you want to allow before a session ends. This input is measured in minutes. Valid values: 1-60"
  default     = 20
}

variable "max_session_duration" {
  type        = number
  description = "The maximum amount of time you want to allow before a session ends. This input is measured in minutes. Valid values: 1-1440"
  default     = 720
}

variable "s3_bucket_name" {
  type        = string
  description = "The name of the bucket to send ssm session logs to"
  default     = ""
}

variable "s3_encryption_enabled" {
  type        = bool
  description = "If set to true, the Amazon S3 bucket you specified in the s3BucketName input must be encrypted."
  default     = true
}

variable "cloudwatch_encryption_enabled" {
  type        = bool
  description = "If set to true, the log group you specified in the cloudWatchLogGroupName input must be encrypted."
  default     = true
}

variable "cloudwatch_streaming_enabled" {
  type        = bool
  description = "If set to true, a continual stream of session data logs are sent to the log group you specified in the cloudWatchLogGroupName input. If set to false, session logs are sent to the log group you specified in the cloudWatchLogGroupName input at the end of your sessions."
  default     = true
}

variable "retention_in_days" {
  type        = number
  description = "Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire."
  default     = 0
}

variable "linux_shell_profile" {
  type        = string
  description = "The shell preferences, environment variables, working directories, and commands you specify for sessions on Linux managed nodes."
  default     = ""
}

variable "windows_shell_profile" {
  type        = string
  description = "The shell preferences, environment variables, working directories, and commands you specify for sessions on Windows managed nodes."
  default     = ""
}

variable "kms_key_id" {
  type        = string
  description = "Provide a KMS Key ID for AWS CMK key not created by this module"
  default     = ""
}

variable "tags" {
  type        = map(string)
  description = "Map of tags to assign to the resources."
  default     = {}
}

variable "create_session_preferences" {
  type        = string
  description = "Whether to create session preferences"
  default     = false
}
