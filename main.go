// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/caberdo/terraform-provider-openmetadata/internal/provider"
)

// Run "go generate" to regenerate the documentation in docs/ from examples/
// and templates/. The provider must be buildable and Terraform must be on PATH.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name openmetadata
func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/caberdo/openmetadata",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}

// version is set at build time by GoReleaser:
//
//	-ldflags "-X main.version={{.Version}}"
var version = "dev"
