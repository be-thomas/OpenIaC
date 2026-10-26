resource "aws_cloudfront_public_key" "purple" {
  name_prefix = "${var.bucket_name}-"
  comment     = "Public key of Purple DS"
  encoded_key = replace(file("${path.module}/cloudfront/purple-public.pem"), "\r\n", "\n")
  lifecycle {
    create_before_destroy = true
  }
}

locals {
  # Query parameters that exist only to attribute traffic in access-log analysis. They never change
  # the object returned - the request path is the S3 key - and CloudFront records the query string
  # verbatim in the access log whether or not it is part of the cache key. Keying on them therefore
  # buys nothing and splits the cache across values that all resolve to the same object, which costs
  # hit rate on exactly the small, frequently requested objects where it matters most. So they are
  # excluded unconditionally, independently of `cloudfront_exclude_tracking_params`.
  #
  # The rule for anything added here: a parameter that only describes the caller does not belong in
  # the cache key, and a parameter that changes the response has to stay in it -
  # `response-content-disposition` sets a response header, and the Tachyon image parameters
  # (`w`, `h`, `webp`, `quality`, `crop`) change the returned image. That is why this is a deny-list
  # and must not be inverted into an allow-list.
  cloudfront_attribution_query_params = [
    "appId",
    "platform",
  ]

  # Click-tracking parameters. Unlike the attribution parameters these are set by third parties on
  # inbound links, so dropping them stays an opt-in per distribution.
  cloudfront_tracking_query_params = [
    "fbclid",
    "gclid",
    "utm_campaign",
    "utm_medium",
    "utm_source",
    "utm_term",
    "msclkid",
    "_ga",
    "mc_cid",
    "dclid",
  ]

  # `allExcept` requires a non-empty item list; the attribution parameters guarantee one.
  cloudfront_excluded_query_params = concat(
    local.cloudfront_attribution_query_params,
    var.cloudfront_exclude_tracking_params ? local.cloudfront_tracking_query_params : [],
  )
}

resource "aws_cloudfront_cache_policy" "s3" {
  name        = var.bucket_name
  min_ttl     = 1
  max_ttl     = 31536000
  default_ttl = 86400
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }
    # No request header belongs in the cache key. CORS is answered at the edge by the response
    # headers policy in `modules/cloudfront`, which derives `Access-Control-Allow-Origin` from the
    # viewer's `Origin` on every response it serves - cache hits included - so the stored object
    # neither needs to carry that header nor may be split per origin to get it right.
    # `Access-Control-Request-Method` and `Access-Control-Request-Headers` are only ever sent on a
    # preflight `OPTIONS`, which is in `allowed_methods` but not in `cached_methods` and so is never
    # cached at all. `Referer` has no consumer: neither S3 nor the Tachyon origin-request Lambda@Edge
    # reads it - Tachyon keys solely on the request path and the image query parameters.
    #
    # All four remain in the origin request policy below, so S3 and Tachyon still receive them
    # unchanged; only the cache key drops them. Keying on them multiplied the entries per object by
    # the product of their distinct values, which costs hit rate on exactly the small, frequently
    # requested objects where it matters most.
    headers_config {
      header_behavior = "none"
    }
    query_strings_config {
      query_string_behavior = "allExcept"
      query_strings {
        items = local.cloudfront_excluded_query_params
      }
    }
    enable_accept_encoding_brotli = false
    enable_accept_encoding_gzip   = true
  }
}

resource "aws_cloudfront_origin_request_policy" "s3" {
  name = var.bucket_name
  cookies_config {
    cookie_behavior = "none"
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = [
        "Origin",
        "Access-Control-Request-Method",
        "Access-Control-Request-Headers",
        "Referer",
      ]
    }
  }
  query_strings_config {
    query_string_behavior = "allExcept"
    query_strings {
      items = local.cloudfront_excluded_query_params
    }
  }
}

resource "aws_cloudfront_key_group" "default" {
  name    = var.bucket_name
  comment = "Public keys for ${var.bucket_name}"
  items = concat(
    [aws_cloudfront_public_key.purple.id],
    var.cloudfront_additional_public_key_ids,
  )
}

module "default_cloudfront" {
  source                              = "./modules/cloudfront"
  bucket_name                         = var.bucket_name
  bucket_prefix                       = var.bucket_prefix
  bucket_regional_domain_name         = module.bucket.s3_bucket_bucket_regional_domain_name
  bucket_region                       = module.bucket.s3_bucket_region
  cloudfront_comment                  = var.cloudfront_comment
  cloudfront_price_class              = var.cloudfront_price_class
  cloudfront_domain                   = var.cloudfront_domain
  cloudfront_minimum_protocol_version = var.cloudfront_minimum_protocol_version
  cloudfront_viewer_protocol_policy   = var.cloudfront_viewer_protocol_policy
  cloudfront_cors_allow_origins       = var.cloudfront_cors_allow_origins
  cloudfront_tachyon_qualified_arn    = module.lambdas.tachyon_qualified_arn
  cloudfront_public_web               = var.cloudfront_public_web
  cloudfront_cache_policy_id          = aws_cloudfront_cache_policy.s3.id
  cloudfront_origin_request_policy_id = aws_cloudfront_origin_request_policy.s3.id
  cloudfront_key_group_id             = aws_cloudfront_key_group.default.id
  cloudfront_logging_config           = var.cloudfront_logging_config
  manage_certificate                  = var.manage_certificate
  acm_certificate_name                = var.acm_certificate_name
  acm_certificate_arn                 = var.acm_certificate_arn
  acm_zone_id                         = var.acm_zone_id
}
