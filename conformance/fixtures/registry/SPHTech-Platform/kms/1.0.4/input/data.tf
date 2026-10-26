data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

#####################################
# Allow IAM access to keys
# See https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
# for more information on key policies
#####################################
data "aws_iam_policy_document" "iam_key_policy" {
  #checkov:skip=CKV_AWS_109:The policy is deliberately wide because it allows IAM control of the KMS
  #checkov:skip=CKV_AWS_111:The policy is deliberately wide because it allows IAM control of the KMS
  #checkov:skip=CKV_AWS_356:The policyi s kept wide for IAM control of KMS
  statement {
    sid = "Allow administration of the key by key creator account"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${local.account_id}:root"]
    }

    actions   = ["kms:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "combined_key_policy" {
  source_policy_documents = concat(
    [data.aws_iam_policy_document.iam_key_policy.json],
    var.key_policy_statements,
  )
}
