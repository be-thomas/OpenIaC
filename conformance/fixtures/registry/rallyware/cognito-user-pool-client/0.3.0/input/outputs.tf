output "id" {
  value       = one(aws_cognito_user_pool_client.default[*].id)
  description = "ID of the user pool client."
}

output "client_secret" {
  value       = one(aws_cognito_user_pool_client.default[*].client_secret)
  description = "Client secret of the user pool client."
}
