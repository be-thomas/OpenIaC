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

variable "security_groups_masters" {
  description = "Security Group IDs to use for external masters load balancer"
  type        = list(string)
}

variable "disable_masters" {
  description = "[MASTERS] Do not create load balancer and its resources"
  default     = false
}

variable "disable_public_agents" {
  description = "[PUBLIC AGENTS] Do not create load balancer and its resources"
  default     = false
}

variable "name_prefix" {
  description = "Name Prefix"
  default     = ""
}

variable "security_groups_masters_internal" {
  description = "Security Group IDs to use for internal communication to masters"
  type        = list(string)
}

variable "security_groups_public_agents" {
  description = "Security Group IDs to use for external public agents load balancer"
  type        = list(string)
}

variable "master_instances" {
  description = "List of master instance IDs"
  type        = list(string)
}

variable "public_agent_instances" {
  description = "List of public agent instance IDs"
  type        = list(string)
}

variable "masters_acm_cert_arn" {
  description = "ACM certifacte to be used for the masters load balancer"
  default     = ""
}

variable "masters_internal_acm_cert_arn" {
  description = "ACM certifacte to be used for the internal masters load balancer"
  default     = ""
}

variable "public_agents_acm_cert_arn" {
  description = "ACM certifacte to be used for the public agents load balancer"
  default     = ""
}

variable "public_agents_additional_listeners" {
  description = "Additional list of listeners for public agents. These listeners are an additon to the default listeners."
  type        = list(string)
  default     = []
}

variable "internal" {
  description = "This ELB is internal only"
  default     = "false"
}

variable "num_masters" {
  description = "Specify the amount of masters. For redundancy you should have at least 3"
}

variable "num_public_agents" {
  description = "Specify the amount of public agents. These agents will host marathon-lb and edgelb"
}

variable "adminrouter_grpc_proxy_port" {
  description = ""
  default     = 12379
}

