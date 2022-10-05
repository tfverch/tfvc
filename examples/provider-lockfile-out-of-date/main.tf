terraform {
  required_version = "~> 1.0"

  required_providers {
    # AWS provider configured with correct constraint but lockfile version is not latest - should warn
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}
