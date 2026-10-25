#######################################################
## KMS Key data.aws_iam_policy_document.kms.json
#######################################################

resource "aws_kms_key" "sns" {
  count                   = var.create_kms_key ? 1 : 0
  description             = "KMS Key for SNStopic ${var.name}"
  policy                  = local.policy
  deletion_window_in_days = var.key_deletion_window
  enable_key_rotation     = var.enable_key_rotation
}

resource "aws_kms_alias" "sns" {
  count         = var.create_kms_key ? 1 : 0
  name          = "alias/sns/${trim(var.name, ".fifo")}"
  target_key_id = aws_kms_key.sns[0].key_id
}

##################
## sns topic
##################
resource "aws_sns_topic" "main" {
  name                                     = var.name
  name_prefix                              = var.name != null ? null : var.name_prefix
  display_name                             = var.display_name
  delivery_policy                          = var.delivery_policy
  application_success_feedback_role_arn    = var.application_success_feedback_role_arn
  application_success_feedback_sample_rate = var.application_success_feedback_sample_rate
  application_failure_feedback_role_arn    = var.application_failure_feedback_role_arn
  http_success_feedback_role_arn           = var.http_success_feedback_role_arn
  http_success_feedback_sample_rate        = var.http_success_feedback_sample_rate
  http_failure_feedback_role_arn           = var.http_failure_feedback_role_arn
  kms_master_key_id                        = var.create_kms_key && var.kms_master_key_arn == null ? aws_kms_key.sns[0].arn : (var.create_kms_key == false && var.kms_master_key_arn != null ? var.kms_master_key_arn : "alias/aws/sns")
  fifo_topic                               = var.fifo_topic
  content_based_deduplication              = var.content_based_deduplication
  lambda_success_feedback_role_arn         = var.lambda_success_feedback_role_arn
  lambda_success_feedback_sample_rate      = var.lambda_success_feedback_sample_rate
  lambda_failure_feedback_role_arn         = var.lambda_failure_feedback_role_arn
  sqs_success_feedback_role_arn            = var.sqs_success_feedback_role_arn
  sqs_success_feedback_sample_rate         = var.sqs_success_feedback_sample_rate
  sqs_failure_feedback_role_arn            = var.sqs_failure_feedback_role_arn
  firehose_success_feedback_role_arn       = var.firehose_success_feedback_role_arn
  firehose_success_feedback_sample_rate    = var.firehose_success_feedback_sample_rate
  firehose_failure_feedback_role_arn       = var.firehose_failure_feedback_role_arn
  tags                                     = var.tags
}

resource "aws_sns_topic_policy" "main" {
  count  = var.create_sns_topic_policy ? 1 : 0
  arn    = aws_sns_topic.main.arn
  policy = var.policy
}

#########################
## sns topic subscription
#########################
resource "aws_sns_topic_subscription" "main" {
  for_each                        = var.sns_topic_subscriptions
  topic_arn                       = aws_sns_topic.main.arn
  endpoint                        = lookup(each.value, "endpoint", null)
  protocol                        = lookup(each.value, "protocol", null)
  subscription_role_arn           = lookup(each.value, "subscription_role_arn", null)
  confirmation_timeout_in_minutes = lookup(each.value, "confirmation_timeout_in_minutes", null)
  delivery_policy                 = lookup(each.value, "delivery_policy", null)
  endpoint_auto_confirms          = lookup(each.value, "endpoint_auto_confirms", null)
  filter_policy                   = lookup(each.value, "filter_policy", null)
  raw_message_delivery            = lookup(each.value, "raw_message_delivery", null)
  redrive_policy                  = lookup(each.value, "redrive_policy", null)
}
