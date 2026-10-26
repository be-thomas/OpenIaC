
variable "name" {
  description = "(Required) Name of the cluster (up to 255 letters, numbers, hyphens, and underscores)"
  type        = string
}

variable "configuration" {
  description = "(Optional) The execute command configuration for the cluster. Detailed below."
  type        = any
  default     = {}
}

variable "default_execute_command_logging" {
  description = "(Optional) Default logging mode for ECS execute command configuration. Valid values are NONE, DEFAULT, OVERRIDE."
  type        = string
  default     = "DEFAULT"
}

variable "container_insights" {
  description = "The value to assign to the setting. Value values are enabled and disabled."
  default     = "enabled"
  type        = string
}

variable "extra_script" {
  type        = string
  description = "The name of the extra script"
  default     = ""
}

variable "install_ssm_agent" {
  type        = bool
  description = "Whether to install ssm agent"
  default     = false
}

variable "ssm_agent_temp_directory" {
  description = "(Optional) Temporary directory for SSM agent installation"
  type        = string
  default     = "/tmp/ssm"
}

variable "ssm_agent_version_centos6" {
  description = "(Optional) SSM agent version for CentOS 6"
  type        = string
  default     = "3.0.1479.0"
}

variable "ssm_agent_s3_urls" {
  description = "(Optional) S3 URLs for SSM agent downloads by architecture and OS"
  type        = map(string)
  default = {
    "amazon_linux_x86_64"  = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm"
    "amazon_linux_arm64"   = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_arm64/amazon-ssm-agent.rpm"
    "centos_rhel_x86_64"   = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm"
    "centos_rhel_arm64"    = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_arm64/amazon-ssm-agent.rpm"
    "centos6_x86_64"       = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/3.0.1479.0/linux_amd64/amazon-ssm-agent.rpm"
    "debian_ubuntu_x86_64" = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/debian_amd64/amazon-ssm-agent.deb"
    "debian_ubuntu_arm64"  = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/debian_arm64/amazon-ssm-agent.deb"
    "ubuntu_snap_x86_64"   = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/ubuntu_amd64/amazon-ssm-agent.snap"
    "ubuntu_snap_arm64"    = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/ubuntu_arm64/amazon-ssm-agent.snap"
    "suse_x86_64"          = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm"
    "suse_arm64"           = "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_arm64/amazon-ssm-agent.rpm"
  }
}

variable "metadata_options" {
  type        = map(string)
  description = "Customize the metadata options of the instance"
  default     = {}
}

variable "tags" {
  description = "Key-value map of resource tags."
  default     = {}
  type        = map(string)
}

# Capacity Providers

variable "add_capacity_providers" {
  description = "Whether or not to deploy aws_cluster_capacity_providers resource"
  type        = bool
  default     = false
}

variable "capacity_providers" {
  description = "(Optional) Set of names of one or more capacity providers to associate with the cluster. Valid values also include FARGATE and FARGATE_SPOT."
  type        = list(string)
  default     = ["FARGATE"]
}

variable "default_capacity_provider_strategy" {
  description = "(Optional) Set of capacity provider strategies to use by default for the cluster."
  type        = map(any)
  default     = {}
}

# Security Group

variable "vpc_id" {
  description = "(Optional, Forces new resource) VPC ID"
  type        = string
  default     = null
}

variable "ingress_rules" {
  description = "(Optional) Ingress rules to add to the security group"
  type        = any
  default     = {}
}
variable "egress_rules" {
  description = "(Optional) Egress rules to add to the security group"
  type        = any
  default     = {}
}

variable "create_security_group" {
  description = "Whether to create a Security Group for ECS cluster."
  default     = true
  type        = bool
}

variable "security_group_description" {
  description = "(Optional) Description for the ECS cluster security group"
  type        = string
  default     = "ECS cluster Security Group"
}

variable "security_group_ingress_description" {
  description = "(Optional) Description for security group ingress rules"
  type        = string
  default     = "Allow custom inbound traffic from specific ports."
}

variable "security_group_egress_description" {
  description = "(Optional) Description for security group egress rules"
  type        = string
  default     = "Allow custom egress traffic"
}

variable "security_group_default_protocol" {
  description = "(Optional) Default protocol for security group rules. Use -1 for all protocols"
  type        = string
  default     = "-1"
}

variable "subnet_id" {
  description = "The subnet ID to launch the instances in"
  type        = string
  default     = null
}

# Launch template

variable "create_ec2_instance" {
  description = "Whether or not to create a cluster ec2 instance(s)"
  type        = bool
  default     = false
}

variable "image_id" {
  description = "(Optional) The AMI from which to launch the instance."
  type        = string
  default     = null
}

variable "instance_type" {
  description = "(Optional) The type of the instance."
  type        = string
  default     = null
}

variable "key_name" {
  description = "(Optional) The key name to use for the instance."
  type        = string
  default     = null
}
variable "user_data" {
  description = "(Optional) The base64-encoded user data to provide when launching the instance."
  type        = string
  default     = null
}

variable "monitoring_enabled" {
  description = "(Optional) The monitoring option for the instance."
  type        = bool
  default     = false
}

variable "associate_public_ip_address" {
  description = "Associate a public ip address with the network interface. Boolean value."
  type        = bool
  default     = true
}

variable "delete_on_termination" {
  description = "Whether the network interface should be destroyed on instance termination. Defaults to false if not set."
  type        = bool
  default     = true
}

# Autoscaling Group
variable "availability_zones" {
  description = "(Optional) A list of one or more availability zones for the group. Used for EC2-Classic, attaching a network interface via id from a launch template"
  type        = list(string)
  default     = []
}

variable "desired_capacity" {
  description = "(Optional) The number of Amazon EC2 instances that should be running in the group."
  type        = number
  default     = 1
}

variable "min_size" {
  description = "(Required) The minimum size of the Auto Scaling Group."
  type        = number
  default     = 1
}

variable "max_size" {
  description = " (Required) The maximum size of the Auto Scaling Group"
  type        = number
  default     = 10
}

variable "block_device_mappings" {
  description = "(Optional) Specify volumes to attach to the instance besides the volumes specified by the AMI. "
  type        = any
  default     = []
}

variable "asg_tags_propagate_at_launch" {
  description = "(Optional) Whether to propagate ASG tags to instances at launch"
  type        = bool
  default     = true
}

### KMS Key
variable "kms_key_id" {
  description = "(Optional) The AWS Key Management Service key ID to encrypt the data between the local client and the container."
  type        = string
  default     = null
}

variable "enable_key_rotation" {
  description = "(Optional) Specifies whether key rotation is enabled. Defaults to false."
  type        = bool
  default     = false
}

variable "key_description" {
  description = "The description of the key as viewed in AWS console."
  type        = string
  default     = null
}

variable "deletion_window_in_days" {
  description = " (Optional) The waiting period, specified in number of days. After the waiting period ends, AWS KMS deletes the KMS key. If you specify a value, it must be between 7 and 30, inclusive. If you do not specify a value, it defaults to 30."
  type        = number
  default     = 30
}

variable "create_kms_key" {
  description = "Whether or not to create a kms key with this module"
  type        = bool
  default     = false
}

variable "kms_policy_iam_statement_id" {
  description = "(Optional) Statement ID for IAM user permissions in KMS key policy"
  type        = string
  default     = "Enable IAM User Permissions"
}

variable "kms_policy_cloudwatch_statement_id" {
  description = "(Optional) Statement ID for CloudWatch logs permissions in KMS key policy"
  type        = string
  default     = "AllowCloudWatchLogs"
}

variable "kms_cloudwatch_actions" {
  description = "(Optional) List of KMS actions allowed for CloudWatch logs"
  type        = list(string)
  default     = ["kms:Encrypt*", "kms:Decrypt*", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:Describe*"]
}

variable "kms_policy_principal" {
  description = "(Optional) AWS principal for KMS policy. Use '*' for all principals or specify specific ARNs"
  type        = string
  default     = "*"
}


variable "launch_template_version" {
  type        = string
  description = "The version of the launch template"
  default     = "$Latest"
}

variable "enable_managed_scaling" {
  type        = bool
  description = "Whether to enable ECS managed scaling for the Auto Scaling Group"
  default     = false
}

variable "ecs_managed_tag_key" {
  description = "(Optional) Tag key for ECS managed scaling. AWS requires this to be 'AmazonECSManaged'"
  type        = string
  default     = "AmazonECSManaged"
}

variable "ecs_managed_tag_value" {
  description = "(Optional) Tag value for ECS managed scaling. AWS requires this to be true"
  type        = bool
  default     = true
}

variable "ecs_managed_tag_propagate_at_launch" {
  description = "(Optional) Whether to propagate the ECS managed tag to instances at launch"
  type        = bool
  default     = true
}


variable "metadata_options_defaults" {
  description = "(Optional) Default values for EC2 instance metadata options"
  type        = map(string)
  default = {
    http_endpoint = "enabled"
    http_tokens   = "optional"
  }
}

variable "protect_from_scale_in_independent" {
  description = "(Optional) Whether to enable scale-in protection independently of managed scaling"
  type        = bool
  default     = null
}

variable "default_ebs_volume_type" {
  description = "(Optional) Default EBS volume type for block device mappings"
  type        = string
  default     = "gp3"
}

variable "default_ebs_encrypted" {
  description = "(Optional) Whether to encrypt EBS volumes by default"
  type        = bool
  default     = true
}

variable "default_ebs_kms_key_id" {
  description = "(Optional) Default KMS key ID for EBS encryption. Uses aws/ebs if not specified"
  type        = string
  default     = null
}

variable "enable_default_ebs_configuration" {
  description = "(Optional) Whether to apply default EBS configuration (gp3, encryption) to block device mappings that don't specify these settings"
  type        = bool
  default     = true
}

variable "create_default_root_volume" {
  description = "(Optional) Whether to create a default root volume with secure defaults if no block_device_mappings are specified"
  type        = bool
  default     = false
}

variable "default_root_volume_size" {
  description = "(Optional) Default root volume size in GB when creating default root volume"
  type        = number
  default     = 20
}

variable "default_root_device_name" {
  description = "(Optional) Default root device name when creating default root volume"
  type        = string
  default     = "/dev/xvda"
}

variable "ec2_volume_sizes" {
  description = "(Optional) Default volume sizes in GB for different device types"
  type        = map(number)
  default = {
    root_volume = 20
    data_volume = 50
    log_volume  = 10
  }
}
