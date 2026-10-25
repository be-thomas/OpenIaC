/**
 * AWS LB Masters
 * ============
 * This module create an application load balancer to acces the masters externally.
 *
 *
 * EXAMPLE
 * -------
 *
 *```hcl
 * module "dcos-lb-masters" {
 *   source  = "terraform-dcos/lb-masters/aws"
 *   version = "~> 0.3.0"
 *
 *   cluster_name = "production"
 *
 *   subnet_ids = ["subnet-12345678"]
 *   security_groups = ["sg-12345678"]
 *
 *   instances = ["i-00123456789e960f8"]
 *   num_instances = 1
 *
 *   https_acm_cert_arn = "arn:aws:acm:us-east-1:123456789123:certificate/ooc4NeiF-1234-5678-9abc-vei5Eeniipo4"
 * }
 *```
 */

provider "aws" {
  version = ">= 2.58"
}

module "masters" {
  source = "dcos-terraform/lb/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name = var.cluster_name

  https_acm_cert_arn = var.https_acm_cert_arn
  elb_name_format    = "%s"

  instances          = var.instances
  num_instances      = var.num_instances
  security_groups    = var.security_groups
  subnet_ids         = var.subnet_ids
  internal           = var.internal
  disable            = var.disable
  name_prefix        = var.name_prefix
  load_balancer_type = "application"

  tags = var.tags
}

