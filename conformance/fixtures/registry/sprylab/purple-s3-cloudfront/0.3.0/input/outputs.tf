// ==========================================================================================================================
// S3
// ==========================================================================================================================

output "s3_bucket_id" {
  value = module.bucket.s3_bucket_id
}

output "s3_bucket_bucket_domain_name" {
  value = module.bucket.s3_bucket_bucket_domain_name
}

output "s3_bucket_bucket_regional_domain_name" {
  value = module.bucket.s3_bucket_bucket_regional_domain_name
}

output "s3_bucket_region" {
  value = module.bucket.s3_bucket_region
}

output "s3_bucket_arn" {
  value = module.bucket.s3_bucket_arn
}
output "s3_bucket_hosted_zone_id" {
  value = module.bucket.s3_bucket_hosted_zone_id
}

output "s3_user_name" {
  value = aws_iam_user.default.name
}

output "s3_access_key" {
  value = aws_iam_access_key.default.id
}

output "s3_secret_key" {
  value     = aws_iam_access_key.default.secret
  sensitive = true
}

output "s3_access_policy_arn" {
  value = aws_iam_policy.bucket.arn
}

// ==========================================================================================================================
// Cloudfront
// ==========================================================================================================================

output "cloudfront_distribution_id" {
  value = module.default_cloudfront.cloudfront_distribution_id
}

output "cloudfront_distribution_arn" {
  value = module.default_cloudfront.cloudfront_distribution_arn
}

output "cloudfront_distribution_domain_name" {
  value = module.default_cloudfront.cloudfront_distribution_domain_name
}

output "cloudfront_distribution_hosted_zone_id" {
  value = module.default_cloudfront.cloudfront_distribution_hosted_zone_id
}

output "cloudfront_public_key_id" {
  value = aws_cloudfront_public_key.purple.id
}

output "cloudfront_key_group_id" {
  value = aws_cloudfront_key_group.default.id
}

output "cloudfront_cache_policy_id" {
  value = aws_cloudfront_cache_policy.s3.id
}

output "cloudfront_origin_request_policy_id" {
  value = aws_cloudfront_origin_request_policy.s3.id
}

// ==========================================================================================================================
// Lambdas
// ==========================================================================================================================

output "tachyon_qualified_arn" {
  value = module.lambdas.tachyon_qualified_arn
}
