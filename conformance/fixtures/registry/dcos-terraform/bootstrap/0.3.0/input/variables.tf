variable "cluster_name" {
  description = "Name of the DC/OS cluster"
}

variable "tags" {
  description = "Add custom tags to all resources"
  type        = map(string)
  default     = {}
}

variable "aws_ami" {
  description = "AMI that will be used for the instances instead of the Mesosphere chosen default images. Custom AMIs must fulfill the Mesosphere DC/OS system-requirements: See https://docs.mesosphere.com/1.12/installing/production/system-requirements/"
  default     = ""
}

variable "aws_instance_type" {
  description = "Instance type"
  default     = "t2.medium"
}

variable "aws_root_volume_size" {
  description = "Root volume size in GB"
  default     = "80"
}

variable "aws_root_volume_type" {
  description = "Root volume type"
  default     = "standard"
}

variable "aws_subnet_ids" {
  description = "Subnets to spawn the instances in. The module tries to distribute the instances"
  type        = list(string)
}

variable "aws_security_group_ids" {
  description = "Firewall IDs to use for these instances"
  type        = list(string)
}

variable "aws_iam_instance_profile" {
  description = "Instance profile to be used for these instances"
  default     = ""
}

variable "aws_associate_public_ip_address" {
  description = "Associate a public IP address with the instances"
  default     = true
}

variable "num_bootstrap" {
  description = "Specify the amount of bootstrap. You should have at most 1"
  default     = 1
}

variable "user_data" {
  description = "User data to be used on these instances (cloud-init)"
  default     = ""
}

// TODO: Maybe use a list instead and provision keys through cloudinit
variable "aws_key_name" {
  description = "Specify the aws ssh key to use. We assume its already loaded in your SSH agent. Set ssh_public_key_file to empty string"
}

variable "hostname_format" {
  description = "Format the hostname inputs are index+1, region, cluster_name"
  default     = "%[3]s-bootstrap%[1]d-%[2]s"
}

variable "dcos_instance_os" {
  description = "Operating system to use. Instead of using your own AMI you could use a provided OS."
  default     = "centos_7.4"
}

variable "name_prefix" {
  description = "Name Prefix"
  default     = ""
}

