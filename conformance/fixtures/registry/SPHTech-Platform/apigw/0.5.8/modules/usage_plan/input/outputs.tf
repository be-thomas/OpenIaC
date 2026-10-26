output "api_keys_id" {
  description = "ID of the api keys"
  value = {
    for k, v in aws_api_gateway_api_key.key : k => v.id
  }
}
