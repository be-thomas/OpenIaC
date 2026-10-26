variable "cluster_name" {
  description = "Name of the DC/OS cluster"
}

variable "tags" {
  description = "Add custom tags to all resources"
  type        = map(string)
  default     = {}
}

variable "subnet_ids" {
  description = "List of subnet IDs created in this network"
  type        = list(string)
}

variable "security_groups" {
  description = "Security Group IDs to use"
  type        = list(string)
  default     = []
}

variable "instances" {
  description = "List of instance IDs"
  type        = list(string)
}

variable "https_acm_cert_arn" {
  description = "ACM certifacte to be used."
  default     = ""
}

variable "elb_name_format" {
  description = "Printf style format for naming the ELB. Gets truncated to 32 characters. (input cluster_name)"
  default     = "load-balancer-%s"
}

variable "listener" {
  description = "List of listeners. By default HTTP and HTTPS are set. If set it overrides the default listeners."
  default     = []
}

variable "disable" {
  description = "Do not create load balancer and its resources"
  default     = false
}

variable "name_prefix" {
  description = "Name Prefix"
  default     = ""
}

variable "additional_listener" {
  description = "List of additional listeners"
  default     = []
}

variable "internal" {
  description = "This ELB is internal only"
  default     = false
}

variable "health_check" {
  description = "Health check definition."
  type        = map(string)

  default = {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "TCP:80"
    interval            = 30
  }
}

variable "cross_zone_load_balancing" {
  description = "Enable cross-zone load balancing"
  default     = true
}

variable "num_instances" {
  description = "How many instances should be created"
}

variable "load_balancer_type" {
  default     = "network"
  description = "Load Balancer type. Allowed values network, application"
}

