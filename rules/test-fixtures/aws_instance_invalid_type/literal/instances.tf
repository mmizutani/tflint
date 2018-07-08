resource "aws_instance" "invalid" {
  instance_type = "t1.2xlarge"
}

resource "aws_instance" "valid" {
  instance_type = "t2.micro"
}
