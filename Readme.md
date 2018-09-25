# Terraform Lambda Provider

Custom provider for Terraform to make creating lambda functions easier.

This provider only compiles and zips up your code. You should rely on other AWS resources to deploy the zip files.

This provider currently supports Go lambda functions, but it was design and should be easy to add support for the other languages.

## Example

```hcl
provider "lambda" {}

resource "lambda_go" "hello" {
  source = "./example/hello"
}

output "hello_base64sha256" {
  value = "${lambda_go.hello.base64sha256}"
}

output "hello_path" {
  value = "${lambda_go.hello.path}"
}

resource "lambda_go" "test" {
  source = "./example/test"
}

output "test_base64sha256" {
  value = "${lambda_go.test.base64sha256}"
}

output "test_path" {
  value = "${lambda_go.test.path}"
}
```

## Motivation

This provider does 2 things: it compiles and zips local code.

I would have liked to use Terraform's own [archive provider](https://github.com/terraform-providers/terraform-provider-archive), but it doesn't correctly set the permissions bits. (See: [this issue](https://github.com/terraform-providers/terraform-provider-archive/issues/17)).

## Installation

```sh
go get github.com/matthewmueller/terraform-provider-zip
go install github.com/matthewmueller/terraform-provider-zip
```

## License

MIT