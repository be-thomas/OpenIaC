terraform {
  required_version = ">= 1.0"
}

variable "region" {
  type    = string
  default = "us-east-1"
}

variable "environment" {
  type    = string
  default = "dev"
}

provider "aws" {
  region = var.region
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name        = "openiac-${var.environment}"
    Environment = var.environment
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.region}a"
  map_public_ip_on_launch = true

  tags = {
    Name = "openiac-public"
  }
}

resource "aws_security_group" "web" {
  name        = "openiac-web"
  description = "Allow HTTP from anywhere"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "openiac-web"
  }
}

resource "aws_instance" "web" {
  ami                    = "ami-0123456789abcdef0"
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.web.id]

  tags = {
    Name = "openiac-web"
  }
}

resource "aws_s3_bucket" "assets" {
  bucket = "openiac-assets-${var.environment}"

  tags = {
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "jobs" {
  name                       = "openiac-jobs"
  visibility_timeout_seconds = 300
  message_retention_seconds  = 1209600
}

output "bucket_arn" {
  value = aws_s3_bucket.assets.arn
}

output "instance_ip" {
  value = aws_instance.web.public_ip
}
