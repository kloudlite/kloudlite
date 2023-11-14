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