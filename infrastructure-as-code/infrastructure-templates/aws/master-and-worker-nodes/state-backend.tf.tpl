terraform {
  required_version = ">= 1.2.0"

  backend "s3" {
    bucket = "${AWS_S3_BUCKET_NAME}"
    key    = "${AWS_S3_BUCKET_FILEPATH}"
    region = "${AWS_S3_BUCKET_REGION}"
  }
}