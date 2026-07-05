package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	frameworkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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

func TestProviderProtocolSchemasAreValid(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("create provider server: %v", err)
	}
	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("get provider schema: %v", err)
	}
	for _, diagnostic := range resp.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("provider schema error: %s: %s", diagnostic.Summary, diagnostic.Detail)
		}
	}
}

func TestProviderResources(t *testing.T) {
	resources := New("test")().Resources(context.Background())
	if len(resources) != 10 {
		t.Fatalf("expected 10 resources, got %d", len(resources))
	}

	typeNames := map[string]bool{}
	for _, newResource := range resources {
		resp := &frameworkresource.MetadataResponse{}
		newResource().Metadata(context.Background(), frameworkresource.MetadataRequest{ProviderTypeName: "dollarbox"}, resp)
		typeNames[resp.TypeName] = true
	}
	for _, expected := range []string{"dollarbox_container", "dollarbox_volume", "dollarbox_invitation", "dollarbox_kubectl_credential", "dollarbox_member", "dollarbox_namespace", "dollarbox_org", "dollarbox_snapshot_policy", "dollarbox_volume_snapshot", "dollarbox_snapshot_restore"} {
		if !typeNames[expected] {
			t.Fatalf("expected resource %s to be registered", expected)
		}
	}
}

func TestProviderDataSources(t *testing.T) {
	dataSources := New("test")().DataSources(context.Background())
	if len(dataSources) != 12 {
		t.Fatalf("expected 12 data sources, got %d", len(dataSources))
	}

	typeNames := map[string]bool{}
	for _, newDataSource := range dataSources {
		resp := &frameworkdatasource.MetadataResponse{}
		newDataSource().Metadata(context.Background(), frameworkdatasource.MetadataRequest{ProviderTypeName: "dollarbox"}, resp)
		typeNames[resp.TypeName] = true
	}
	for _, expected := range []string{"dollarbox_org", "dollarbox_containers", "dollarbox_volumes", "dollarbox_invitations", "dollarbox_kubectl_credentials", "dollarbox_member", "dollarbox_members", "dollarbox_namespace", "dollarbox_namespaces", "dollarbox_orgs", "dollarbox_volume_snapshot", "dollarbox_volume_snapshots"} {
		if !typeNames[expected] {
			t.Fatalf("expected data source %s to be registered", expected)
		}
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

func TestClientMembers(t *testing.T) {
	var sawPatch atomic.Bool
	var sawDelete atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/members/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{
						"id":        "42",
						"email":     "admin@example.com",
						"role":      "admin",
						"joined_at": "2026-01-01T00:00:00Z",
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/members/42/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":        42,
				"email":     "admin@example.com",
				"role":      "member",
				"joined_at": "2026-01-01T00:00:00Z",
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/members/42/":
			var payload memberPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode member payload: %v", err)
			}
			if payload.Role != "admin" {
				t.Fatalf("expected role admin, got %q", payload.Role)
			}
			sawPatch.Store(true)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    42,
				"email": "admin@example.com",
				"role":  "admin",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/members/42/":
			sawDelete.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewAPIClient(ClientConfig{Endpoint: server.URL, Token: "dbx_prefix_secret", Org: "acme"})
	members, err := client.ListMembers(context.Background())
	if err != nil {
		t.Fatalf("ListMembers returned error: %v", err)
	}
	if len(members) != 1 || members[0].ID != "42" || members[0].Email != "admin@example.com" {
		t.Fatalf("unexpected members: %#v", members)
	}

	member, err := client.GetMember(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetMember returned error: %v", err)
	}
	if member.ID != "42" || member.Role != "member" {
		t.Fatalf("unexpected member: %#v", member)
	}

	member, err = client.UpdateMember(context.Background(), "42", memberPayload{Role: "admin"})
	if err != nil {
		t.Fatalf("UpdateMember returned error: %v", err)
	}
	if !sawPatch.Load() || member.Role != "admin" {
		t.Fatalf("expected member update to set admin role, sawPatch=%v member=%#v", sawPatch.Load(), member)
	}

	if err := client.DeleteMember(context.Background(), "42"); err != nil {
		t.Fatalf("DeleteMember returned error: %v", err)
	}
	if !sawDelete.Load() {
		t.Fatal("expected DeleteMember to call DELETE")
	}
}

func TestClientOrgs(t *testing.T) {
	var sawPatch atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{
						"slug":            "acme",
						"name":            "Acme",
						"billing_email":   "billing@example.com",
						"status":          "active",
						"billing_mode":    "standard",
						"kubectl_enabled": true,
						"api_enabled":     true,
						"created_at":      "2026-01-01T00:00:00Z",
					},
				},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/orgs/acme/":
			var payload orgPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode org payload: %v", err)
			}
			if payload.Name != "Acme Updated" || payload.BillingEmail != "billing-updated@example.com" {
				t.Fatalf("unexpected org payload: %#v", payload)
			}
			sawPatch.Store(true)
			_ = json.NewEncoder(w).Encode(apiOrg{
				Slug:           "acme",
				Name:           "Acme Updated",
				BillingEmail:   "billing-updated@example.com",
				Status:         "active",
				BillingMode:    "standard",
				KubectlEnabled: true,
				APIEnabled:     true,
				CreatedAt:      "2026-01-01T00:00:00Z",
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewAPIClient(ClientConfig{Endpoint: server.URL, Token: "dbx_prefix_secret", Org: "acme"})
	orgs, err := client.ListOrgs(context.Background())
	if err != nil {
		t.Fatalf("ListOrgs returned error: %v", err)
	}
	if len(orgs) != 1 || orgs[0].Slug != "acme" {
		t.Fatalf("unexpected orgs: %#v", orgs)
	}

	org, err := client.UpdateOrg(context.Background(), "acme", orgPayload{Name: "Acme Updated", BillingEmail: "billing-updated@example.com"})
	if err != nil {
		t.Fatalf("UpdateOrg returned error: %v", err)
	}
	if !sawPatch.Load() || org.Name != "Acme Updated" || org.BillingEmail != "billing-updated@example.com" {
		t.Fatalf("expected org update, sawPatch=%v org=%#v", sawPatch.Load(), org)
	}
}

func TestClientNamespaces(t *testing.T) {
	var sawPost atomic.Bool
	var sawPatch atomic.Bool
	var sawDelete atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{
						"id":                   7,
						"slug":                 "dev",
						"allocated_containers": 2,
						"allocated_volume_gb":  10,
						"status":               "active",
						"k8s_namespace":        "acme-dev",
						"created_at":           "2026-01-01T00:00:00Z",
						"updated_at":           "2026-01-01T00:00:00Z",
					},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/namespaces/":
			var payload namespacePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode namespace payload: %v", err)
			}
			if payload.Slug != "dev" || payload.AllocatedContainers != 2 || payload.AllocatedVolumeGB != 10 {
				t.Fatalf("unexpected namespace payload: %#v", payload)
			}
			sawPost.Store(true)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                   7,
				"slug":                 "dev",
				"allocated_containers": 2,
				"allocated_volume_gb":  10,
				"status":               "pending",
				"k8s_namespace":        "acme-dev",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces/7/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                   7,
				"slug":                 "dev",
				"allocated_containers": 2,
				"allocated_volume_gb":  10,
				"status":               "active",
				"k8s_namespace":        "acme-dev",
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/namespaces/7/":
			var payload namespacePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode namespace payload: %v", err)
			}
			if payload.AllocatedContainers != 3 || payload.AllocatedVolumeGB != 20 {
				t.Fatalf("unexpected namespace update payload: %#v", payload)
			}
			sawPatch.Store(true)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                   7,
				"slug":                 "dev",
				"allocated_containers": 3,
				"allocated_volume_gb":  20,
				"status":               "active",
				"k8s_namespace":        "acme-dev",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/namespaces/7/":
			sawDelete.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewAPIClient(ClientConfig{Endpoint: server.URL, Token: "dbx_prefix_secret", Org: "acme"})
	namespaces, err := client.ListNamespaces(context.Background())
	if err != nil {
		t.Fatalf("ListNamespaces returned error: %v", err)
	}
	if len(namespaces) != 1 || namespaces[0].ID != 7 || namespaces[0].Slug != "dev" {
		t.Fatalf("unexpected namespaces: %#v", namespaces)
	}

	namespace, err := client.CreateNamespace(context.Background(), namespacePayload{Slug: "dev", AllocatedContainers: 2, AllocatedVolumeGB: 10})
	if err != nil {
		t.Fatalf("CreateNamespace returned error: %v", err)
	}
	if !sawPost.Load() || namespace.ID != 7 {
		t.Fatalf("expected namespace create, sawPost=%v namespace=%#v", sawPost.Load(), namespace)
	}

	namespace, err = client.GetNamespace(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetNamespace returned error: %v", err)
	}
	if namespace.Status != "active" {
		t.Fatalf("unexpected namespace: %#v", namespace)
	}

	namespace, err = client.UpdateNamespace(context.Background(), "7", namespacePayload{AllocatedContainers: 3, AllocatedVolumeGB: 20})
	if err != nil {
		t.Fatalf("UpdateNamespace returned error: %v", err)
	}
	if !sawPatch.Load() || namespace.AllocatedContainers != 3 || namespace.AllocatedVolumeGB != 20 {
		t.Fatalf("expected namespace update, sawPatch=%v namespace=%#v", sawPatch.Load(), namespace)
	}

	if err := client.DeleteNamespace(context.Background(), "7"); err != nil {
		t.Fatalf("DeleteNamespace returned error: %v", err)
	}
	if !sawDelete.Load() {
		t.Fatal("expected DeleteNamespace to call DELETE")
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
