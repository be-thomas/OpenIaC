/**
 * DC/OS VPC
 * ===========
 * This is an module to create a AWS VPC specially used for DC/OS
 *
 * This modules creates a subnet for every .
 *
 * EXAMPLE
 * -------
 *```hcl
 * module "dcos-vpc" {
 *   source  = "dcos-terraform/vpc/aws"
 *   version = "~> 0.3.0"
 *
 *   cluster_name = "production"
 *   availability_zones = ["us-east-1b"]
 *   subnet_range = "172.16.0.0/16"
 *   # providers {
 *   # aws = "aws.my-provider"
 *   # }
 * }
 *```
 */
provider "aws" {
  version = ">= 2.0"
}

// if availability zones is not set request the available in this region
data "aws_availability_zones" "available" {
}

locals {
  availability_zones = coalescelist(var.availability_zones, data.aws_availability_zones.available.names)
  # numbered_availability_zones = { for index, zone in local.availability_zones : cidrsubnet(var.subnet_range, 4, index) => zone }
  #numbered_availability_zones = { for index in local.availability_zones : cidrsubnet(var.subnet_range, 4, index) }
}

# Create a VPC to launch our instances into
resource "aws_vpc" "dcos_vpc" {
  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )

  cidr_block           = var.subnet_range
  enable_dns_support   = true
  enable_dns_hostnames = true
}

# Create a subnet to launch our instances into
resource "aws_subnet" "dcos_subnet" {
  count = length(local.availability_zones)

  vpc_id = aws_vpc.dcos_vpc.id

  cidr_block        = cidrsubnet(var.subnet_range, 4, count.index)
  availability_zone = element(coalescelist(local.availability_zones, data.aws_availability_zones.available.names), count.index)

  map_public_ip_on_launch = true

  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )
}

# Create an internet gateway to give our subnet access to the outside world
resource "aws_internet_gateway" "default" {
  vpc_id = aws_vpc.dcos_vpc.id

  tags = merge(
    var.tags,
    {
      "Name"    = var.cluster_name
      "Cluster" = var.cluster_name
    },
  )
}

# Grant the VPC internet access on its main route table
resource "aws_route" "internet_access" {
  route_table_id         = aws_vpc.dcos_vpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.default.id
}
