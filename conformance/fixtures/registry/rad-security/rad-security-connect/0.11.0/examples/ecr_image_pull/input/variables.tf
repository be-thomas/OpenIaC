variable "rad_access_key_id" {
  type        = string
  description = "RAD Security API access key ID"
}

variable "rad_secret_key" {
  type        = string
  description = "RAD Security API secret key"
  sensitive   = true
}

variable "aws_region" {
  type        = string
  description = "AWS region where the module is applied"
}
