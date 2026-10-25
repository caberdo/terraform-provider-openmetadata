//go:build tools

// Package tools pins build-time tooling so `go mod tidy` keeps it in go.mod.
// It is never compiled into the provider.
package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
