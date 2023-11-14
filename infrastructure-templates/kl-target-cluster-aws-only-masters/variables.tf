variable "aws_access_key" {
  description = "AWS Access Key"
  type        = string
}
variable "aws_secret_key" {
  description = "AWS Secret Key"
  type        = string
}

variable "aws_assume_role" {
  type = object({
    enabled     = bool
    role_arn    = string
    external_id = optional(string, null)
  })
}

