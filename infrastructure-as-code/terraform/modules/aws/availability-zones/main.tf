data "aws_availability_zones" "az" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}