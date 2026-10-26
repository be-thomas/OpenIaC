variable "name" {
  description = "(Required) This is the name of the hosted zone."
  type        = string
}

variable "comment" {
  description = "A comment for the hosted zone."
  type        = string
  default     = "Managed by Terraform"

}

variable "delegation_set_id" {
  description = "(Optional) The ID of the reusable delegation set whose NS records you want to assign to the hosted zone. Conflicts with vpc as delegation sets can only be used for public zones."
  type        = string
  default     = null
}

variable "force_destroy" {
  description = " (Optional) Whether to destroy all records (possibly managed outside of Terraform) in the zone when destroying the zone."
  type        = bool
  default     = true
}

variable "vpc" {
  description = "(Optional) Configuration block(s) specifying VPC(s) to associate with a private hosted zone. Conflicts with the delegation_set_id "
  type        = any
  default     = []
}

#Tags
variable "tags" {
  description = "A map of tags to assign to the zone"
  type        = map(string)
  default     = {}
}

variable "records" {
  description = "A list of records to create"
  type        = any
  default     = []
}

variable "retention_in_days" {
  description = "Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653, and 0. If you select 0, the events in the log group are always retained and never expire."
  type        = number
  default     = 30
}

variable "kms_key_id" {
  description = "The ARN of the KMS Key to use when encrypting log data. Please note, after the AWS KMS CMK is disassociated from the log group, AWS CloudWatch Logs stops encrypting newly ingested data for the log group. All previously ingested data remains encrypted, and AWS CloudWatch Logs requires permissions for the CMK whenever the encrypted data is requested."
  type        = string
  default     = null
}
