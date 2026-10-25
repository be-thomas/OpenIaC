# Note: Resource Policies are natively embedded directly into the OpenAPI specification
# body via x-amazon-apigateway-policy in main.tf (local.body_spec).
# This removes legacy policy attachments from state without issuing a delete. Destroying
# the legacy resource would otherwise clear the policy applied by the OpenAPI import.
removed {
  from = aws_api_gateway_rest_api_policy.policy_attachment

  lifecycle {
    destroy = false
  }
}
