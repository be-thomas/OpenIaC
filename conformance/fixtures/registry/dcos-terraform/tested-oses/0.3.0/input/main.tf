/**
 * AWS Tested OSes
 * ============
 * This module returns for an given OS the AMI aswell as its default user
 *
 *
 * EXAMPLE
 * -------
 *```hcl
 * module "dcos-tested-oses" {
 *   source  = "terraform-dcos/tested-oses/aws"
 *   version = "~> 0.3.0"
 * }
 *```
 */

provider "aws" {
  version = ">= 2.0"
}

data "aws_region" "current" {
}

data "template_file" "aws_ami" {
  template = "$${aws_ami_result}"

  vars = {
    aws_ami_result = var.aws_ami[format(
      "%s_%s",
      var.os,
      coalesce(var.region, data.aws_region.current.name),
    )]
  }
}

data "template_file" "aws_ami_user" {
  template = "$${aws_user_result}"

  vars = {
    aws_user_result = var.aws_default_os_user[element(split("_", var.os), 0)]
  }
}

