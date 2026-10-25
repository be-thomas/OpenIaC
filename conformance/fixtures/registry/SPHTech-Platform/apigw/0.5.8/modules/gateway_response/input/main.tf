resource "aws_api_gateway_gateway_response" "this" {
  rest_api_id   = var.rest_api_id
  response_type = var.response_type
  status_code   = var.status_code

  response_templates  = var.gateway_responses
  response_parameters = var.response_parameters
}
