resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name

  tags = merge(var.bucket_tags, {
    Name      = var.bucket_name
    CreatedBy = "kloudlite-infrastructure-as-code"
  })
}