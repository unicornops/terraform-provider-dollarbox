package provider

import (
	"context"
	"testing"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func TestProviderMetadata(t *testing.T) {
	resp := &frameworkprovider.MetadataResponse{}

	New("test")().Metadata(context.Background(), frameworkprovider.MetadataRequest{}, resp)

	if resp.TypeName != "dollarbox" {
		t.Fatalf("expected provider type name dollarbox, got %q", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Fatalf("expected provider version test, got %q", resp.Version)
	}
}

func TestProviderSchema(t *testing.T) {
	resp := &frameworkprovider.SchemaResponse{}

	New("test")().Schema(context.Background(), frameworkprovider.SchemaRequest{}, resp)

	endpoint, ok := resp.Schema.Attributes["endpoint"].(schema.StringAttribute)
	if !ok {
		t.Fatal("expected endpoint to be a string attribute")
	}
	if !endpoint.Optional {
		t.Fatal("expected endpoint to be optional")
	}

	token, ok := resp.Schema.Attributes["token"].(schema.StringAttribute)
	if !ok {
		t.Fatal("expected token to be a string attribute")
	}
	if !token.Optional {
		t.Fatal("expected token to be optional")
	}
	if !token.Sensitive {
		t.Fatal("expected token to be sensitive")
	}

	org, ok := resp.Schema.Attributes["org"].(schema.StringAttribute)
	if !ok {
		t.Fatal("expected org to be a string attribute")
	}
	if !org.Optional {
		t.Fatal("expected org to be optional")
	}
}
