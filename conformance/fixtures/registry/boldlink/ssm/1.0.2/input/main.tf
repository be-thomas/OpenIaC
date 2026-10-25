resource "aws_cloudwatch_log_group" "ssm_log_group" {
  name              = "/aws/ssm/${var.name}"
  retention_in_days = var.retention_in_days
  kms_key_id        = var.kms_key_id
  tags              = var.tags
}

### custom document
resource "aws_ssm_document" "custom" {
  count           = var.create_custom_document ? 1 : 0
  name            = var.document_name
  document_type   = var.document_type
  document_format = var.document_format
  tags            = var.tags
  content         = var.content
}

resource "aws_ssm_association" "main" {
  count            = var.association_name != null ? 1 : 0
  name             = aws_ssm_document.custom[0].name
  document_version = try(var.document_version, "$LATEST")
  association_name = var.association_name

  dynamic "targets" {
    for_each = var.targets
    content {
      key    = try(targets.value.key, null)
      values = try(targets.value.values, null)
    }
  }
}

## Session preferences
resource "aws_ssm_document" "session_preferences" {
  count           = var.create_session_preferences ? 1 : 0
  name            = var.name
  document_type   = "Session"
  document_format = "JSON"
  tags            = var.tags
  content = jsonencode({
    "schemaVersion" : "1.0",
    "description" : "Document to hold regional settings for Session Manager",
    "sessionType" : "Standard_Stream",
    "inputs" : {
      "s3BucketName" : var.s3_bucket_name,
      "s3KeyPrefix" : var.s3_key_prefix,
      "s3EncryptionEnabled" : var.s3_encryption_enabled,
      "cloudWatchLogGroupName" : aws_cloudwatch_log_group.ssm_log_group.name,
      "cloudWatchEncryptionEnabled" : var.cloudwatch_encryption_enabled,
      "cloudWatchStreamingEnabled" : var.cloudwatch_streaming_enabled,
      "kmsKeyId" : var.kms_key_id,
      "runAsEnabled" : var.run_as_enabled,
      "runAsDefaultUser" : var.run_as_default_user,
      "idleSessionTimeout" : var.idle_session_timeout,
      "maxSessionDuration" : var.max_session_duration,
      "shellProfile" : {
        "windows" : var.windows_shell_profile,
        "linux" : var.linux_shell_profile
      }
    }
  })
}
