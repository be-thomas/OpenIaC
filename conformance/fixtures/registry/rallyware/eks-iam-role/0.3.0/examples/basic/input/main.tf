module "eks_iam_role" {
  source = "rallyware/eks-iam-role/aws"
  # version     = "x.x.x"

  namespace   = var.namespace
  environment = var.environment
  stage       = var.stage
  name        = var.name
  delimiter   = var.delimiter
  attributes  = var.attributes
  tags        = var.tags

  aws_account_number          = local.account_id
  eks_cluster_oidc_issuer_url = module.eks_cluster.eks_cluster_identity_oidc_issuer

  # Create a role for the service account named `autoscaler` in the Kubernetes namespace `kube-system`
  service_account_name      = "autoscaler"
  service_account_namespace = "kube-system"
  # JSON IAM policy document to assign to the service account role
  aws_iam_policy_document = data.aws_iam_policy_document.autoscaler.json
}


data "aws_iam_policy_document" "autoscaler" {
  statement {
    sid = "AllowToScaleEKSNodeGroupAutoScalingGroup"

    actions = [
      "ec2:DescribeLaunchTemplateVersions",
      "autoscaling:TerminateInstanceInAutoScalingGroup",
      "autoscaling:SetDesiredCapacity",
      "autoscaling:DescribeTags",
      "autoscaling:DescribeLaunchConfigurations",
      "autoscaling:DescribeAutoScalingInstances",
      "autoscaling:DescribeAutoScalingGroups"
    ]

    effect    = "Allow"
    resources = ["*"]
  }
}
