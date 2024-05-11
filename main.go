package main

import (
	"context"
	"terraform-provider-instatus/instatus"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name instatus

func main() {
	providerserver.Serve(context.Background(), instatus.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/brunoscota/instatus",
	})
}
