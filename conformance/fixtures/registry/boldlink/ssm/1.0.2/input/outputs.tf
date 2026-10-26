output "name" {
  description = "The name of the session document"
  value       = aws_ssm_document.session_preferences[*].name
}

output "document_type" {
  description = "The type of the document"
  value       = aws_ssm_document.session_preferences[*].document_type
}
