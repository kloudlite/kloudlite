variable "awsprops" {
    type = "map"
    default = {
    region = "us-east-1"
    vpc = "vpc-5234832d"
    ami = "ami-0c1bea58988a989155"
    itype = "t2.micro"
    subnet = "subnet-81896c8e"
    publicip = true
    keyname = "myseckey"
    secgroupname = "IAC-Sec-Group"
  }
}

