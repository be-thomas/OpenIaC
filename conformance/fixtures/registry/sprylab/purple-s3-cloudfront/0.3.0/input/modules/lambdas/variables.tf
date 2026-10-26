// ==========================================================================================================================
// S3
// ==========================================================================================================================

variable "bucket_name" {
  type = string
}

variable "bucket_arn" {
  type = string
}

// ==========================================================================================================================
// Tachyon
// ==========================================================================================================================

variable "tachyon_memory_size" {
  description = <<-EOT
    Memory in MB for the Tachyon image-resizing Lambda@Edge.

    Image resizing is memory-hungry: decoding an original into a bitmap needs far more memory than the
    encoded file size suggests, so even a modest source image can exhaust a small allocation. When the
    function exceeds its allocation Lambda kills it rather than letting it return an error.

    That matters more here than for an ordinary Lambda. As an origin-request function it sits between the
    viewer and the origin, so while it is stuck CloudFront has accepted the viewer connection and has
    nothing to send. Callers see a connected but silent socket until the timeout expires, and any client
    without its own read timeout will hang for the full duration. Headroom here is worth more than the
    memory it costs.

    Lambda CPU scales with memory, so this also governs resize duration; for CPU-bound work the billed
    GB-ms stays roughly flat as memory rises and duration falls. Raise it further for buckets holding
    unusually large originals.
  EOT
  type        = number
  default     = 2048

  validation {
    condition     = var.tachyon_memory_size >= 512
    error_message = "tachyon_memory_size must be at least 512 MB; below that Tachyon cannot resize typical originals."
  }
}
