// ==========================================================================================================================
// S3
// ==========================================================================================================================

variable "bucket_name" {
  type = string
}

variable "bucket_region" {
  type    = string
  default = "eu-central-1"
}

variable "bucket_prefix" {
  type    = string
  default = ""
}

variable "bucket_iam_user_name" {
  type    = string
  default = null
}

variable "bucket_additional_cloudfront_arns" {
  type    = list(string)
  default = []
}

variable "bucket_automatic_cleanup_enabled" {
  type    = bool
  default = true
}

variable "bucket_automatic_cleanup_days" {
  description = <<-EOT
    How many days a non-current object version - one that has been overwritten or deleted - is kept
    before it expires. This is the window in which an overwrite or deletion can be undone.

    Versioning stores roughly the monthly overwrite volume times (days / 30) in non-current versions.
    Versions above bucket_noncurrent_version_transition_min_size move to Glacier Instant Retrieval,
    whose 90-day minimum storage duration this default matches; a shorter value still expires them on
    time, but they are billed for 90 days.
  EOT
  type        = number
  default     = 90
}

variable "bucket_noncurrent_version_transition_min_size" {
  description = <<-EOT
    Size in bytes above which a non-current version is moved to Glacier Instant Retrieval. Smaller
    versions stay in the class they had while current.

    Each move costs a lifecycle transition request (about $0.02 per 1,000), which a version only earns
    back through the storage saving if it is large enough. The saving depends on the retention: over
    the default 90 days it is about $0.035/GB against Intelligent-Tiering, which puts the break-even
    at roughly 0.35-0.55 MB. With a shorter bucket_automatic_cleanup_days the saving shrinks, because
    Glacier Instant Retrieval bills a minimum of 90 days - at 35 days the break-even is about 1.5 MB.
  EOT
  type        = number
  default     = 524288
}

variable "bucket_automatic_cleanup_multipart_upload_days" {
  description = "How many days should aborted multipart uploads be kept"
  type        = number
  default     = 7
}

// ==========================================================================================================================
// Tachyon
// ==========================================================================================================================

variable "tachyon_memory_size" {
  description = <<-EOT
    Memory in MB for the Tachyon image-resizing Lambda@Edge. Raise this for buckets holding unusually large
    originals: Lambda CPU scales with memory, so it governs both out-of-memory kills and resize duration.
    512 MB is not enough for larger source images.
  EOT
  type        = number
  default     = 2048
}

// ==========================================================================================================================
// Cloudfront
// ==========================================================================================================================

variable "cloudfront_comment" {
  type    = string
  default = null
}

variable "cloudfront_price_class" {
  type    = string
  default = "PriceClass_100"
}

variable "cloudfront_domain" {
  type    = string
  default = ""
}

variable "cloudfront_minimum_protocol_version" {
  type    = string
  default = "TLSv1.2_2021"
}

variable "cloudfront_viewer_protocol_policy" {
  description = <<-EOT
    How CloudFront answers a plain HTTP viewer request, for every cache behavior of the distribution.

    The default redirects to HTTPS. Purple never issues HTTP URLs - signed CDN URLs are built with an
    `https://` scheme - so a request that arrives over HTTP comes from a third party, an old hardcoded
    link, or an absolute `http://` asset URL inside older content. Those keep working through the 301
    at the cost of one extra round trip, while a signed URL that would otherwise have travelled in
    cleartext - signature included, replayable until it expires - no longer does.

    `https-only` answers HTTP with 403 instead of redirecting, which breaks those callers rather than
    carrying them over; use it only for a distribution known to have none. `allow-all` was the default
    before 0.1.15 and is what to set if a legacy client turns out not to follow the redirect.
  EOT
  type        = string
  default     = "redirect-to-https"

  validation {
    condition     = contains(["redirect-to-https", "https-only", "allow-all"], var.cloudfront_viewer_protocol_policy)
    error_message = "Must be one of redirect-to-https, https-only, allow-all."
  }
}

variable "cloudfront_cors_allow_origins" {
  type    = list(string)
  default = null
}

variable "cloudfront_public_web" {
  type    = bool
  default = false
}

variable "cloudfront_logging_config" {
  description = "The logging configuration that controls how logs are written to your distribution (maximum one)."
  type        = any
  default     = null
}

variable "cloudfront_exclude_tracking_params" {
  description = "Exclude common third-party click-tracking parameters (utm_*, fbclid, gclid, etc.) from the CloudFront cache key to improve cache hit rates for social media traffic. The attribution parameters (appId, platform) are always excluded from the cache key and are not affected by this toggle."
  type        = bool
  default     = false
}

variable "cloudfront_additional_public_key_ids" {
  description = <<-EOT
    IDs of existing CloudFront public keys to trust alongside the key this module creates. They are
    added to the module's key group, so every cache behavior that requires signed URLs accepts them.
    The keys stay unmanaged - only their IDs are referenced - so a key shared with other distributions
    is never owned by this module's state.

    Set this when adopting the module for a distribution that already serves signed URLs. CloudFront
    verifies a signed URL by looking up the `Key-Pair-Id` it carries among the keys in the key groups
    attached to the behavior, so a newly created key is a different key even when its material is
    byte-identical to the one in use: without listing the existing ID here, every URL signed with it is
    rejected from the moment the distribution deploys, until whoever issues those URLs has been moved
    to the new key. Listing it keeps both valid and makes that move a separate, reversible step.

    CloudFront allows 5 public keys per key group and the module already uses one of them, so at most
    four IDs fit here.
  EOT
  type        = list(string)
  default     = []
}

// ==========================================================================================================================
// ACM
// ==========================================================================================================================

variable "manage_certificate" {
  type    = bool
  default = true
}

variable "acm_certificate_name" {
  type    = string
  default = null
}

variable "acm_certificate_arn" {
  type    = string
  default = null
}

variable "acm_zone_id" {
  type    = string
  default = null
}
