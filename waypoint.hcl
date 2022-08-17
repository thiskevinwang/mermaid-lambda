project = "mermaid"

app "mermaid" {
  build {
    use "docker" {
      buildkit = true
      platform           = "arm64"
      dockerfile         = "${path.app}/Dockerfile"
      disable_entrypoint = true
    }

    registry {
      use "aws-ecr" {
        region     = var.region
        repository = var.repository
        tag        = var.tag
      }
    }
  }

  # Build and deploy are optional.
  # My current use case is Waypoint simply builds an ECR Docker Image.
  #
  # From a separate AWS CDK stack, I reuse this ECR image, and create a brand new Lambda function, 
  # managed by the AWS CDK stack.

  // deploy {
  //   use "aws-lambda" {
  //     region = var.region
  //     memory = 1024
  //     static_environment = {
  //     }
  //   }
  // }

  // release {
  //   use "lambda-function-url" {

  //   }
  // }
}

variable "region" {
  default     = "us-east-1"
  type        = string
  description = "AWS Region"
}
variable "repository" {
  default     = "mermaid"
  type        = string
  description = "AWS ECR Repository Name"
}
variable "tag" {
  default     = "latest"
  type        = string
  description = "A tag"
}