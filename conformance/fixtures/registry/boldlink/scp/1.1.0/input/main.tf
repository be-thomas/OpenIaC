### SCPs module
resource "aws_organizations_policy" "main" {
  name        = var.name
  content     = var.content
  description = var.description
  type        = var.type
  tags        = var.tags
}

resource "aws_organizations_policy_attachment" "main" {
  count     = length(var.target_ids)
  policy_id = aws_organizations_policy.main.id
  target_id = var.target_ids[count.index]
}
