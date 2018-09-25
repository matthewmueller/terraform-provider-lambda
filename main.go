package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/matthewmueller/terraform-provider-lambda/lambda"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: lambda.Provider,
	})
}
