/**
 * AWS DC/OS Master Instances
 * ============
 * This module creates typical master instances used by DC/OS
 *
 * EXAMPLE
 * -------
 *
 *```hcl
 * module "dcos-master-instances" {
 *   source  = "dcos-terraform/masters/aws"
 *   version = "~> 0.3.0"
 *
 *   cluster_name = "production"
 *   subnet_ids = ["subnet-12345678"]
 *   security_group_ids = ["sg-12345678"]"
 *   aws_key_name = "my-ssh-key"
 *
 *   num_masters = "3"
 * }
 *```
 */

provider "aws" {
  version = ">= 2.58"
}

// Instances is spawning the VMs to be used with DC/OS (bootstrap)
module "dcos-master-instances" {
  source  = "dcos-terraform/instance/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name                = var.cluster_name
  hostname_format             = var.hostname_format
  num                         = var.num_masters
  ami                         = var.aws_ami
  user_data                   = var.user_data
  instance_type               = var.aws_instance_type
  subnet_ids                  = var.aws_subnet_ids
  security_group_ids          = var.aws_security_group_ids
  key_name                    = var.aws_key_name
  root_volume_size            = var.aws_root_volume_size
  root_volume_type            = "gp2"
  extra_volumes               = var.aws_extra_volumes
  tags                        = var.tags
  associate_public_ip_address = var.aws_associate_public_ip_address
  dcos_instance_os            = var.dcos_instance_os
  iam_instance_profile        = var.aws_iam_instance_profile
  name_prefix                 = var.name_prefix
}
