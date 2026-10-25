/**
 * AWS DC/OS Bootstrap Instance
 * ============
 * This module create a typical bootstrap instance for DC/OS
 *
 *
 * EXAMPLE
 * -------
 *
 *```hcl
 * module "dcos-bootstrap-instance" {
 *   source  = "dcos-terraform/bootstrap/aws"
 *   version = "~> 0.2.0"
 *
 *   cluster_name = "production"
 *   subnet_ids = ["subnet-12345678"]
 *   security_group_ids = ["sg-12345678"]
 *   aws_key_name = "my-ssh-key"
 * }
 *```
 */

provider "aws" {
  version = ">= 2.58"
}

// Instances is spawning the VMs to be used with DC/OS (bootstrap)
module "dcos-bootstrap-instance" {
  source  = "dcos-terraform/instance/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name                = var.cluster_name
  hostname_format             = var.hostname_format
  num                         = var.num_bootstrap
  ami                         = var.aws_ami
  user_data                   = var.user_data == "" && var.aws_ami == "" ? file("${path.module}/cloudinit.tpl") : var.user_data
  instance_type               = var.aws_instance_type
  subnet_ids                  = var.aws_subnet_ids
  security_group_ids          = var.aws_security_group_ids
  key_name                    = var.aws_key_name
  root_volume_size            = var.aws_root_volume_size
  root_volume_type            = var.aws_root_volume_type
  associate_public_ip_address = var.aws_associate_public_ip_address
  tags                        = var.tags
  dcos_instance_os            = var.dcos_instance_os
  iam_instance_profile        = var.aws_iam_instance_profile
  name_prefix                 = var.name_prefix
}
