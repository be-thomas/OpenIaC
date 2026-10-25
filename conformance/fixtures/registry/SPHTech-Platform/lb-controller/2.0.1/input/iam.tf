module "lb_controller_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts"
  version = "~> 6.0"

  name        = coalesce(var.iam_role_name, "${var.cluster_name}-lb")
  description = "EKS Cluster ${var.cluster_name} LoadBalancer Controller"

  attach_load_balancer_controller_policy = true
  oidc_providers = {
    main = {
      provider_arn               = var.oidc_provider_arn
      namespace_service_accounts = ["kube-system:${local.service_account_name}"]
    }
  }
}

moved {
  from = module.lb_controller_role.aws_iam_policy.load_balancer_controller[0]
  to   = module.lb_controller_role.aws_iam_policy.this[0]
}

moved {
  from = module.lb_controller_role.aws_iam_role_policy_attachment.load_balancer_controller[0]
  to   = module.lb_controller_role.aws_iam_role_policy_attachment.this[0]
}
