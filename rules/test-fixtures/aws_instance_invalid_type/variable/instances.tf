variable "instance_type" {}

resource "aws_instance" "invalid" {
  instance_type = "${var.instance_type}"
}
