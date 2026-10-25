variable "admin_ips" {
  description = "List of CIDR admin IPs"
  type        = list(string)
}

variable "public_agents_access_ips" {
  description = "List of ips allowed access to public agents. admin_ips are joined to this list"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "cluster_name" {
  description = "Name of the DC/OS cluster"
  default     = "aws-example"
}

variable "tags" {
  description = "Add custom tags to all resources"
  type        = map(string)
  default     = {}
}

variable "subnet_range" {
  description = "Private IP space to be used in CIDR format"
}

variable "vpc_id" {
  description = "AWS VPC ID"
}

variable "public_agents_additional_ports" {
  description = "List of additional ports allowed for public access on public agents (80 and 443 open by default)"
  type        = list(string)
  default     = []
}

variable "public_agents_allow_registered" {
  description = "Allow registered / user ports (1024-49151 see: RFC6335) on public agents public IPs"
  default     = false
}

variable "public_agents_allow_dynamic" {
  description = "Allow dynamic / ephemeral ports (49152-65535 see: RFC6335) on public agents public IPs"
  default     = false
}

variable "accepted_internal_networks" {
  description = "Subnet ranges for all internal networks"
  type        = list(string)
  default     = []
}

variable "adminrouter_grpc_proxy_port" {
  description = ""
  default     = 12379
}

variable "open_admin_router" {
  description = "Open admin router to public (80+443 on load balancer). WARNING: attackers could take over your cluster"
  default     = false
}

variable "open_instance_ssh" {
  description = "Open SSH on instances to public. WARNING: make sure you use a strong SSH key"
  default     = false
}
