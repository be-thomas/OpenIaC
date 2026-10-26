/**
 * AWS LB DC/OS
 * ============
 * This module creates three load balancers for DC/OS.
 *
 * External masters application load balancer
 * ------------------------------------------
 * This load balancer keeps an redundant entry point to the masters
 *
 * Internal masters network load balancer
 * --------------------------------------
 * this load balancer is used for internal communication to masters
 *
 * External public agents network load balancer
 * --------------------------------------------
 * This load balancer keeps a single entry point to your public agents no matter how many you're running.
 *
 * EXAMPLE
 * -------
 *
 *```hcl
 * module "dcos-lbs" {
 *   source  = "dcos-terraform/lb-dcos/aws"
 *   version = "~> 0.3.0"
 *
 *   cluster_name = "production"
 *
 *   subnet_ids = ["subnet-12345678"]
 *   security_groups_masters = ["sg-12345678"]
 *   security_groups_masters_internal = ["sg-12345678"]
 *   security_groups_public_agents = ["sg-12345678"]
 *   master_instances = ["i-00123456789e960f8"]
 *   public_agent_instances = ["i-00123456789e960f8"]
 *
 *   masters_acm_cert_arn = "arn:aws:acm:us-east-1:123456789123:certificate/ooc4NeiF-1234-5678-9abc-vei5Eeniipo4"
 * }
 *```
 */

provider "aws" {
  version = ">= 2.58"
}

module "dcos-lb-masters" {
  source  = "dcos-terraform/lb-masters/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name       = var.cluster_name
  subnet_ids         = var.subnet_ids
  security_groups    = concat(var.security_groups_masters, var.security_groups_masters_internal)
  instances          = var.master_instances
  https_acm_cert_arn = var.masters_acm_cert_arn
  internal           = var.internal
  disable            = var.disable_masters
  num_instances      = var.num_masters
  name_prefix        = var.name_prefix
  tags               = var.tags
}

module "dcos-lb-masters-internal" {
  source  = "dcos-terraform/lb-masters-internal/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name                = var.cluster_name
  subnet_ids                  = var.subnet_ids
  security_groups             = var.security_groups_masters_internal
  instances                   = var.master_instances
  disable                     = var.disable_masters
  https_acm_cert_arn          = var.masters_internal_acm_cert_arn
  num_instances               = var.num_masters
  name_prefix                 = var.name_prefix
  tags                        = var.tags
  adminrouter_grpc_proxy_port = var.adminrouter_grpc_proxy_port
}

module "dcos-lb-public-agents" {
  source  = "dcos-terraform/lb-public-agents/aws"
  version = "~> 0.3.0"

  providers = {
    aws = aws
  }

  cluster_name        = var.cluster_name
  subnet_ids          = var.subnet_ids
  security_groups     = var.security_groups_public_agents
  instances           = var.public_agent_instances
  https_acm_cert_arn  = var.public_agents_acm_cert_arn
  internal            = var.internal
  disable             = var.disable_public_agents
  num_instances       = var.num_public_agents
  name_prefix         = var.name_prefix
  additional_listener = var.public_agents_additional_listeners
  tags                = var.tags
}

