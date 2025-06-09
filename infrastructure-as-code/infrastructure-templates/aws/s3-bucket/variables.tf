variable "aws_access_key" { type = string }
variable "aws_secret_key" { type = string }

variable "aws_region" { type = string }

variable "bucket_name" {
  description = "The name of the S3 bucket to create"
  type        = string
}

variable "tracker_id" {
  description = "tracker id, for which this resource is being created"
  type        = string
}

variable "aws_assume_role" {
  type = object({
    enabled     = bool
    role_arn    = optional(string, null)
    external_id = optional(string, null)
  })
  default = { enabled = false }
}

