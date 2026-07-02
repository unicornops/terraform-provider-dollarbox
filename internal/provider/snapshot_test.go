package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestSnapshotClientMethodsAndPagination(t *testing.T) {
	var sawPolicyPut, sawPolicyDelete, sawSnapshotPost, sawSnapshotDelete, sawRestorePost bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshot-policy/" && r.Method == http.MethodPut:
			var payload snapshotPolicyPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.RetentionDays != 5 {
				t.Fatalf("unexpected policy payload: %#v, %v", payload, err)
			}
			sawPolicyPut = true
			writeSnapshotPolicy(t, w, "activating", 5)
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshot-policy/" && r.Method == http.MethodGet:
			writeSnapshotPolicy(t, w, "active", 5)
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshot-policy/" && r.Method == http.MethodDelete:
			sawPolicyDelete = true
			writeSnapshotPolicy(t, w, "retiring", 5)
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshots/" && r.Method == http.MethodGet && r.URL.Query().Get("cursor") == "":
			_ = json.NewEncoder(w).Encode(map[string]any{"next": "/api/v1/namespaces/7/volumes/data/snapshots/?cursor=next", "results": []apiVolumeSnapshot{{ID: "snapshot-1", Status: "ready", Labels: map[string]string{}}}})
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshots/" && r.Method == http.MethodGet && r.URL.Query().Get("cursor") == "next":
			_ = json.NewEncoder(w).Encode(map[string]any{"next": nil, "results": []apiVolumeSnapshot{{ID: "snapshot-2", Status: "ready", Labels: map[string]string{}}}})
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshots/" && r.Method == http.MethodPost:
			var payload snapshotCreatePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name != "before-upgrade" || payload.Labels["env"] != "prod" {
				t.Fatalf("unexpected snapshot payload: %#v, %v", payload, err)
			}
			sawSnapshotPost = true
			_ = json.NewEncoder(w).Encode(apiVolumeSnapshot{ID: "snapshot-1", Name: payload.Name, Labels: payload.Labels, Status: "pending"})
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshots/snapshot-1/" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(apiVolumeSnapshot{ID: "snapshot-1", Status: "ready", Labels: map[string]string{}})
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshots/snapshot-1/" && r.Method == http.MethodDelete:
			sawSnapshotDelete = true
			_ = json.NewEncoder(w).Encode(apiVolumeSnapshot{ID: "snapshot-1", Status: "deleting", Labels: map[string]string{}})
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshots/snapshot-1/restore/" && r.Method == http.MethodPost:
			var payload snapshotRestorePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.PVCName != "restored-data" {
				t.Fatalf("unexpected restore payload: %#v, %v", payload, err)
			}
			sawRestorePost = true
			_ = json.NewEncoder(w).Encode(apiSnapshotRestore{ID: "restore-1", SnapshotID: "snapshot-1", PVCName: payload.PVCName, Status: "pending"})
		case r.URL.Path == "/api/v1/namespaces/7/volumes/data/snapshot-restores/restore-1/" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(apiSnapshotRestore{ID: "restore-1", SnapshotID: "snapshot-1", PVCName: "restored-data", Status: "bound"})
		default:
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
	defer server.Close()

	client := NewAPIClient(ClientConfig{Endpoint: server.URL, Token: "token", Org: "acme"})
	ctx := context.Background()
	if _, err := client.PutSnapshotPolicy(ctx, "7", "data", snapshotPolicyPayload{RetentionDays: 5}); err != nil {
		t.Fatal(err)
	}
	policy, err := client.GetSnapshotPolicy(ctx, "7", "data")
	if err != nil || policy.Status != "active" {
		t.Fatalf("unexpected policy: %#v, %v", policy, err)
	}
	if _, err := client.DeleteSnapshotPolicy(ctx, "7", "data"); err != nil {
		t.Fatal(err)
	}
	snapshots, err := client.ListVolumeSnapshots(ctx, "7", "data")
	if err != nil || len(snapshots) != 2 || snapshots[1].ID != "snapshot-2" {
		t.Fatalf("unexpected snapshots: %#v, %v", snapshots, err)
	}
	if _, err := client.CreateVolumeSnapshot(ctx, "7", "data", snapshotCreatePayload{Name: "before-upgrade", Labels: map[string]string{"env": "prod"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetVolumeSnapshot(ctx, "7", "data", "snapshot-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteVolumeSnapshot(ctx, "7", "data", "snapshot-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateSnapshotRestore(ctx, "7", "data", "snapshot-1", snapshotRestorePayload{PVCName: "restored-data"}); err != nil {
		t.Fatal(err)
	}
	restore, err := client.GetSnapshotRestore(ctx, "7", "data", "restore-1")
	if err != nil || restore.Status != "bound" {
		t.Fatalf("unexpected restore: %#v, %v", restore, err)
	}
	if !sawPolicyPut || !sawPolicyDelete || !sawSnapshotPost || !sawSnapshotDelete || !sawRestorePost {
		t.Fatalf("missing request: policyPut=%v policyDelete=%v snapshotPost=%v snapshotDelete=%v restorePost=%v", sawPolicyPut, sawPolicyDelete, sawSnapshotPost, sawSnapshotDelete, sawRestorePost)
	}
}

func TestSnapshotPollingTerminalFailures(t *testing.T) {
	client := NewAPIClient(ClientConfig{Endpoint: "http://example.invalid"})
	if _, err := waitForSnapshotPolicyActive(context.Background(), client, "7", "data", apiSnapshotPolicy{Status: "billing_blocked", Error: "payment required"}); err == nil || !strings.Contains(err.Error(), "payment required") {
		t.Fatalf("expected billing failure, got %v", err)
	}
	if _, err := waitForVolumeSnapshotReady(context.Background(), client, "7", "data", apiVolumeSnapshot{Status: "failed", Error: "csi failed"}); err == nil || !strings.Contains(err.Error(), "csi failed") {
		t.Fatalf("expected snapshot failure, got %v", err)
	}
	if _, err := waitForSnapshotRestoreBound(context.Background(), client, "7", "data", apiSnapshotRestore{Status: "failed", Error: "quota exceeded"}); err == nil || !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("expected restore failure, got %v", err)
	}
}

func TestSnapshotDeletionAcceptsNotFound(t *testing.T) {
	oldInterval := snapshotPollInterval
	snapshotPollInterval = time.Millisecond
	defer func() { snapshotPollInterval = oldInterval }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found","message":"gone"}}`))
	}))
	defer server.Close()
	client := NewAPIClient(ClientConfig{Endpoint: server.URL})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitForSnapshotPolicyDeleted(ctx, client, "7", "data", apiSnapshotPolicy{Status: "retiring"}); err != nil {
		t.Fatalf("policy deletion returned error: %v", err)
	}
	if err := waitForVolumeSnapshotDeleted(ctx, client, "7", "data", apiVolumeSnapshot{ID: "snapshot-1", Status: "deleting"}); err != nil {
		t.Fatalf("snapshot deletion returned error: %v", err)
	}
}

func TestSnapshotSchemas(t *testing.T) {
	policyResp := &frameworkresource.SchemaResponse{}
	NewSnapshotPolicyResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, policyResp)
	retention, ok := policyResp.Schema.Attributes["retention_days"].(resourceschema.Int64Attribute)
	if !ok || !retention.Optional || !retention.Computed || retention.Default == nil || len(retention.Validators) != 1 {
		t.Fatalf("unexpected retention_days schema: %#v", retention)
	}
	if _, ok := policyResp.Schema.Blocks["timeouts"]; !ok {
		t.Fatal("snapshot policy schema has no timeouts block")
	}

	snapshotResp := &frameworkresource.SchemaResponse{}
	NewVolumeSnapshotResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, snapshotResp)
	name, ok := snapshotResp.Schema.Attributes["name"].(resourceschema.StringAttribute)
	if !ok || !name.Optional || !name.Computed || len(name.PlanModifiers) == 0 {
		t.Fatalf("unexpected snapshot name schema: %#v", name)
	}
	if _, ok := snapshotResp.Schema.Blocks["timeouts"]; !ok {
		t.Fatal("volume snapshot schema has no timeouts block")
	}

	restoreResp := &frameworkresource.SchemaResponse{}
	NewSnapshotRestoreResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, restoreResp)
	if target, ok := restoreResp.Schema.Attributes["target_pvc_name"].(resourceschema.StringAttribute); !ok || !target.Required || len(target.PlanModifiers) == 0 {
		t.Fatalf("unexpected target_pvc_name schema: %#v", target)
	}
}

func TestParseSnapshotImportID(t *testing.T) {
	parts, err := parseSnapshotImportID("7/data/snapshot-1", 3)
	if err != nil || strings.Join(parts, "/") != "7/data/snapshot-1" {
		t.Fatalf("unexpected parse result: %#v, %v", parts, err)
	}
	for _, id := range []string{"data/snapshot-1", "not-an-integer/data/snapshot-1", "7//snapshot-1"} {
		if _, err := parseSnapshotImportID(id, 3); err == nil {
			t.Fatalf("expected %q to fail", id)
		}
	}
}

func writeSnapshotPolicy(t *testing.T, w http.ResponseWriter, status string, retention int64) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(apiSnapshotPolicy{PVCName: "data", RetentionDays: retention, ProtectedGB: 10, BilledGB: 10, MonthlyCostCents: 100, Status: status}); err != nil {
		t.Fatal(err)
	}
}
