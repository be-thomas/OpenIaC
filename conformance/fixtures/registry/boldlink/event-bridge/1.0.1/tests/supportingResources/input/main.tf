module "ecs_vpc" {
  source                 = "boldlink/vpc/aws"
  version                = "3.1.0"
  name                   = var.name
  cidr_block             = var.cidr_block
  enable_dns_support     = var.enable_dns_support
  enable_dns_hostnames   = var.enable_dns_hostnames
  enable_public_subnets  = var.enable_public_subnets
  enable_private_subnets = var.enable_private_subnets
  tags                   = var.tags

  public_subnets = {
    public = {
      cidrs                   = local.public_subnets
      map_public_ip_on_launch = var.map_public_ip_on_launch
      nat                     = var.nat
    }
  }

  private_subnets = {
    private = {
      cidrs = local.private_subnets
    }
  }
}

module "kms_key" {
  source                  = "boldlink/kms/aws"
  version                 = "1.1.0"
  description             = var.kms_description
  create_kms_alias        = var.create_kms_alias
  alias_name              = "alias/${var.name}"
  enable_key_rotation     = var.enable_key_rotation
  kms_policy              = local.kms_policy
  deletion_window_in_days = var.deletion_window_in_days
  tags                    = local.tags
}

resource "aws_cloudwatch_log_group" "cluster" {
  #checkov:skip=CKV_AWS_338: "Ensure CloudWatch log groups retains logs for at least 1 year"
  name              = "/aws/ecs-cluster/${var.name}-log-group"
  retention_in_days = var.retention_in_days
  kms_key_id        = module.kms_key.arn
  tags              = local.tags
}


module "cluster" {
  source  = "boldlink/ecs-cluster/aws"
  version = "1.1.1"
  name    = var.name
  tags    = local.tags
  configuration = {
    execute_command_configuration = {
      kms_key_id = module.kms_key.key_id
      log_configuration = {
        cloud_watch_encryption_enabled = var.cloud_watch_encryption_enabled
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.cluster.name
        s3_bucket_encryption_enabled   = var.s3_bucket_encryption_enabled
      }

      logging = var.logging_type
    }
  }
}

module "ecs_service" {
  source                            = "boldlink/ecs-service/aws"
  version                           = "1.6.0"
  name                              = var.name
  family                            = "${var.name}-task-definition"
  network_mode                      = var.network_mode
  cluster                           = module.cluster.arn
  vpc_id                            = module.ecs_vpc.vpc_id
  task_assume_role_policy           = data.aws_iam_policy_document.ecs_assume_role_policy.json
  task_execution_assume_role_policy = data.aws_iam_policy_document.ecs_assume_role_policy.json
  task_execution_role_policy        = local.task_execution_role_policy_doc
  container_definitions             = local.default_container_definitions
  #kms_key_id                        = module.kms_key.arn
  tags                  = local.tags
  service_ingress_rules = var.service_ingress_rules
  depends_on            = [module.kms_key]

  network_configuration = {
    subnets = flatten(module.ecs_vpc.private_subnet_ids)
  }
}
