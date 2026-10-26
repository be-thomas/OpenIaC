terraform {
  required_version = ">= 1.0.8"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0.0"
    }

    rad-security = {
      source  = "rad-security/rad-security"
      version = ">= 1.0.3"
    }
  }
}
