module "aws-s3-bucket" {
  source      = "../../terraform/modules/aws/s3-bucket"
  bucket_name = var.bucket_name
  bucket_tags = {
    Purpose   = "kloudlite-terraform-state-files"
    TrackerId = var.tracker_id
  }
}
