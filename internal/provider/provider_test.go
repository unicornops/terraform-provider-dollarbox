package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	frameworkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
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

func TestProviderResources(t *testing.T) {
	resources := New("test")().Resources(context.Background())
	if len(resources) != 4 {
		t.Fatalf("expected 4 resources, got %d", len(resources))
	}

	typeNames := map[string]bool{}
	for _, newResource := range resources {
		resp := &frameworkresource.MetadataResponse{}
		newResource().Metadata(context.Background(), frameworkresource.MetadataRequest{ProviderTypeName: "dollarbox"}, resp)
		typeNames[resp.TypeName] = true
	}
	for _, expected := range []string{"dollarbox_container", "dollarbox_volume", "dollarbox_invitation", "dollarbox_kubectl_credential"} {
		if !typeNames[expected] {
			t.Fatalf("expected resource %s to be registered", expected)
		}
	}
}

func TestProviderDataSources(t *testing.T) {
	dataSources := New("test")().DataSources(context.Background())
	if len(dataSources) != 1 {
		t.Fatalf("expected 1 data source, got %d", len(dataSources))
	}

	resp := &frameworkdatasource.MetadataResponse{}
	dataSources[0]().Metadata(context.Background(), frameworkdatasource.MetadataRequest{ProviderTypeName: "dollarbox"}, resp)
	if resp.TypeName != "dollarbox_org" {
		t.Fatalf("expected data source dollarbox_org, got %q", resp.TypeName)
	}
}

func TestClientSendsAuthAndOrgHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer dbx_prefix_secret" {
			t.Fatalf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get(orgHeader) != "acme" {
			t.Fatalf("unexpected org header %q", r.Header.Get(orgHeader))
		}
		if r.URL.Path != "/api/v1/orgs/acme/" {
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(apiOrg{
			Slug:           "acme",
			Name:           "Acme",
			BillingEmail:   "billing@example.com",
			Status:         "active",
			BillingMode:    "standard",
			KubectlEnabled: true,
			APIEnabled:     true,
			CreatedAt:      "2026-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewAPIClient(ClientConfig{
		Endpoint: server.URL,
		Token:    "dbx_prefix_secret",
		Org:      "acme",
	})
	org, err := client.GetOrg(context.Background(), "acme")
	if err != nil {
		t.Fatalf("GetOrg returned error: %v", err)
	}
	if org.Slug != "acme" {
		t.Fatalf("expected org slug acme, got %q", org.Slug)
	}
}

func TestClientDecodesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":"quota_exceeded","message":"Quota exceeded.","detail":{}}}`))
	}))
	defer server.Close()

	client := NewAPIClient(ClientConfig{Endpoint: server.URL, Token: "dbx_prefix_secret"})
	_, err := client.GetContainer(context.Background(), "123")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests || apiErr.Code != "quota_exceeded" {
		t.Fatalf("unexpected API error: %#v", apiErr)
	}
}
