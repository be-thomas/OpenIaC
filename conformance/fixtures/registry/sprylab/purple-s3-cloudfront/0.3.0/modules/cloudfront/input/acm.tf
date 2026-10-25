module "acm" {
  source               = "terraform-aws-modules/acm/aws"
  version              = "6.3.0"
  region               = "us-east-1"
  create_certificate   = var.cloudfront_domain != "" && var.manage_certificate
  domain_name          = var.cloudfront_domain
  zone_id              = var.acm_zone_id
  validation_method    = "DNS"
  validate_certificate = true
  wait_for_validation  = true

  tags = {
    Name = try(var.acm_certificate_name, "Certificate ${var.cloudfront_domain}")
  }
}
