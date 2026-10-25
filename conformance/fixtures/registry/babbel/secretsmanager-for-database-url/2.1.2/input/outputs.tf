output "secretsmanager_secret" {
  value = aws_secretsmanager_secret.endpoint

  description = "The AWS SecretsManager containing the `DATABASE_URL`."
}

output "secretsmanager_secret_reader_endpoint" {
  value = aws_secretsmanager_secret.reader_endpoint

  description = "The AWS SecretsManager containing the `DATABASE_REPLICA_URL`."
}
