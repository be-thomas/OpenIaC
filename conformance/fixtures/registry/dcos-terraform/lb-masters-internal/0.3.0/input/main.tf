/**
 * AWS LB Internal
 * ============
 * This module create a network load balancer for cluster internal access to masters
 *
 *
 * EXAMPLE
 * -------
 *
 *```hcl
 * module "dcos-lb-masters-internal" {
 *   source  = "terraform-dcos/lb-masters-internal/aws"
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

module "masters-internal" {
  source  = "dcos-terraform/lb/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name = var.cluster_name

  internal = true

  listener = [
    {
      port     = 5050
      protocol = "tcp"
    },
    {
      port     = 2181
      protocol = "tcp"
    },
    {
      port     = 8181
      protocol = "tcp"
    },
    {
      port     = 80
      protocol = "tcp"
    },
    {
      port     = 443
      protocol = "tcp"
    },
    {
      port     = 8080
      protocol = "tcp"
    },
    {
      port     = var.adminrouter_grpc_proxy_port
      protocol = "tcp"
    },
  ]

  https_acm_cert_arn = var.https_acm_cert_arn
  elb_name_format    = "int-%s"
  instances          = var.instances
  num_instances      = var.num_instances
  security_groups    = var.security_groups
  subnet_ids         = var.subnet_ids
  disable            = var.disable
  name_prefix        = var.name_prefix
  tags               = var.tags
}

