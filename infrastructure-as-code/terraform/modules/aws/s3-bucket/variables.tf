variable "bucket_name" {
  description = "The name of the S3 bucket to create"
  type        = string
}

variable "bucket_tags" {
  description = "The tags to apply to the S3 bucket resource"
  type        = map(string)
}