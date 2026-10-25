variable "rest_api_id" {
  description = "The ID of the REST API"
  type        = string
  default     = ""
}

variable "response_type" {
  description = "The response type of the gateway response"
  type        = string
  default     = ""
}

variable "status_code" {
  description = "The status code of the gateway response"
  type        = string
  default     = ""
}

variable "gateway_responses" {
  description = "map of gateway responses"
  type        = map(any)
  default     = {}
}

variable "response_parameters" {
  description = "map of response parameters"
  type        = map(any)
  default     = {}
}
