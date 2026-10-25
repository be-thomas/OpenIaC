/**
 * AWS LB Public Agents
 * ============
 * This module create a network load balancer to acces public agents externally
 *
 *
 * EXAMPLE
 * -------
 *
 *```hcl
 * module "dcos-lb-public-agents" {
 *   source  = "terraform-dcos/lb-public-agents/aws"
 *   version = "~> 0.3.0"
 *
 *   cluster_name = "production"
 *
 *   subnet_ids = ["subnet-12345678"]
 *   security_groups = ["sg-12345678"]
 *
 *   instances = ["i-00123456789e960f8"]
 *
 *   https_acm_cert_arn = "arn:aws:acm:us-east-1:123456789123:certificate/ooc4NeiF-1234-5678-9abc-vei5Eeniipo4"
 * }
 *```
 */

provider "aws" {
  version = ">= 2.58"
}

module "public-agents" {
  source  = "dcos-terraform/lb/aws"
  version = "~> 0.3.0"
  
  providers = {
    aws = aws
    # health_check {
    #   healthy_threshold   = 2
    #   unhealthy_threshold = 2
    #   timeout             = 2
    #   target              = "HTTP:9090/_haproxy_health_check"
    #   interval            = 5
    # }
  }

  cluster_name = var.cluster_name

  https_acm_cert_arn  = var.https_acm_cert_arn
  elb_name_format     = "ext-%s"
  additional_listener = var.additional_listener
  instances           = var.instances
  num_instances       = var.num_instances
  security_groups     = var.security_groups
  subnet_ids          = var.subnet_ids
  internal            = var.internal
  disable             = var.disable
  name_prefix         = var.name_prefix
  tags                = var.tags
}

