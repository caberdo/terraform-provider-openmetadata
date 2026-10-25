package acctest

import "testing"

func TestProtoV6ProviderFactories(t *testing.T) {
	factories := ProtoV6ProviderFactories()

	factory, ok := factories[ProviderTypeName]
	if !ok {
		t.Fatalf("factories do not contain %q: %v", ProviderTypeName, factories)
	}

	server, err := factory()
	if err != nil {
		t.Fatalf("factory returned an error: %v", err)
	}
	if server == nil {
		t.Fatal("factory returned a nil server")
	}
}
