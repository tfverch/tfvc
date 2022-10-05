terraform {
  required_version = "~> 1.0"

  required_providers {
    # AWS provider configured with no version constraint - should fail
    aws = {
      source  = "hashicorp/aws"
    }
  }
}
