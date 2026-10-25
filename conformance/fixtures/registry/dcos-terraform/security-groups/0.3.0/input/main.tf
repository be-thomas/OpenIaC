/**
 * AWS Security Group
 * ============
 * This module create DC/OS security groups
 *
 * The firewall module creates four different policies to be used by provisioning DC/OS Infrastructure
 *
 * EXAMPLE
 * -------
 *```hcl
 * module "dcos-security-groups" {
 *   source  = "dcos-terraform/security-groups/aws"
 *   version = "~> 0.3.0"
 *
 *   vpc_id = "vpc-12345678"
 *   cluster_name = "production"
 *   subnet_range = "172.16.0.0/16"
 *   admin_ips = ["1.2.3.4/32"]
 * }
 *```
 */

locals {
  public_agents_ports            = [80, 443]
  public_agents_additional_ports = concat(local.public_agents_ports, var.public_agents_additional_ports)
  admin_ports                    = concat([22, 8181, 9090], local.public_agents_ports, [var.adminrouter_grpc_proxy_port])
}

provider "aws" {
  version = ">= 2.0"
}

resource "aws_security_group" "internal" {
  name        = "dcos-${var.cluster_name}-internal-firewall"
  description = "Allow all internal traffic"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )
}

resource "aws_security_group_rule" "internal_ingress_rule" {
  type        = "ingress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = distinct(concat([var.subnet_range], var.accepted_internal_networks))

  security_group_id = aws_security_group.internal.id
}

resource "aws_security_group_rule" "internal_egress_rule" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.internal.id
}

resource "aws_security_group" "master_lb" {
  name        = "dcos-${var.cluster_name}-master-lb-firewall"
  description = "Allow incoming traffic on masters load balancer"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )

  dynamic "ingress" {
    for_each = local.public_agents_ports
    content {
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = var.open_admin_router ? ["0.0.0.0/0"] : var.admin_ips
    }
  }
}

resource "aws_security_group" "public_agents" {
  name        = "dcos-${var.cluster_name}-public-agents-lb-firewall"
  description = "Allow incoming traffic on Public Agents load balancer"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )
}

resource "aws_security_group_rule" "additional_rules" {
  count       = length(local.public_agents_additional_ports)
  type        = "ingress"
  protocol    = "tcp"
  from_port   = local.public_agents_additional_ports[count.index]
  to_port     = local.public_agents_additional_ports[count.index]
  cidr_blocks = distinct(var.public_agents_access_ips)

  security_group_id = aws_security_group.public_agents.id
}

resource "aws_security_group_rule" "allow_registered" {
  count       = var.public_agents_allow_registered ? 1 : 0
  type        = "ingress"
  protocol    = "TCP"
  from_port   = "1024"
  to_port     = "49151"
  cidr_blocks = distinct(var.public_agents_access_ips)

  security_group_id = aws_security_group.public_agents.id
}

resource "aws_security_group_rule" "allow_registered_udp" {
  count       = var.public_agents_allow_registered ? 1 : 0
  type        = "ingress"
  protocol    = "UDP"
  from_port   = "1024"
  to_port     = "49151"
  cidr_blocks = distinct(var.public_agents_access_ips)

  security_group_id = aws_security_group.public_agents.id
}

resource "aws_security_group_rule" "allow_dynamic" {
  count       = var.public_agents_allow_dynamic ? 1 : 0
  type        = "ingress"
  protocol    = "TCP"
  from_port   = "49152"
  to_port     = "65535"
  cidr_blocks = distinct(var.public_agents_access_ips)

  security_group_id = aws_security_group.public_agents.id
}

resource "aws_security_group_rule" "allow_dynamic_udp" {
  count       = var.public_agents_allow_dynamic ? 1 : 0
  type        = "ingress"
  protocol    = "UDP"
  from_port   = "49152"
  to_port     = "65535"
  cidr_blocks = distinct(var.public_agents_access_ips)

  security_group_id = aws_security_group.public_agents.id
}

resource "aws_security_group" "admin" {
  name        = "dcos-${var.cluster_name}-admin-firewall"
  description = "Allow incoming traffic from admin_ips"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )

  ingress {
    from_port   = 8
    to_port     = 0
    protocol    = "icmp"
    cidr_blocks = var.admin_ips
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.open_admin_router ? ["0.0.0.0/0"] : var.admin_ips
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.admin_ips
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = var.admin_ips
  }

  ingress {
    from_port   = var.adminrouter_grpc_proxy_port
    to_port     = var.adminrouter_grpc_proxy_port
    protocol    = "tcp"
    cidr_blocks = var.admin_ips
  }

  ingress {
    from_port   = 3389
    to_port     = 3389
    protocol    = "tcp"
    cidr_blocks = var.admin_ips
  }

  dynamic "ingress" {
    for_each = local.admin_ports
    content {
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = var.open_admin_router ? ["0.0.0.0/0"] : var.admin_ips
    }
  }
}
