// Package providerdata carries the shared OpenMetadata client from the
// provider's Configure method to every resource and data source.
//
// Keeping the carrier in its own package means entity packages depend on one
// small struct rather than on the provider package, which keeps the dependency
// graph acyclic and merges conflict-free.
package providerdata

import "github.com/caberdo/terraform-provider-openmetadata/internal/client"

// ProviderData is the value passed through ResourceData and DataSourceData.
type ProviderData struct {
	Client *client.Client
}
