variable "string_var" {}
variable "integer_var" {}
variable "list_var" {}
variable "map_var" {}
variable "no_value_var" {}

resource "null_resource" "test" {
  // string
  literal      = "literal_val"
  string       = "${var.string_var}"
  new_string   = var.string_var
  list_element = "${var.list_var[0]}"
  map_element  = "${var.map_var["one"]}"
  conditional  = "${true ? "production" : "development"}"
  function     = "${md5("foo")}"

  // integer
  integer = "${var.integer_var}"

  // list
  list = ["one", "two", "three"]

  // map
  map = {
    one = 1
    two = 2
  }

  undefined = "${var.undefined_var}"
  no_value  = "${var.no_value_var}"
  env       = "${terraform.env}"
  workspace = "${terraform.workspace}"
}
