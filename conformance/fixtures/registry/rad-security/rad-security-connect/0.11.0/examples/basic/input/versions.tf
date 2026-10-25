terraform {
  required_version = ">= 1.0.8"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0.0"
    }

    ksoc = {
      source  = "ksoclabs/ksoc"
      version = ">= 0.1.0"
    }
  }
}
